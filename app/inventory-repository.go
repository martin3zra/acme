package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/martin3zra/playsql"
)

// Sentinel errors for stock transfers, mapped to form fields by the handler.
var (
	ErrSameWarehouse     = errors.New("source and destination warehouse must be different")
	ErrInsufficientStock = errors.New("insufficient stock at the source warehouse")
	ErrTransferNotFound  = errors.New("transfer not found")
	ErrInvalidTransition = errors.New("transfer cannot move to the requested status")
	ErrNoTransferLines   = errors.New("a transfer needs at least one product line")
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

type transferRow struct {
	ID            int64   `json:"id"`
	UUID          string  `json:"uuid"`
	FromWarehouse string  `json:"from_warehouse"`
	ToWarehouse   string  `json:"to_warehouse"`
	Status        string  `json:"status"`
	Notes         string  `json:"notes"`
	RequestedBy   string  `json:"requested_by"`
	LineCount     int     `json:"line_count"`
	TotalQty      float64 `json:"total_qty"`
	TotalCost     float64 `json:"total_cost"`
	CreatedAt     string  `json:"created_at"`
	DispatchedAt  string  `json:"dispatched_at"`
	ReceivedAt    string  `json:"received_at"`
}

type transferLineRow struct {
	ID          int64   `json:"id"`
	VariantID   int64   `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	Reference   string  `json:"reference"`
	ItemName    string  `json:"item_name"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Qty         float64 `json:"qty"`
	UnitCost    float64 `json:"unit_cost"`
	LineTotal   float64 `json:"line_total"`
}

// transferItemOption is a sellable-variant hit for the transfer line search. ID
// is the item id and VariantID the specific variant this row transacts; the
// search expands one row per trackable variant, so a variant product surfaces
// each of its variants ("Item — Variant"). Cost is that variant's cost_price.
type transferItemOption struct {
	ID          int64   `json:"id"`
	VariantID   int64   `json:"variant_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Reference   string  `json:"reference"`
	SKU         string  `json:"sku"`
	Cost        float64 `json:"cost"`
	Unit        struct {
		ID   *int64  `json:"id"`
		Name *string `json:"name"`
	} `json:"unit"`
}

// ErrWarehouseNotInCompany is returned when a request names a warehouse the current
// company does not own, or one that has been deleted.
var ErrWarehouseNotInCompany = errors.New("warehouse does not belong to this company")

// assertWarehousesInCompany fails unless every id names a live warehouse owned by
// the company. Ids of zero are ignored: those columns are NOT NULL and the foreign
// key would reject them anyway.
//
// Warehouse ids arrive straight from the request on invoice lines and on a
// transfer's from/to, and nothing downstream checked them. recordMovement took the
// id on trust, so a foreign warehouse produced an inventory_balances row keyed to
// the caller's company_id and someone else's warehouse_id. findStocks joins
// `w.company_id = ib.company_id`, so that row leaked nothing — it simply vanished,
// leaving stock recorded and invisible to everyone. Purchases were never exposed:
// they take their warehouse from firstWarehouseID, which is already company-scoped.
//
// warehouseRead carries play:"softdelete", so a deleted warehouse does not count.
func assertWarehousesInCompany(tx *sql.Tx, companyID int, ids ...int) error {
	wanted := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			wanted[id] = struct{}{}
		}
	}
	if len(wanted) == 0 {
		return nil
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	values := make([]any, 0, len(wanted))
	for id := range wanted {
		values = append(values, id)
	}

	var rows []warehouseRead
	if err := ptx.Model(&warehouseRead{}).
		Select("id").
		WhereEq("company_id", companyID).
		WhereIn("id", values...).
		Get(context.Background(), &rows); err != nil {
		return err
	}
	if len(rows) != len(wanted) {
		return ErrWarehouseNotInCompany
	}
	return nil
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
	// The warehouse comes from the caller, which in turn takes it from the request.
	if err := assertWarehousesInCompany(tx, companyID, warehouseID); err != nil {
		return fmt.Errorf("recordMovement: %w", err)
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// Skip non-tracked variants.
	var variant itemVariantRead
	if err := ptx.Model(&itemVariantRead{}).
		Select("track_inventory").
		WhereEq("id", variantID).
		WhereEq("company_id", companyID).
		First(context.Background(), &variant); err != nil {
		return fmt.Errorf("recordMovement: lookup track_inventory: %w", err)
	}
	if !variant.TrackInventory {
		return nil
	}

	// Resolve base_qty for unit conversion. An unknown unit leaves the multiplier at
	// 1, as before. The old COALESCE(base_qty, 1) was dead: the column is NOT NULL.
	baseQty := 1
	if unitID > 0 {
		var u unitRead
		err := ptx.Model(&unitRead{}).
			Select("base_qty").
			WhereEq("id", unitID).
			WhereEq("company_id", companyID).
			First(context.Background(), &u)
		if err != nil && !errors.Is(err, playsql.ErrNotFound) {
			return fmt.Errorf("recordMovement: lookup base_qty: %w", err)
		}
		if err == nil {
			baseQty = u.BaseQty
		}
	}

	finalQty := qty * float64(baseQty)

	// Insert movement record.
	if err := ptx.Insert(context.Background(), &InventoryMovement{
		CompanyID:       companyID,
		VariantID:       variantID,
		WarehouseID:     warehouseID,
		TransactionKind: kind,
		Qty:             finalQty,
		UnitCost:        unitCost,
		ReferenceType:   referenceType,
		ReferenceID:     referenceID,
		CreatedAt:       time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("recordMovement: insert movement: %w", err)
	}

	// Stays raw: the conflict branch adds EXCLUDED.quantity to the stored quantity.
	// playsql's Upsert can only assign `col = EXCLUDED.col`, not accumulate.
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
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	var movements []inventoryMovementRead
	if err := ptx.Model(&inventoryMovementRead{}).
		Select("variant_id", "warehouse_id", "qty", "unit_cost").
		WhereEq("company_id", companyID).
		WhereEq("reference_type", referenceType).
		WhereEq("reference_id", referenceID).
		WhereNotIn("transaction_kind", string(InventoryMovementKinds.SaleReturn), string(InventoryMovementKinds.PurchaseReturn)).
		Get(context.Background(), &movements); err != nil {
		return fmt.Errorf("reverseMovements: query: %w", err)
	}

	now := time.Now().UTC()
	for _, m := range movements {
		reversal := -m.Qty

		if err := ptx.Insert(context.Background(), &InventoryMovement{
			CompanyID:       companyID,
			VariantID:       m.VariantID,
			WarehouseID:     m.WarehouseID,
			TransactionKind: reversalKind,
			Qty:             reversal,
			UnitCost:        m.UnitCost,
			ReferenceType:   referenceType,
			ReferenceID:     referenceID,
			CreatedAt:       now,
		}); err != nil {
			return fmt.Errorf("reverseMovements: insert reversal: %w", err)
		}

		// Stays raw: `quantity = quantity + $1` is a self-referencing increment.
		_, err = tx.Exec(
			`UPDATE inventory_balances
			    SET quantity = quantity + $1, updated_at = NOW()
			  WHERE company_id = $2 AND variant_id = $3 AND warehouse_id = $4`,
			reversal, companyID, m.VariantID, m.WarehouseID,
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

// findTransfers returns the manual transfer requests for a company, newest first,
// with aggregate line count, total qty and total cost per transfer.
func (s *Server) findTransfers(ctx context.Context, companyID int) ([]*transferRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id,
		       t.uuid::text,
		       wf.name                AS from_warehouse,
		       wt.name                AS to_warehouse,
		       t.status::text         AS status,
		       COALESCE(t.notes, '')  AS notes,
		       COALESCE(u.name, '')   AS requested_by,
		       COALESCE(l.cnt, 0)        AS line_count,
		       COALESCE(l.total_qty, 0)  AS total_qty,
		       COALESCE(l.total_cost, 0) AS total_cost,
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI') AS created_at,
		       COALESCE(TO_CHAR(t.dispatched_at, 'YYYY-MM-DD HH24:MI'), '') AS dispatched_at,
		       COALESCE(TO_CHAR(t.received_at,   'YYYY-MM-DD HH24:MI'), '') AS received_at
		  FROM inventory_transfers t
		  JOIN warehouses     wf ON wf.id = t.from_warehouse_id  AND wf.company_id = t.company_id
		  JOIN warehouses     wt ON wt.id = t.to_warehouse_id    AND wt.company_id = t.company_id
		  LEFT JOIN users     u  ON u.id  = t.requested_by
		  LEFT JOIN (
		        SELECT transfer_id,
		               COUNT(*)               AS cnt,
		               SUM(qty)               AS total_qty,
		               SUM(qty * unit_cost)   AS total_cost
		          FROM inventory_transfer_lines
		         WHERE company_id = $1
		         GROUP BY transfer_id
		  ) l ON l.transfer_id = t.id
		 WHERE t.company_id = $1
		 ORDER BY t.created_at DESC`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*transferRow
	for rows.Next() {
		var r transferRow
		if err := rows.Scan(
			&r.ID, &r.UUID, &r.FromWarehouse, &r.ToWarehouse,
			&r.Status, &r.Notes, &r.RequestedBy,
			&r.LineCount, &r.TotalQty, &r.TotalCost,
			&r.CreatedAt, &r.DispatchedAt, &r.ReceivedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &r)
	}
	return result, rows.Err()
}

// findTransferByUUID returns a single transfer header (with aggregates) for the
// detail view.
func (s *Server) findTransferByUUID(ctx context.Context, companyID int, uuid string) (*transferRow, error) {
	var r transferRow
	err := s.db.QueryRowContext(ctx, `
		SELECT t.id,
		       t.uuid::text,
		       wf.name, wt.name,
		       t.status::text,
		       COALESCE(t.notes, ''),
		       COALESCE(u.name, ''),
		       COALESCE(l.cnt, 0), COALESCE(l.total_qty, 0), COALESCE(l.total_cost, 0),
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI'),
		       COALESCE(TO_CHAR(t.dispatched_at, 'YYYY-MM-DD HH24:MI'), ''),
		       COALESCE(TO_CHAR(t.received_at,   'YYYY-MM-DD HH24:MI'), '')
		  FROM inventory_transfers t
		  JOIN warehouses wf ON wf.id = t.from_warehouse_id AND wf.company_id = t.company_id
		  JOIN warehouses wt ON wt.id = t.to_warehouse_id   AND wt.company_id = t.company_id
		  LEFT JOIN users u ON u.id = t.requested_by
		  LEFT JOIN (
		        SELECT transfer_id, COUNT(*) AS cnt, SUM(qty) AS total_qty, SUM(qty * unit_cost) AS total_cost
		          FROM inventory_transfer_lines WHERE company_id = $1 GROUP BY transfer_id
		  ) l ON l.transfer_id = t.id
		 WHERE t.company_id = $1 AND t.uuid = $2`,
		companyID, uuid,
	).Scan(
		&r.ID, &r.UUID, &r.FromWarehouse, &r.ToWarehouse,
		&r.Status, &r.Notes, &r.RequestedBy,
		&r.LineCount, &r.TotalQty, &r.TotalCost,
		&r.CreatedAt, &r.DispatchedAt, &r.ReceivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTransferNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// findTransferLines returns the product detail lines for a transfer.
func (s *Server) findTransferLines(ctx context.Context, companyID int, transferID int64) ([]*transferLineRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tl.id,
		       tl.variant_id,
		       iv.name                                  AS variant_name,
		       COALESCE(iv.sku, '')                     AS sku,
		       COALESCE(i.identifiers->>'reference', '') AS reference,
		       i.name                                   AS item_name,
		       COALESCE(tl.description, '')             AS description,
		       COALESCE(un.name, '')                    AS unit,
		       tl.qty,
		       tl.unit_cost,
		       (tl.qty * tl.unit_cost)                  AS line_total
		  FROM inventory_transfer_lines tl
		  JOIN items_variants iv ON iv.id = tl.variant_id AND iv.company_id = tl.company_id
		  JOIN items          i  ON i.id  = iv.item_id    AND i.company_id  = tl.company_id
		  LEFT JOIN units     un ON un.id = tl.unit_id
		 WHERE tl.company_id = $1 AND tl.transfer_id = $2
		 ORDER BY tl.id`,
		companyID, transferID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*transferLineRow
	for rows.Next() {
		var r transferLineRow
		if err := rows.Scan(
			&r.ID, &r.VariantID, &r.VariantName, &r.SKU, &r.Reference, &r.ItemName,
			&r.Description, &r.Unit, &r.Qty, &r.UnitCost, &r.LineTotal,
		); err != nil {
			return nil, err
		}
		result = append(result, &r)
	}
	return result, rows.Err()
}

// storeTransfer records a new manual transfer request with its product lines.
// No stock moves yet — stock leaves the source on dispatch and arrives at the
// destination on receive. Each line resolves to a concrete variant (explicit, or
// the item's default) so a variant product moves the intended variant's stock.
func (s *Server) storeTransfer(ctx context.Context, form *StoreTransferForm) error {
	if form.FromWarehouseID == form.ToWarehouseID {
		return ErrSameWarehouse
	}
	if len(form.Lines) == 0 {
		return ErrNoTransferLines
	}

	companyID := CurrentCompany(ctx).ID

	var requestedBy any
	if u := AuthUserFromContext(ctx); u != nil && !u.IsEmpty() {
		requestedBy = u.GetAuthIdentifier()
	}

	createdAt := form.Date
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := assertWarehousesInCompany(tx, companyID, form.FromWarehouseID, form.ToWarehouseID); err != nil {
		return err
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// created_at is supplied, so playsql does not stamp over it. A nil map value
	// writes NULL, which is what NULLIF($4, '') produced for an empty note.
	id, err := ptx.Model(&inventoryTransferRead{}).Insert(context.Background(), map[string]any{
		"company_id":        companyID,
		"from_warehouse_id": form.FromWarehouseID,
		"to_warehouse_id":   form.ToWarehouseID,
		"notes":             nullIfEmpty(form.Notes),
		"status":            string(TransferStatuses.Requested),
		"requested_by":      requestedBy,
		"created_at":        createdAt,
	})
	if err != nil {
		return err
	}
	transferID := int(id)

	// Resolve each line to a concrete variant with the same rules as sales and
	// purchases: an explicit variant is validated against its item, a plain item
	// falls to its default, and a has_variants item without a variant is rejected
	// rather than silently defaulting. Stubs carry (item, variant) for the shared
	// resolver, which writes the resolved id back onto each stub.
	stubs := make([]*Line, len(form.Lines))
	for i, l := range form.Lines {
		stubs[i] = &Line{ID: l.ID, VariantID: l.VariantID}
	}
	if err := resolveVariantsForLines(tx, companyID, stubs); err != nil {
		return err
	}

	rows := make([]map[string]any, 0, len(form.Lines))
	for i, l := range form.Lines {
		var unitID any
		if l.Unit > 0 {
			unitID = l.Unit
		}
		rows = append(rows, map[string]any{
			"company_id":  companyID,
			"transfer_id": transferID,
			"variant_id":  stubs[i].VariantID,
			"qty":         l.Qty,
			"unit_id":     unitID,
			"unit_cost":   l.Cost,
			"description": nullIfEmpty(l.Description),
		})
	}
	if _, err := ptx.Model(&inventoryTransferLine{}).InsertMany(context.Background(), rows); err != nil {
		return err
	}

	return tx.Commit()
}

// nullIfEmpty mirrors SQL's NULLIF(x, ”): an empty string becomes a NULL column
// value rather than an empty one.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// transferHeader is the minimal header row needed to drive a status transition.
type transferHeader struct {
	ID     int
	FromWh int
	ToWh   int
	Status TransferStatus
}

// transferLineMovement is a line reduced to what stock movement needs.
type transferLineMovement struct {
	VariantID int
	UnitID    int
	Qty       float64
	UnitCost  float64
}

// loadTransferForUpdate locks a transfer row by uuid for the current company.
func loadTransferForUpdate(tx *sql.Tx, companyID int, uuid string) (*transferHeader, error) {
	var h transferHeader
	var status string
	err := tx.QueryRow(
		`SELECT id, from_warehouse_id, to_warehouse_id, status::text
		   FROM inventory_transfers
		  WHERE company_id = $1 AND uuid = $2
		  FOR UPDATE`,
		companyID, uuid,
	).Scan(&h.ID, &h.FromWh, &h.ToWh, &status)
	if err == sql.ErrNoRows {
		return nil, ErrTransferNotFound
	}
	if err != nil {
		return nil, err
	}
	h.Status = TransferStatus(status)
	return &h, nil
}

// loadTransferLineMovements reads the lines of a transfer for stock movement. A NULL
// unit_id scans as 0, which is what the old COALESCE(unit_id, 0) produced.
func loadTransferLineMovements(tx *sql.Tx, companyID, transferID int) ([]transferLineMovement, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return nil, err
	}

	var rows []inventoryTransferLine
	if err := ptx.Model(&inventoryTransferLine{}).
		Select("variant_id", "unit_id", "qty", "unit_cost").
		WhereEq("company_id", companyID).
		WhereEq("transfer_id", transferID).
		Get(context.Background(), &rows); err != nil {
		return nil, err
	}

	lines := make([]transferLineMovement, 0, len(rows))
	for _, r := range rows {
		l := transferLineMovement{VariantID: r.VariantID, Qty: r.Qty, UnitCost: r.UnitCost}
		if r.UnitID != nil {
			l.UnitID = *r.UnitID
		}
		lines = append(lines, l)
	}
	return lines, nil
}

// unitBaseQty returns a unit's base_qty multiplier (1 when no/unknown unit).
//
// The old COALESCE(base_qty, 1) was dead: the column is NOT NULL. What made an
// unknown unit fall back to 1 was the sql.ErrNoRows guard, kept here as
// playsql.ErrNotFound.
func unitBaseQty(tx *sql.Tx, companyID, unitID int) (float64, error) {
	if unitID <= 0 {
		return 1, nil
	}

	ptx, err := playTx(tx)
	if err != nil {
		return 0, err
	}

	var u unitRead
	err = ptx.Model(&unitRead{}).
		Select("base_qty").
		WhereEq("id", unitID).
		WhereEq("company_id", companyID).
		First(context.Background(), &u)
	if errors.Is(err, playsql.ErrNotFound) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return float64(u.BaseQty), nil
}

// setTransferStatus transitions a transfer, stamping any extra timestamp columns the
// transition owns. playsql stamps updated_at because inventoryTransferRead maps it.
//
// The statements this replaces matched on `id` alone. The id came from a
// company-scoped lock, so the result is the same, but the company_id predicate is
// carried here explicitly rather than implied.
func setTransferStatus(tx *sql.Tx, companyID, transferID int, status TransferStatus, extra map[string]any) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	changes := map[string]any{"status": string(status)}
	for k, v := range extra {
		changes[k] = v
	}

	_, err = ptx.Model(&inventoryTransferRead{}).
		WhereEq("company_id", companyID).
		WhereEq("id", transferID).
		Update(context.Background(), changes)
	return err
}

// dispatchTransfer moves a transfer requested -> in_transit, taking stock OUT of
// the source warehouse for every line. Fails if any line lacks enough on hand.
func (s *Server) dispatchTransfer(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	h, err := loadTransferForUpdate(tx, companyID, uuid)
	if err != nil {
		return err
	}
	if !h.Status.CanTransitionTo(TransferStatuses.InTransit) {
		return ErrInvalidTransition
	}

	lines, err := loadTransferLineMovements(tx, companyID, h.ID)
	if err != nil {
		return err
	}

	for _, l := range lines {
		base, err := unitBaseQty(tx, companyID, l.UnitID)
		if err != nil {
			return err
		}
		needed := l.Qty * base

		// Lock the source balance and ensure enough on hand. Missing row = zero.
		var available float64
		err = tx.QueryRow(
			`SELECT quantity FROM inventory_balances
			  WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3
			  FOR UPDATE`,
			companyID, l.VariantID, h.FromWh,
		).Scan(&available)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if available < needed {
			return ErrInsufficientStock
		}

		// OUT from source (negative). recordMovement applies unit conversion.
		if err := s.recordMovement(
			tx, companyID,
			l.VariantID, h.FromWh, l.UnitID,
			-l.Qty, l.UnitCost,
			InventoryMovementKinds.Transfer,
			string(InventoryMovementKinds.Transfer), h.ID,
		); err != nil {
			return err
		}
	}

	if err := setTransferStatus(tx, companyID, h.ID, TransferStatuses.InTransit, map[string]any{
		"dispatched_at": time.Now().UTC(),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// receiveTransfer moves a transfer in_transit -> received, taking stock IN to the
// destination warehouse for every line.
func (s *Server) receiveTransfer(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	h, err := loadTransferForUpdate(tx, companyID, uuid)
	if err != nil {
		return err
	}
	if !h.Status.CanTransitionTo(TransferStatuses.Received) {
		return ErrInvalidTransition
	}

	lines, err := loadTransferLineMovements(tx, companyID, h.ID)
	if err != nil {
		return err
	}

	for _, l := range lines {
		// IN to destination (positive).
		if err := s.recordMovement(
			tx, companyID,
			l.VariantID, h.ToWh, l.UnitID,
			l.Qty, l.UnitCost,
			InventoryMovementKinds.Transfer,
			string(InventoryMovementKinds.Transfer), h.ID,
		); err != nil {
			return err
		}
	}

	if err := setTransferStatus(tx, companyID, h.ID, TransferStatuses.Received, map[string]any{
		"received_at": time.Now().UTC(),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// cancelTransfer moves a transfer requested -> cancelled. Only allowed before
// dispatch; no stock has moved yet so there is nothing to reverse.
func (s *Server) cancelTransfer(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	h, err := loadTransferForUpdate(tx, companyID, uuid)
	if err != nil {
		return err
	}
	if !h.Status.CanTransitionTo(TransferStatuses.Cancelled) {
		return ErrInvalidTransition
	}

	if err := setTransferStatus(tx, companyID, h.ID, TransferStatuses.Cancelled, nil); err != nil {
		return err
	}

	return tx.Commit()
}

// findTransferItems searches trackable products for the transfer line picker.
// Each hit is an item with its default variant's sku and cost_price.
func (s *Server) findTransferItems(ctx context.Context, companyID int, search string) ([]*transferItemOption, error) {
	rows, err := s.db.QueryContext(ctx, transferItemQuery+" ORDER BY i.name, iv.is_default DESC, iv.name LIMIT 25", companyID, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*transferItemOption
	for rows.Next() {
		opt, err := scanTransferItem(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, opt)
	}
	return result, rows.Err()
}

// findTransferItem returns the first trackable product matching the search term,
// used by the reference-enter quick add on the transfer form.
func (s *Server) findTransferItem(ctx context.Context, companyID int, search string) (*transferItemOption, error) {
	rows, err := s.db.QueryContext(ctx, transferItemQuery+" ORDER BY i.name, iv.is_default DESC, iv.name LIMIT 1", companyID, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	return scanTransferItem(rows)
}

const transferItemQuery = `
	SELECT i.id,
	       iv.id                                     AS variant_id,
	       CASE WHEN iv.is_default THEN i.name
	            ELSE i.name || ' — ' || iv.name END  AS name,
	       COALESCE(i.description, '')               AS description,
	       COALESCE(i.identifiers->>'reference', '') AS reference,
	       COALESCE(iv.sku, '')                      AS sku,
	       COALESCE(iv.cost_price, 0)                AS cost,
	       iu.unit_id,
	       iu.unit_name
	  FROM items i
	  JOIN items_variants iv
	    ON iv.item_id = i.id AND iv.company_id = i.company_id
	   AND iv.track_inventory = TRUE AND iv.deleted_at IS NULL
	  LEFT JOIN LATERAL (
	        SELECT iu.unit_id, u.name AS unit_name
	          FROM items_units iu JOIN units u ON u.id = iu.unit_id
	         WHERE iu.item_id = i.id LIMIT 1
	  ) iu ON TRUE
	 WHERE i.company_id = $1 AND i.deleted_at IS NULL
	   AND (i.name ILIKE $2 OR i.identifiers->>'reference' ILIKE $2
	        OR iv.sku ILIKE $2 OR iv.name ILIKE $2
	        OR iv.reference ILIKE $2 OR iv.barcode ILIKE $2)`

func scanTransferItem(rows *sql.Rows) (*transferItemOption, error) {
	var o transferItemOption
	var unitID sql.NullInt64
	var unitName sql.NullString
	if err := rows.Scan(&o.ID, &o.VariantID, &o.Name, &o.Description, &o.Reference, &o.SKU, &o.Cost, &unitID, &unitName); err != nil {
		return nil, err
	}
	if unitID.Valid {
		id := unitID.Int64
		o.Unit.ID = &id
	}
	if unitName.Valid {
		name := unitName.String
		o.Unit.Name = &name
	}
	return &o, nil
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
