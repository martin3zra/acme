package app

import (
	"context"
	"errors"
	"strings"

	"github.com/martin3zra/playsql"
)

func normalizeAttributeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Stays raw: the predicate is `lower(trim(name)) = $2`. playsql's Where builds a
// quoted identifier, and WhereRaw takes no bind parameters, so the only way to
// express this through the builder would be to interpolate a user-supplied name
// into SQL.
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

// findAttributes returns all active attributes for the current company. The
// `deleted_at IS NULL` comes from attributeRead's softdelete tag.
func (s *Server) findAttributes(ctx context.Context) ([]*attribute, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []attributeRead
	if err := pdb.Model(&attributeRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*attribute, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toAttribute())
	}
	return data, nil
}

// findAttribute resolves one attribute by a single column.
func (s *Server) findAttribute(ctx context.Context, column string, value any) (*attribute, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row attributeRead
	err = pdb.Model(&attributeRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq(column, value).
		First(ctx, &row)
	if errors.Is(err, playsql.ErrNotFound) {
		return nil, errors.New("attribute not found")
	}
	if err != nil {
		return nil, err
	}
	return row.toAttribute(), nil
}

// findAttributeByID returns a single attribute by UUID
func (s *Server) findAttributeByID(ctx context.Context, id string) (*attribute, error) {
	return s.findAttribute(ctx, "uuid", id)
}

// findAttributeByIntID returns a single attribute by integer ID
func (s *Server) findAttributeByIntID(ctx context.Context, id int) (*attribute, error) {
	return s.findAttribute(ctx, "id", id)
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

	pdb, err := s.play()
	if err != nil {
		return err
	}

	// The statement this replaces ended in `RETURNING id` and then called Row.Err()
	// without ever calling Scan. Row.Err() does not consume the row, so the *Rows and
	// its connection were held until the garbage collector ran a finalizer. Nothing
	// used the returned id, so it is simply not requested.
	//
	// uuid is DB-generated and stays unmapped; created_at/updated_at are stamped by
	// playsql, replacing the literal NOW()s.
	_, err = pdb.Model(&attributeRead{}).Insert(context.Background(), map[string]any{
		"company_id":   companyID,
		"name":         form.Name,
		"type":         form.Type,
		"display_name": form.DisplayName,
		"description":  form.Description,
	})
	return err
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

	pdb, err := s.play()
	if err != nil {
		return err
	}

	// softdelete supplies `deleted_at IS NULL`; updated_at is stamped.
	affected, err := pdb.Model(&attributeRead{}).
		WhereEq("company_id", companyID).
		WhereEq("id", id).
		Update(ctx, map[string]any{
			"name":         form.Name,
			"type":         form.Type,
			"display_name": form.DisplayName,
			"description":  form.Description,
		})
	return mustAffectRows(affected, err, "attribute")
}

// findAttributesWithValues returns all active attributes including active values.
//
// This used to issue one query per attribute. With("Values") eager-loads them in a
// single second query; the relation target is softdelete-tagged, so deleted values
// are excluded as they were before.
func (s *Server) findAttributesWithValues(ctx context.Context) ([]*attribute, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []attributeRead
	if err := pdb.Model(&attributeRead{}).
		WithConstraint("Values", func(q *playsql.Builder) {
			q.OrderBy("sort_order", playsql.Asc).OrderBy("value", playsql.Asc)
		}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*attribute, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toAttribute())
	}
	return data, nil
}
