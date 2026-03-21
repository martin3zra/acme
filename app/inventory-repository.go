package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ── Internal structs ─────────────────────────────────────────────────────────

type stockBalance struct {
	VariantID   int64   `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	ItemID      int64   `json:"item_id"`
	ItemName    string  `json:"item_name"`
	WarehouseID int64   `json:"warehouse_id"`
	Warehouse   string  `json:"warehouse"`
	Quantity    float64 `json:"quantity"`
	UpdatedAt   string  `json:"updated_at"`
}

type inventoryMovementRow struct {
	ID            int64   `json:"id"`
	VariantID     int64   `json:"variant_id"`
	VariantName   string  `json:"variant_name"`
	SKU           string  `json:"sku"`
	ItemName      string  `json:"item_name"`
	WarehouseID   int64   `json:"warehouse_id"`
	Warehouse     string  `json:"warehouse"`
	Kind          string  `json:"kind"`
	Qty           float64 `json:"qty"`
	UnitCost      float64 `json:"unit_cost"`
	ReferenceType string  `json:"reference_type"`
	ReferenceID   int64   `json:"reference_id"`
	CreatedAt     string  `json:"created_at"`
}

type adjustmentRow struct {
	ID          int64   `json:"id"`
	VariantID   int64   `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	ItemName    string  `json:"item_name"`
	WarehouseID int64   `json:"warehouse_id"`
	Warehouse   string  `json:"warehouse"`
	Qty         float64 `json:"qty"`
	Reason      string  `json:"reason"`
	Notes       string  `json:"notes"`
	CreatedAt   string  `json:"created_at"`
}

// ── Core movement helpers ────────────────────────────────────────────────────

// recordMovement inserts one row into inventory_movements and upserts
// inventory_balances for the given variant/warehouse combination.
//
// qty must be signed: positive = stock IN, negative = stock OUT.
// The qty is multiplied by the unit's base_qty before being stored.
// If the variant has track_inventory = false the call is a no-op.
func (s *Server) recordMovement(
	tx *sql.Tx,
	companyID, variantID, warehouseID, unitID int,
	qty float64,
	unitCost float64,
	kind InventoryMovementKind,
	referenceType string,
	referenceID int,
) error {
	// Skip non-tracked variants.
	var track bool
	if err := tx.QueryRow(
		"SELECT track_inventory FROM items_variants WHERE id = $1 AND company_id = $2",
		variantID, companyID,
	).Scan(&track); err != nil {
		return fmt.Errorf("recordMovement: lookup track_inventory: %w", err)
	}
	if !track {
		return nil
	}

	// Resolve base_qty for unit conversion.
	baseQty := 1
	if unitID > 0 {
		if err := tx.QueryRow(
			"SELECT COALESCE(base_qty, 1) FROM units WHERE id = $1 AND company_id = $2",
			unitID, companyID,
		).Scan(&baseQty); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("recordMovement: lookup base_qty: %w", err)
		}
	}

	finalQty := qty * float64(baseQty)

	// Insert movement record.
	_, err := tx.Exec(
		`INSERT INTO inventory_movements
		    (company_id, variant_id, warehouse_id, transaction_kind, qty, unit_cost, reference_type, reference_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		companyID, variantID, warehouseID, kind, finalQty, unitCost,
		referenceType, referenceID, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("recordMovement: insert movement: %w", err)
	}

	// Upsert balance.
	_, err = tx.Exec(
		`INSERT INTO inventory_balances (company_id, variant_id, warehouse_id, quantity, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (company_id, variant_id, warehouse_id)
		 DO UPDATE SET quantity = inventory_balances.quantity + EXCLUDED.quantity,
		               updated_at = NOW()`,
		companyID, variantID, warehouseID, finalQty,
	)
	if err != nil {
		return fmt.Errorf("recordMovement: upsert balance: %w", err)
	}

	return nil
}

// reverseMovements creates return entries (sale_return or purchase_return) for
// every movement previously recorded for the given reference. The reversalKind
// determines which enum value is used for the new rows.
func (s *Server) reverseMovements(tx *sql.Tx, companyID int, referenceType string, referenceID int, reversalKind InventoryMovementKind) error {
	rows, err := tx.Query(
		`SELECT variant_id, warehouse_id, qty, unit_cost
		   FROM inventory_movements
		  WHERE company_id = $1 AND reference_type = $2 AND reference_id = $3
		    AND transaction_kind NOT IN ('sale_return', 'purchase_return')`,
		companyID, referenceType, referenceID,
	)
	if err != nil {
		return fmt.Errorf("reverseMovements: query: %w", err)
	}
	defer rows.Close()

	type mvt struct{ variantID, warehouseID int; qty, unitCost float64 }
	var movements []mvt
	for rows.Next() {
		var m mvt
		if err := rows.Scan(&m.variantID, &m.warehouseID, &m.qty, &m.unitCost); err != nil {
			return fmt.Errorf("reverseMovements: scan: %w", err)
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, m := range movements {
		reversal := -m.qty

		_, err := tx.Exec(
			`INSERT INTO inventory_movements
			    (company_id, variant_id, warehouse_id, transaction_kind, qty, unit_cost, reference_type, reference_id, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			companyID, m.variantID, m.warehouseID, reversalKind,
			reversal, m.unitCost, referenceType, referenceID, now,
		)
		if err != nil {
			return fmt.Errorf("reverseMovements: insert reversal: %w", err)
		}

		_, err = tx.Exec(
			`UPDATE inventory_balances
			    SET quantity = quantity + $1, updated_at = NOW()
			  WHERE company_id = $2 AND variant_id = $3 AND warehouse_id = $4`,
			reversal, companyID, m.variantID, m.warehouseID,
		)
		if err != nil {
			return fmt.Errorf("reverseMovements: update balance: %w", err)
		}
	}

	return nil
}

