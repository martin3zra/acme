package app

import (
	"context"
	"database/sql"
	"errors"
)

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

// findAttributeValueByID returns a single attribute value by ID
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

// storeAttributeValue creates a new attribute value
func (s *Server) storeAttributeValue(ctx context.Context, form *StoreAttributeValueForm) error {
	companyID := CurrentCompany(ctx).ID

	stmt, err := s.db.PrepareContext(
		ctx,
		`INSERT INTO attribute_values (company_id, uuid, attribute_id, value, display_name, sort_order, created_at, updated_at)
		 VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	sortOrder := form.SortOrder
	if sortOrder == 0 {
		sortOrder = 0
	}

	return stmt.QueryRowContext(ctx, companyID, form.AttributeID, form.Value, form.DisplayName, sortOrder).Err()
}

// updateAttributeValue updates an existing attribute value
func (s *Server) updateAttributeValue(ctx context.Context, id int, form *StoreAttributeValueForm) error {
	companyID := CurrentCompany(ctx).ID

	_, err := s.db.ExecContext(
		ctx,
		`UPDATE attribute_values 
		 SET display_name = $1, sort_order = $2, updated_at = NOW()
		 WHERE company_id = $3 AND id = $4 AND deleted_at IS NULL`,
		form.DisplayName, form.SortOrder, companyID, id,
	)

	return err
}

// deleteAttributeValue soft-deletes an attribute value
func (s *Server) deleteAttributeValue(ctx context.Context, id int) error {
	companyID := CurrentCompany(ctx).ID

	_, err := s.db.ExecContext(
		ctx,
		`UPDATE attribute_values 
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE company_id = $1 AND id = $2`,
		companyID, id,
	)

	return err
}
