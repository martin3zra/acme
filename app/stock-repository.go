package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// findStockLevels returns stock levels with optional filtering
func (s *Server) findStockLevels(ctx context.Context, filters map[string]interface{}) ([]*stockLevel, error) {
	companyID := CurrentCompany(ctx).ID
	query := `SELECT sl.id, sl.uuid, sl.warehouse_id, sl.variant_id, sl.quantity, sl.reorder_level, sl.reorder_quantity,
	                 sl.created_at, sl.updated_at
	          FROM stock_levels sl
	          WHERE sl.company_id = $1`
	args := []interface{}{companyID}
	argNum := 2

	// Optional warehouse filter
	if warehouseID, exists := filters["warehouse_id"]; exists {
		query += fmt.Sprintf(" AND sl.warehouse_id = $%d", argNum)
		args = append(args, warehouseID)
		argNum++
	}

	// Optional variant filter
	if variantID, exists := filters["variant_id"]; exists {
		query += fmt.Sprintf(" AND sl.variant_id = $%d", argNum)
		args = append(args, variantID)
		argNum++
	}

	query += " ORDER BY sl.warehouse_id, sl.variant_id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*stockLevel, 0)
	for rows.Next() {
		level := new(stockLevel)
		if err = rows.Scan(
			&level.ID, &level.UUID, &level.WarehouseID, &level.VariantID, &level.Quantity,
			&level.ReorderLevel, &level.ReorderQuantity, &level.CreatedAt, &level.UpdatedAt,
		); err != nil {
			return data, err
		}
		data = append(data, level)
	}

	return data, rows.Err()
}

// findStockByWarehouseAndVariant returns a single stock level record
func (s *Server) findStockByWarehouseAndVariant(ctx context.Context, warehouseID, variantID int) (*stockLevel, error) {
	level := new(stockLevel)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT sl.id, sl.uuid, sl.warehouse_id, sl.variant_id, sl.quantity, sl.reorder_level, sl.reorder_quantity,
		        sl.created_at, sl.updated_at
		 FROM stock_levels sl
		 WHERE sl.company_id = $1 AND sl.warehouse_id = $2 AND sl.variant_id = $3`,
		CurrentCompany(ctx).ID, warehouseID, variantID,
	).Scan(
		&level.ID, &level.UUID, &level.WarehouseID, &level.VariantID, &level.Quantity,
		&level.ReorderLevel, &level.ReorderQuantity, &level.CreatedAt, &level.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("stock level not found")
	}

	return level, err
}

// getStockForVariant returns the total quantity for a variant across all warehouses
func (s *Server) getStockForVariant(ctx context.Context, variantID int) (int, error) {
	var totalQuantity sql.NullInt64

	err := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(quantity), 0) FROM stock_levels
		 WHERE company_id = $1 AND variant_id = $2`,
		CurrentCompany(ctx).ID, variantID,
	).Scan(&totalQuantity)

	if err != nil {
		return 0, err
	}

	return int(totalQuantity.Int64), nil
}

// getStockForItem returns all stock levels for an item's variants
func (s *Server) getStockForItem(ctx context.Context, itemID int) ([]*stockLevel, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT sl.id, sl.uuid, sl.warehouse_id, sl.variant_id, sl.quantity, sl.reorder_level, sl.reorder_quantity,
		        sl.created_at, sl.updated_at
		 FROM stock_levels sl
		 JOIN items_variants iv ON sl.variant_id = iv.id
		 WHERE sl.company_id = $1 AND iv.item_id = $2 AND iv.deleted_at IS NULL
		 ORDER BY sl.warehouse_id, sl.variant_id`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*stockLevel, 0)
	for rows.Next() {
		level := new(stockLevel)
		if err = rows.Scan(
			&level.ID, &level.UUID, &level.WarehouseID, &level.VariantID, &level.Quantity,
			&level.ReorderLevel, &level.ReorderQuantity, &level.CreatedAt, &level.UpdatedAt,
		); err != nil {
			return data, err
		}
		data = append(data, level)
	}

	return data, rows.Err()
}

// getStockForWarehouse returns all stock levels in a warehouse
func (s *Server) getStockForWarehouse(ctx context.Context, warehouseID int) ([]*stockLevel, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT sl.id, sl.uuid, sl.warehouse_id, sl.variant_id, sl.quantity, sl.reorder_level, sl.reorder_quantity,
		        sl.created_at, sl.updated_at
		 FROM stock_levels sl
		 WHERE sl.company_id = $1 AND sl.warehouse_id = $2
		 ORDER BY sl.variant_id`,
		CurrentCompany(ctx).ID, warehouseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*stockLevel, 0)
	for rows.Next() {
		level := new(stockLevel)
		if err = rows.Scan(
			&level.ID, &level.UUID, &level.WarehouseID, &level.VariantID, &level.Quantity,
			&level.ReorderLevel, &level.ReorderQuantity, &level.CreatedAt, &level.UpdatedAt,
		); err != nil {
			return data, err
		}
		data = append(data, level)
	}

	return data, rows.Err()
}

// adjustStock updates stock quantity (creates if not exists)
func (s *Server) adjustStock(ctx context.Context, warehouseID, variantID int, quantity int) error {
	companyID := CurrentCompany(ctx).ID

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO stock_levels (company_id, warehouse_id, variant_id, quantity, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (company_id, warehouse_id, variant_id) 
		 DO UPDATE SET quantity = quantity + $4, updated_at = NOW()`,
		companyID, warehouseID, variantID, quantity,
	)

	return err
}

// initializeStock creates a stock level record with initial quantity
func (s *Server) initializeStock(tx *sql.Tx, companyID, warehouseID, variantID int, initialQty int) error {
	stmt, err := tx.Prepare(
		`INSERT INTO stock_levels (company_id, warehouse_id, variant_id, quantity, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (company_id, warehouse_id, variant_id) DO NOTHING`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(companyID, warehouseID, variantID, initialQty)
	return err
}
