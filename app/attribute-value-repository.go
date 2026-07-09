package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func normalizeAttributeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (s *Server) hasDuplicateAttributeValue(ctx context.Context, companyID, attributeID int, value string, exceptUUID *string) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1
		FROM attribute_values
		WHERE company_id = $1
		  AND attribute_id = $2
		  AND deleted_at IS NULL
		  AND lower(trim(value)) = $3
	)`

	args := []any{companyID, attributeID, normalizeAttributeValue(value)}
	if exceptUUID != nil {
		query = `SELECT EXISTS(
			SELECT 1
			FROM attribute_values
			WHERE company_id = $1
			  AND attribute_id = $2
			  AND deleted_at IS NULL
			  AND lower(trim(value)) = $3
			  AND uuid != $4
		)`
		args = append(args, *exceptUUID)
	}

	var exists bool
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// findAttributeValuesByAttribute returns all values for an attribute
func (s *Server) findAttributeValuesByAttribute(ctx context.Context, attributeID int) ([]*attributeValue, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT av.id, av.uuid, av.attribute_id, av.value, av.display_name, av.sort_order,
		        av.created_at, av.updated_at, av.deleted_at
		 FROM attribute_values av
		 WHERE av.company_id = $1 AND av.attribute_id = $2 AND av.deleted_at IS NULL
		 ORDER BY av.sort_order, av.value`,
		CurrentCompany(ctx).ID, attributeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*attributeValue, 0)
	for rows.Next() {
		av := new(attributeValue)
		if err = rows.Scan(
			&av.ID, &av.UUID, &av.AttributeID, &av.Value, &av.DisplayName, &av.SortOrder,
			&av.CreatedAt, &av.UpdatedAt, &av.DeletedAt,
		); err != nil {
			return data, err
		}
		data = append(data, av)
	}

	return data, rows.Err()
}

// findAttributeValueByID returns a single attribute value by integer ID
func (s *Server) findAttributeValueByID(ctx context.Context, id int) (*attributeValue, error) {
	av := new(attributeValue)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT av.id, av.uuid, av.attribute_id, av.value, av.display_name, av.sort_order,
		        av.created_at, av.updated_at, av.deleted_at
		 FROM attribute_values av
		 WHERE av.company_id = $1 AND av.id = $2 AND av.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&av.ID, &av.UUID, &av.AttributeID, &av.Value, &av.DisplayName, &av.SortOrder,
		&av.CreatedAt, &av.UpdatedAt, &av.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("attribute value not found")
	}

	return av, err
}

// findAttributeValueByUUID returns a single attribute value by UUID
func (s *Server) findAttributeValueByUUID(ctx context.Context, uuid string) (*attributeValue, error) {
	av := new(attributeValue)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT av.id, av.uuid, av.attribute_id, av.value, av.display_name, av.sort_order,
		        av.created_at, av.updated_at, av.deleted_at
		 FROM attribute_values av
		 WHERE av.company_id = $1 AND av.uuid = $2 AND av.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, uuid,
	).Scan(
		&av.ID, &av.UUID, &av.AttributeID, &av.Value, &av.DisplayName, &av.SortOrder,
		&av.CreatedAt, &av.UpdatedAt, &av.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("attribute value not found")
	}

	return av, err
}

// storeAttributeValue creates a new attribute value (or revives a soft-deleted one)
func (s *Server) storeAttributeValue(ctx context.Context, form *StoreAttributeValueForm) error {
	companyID := CurrentCompany(ctx).ID
	form.Value = strings.TrimSpace(form.Value)
	form.DisplayName = strings.TrimSpace(form.DisplayName)

	hasDuplicate, err := s.hasDuplicateAttributeValue(ctx, companyID, form.AttributeID, form.Value, nil)
	if err != nil {
		return err
	}

	if hasDuplicate {
		return errors.New("attribute value already exists")
	}

	stmt, err := s.db.PrepareContext(
		ctx,
		`INSERT INTO attribute_values (company_id, uuid, attribute_id, value, display_name, sort_order, created_at, updated_at)
		 VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, NOW(), NOW())
		 ON CONFLICT (company_id, attribute_id, value)
		 DO UPDATE
		 SET
		    deleted_at = NULL,
		    display_name = EXCLUDED.display_name,
		    sort_order = EXCLUDED.sort_order,
		    updated_at = NOW()
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var id int
	return stmt.QueryRowContext(ctx, companyID, form.AttributeID, form.Value, form.DisplayName, form.SortOrder).Scan(&id)
}

// updateAttributeValue updates an existing attribute value
func (s *Server) updateAttributeValue(ctx context.Context, uuid string, form *StoreAttributeValueForm) error {
	companyID := CurrentCompany(ctx).ID
	form.Value = strings.TrimSpace(form.Value)
	form.DisplayName = strings.TrimSpace(form.DisplayName)

	hasDuplicate, err := s.hasDuplicateAttributeValue(ctx, companyID, form.AttributeID, form.Value, &uuid)
	if err != nil {
		return err
	}

	if hasDuplicate {
		return errors.New("attribute value already exists")
	}

	res, err := s.db.ExecContext(
		ctx,
		`UPDATE attribute_values
		 SET value = $1, display_name = $2, sort_order = $3, updated_at = NOW()
		 WHERE company_id = $4 AND uuid = $5 AND deleted_at IS NULL`,
		form.Value, form.DisplayName, form.SortOrder, companyID, uuid,
	)

	return mustAffectRow(res, err, "attribute value")
}

// deleteAttributeValue soft-deletes an attribute value
func (s *Server) deleteAttributeValue(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	res, err := s.db.ExecContext(
		ctx,
		`UPDATE attribute_values
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE company_id = $1 AND uuid = $2`,
		companyID, uuid,
	)

	return mustAffectRow(res, err, "attribute value")
}
