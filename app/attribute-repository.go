package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func normalizeAttributeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (s *Server) hasDuplicateAttributeName(ctx context.Context, companyID int, name string, exceptID *int) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1
		FROM attributes
		WHERE company_id = $1
		  AND deleted_at IS NULL
		  AND lower(trim(name)) = $2
	)`

	args := []any{companyID, normalizeAttributeName(name)}
	if exceptID != nil {
		query = `SELECT EXISTS(
			SELECT 1
			FROM attributes
			WHERE company_id = $1
			  AND deleted_at IS NULL
			  AND lower(trim(name)) = $2
			  AND id != $3
		)`
		args = append(args, *exceptID)
	}

	var exists bool
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// findAttributes returns all active attributes for the current company
func (s *Server) findAttributes(ctx context.Context) ([]*attribute, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT a.id, a.uuid, a.name, a.type, a.display_name, a.description, 
		        a.created_at, a.updated_at, a.deleted_at
		 FROM attributes a
		 WHERE a.company_id = $1 AND a.deleted_at IS NULL
		 ORDER BY a.name`,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*attribute, 0)
	for rows.Next() {
		attr := new(attribute)
		if err = rows.Scan(
			&attr.ID, &attr.UUID, &attr.Name, &attr.Type, &attr.DisplayName, &attr.Description,
			&attr.CreatedAt, &attr.UpdatedAt, &attr.DeletedAt,
		); err != nil {
			return data, err
		}
		data = append(data, attr)
	}

	return data, rows.Err()
}

// findAttributeByID returns a single attribute by ID
func (s *Server) findAttributeByID(ctx context.Context, id string) (*attribute, error) {
	attr := new(attribute)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT a.id, a.uuid, a.name, a.type, a.display_name, a.description,
		        a.created_at, a.updated_at, a.deleted_at
		 FROM attributes a
		 WHERE a.company_id = $1 AND a.uuid = $2 AND a.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&attr.ID, &attr.UUID, &attr.Name, &attr.Type, &attr.DisplayName, &attr.Description,
		&attr.CreatedAt, &attr.UpdatedAt, &attr.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("attribute not found")
	}

	return attr, err
}

// findAttributeByIntID returns a single attribute by integer ID
func (s *Server) findAttributeByIntID(ctx context.Context, id int) (*attribute, error) {
	attr := new(attribute)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT a.id, a.uuid, a.name, a.type, a.display_name, a.description,
		        a.created_at, a.updated_at, a.deleted_at
		 FROM attributes a
		 WHERE a.company_id = $1 AND a.id = $2 AND a.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&attr.ID, &attr.UUID, &attr.Name, &attr.Type, &attr.DisplayName, &attr.Description,
		&attr.CreatedAt, &attr.UpdatedAt, &attr.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("attribute not found")
	}

	return attr, err
}

// storeAttribute creates a new attribute
func (s *Server) storeAttribute(ctx context.Context, form *StoreAttributeForm) error {
	companyID := CurrentCompany(ctx).ID
	form.Name = strings.TrimSpace(form.Name)
	form.DisplayName = strings.TrimSpace(form.DisplayName)
	form.Description = strings.TrimSpace(form.Description)

	hasDuplicate, err := s.hasDuplicateAttributeName(ctx, companyID, form.Name, nil)
	if err != nil {
		return err
	}

	if hasDuplicate {
		return errors.New("attribute name already exists")
	}

	stmt, err := s.db.PrepareContext(
		ctx,
		`INSERT INTO attributes (company_id, uuid, name, type, display_name, description, created_at, updated_at)
		 VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRowContext(ctx, companyID, form.Name, form.Type, form.DisplayName, form.Description).Err()
}

// updateAttribute updates an existing attribute
func (s *Server) updateAttribute(ctx context.Context, id int, form *StoreAttributeForm) error {
	companyID := CurrentCompany(ctx).ID
	form.Name = strings.TrimSpace(form.Name)
	form.DisplayName = strings.TrimSpace(form.DisplayName)
	form.Description = strings.TrimSpace(form.Description)

	hasDuplicate, err := s.hasDuplicateAttributeName(ctx, companyID, form.Name, &id)
	if err != nil {
		return err
	}

	if hasDuplicate {
		return errors.New("attribute name already exists")
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE attributes 
		 SET name = $1, type = $2, display_name = $3, description = $4, updated_at = NOW()
		 WHERE company_id = $5 AND id = $6 AND deleted_at IS NULL`,
		form.Name, form.Type, form.DisplayName, form.Description, companyID, id,
	)

	return err
}

// findAttributesWithValues returns all active attributes including active values
func (s *Server) findAttributesWithValues(ctx context.Context) ([]*attribute, error) {
	attrs, err := s.findAttributes(ctx)
	if err != nil {
		return nil, err
	}

	for _, attr := range attrs {
		values, valueErr := s.findAttributeValuesByAttribute(ctx, attr.ID)
		if valueErr != nil {
			return nil, valueErr
		}

		attr.Values = values
	}

	return attrs, nil
}