// ── Query helpers ─────────────────────────────────────────────────────────────

// findStocks returns the current inventory balances for a company.
func (s *Server) findStocks(ctx context.Context, companyID int) ([]*stockBalance, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ib.variant_id,
		       iv.name       AS variant_name,
		       COALESCE(iv.sku, '')  AS sku,
		       i.id          AS item_id,
		       i.name        AS item_name,
		       ib.warehouse_id,
		       w.name        AS warehouse,
		       ib.quantity,
		       TO_CHAR(ib.updated_at, 'YYYY-MM-DD HH24:MI') AS updated_at
		  FROM inventory_balances ib
		  JOIN items_variants iv ON iv.id = ib.variant_id AND iv.company_id = ib.company_id
		  JOIN items          i  ON i.id  = iv.item_id   AND i.company_id  = ib.company_id
		  JOIN warehouses     w  ON w.id  = ib.warehouse_id AND w.company_id = ib.company_id
		 WHERE ib.company_id = $1
		   AND iv.deleted_at IS NULL
		   AND i.deleted_at  IS NULL
		   AND w.deleted_at  IS NULL
		 ORDER BY i.name, iv.name, w.name`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*stockBalance
	for rows.Next() {
		var sb stockBalance
		if err := rows.Scan(
			&sb.VariantID, &sb.VariantName, &sb.SKU,
			&sb.ItemID, &sb.ItemName,
			&sb.WarehouseID, &sb.Warehouse,
			&sb.Quantity, &sb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &sb)
	}
	return result, rows.Err()
}

// findMovements returns the inventory movement log for a company.
func (s *Server) findMovements(ctx context.Context, companyID int) ([]*inventoryMovementRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id,
		       m.variant_id,
		       iv.name                AS variant_name,
		       COALESCE(iv.sku, '')   AS sku,
		       i.name                 AS item_name,
		       m.warehouse_id,
		       w.name                 AS warehouse,
		       m.transaction_kind::text AS kind,
		       m.qty,
		       m.unit_cost,
		       COALESCE(m.reference_type, '') AS reference_type,
		       COALESCE(m.reference_id,   0)  AS reference_id,
		       TO_CHAR(m.created_at, 'YYYY-MM-DD HH24:MI') AS created_at
		  FROM inventory_movements m
		  JOIN items_variants iv ON iv.id = m.variant_id   AND iv.company_id = m.company_id
		  JOIN items          i  ON i.id  = iv.item_id     AND i.company_id  = m.company_id
		  JOIN warehouses     w  ON w.id  = m.warehouse_id AND w.company_id  = m.company_id
		 WHERE m.company_id = $1
		   AND iv.deleted_at IS NULL
		   AND i.deleted_at  IS NULL
		 ORDER BY m.created_at DESC
		 LIMIT 500`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*inventoryMovementRow
	for rows.Next() {
		var r inventoryMovementRow
		if err := rows.Scan(
			&r.ID, &r.VariantID, &r.VariantName, &r.SKU, &r.ItemName,
			&r.WarehouseID, &r.Warehouse,
			&r.Kind, &r.Qty, &r.UnitCost,
			&r.ReferenceType, &r.ReferenceID,
			&r.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &r)
	}
	return result, rows.Err()
}

