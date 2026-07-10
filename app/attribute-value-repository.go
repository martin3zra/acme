package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/martin3zra/playsql"
)

func normalizeAttributeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// Stays raw for the same reason as hasDuplicateAttributeName: the predicate is
// `lower(trim(value)) = $3`, which the builder cannot parameterise.
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

// findAttributeValuesByAttribute returns all values for an attribute. The
// `deleted_at IS NULL` comes from attributeValueRead's softdelete tag.
func (s *Server) findAttributeValuesByAttribute(ctx context.Context, attributeID int) ([]*attributeValue, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []attributeValueRead
	if err := pdb.Model(&attributeValueRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("attribute_id", attributeID).
		OrderBy("sort_order", playsql.Asc).
		OrderBy("value", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*attributeValue, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toAttributeValue())
	}
	return data, nil
}

// findAttributeValue resolves one value by a single column.
func (s *Server) findAttributeValue(ctx context.Context, column string, value any) (*attributeValue, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row attributeValueRead
	err = pdb.Model(&attributeValueRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq(column, value).
		First(ctx, &row)
	if errors.Is(err, playsql.ErrNotFound) {
		return nil, errors.New("attribute value not found")
	}
	if err != nil {
		return nil, err
	}
	return row.toAttributeValue(), nil
}

// findAttributeValueByID returns a single attribute value by integer ID
func (s *Server) findAttributeValueByID(ctx context.Context, id int) (*attributeValue, error) {
	return s.findAttributeValue(ctx, "id", id)
}

// findAttributeValueByUUID returns a single attribute value by UUID
func (s *Server) findAttributeValueByUUID(ctx context.Context, uuid string) (*attributeValue, error) {
	return s.findAttributeValue(ctx, "uuid", uuid)
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

	pdb, err := s.play()
	if err != nil {
		return err
	}

	// Reviving a soft-deleted value is what the conflict branch is for, so deleted_at
	// is written explicitly as nil: EXCLUDED.deleted_at is then NULL, and the update
	// clears it. playsql stamps created_at/updated_at, and uuid keeps its DB default.
	//
	// The old statement's `RETURNING id` was scanned into a variable nothing read.
	_, err = pdb.Model(&attributeValueRead{}).Upsert(context.Background(),
		[]map[string]any{{
			"company_id":   companyID,
			"attribute_id": form.AttributeID,
			"value":        form.Value,
			"display_name": form.DisplayName,
			"sort_order":   form.SortOrder,
			"deleted_at":   nil,
		}},
		[]string{"company_id", "attribute_id", "value"},
		[]string{"display_name", "sort_order", "deleted_at", "updated_at"},
	)
	return err
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

	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&attributeValueRead{}).
		WhereEq("company_id", companyID).
		WhereEq("uuid", uuid).
		Update(ctx, map[string]any{
			"value":        form.Value,
			"display_name": form.DisplayName,
			"sort_order":   form.SortOrder,
		})
	return mustAffectRows(affected, err, "attribute value")
}

// deleteAttributeValue soft-deletes an attribute value
func (s *Server) deleteAttributeValue(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	pdb, err := s.play()
	if err != nil {
		return err
	}

	// Soft delete through Update, not Delete: Builder.Delete stamps deleted_at only,
	// and the statement it replaced bumped updated_at too. The softdelete tag also
	// adds `deleted_at IS NULL`, so re-deleting an already-deleted value is now a
	// not-found rather than a silent second write.
	affected, err := pdb.Model(&attributeValueRead{}).
		WhereEq("company_id", companyID).
		WhereEq("uuid", uuid).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return mustAffectRows(affected, err, "attribute value")
}