// findAdjustments returns manual stock adjustments for a company.
func (s *Server) findAdjustments(ctx context.Context, companyID int) ([]*adjustmentRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id,
		       m.variant_id,
		       iv.name                AS variant_name,
		       COALESCE(iv.sku, '')   AS sku,
		       i.name                 AS item_name,
		       m.warehouse_id,
		       w.name                 AS warehouse,
		       m.qty,
		       COALESCE(m.reference_type, '') AS reason,
		       '' AS notes,
		       TO_CHAR(m.created_at, 'YYYY-MM-DD HH24:MI') AS created_at
		  FROM inventory_movements m
		  JOIN items_variants iv ON iv.id = m.variant_id   AND iv.company_id = m.company_id
		  JOIN items          i  ON i.id  = iv.item_id     AND i.company_id  = m.company_id
		  JOIN warehouses     w  ON w.id  = m.warehouse_id AND w.company_id  = m.company_id
		 WHERE m.company_id = $1
		   AND m.transaction_kind = 'adjustment'
		   AND iv.deleted_at IS NULL
		   AND i.deleted_at  IS NULL
		 ORDER BY m.created_at DESC`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*adjustmentRow
	for rows.Next() {
		var r adjustmentRow
		if err := rows.Scan(
			&r.ID, &r.VariantID, &r.VariantName, &r.SKU, &r.ItemName,
			&r.WarehouseID, &r.Warehouse,
			&r.Qty, &r.Reason, &r.Notes,
			&r.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &r)
	}
	return result, rows.Err()
}

// storeAdjustment records a manual inventory adjustment.
func (s *Server) storeAdjustment(ctx context.Context, form *StoreAdjustmentForm) error {
	companyID := CurrentCompany(ctx).ID

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Use unitID=0 so no unit conversion is applied for adjustments —
	// adjustments are always in base units.
	if err := s.recordMovement(
		tx, companyID,
		form.VariantID, form.WarehouseID, 0,
		form.Qty, 0,
		InventoryMovementKinds.Adjustment,
		form.Reason, 0,
	); err != nil {
		return err
	}

	return tx.Commit()
}

type variantOption struct {
ID       int64  `json:"id"`
Name     string `json:"name"`
ItemName string `json:"item_name"`
SKU      string `json:"sku"`
}

// findTrackableVariants returns all enabled variants with track_inventory=true.
func (s *Server) findTrackableVariants(ctx context.Context, companyID int) ([]*variantOption, error) {
rows, err := s.db.QueryContext(ctx, `
SELECT iv.id,
       iv.name,
       i.name  AS item_name,
       COALESCE(iv.sku, '') AS sku
  FROM items_variants iv
  JOIN items i ON i.id = iv.item_id AND i.company_id = iv.company_id
 WHERE iv.company_id = $1
   AND iv.track_inventory = TRUE
   AND iv.deleted_at IS NULL
   AND i.deleted_at  IS NULL
   AND iv.status = 'enabled'
 ORDER BY i.name, iv.name`,
companyID,
)
if err != nil {
return nil, err
}
defer rows.Close()

var result []*variantOption
for rows.Next() {
var v variantOption
if err := rows.Scan(&v.ID, &v.Name, &v.ItemName, &v.SKU); err != nil {
return nil, err
}
result = append(result, &v)
}
return result, rows.Err()
}
