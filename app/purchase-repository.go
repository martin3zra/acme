package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type linkedPurchaseReceipt struct {
	ID     int       `json:"id"`
	UUID   string    `json:"uuid"`
	Number string    `json:"number"`
	Date   time.Time `json:"date"`
}

type purchase struct {
	CompanyID      int                      `json:"company_id"`
	ID             int                      `json:"id"`
	UUID           string                   `json:"uuid"`
	Number         string                   `json:"number"` // purchases.code
	Vendor         vendor                   `json:"vendor"`
	WarehouseID    int                      `json:"warehouse_id"`
	Date           time.Time                `json:"date"`
	DueOn          *time.Time               `json:"due_on"` // purchases.due_date
	Terms          string                   `json:"terms"`  // computed from date/due_on
	Amount         float64                  `json:"amount"` // purchases.subtotal
	Discount       Discount                 `json:"discount"`
	Tax            float64                  `json:"tax"` // purchases.tax_amount
	Total          float64                  `json:"total"`
	AmountDue      float64                  `json:"amount_due"`               // computed
	InvoiceNumber  string                   `json:"invoice_number,omitempty"` // vendor-supplied reference
	Status         string                   `json:"status"`                   // purchases.purchase_status
	PaymentStatus  PaidStatus               `json:"payment_status"`
	Notes          string                   `json:"notes"` // purchases.notes
	Kind           PurchaseTransactionKind  `json:"transaction_kind"`
	Source         *PurchaseSource          `json:"source,omitempty"`
	LinkedReceipts []*linkedPurchaseReceipt `json:"linked_receipts,omitempty"`

	EntityStatus foundation.Status `json:"-"` // purchases.status
}

// Both header reads drop `INNER JOIN companies` (existence only — company_id is a
// NOT NULL FK, and the company_id predicate already scopes the query) and turn
// `INNER JOIN vendors` into a belongsTo eager load. That eager load needs
// WithTrashed: vendorRead is softdelete-tagged, but the join never filtered
// vendors.deleted_at, and a purchase from a since-deleted vendor must still render.
//
// purchaseRead's softdelete tag supplies the `p.deleted_at IS NULL` both carried.

func (s *Server) findPurchases(ctx context.Context, kind PurchaseTransactionKind) ([]*purchase, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []purchaseRead
	if err := pdb.Model(&purchaseRead{}).
		Select(purchaseListColumns...).
		WithConstraint("Vendor", withTrashedRelation).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("transaction_kind", string(kind)).
		OrderBy("id", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*purchase, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toPurchase())
	}
	return data, nil
}

func (s *Server) findPurchaseByUUID(ctx context.Context, companyID int, uuid string) (*purchase, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row purchaseRead
	if err := pdb.Model(&purchaseRead{}).
		WithConstraint("Vendor", withTrashedRelation).
		WhereEq("company_id", companyID).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}
	return row.toPurchase(), nil
}

func (s *Server) findPurchaseLines(ctx context.Context, companyID, purchaseID int) ([]*line, error) {
	rows, err := s.db.Query(`
    SELECT it.id,
    pi.variant_id::bigint,
    iv.name,
    COALESCE(iv.sku, ''),
    pi.qty::bigint,
    pi.unit_price::float8,
    COALESCE(pi.unit_id, items_units.unit_id),
    CASE WHEN iv.is_default THEN it.name ELSE it.name || ' — ' || iv.name END,
    it.description,
    COALESCE(unit_selected.name, items_units.name),
    pi.created_at,
    pi.updated_at,
    pi.deleted_at,
    'unchanged' as action,
    (pi.qty * pi.unit_price)::float8 as amount,
    pi.line_total::float8,
    taxes.id as tax_id,
    taxes.name as tax_name,
    taxes.rate,
    pi.tax_amount::float8,
    it.identifiers
    FROM purchase_items AS pi
    INNER JOIN companies AS com ON (pi.company_id = com.id)
    INNER JOIN purchases AS p ON (pi.purchase_id = p.id AND pi.company_id = p.company_id)
    INNER JOIN items_variants AS iv ON (pi.variant_id = iv.id AND pi.company_id = iv.company_id)
    INNER JOIN items AS it ON (iv.item_id = it.id AND iv.company_id = it.company_id)
    LEFT JOIN units unit_selected ON (pi.unit_id = unit_selected.id)
    LEFT JOIN LATERAL (
      SELECT items_units.unit_id, units.name
      FROM items_units
      INNER JOIN units ON (items_units.unit_id = units.id)
      WHERE items_units.item_id = it.id limit 1
    ) items_units ON true
    INNER JOIN taxes ON (it.company_id = taxes.company_id AND COALESCE(pi.tax_id, it.tax_id) = taxes.id)
    WHERE pi.company_id = $1
    AND pi.purchase_id = $2
    AND pi.deleted_at IS NULL`, companyID, purchaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*line, 0)
	for rows.Next() {
		l := new(line)
		if err = rows.Scan(
			&l.ID,
			&l.VariantID,
			&l.VariantName,
			&l.VariantSKU,
			&l.Qty,
			&l.Price,
			&l.Unit.ID,
			&l.Name,
			&l.Description,
			&l.Unit.Name,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
			&l.Action,
			&l.Amount,
			&l.Total,
			&l.Tax.ID,
			&l.Tax.Name,
			&l.Tax.Rate,
			&l.Tax.Amount,
			&l.Identifier,
		); err != nil {
			return nil, err
		}
		data = append(data, l)
	}

	return data, nil
}

// enrichLinesWithRemainingQty sets RemainingQty on each line for a purchase order,
// calculated as ordered qty minus total qty already received across all linked receipts.
func (s *Server) enrichLinesWithRemainingQty(ctx context.Context, companyID int, purchaseOrderUUID string, lines []*line) error {
	rows, err := s.db.QueryContext(ctx,
		"SELECT pi.variant_id, COALESCE(SUM(pi.qty), 0)::bigint "+
			"FROM purchase_items pi "+
			"JOIN purchases p ON pi.purchase_id = p.id AND pi.company_id = p.company_id "+
			"WHERE p.source->>'id' = $1 AND p.transaction_kind = 'purchase_receipt' "+
			"AND p.company_id = $2 AND pi.deleted_at IS NULL AND p.deleted_at IS NULL "+
			"GROUP BY pi.variant_id",
		purchaseOrderUUID, companyID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	received := make(map[int64]int64)
	for rows.Next() {
		var variantID, qty int64
		if err = rows.Scan(&variantID, &qty); err != nil {
			return err
		}
		received[variantID] = qty
	}

	for _, l := range lines {
		remaining := l.Qty - received[l.VariantID]
		if remaining < 0 {
			remaining = 0
		}
		l.RemainingQty = &remaining
	}
	return nil
}

// findLinkedReceiptsForOrder finds the receipts booked against a purchase order.
//
// WhereJSON renders the Postgres path form `source #>> '{id}'`, which is equivalent
// to the old `source->>'id'` for a top-level key. There is no expression index on
// purchases.source, so nothing regresses.
func (s *Server) findLinkedReceiptsForOrder(ctx context.Context, companyID int, purchaseOrderUUID string) ([]*linkedPurchaseReceipt, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []purchaseRead
	if err := pdb.Model(&purchaseRead{}).
		Select("id", "uuid", "code", "date").
		WhereEq("company_id", companyID).
		WhereJSON("source", "id", "=", purchaseOrderUUID).
		WhereEq("transaction_kind", string(PurchaseTransactionKinds.PurchaseReceipt)).
		OrderBy("id", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*linkedPurchaseReceipt, 0, len(rows))
	for _, r := range rows {
		data = append(data, &linkedPurchaseReceipt{
			ID: r.ID, UUID: r.UUID, Number: r.Code, Date: r.Date,
		})
	}
	return data, nil
}

func (s *Server) storePurchase(ctx context.Context, form *StorePurchaseForm) (string, error) {
	companyID := CurrentCompany(ctx).ID
	var purchaseID string

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var err error
		purchaseID, err = s.storePurchaseInternal(tx, companyID, form)
		return err
	})

	return purchaseID, err
}

// firstWarehouseID returns the company's lowest-numbered live warehouse. The
// `deleted_at IS NULL` comes from warehouseRead's softdelete tag.
func firstWarehouseID(tx *sql.Tx, companyID int) (int, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return 0, err
	}

	var row warehouseRead
	if err := ptx.Model(&warehouseRead{}).
		Select("id").
		WhereEq("company_id", companyID).
		OrderBy("id", playsql.Asc).
		First(context.Background(), &row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

// resolveVariantsForLines resolves the variant every sales/purchase line
// transacts and writes it back onto each line's VariantID, in two queries total
// regardless of line count. Rules per line:
//   - names a variant: it must belong to the line's item (company, item, not deleted);
//   - names none: the item's default variant is used, EXCEPT an item flagged
//     has_variants MUST name a variant explicitly (no silent guess across variants).
func resolveVariantsForLines(tx *sql.Tx, companyID int, lines []*Line) error {
	if len(lines) == 0 {
		return nil
	}

	// Collect the distinct item set (for default resolution) and the explicit
	// variant ids named by lines (for ownership validation).
	itemIDSet := make(map[int]struct{}, len(lines))
	variantIDSet := make(map[int]struct{})
	for _, l := range lines {
		itemIDSet[l.ID] = struct{}{}
		if l.VariantID != 0 {
			variantIDSet[l.VariantID] = struct{}{}
		}
	}
	itemIDs := make([]int, 0, len(itemIDSet))
	for id := range itemIDSet {
		itemIDs = append(itemIDs, id)
	}

	// Query 1: per item, its has_variants flag and default variant id.
	type itemInfo struct {
		hasVariants    bool
		defaultVariant sql.NullInt64
	}
	items := make(map[int]itemInfo, len(itemIDs))
	rows, err := tx.Query(
		`SELECT i.id, i.has_variants,
		        (SELECT iv.id FROM items_variants iv
		           WHERE iv.company_id = i.company_id AND iv.item_id = i.id AND iv.deleted_at IS NULL
		           ORDER BY iv.is_default DESC, iv.id
		           LIMIT 1)
		   FROM items i
		  WHERE i.company_id = $1 AND i.id = ANY($2)`,
		companyID, pq.Array(itemIDs),
	)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var info itemInfo
		if err := rows.Scan(&id, &info.hasVariants, &info.defaultVariant); err != nil {
			rows.Close()
			return err
		}
		items[id] = info
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Query 2: for every explicitly named variant, the item it belongs to
	// (deleted variants excluded, so a soft-deleted variant can't be transacted).
	variantOwner := make(map[int]int, len(variantIDSet))
	if len(variantIDSet) > 0 {
		variantIDs := make([]int, 0, len(variantIDSet))
		for id := range variantIDSet {
			variantIDs = append(variantIDs, id)
		}
		vrows, err := tx.Query(
			`SELECT id, item_id FROM items_variants
			  WHERE company_id = $1 AND id = ANY($2) AND deleted_at IS NULL`,
			companyID, pq.Array(variantIDs),
		)
		if err != nil {
			return err
		}
		for vrows.Next() {
			var variantID, itemID int
			if err := vrows.Scan(&variantID, &itemID); err != nil {
				vrows.Close()
				return err
			}
			variantOwner[variantID] = itemID
		}
		if err := vrows.Err(); err != nil {
			vrows.Close()
			return err
		}
		vrows.Close()
	}

	// Apply the rules in-memory, mutating each line's resolved variant.
	for _, l := range lines {
		if l.VariantID != 0 {
			if owner, ok := variantOwner[l.VariantID]; !ok || owner != l.ID {
				return fmt.Errorf("variant %d does not belong to item %d", l.VariantID, l.ID)
			}
			continue
		}
		info, ok := items[l.ID]
		if !ok {
			return fmt.Errorf("missing item variant for item_id=%d", l.ID)
		}
		if info.hasVariants {
			return fmt.Errorf("item %d requires an explicit variant", l.ID)
		}
		if !info.defaultVariant.Valid {
			return fmt.Errorf("missing item variant for item_id=%d", l.ID)
		}
		l.VariantID = int(info.defaultVariant.Int64)
	}
	return nil
}

func resolveItemTaxIDs(tx *sql.Tx, companyID int, itemIDs []int) (map[int]*int, error) {
	rows, err := tx.Query(
		`SELECT i.id, i.tax_id
     FROM items i
     WHERE i.company_id = $1 AND i.id = ANY($2)`,
		companyID,
		pq.Array(itemIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int]*int, len(itemIDs))
	for rows.Next() {
		var itemID int
		var taxID sql.NullInt64
		if err := rows.Scan(&itemID, &taxID); err != nil {
			return nil, err
		}
		if taxID.Valid {
			v := int(taxID.Int64)
			m[itemID] = &v
		} else {
			m[itemID] = nil
		}
	}
	return m, nil
}

// sourceValue unwraps the optional jsonb source blob. A nil pointer becomes a nil
// map value, which playsql writes as SQL NULL.
func sourceValue(source *[]byte) any {
	if source == nil {
		return nil
	}
	return *source
}

func (s *Server) storePurchaseInternal(tx *sql.Tx, companyID int, form *StorePurchaseForm) (string, error) {
	seqInfo, err := GetNextSequence(tx, companyID, string(form.Kind))
	if err != nil {
		return "", err
	}

	warehouseID, err := firstWarehouseID(tx, companyID)
	if err != nil {
		return "", err
	}

	var source *[]byte
	if form.Source != nil && form.Source.ID != "" {
		j := foundation.AsJSON(form.Source)
		source = &j
	}

	discountAmount := float64(0)
	for _, l := range form.Lines {
		discountAmount += l.discount
	}

	isConversionReceipt := form.Kind == PurchaseTransactionKinds.PurchaseReceipt && form.Source != nil && form.Source.ID != ""

	// Receipts and vendor bills are always created in 'draft' so the user can
	// review them before explicitly confirming. Confirmation triggers inventory
	// movements and advances the status to 'received' (receipt) or 'posted' (bill).
	needsDraftStatus := form.Kind == PurchaseTransactionKinds.PurchaseReceipt ||
		form.Kind == PurchaseTransactionKinds.VendorBill

	ptx, err := playTx(tx)
	if err != nil {
		return "", err
	}

	// The old statement concatenated columns and renumbered placeholders by hand to
	// leave purchase_status and invoice_number at their defaults. A map insert says
	// the same thing: a key that is absent is a column that is not written.
	row := map[string]any{
		"company_id":       companyID,
		"vendor_id":        form.VendorID,
		"warehouse_id":     warehouseID,
		"transaction_kind": string(form.Kind),
		"notes":            form.Notes,
		"subtotal":         form.amount,
		"discount_amount":  discountAmount,
		"tax_amount":       form.tax,
		"total":            form.total,
		"payment_status":   string(form.paymentStatus),
		"code":             seqInfo.Code,
		"source":           sourceValue(source),
		"date":             form.Date,
		"due_date":         form.dueOn,
	}
	if needsDraftStatus {
		row["purchase_status"] = "draft"
	}
	if form.InvoiceNumber != "" {
		row["invoice_number"] = form.InvoiceNumber
	}

	id, err := ptx.Model(&purchaseRead{}).Insert(context.Background(), row)
	if err != nil {
		return "", err
	}
	purchaseID := int(id)

	// Insert returns the pk only; uuid is DB-generated, so it is read back.
	var stored purchaseRead
	if err := ptx.Model(&purchaseRead{}).
		Select("uuid").
		WhereEq("id", purchaseID).
		First(context.Background(), &stored); err != nil {
		return "", err
	}
	purchaseUUID := stored.UUID

	if err := s.attachPurchaseLines(tx, companyID, purchaseID, form); err != nil {
		return "", err
	}

	// When saving a vendor bill, automatically create the AP entry.
	if form.Kind == PurchaseTransactionKinds.VendorBill {
		if err := s.createAPForVendorBill(tx, companyID, purchaseID, form.VendorID, form); err != nil {
			return "", err
		}
	}

	if isConversionReceipt {
		poStatus, err := resolveReceivedStatus(tx, companyID, form.Source.ID)
		if err != nil {
			return "", err
		}

		if _, err := ptx.Model(&purchaseRead{}).
			WhereEq("company_id", companyID).
			WhereEq("uuid", form.Source.ID).
			Update(context.Background(), map[string]any{
				"purchase_status": poStatus,
				"source": foundation.AsJSON(map[string]any{
					"type": string(form.Kind),
					"id":   purchaseUUID,
					"code": seqInfo.Code,
				}),
			}); err != nil {
			return "", err
		}

		// Invalidate the source PO's cached preview so the updated status is shown.
		poCache := cache.NewPgCache(tx)
		_ = poCache.Delete(context.Background(), fmt.Sprintf("preview:purchase:%s", form.Source.ID))
	}

	// When saving a vendor bill that was converted from a receipt, stamp the
	// receipt's source.target so the receipt list view can show a forward link.
	isConversionVendorBill := form.Kind == PurchaseTransactionKinds.VendorBill && form.Source != nil && form.Source.ID != ""
	if isConversionVendorBill {
		// The old COALESCE(source, '{}') is unnecessary: playsql scans a SQL NULL as
		// the field's zero value, and unmarshalling an empty slice leaves the map nil —
		// which the branch below already handles.
		var receipt purchaseRead
		if err := ptx.Model(&purchaseRead{}).
			Select("source").
			WhereEq("company_id", companyID).
			WhereEq("uuid", form.Source.ID).
			First(context.Background(), &receipt); err != nil {
			return "", err
		}

		var existingSource map[string]any
		if jsonErr := json.Unmarshal(receipt.Source, &existingSource); jsonErr != nil || existingSource == nil {
			existingSource = map[string]any{}
		}
		existingSource["target"] = map[string]any{
			"type": string(PurchaseTransactionKinds.VendorBill),
			"id":   purchaseUUID,
			"code": seqInfo.Code,
		}

		if _, err := ptx.Model(&purchaseRead{}).
			WhereEq("company_id", companyID).
			WhereEq("uuid", form.Source.ID).
			Update(context.Background(), map[string]any{
				"source": foundation.AsJSON(existingSource),
			}); err != nil {
			return "", err
		}
	}

	// Inventory IN movements are NOT recorded at creation time.
	// They are deferred to the explicit "confirm" action (PUT /purchases/:id/confirm)
	// which transitions the document from 'draft' to 'received' (receipt) or 'posted'
	// (vendor bill) and atomically records the movements.

	return purchaseUUID, nil
}

func (s *Server) createAPForVendorBill(tx *sql.Tx, companyID, purchaseID, vendorID int, form *StorePurchaseForm) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	// amount_total is the pre-tax subtotal; amount_payable is a generated column
	// (amount_total + tax_amount - discount_amount) and is not written here.
	ap := &accountsPayableInsert{
		CompanyID:      companyID,
		VendorID:       vendorID,
		PurchaseID:     purchaseID,
		InvoiceNumber:  form.InvoiceNumber,
		InvoiceDate:    form.Date,
		DueDate:        form.dueOn,
		AmountTotal:    form.amount,
		TaxAmount:      form.tax,
		DiscountAmount: 0,
		AmountPaid:     0,
		Currency:       "DOP",
		PaymentTerms:   form.Terms,
		Status:         PayableStatuses.Pending,
		PaidStatus:     PaidStatuses.UnPaid,
		CreatedBy:      form.User().GetAuthIdentifier(),
	}
	if err = ptx.Insert(context.Background(), ap); err != nil {
		return err
	}
	apID := int(ap.ID)

	if err := s.registerPayable(tx, companyID, apID, vendorID); err != nil {
		return err
	}

	return s.updateVendorAmountPayable(tx, companyID, vendorID, form.total)
}

func (s *Server) attachPurchaseLines(tx *sql.Tx, companyID, purchaseID int, form *StorePurchaseForm) error {
	itemIDs := make([]int, 0, len(form.Lines))
	for _, l := range form.Lines {
		itemIDs = append(itemIDs, l.ID)
	}

	taxIDs, err := resolveItemTaxIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}
	if err := resolveVariantsForLines(tx, companyID, form.Lines); err != nil {
		return err
	}

	rows := make([]map[string]any, 0, len(form.Lines))
	for _, l := range form.Lines {
		rows = append(rows, map[string]any{
			"company_id":  companyID,
			"purchase_id": purchaseID,
			"variant_id":  l.VariantID,
			"qty":         l.Qty,
			"unit_price":  l.Price,
			"line_total":  l.total,
			"unit_id":     l.Unit,
			"discount":    l.discount,
			"tax_id":      taxIDs[l.ID],
			"tax_amount":  l.tax,
		})
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	_, err = ptx.Model(&PurchaseItem{}).InsertMany(context.Background(), rows)
	return err
}

func (s *Server) updatePurchase(ctx context.Context, uuid string, form *UpdatePurchaseForm) error {
	companyID := CurrentCompany(ctx).ID
	purchase, err := s.findPurchaseByUUID(ctx, companyID, uuid)
	if err != nil {
		return err
	}

	// Block edits on closed purchase orders.
	if purchase.Kind == PurchaseTransactionKinds.PurchaseOrder &&
		PurchaseStatus(purchase.Status) == PurchaseStatuses.Closed {
		return fmt.Errorf("purchase order %s is closed and cannot be edited", purchase.Number)
	}

	// Lock header fields (vendor, date) once a receipt exists for the PO.
	if purchase.Kind == PurchaseTransactionKinds.PurchaseOrder {
		var receiptCount int
		if err := s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM purchases WHERE company_id = $1 AND source->>'id' = $2 AND transaction_kind = 'purchase_receipt' AND deleted_at IS NULL",
			companyID, uuid,
		).Scan(&receiptCount); err != nil {
			return err
		}
		if receiptCount > 0 {
			if purchase.Vendor.ID != form.VendorID {
				return fmt.Errorf("vendor cannot be changed after a purchase receipt has been created for this order")
			}
			if !purchase.Date.Equal(form.Date) {
				return fmt.Errorf("order date cannot be changed after a purchase receipt has been created for this order")
			}
		}
	}

	discountAmount := float64(0)
	for _, l := range form.Lines {
		if l.Action == LineActions.Deleted {
			continue
		}
		discountAmount += l.discount
	}

	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// purchase.ID came from a company-scoped read, so this must match a row.
		affected, err := ptx.Model(&purchaseRead{}).
			WhereEq("company_id", companyID).
			WhereEq("id", purchase.ID).
			Update(context.Background(), map[string]any{
				"vendor_id":        form.VendorID,
				"date":             form.Date,
				"due_date":         form.dueOn,
				"subtotal":         form.amount,
				"discount_amount":  discountAmount,
				"tax_amount":       form.tax,
				"total":            form.total,
				"notes":            form.Notes,
				"payment_status":   string(form.paymentStatus),
				"transaction_kind": string(form.Kind),
				"invoice_number":   form.InvoiceNumber,
			})
		if err := mustAffectRows(affected, err, "purchase"); err != nil {
			return err
		}

		if err := s.processPurchaseLines(tx, companyID, purchase.ID, form); err != nil {
			return err
		}

		// Re-sync the linked AP record when a vendor bill is updated.
		if purchase.Kind == PurchaseTransactionKinds.VendorBill {
			_, err = tx.Exec(
				`UPDATE accounts_payable
				   SET invoice_number = $3,
				       invoice_date   = $4,
				       due_date       = $5,
				       amount_total   = $6,
				       tax_amount     = $7,
				       updated_at     = NOW()
				 WHERE company_id = $1 AND purchase_id = $2`,
				companyID, purchase.ID,
				form.InvoiceNumber,
				form.Date,
				form.dueOn,
				form.total,
				form.tax,
			)
			if err != nil {
				return err
			}
			// Adjust vendor.amount_payable by the delta.
			delta := form.total - purchase.Total
			if delta != 0 {
				if err = s.updateVendorAmountPayable(tx, companyID, purchase.Vendor.ID, delta); err != nil {
					return err
				}
			}
		}

		// Re-evaluate the source purchase order status when a receipt is updated.
		if purchase.Kind == PurchaseTransactionKinds.PurchaseReceipt && purchase.Source != nil && purchase.Source.ID != "" {
			poStatus, err := resolveReceivedStatus(tx, companyID, purchase.Source.ID)
			if err != nil {
				return err
			}
			_, err = tx.Exec(
				"UPDATE purchases SET purchase_status = $3, updated_at = NOW() WHERE company_id = $1 AND uuid = $2",
				companyID, purchase.Source.ID, poStatus,
			)
			if err != nil {
				return err
			}
			// Invalidate the source PO's cached preview within the same transaction.
			poCache := cache.NewPgCache(tx)
			_ = poCache.Delete(context.Background(), fmt.Sprintf("preview:purchase:%s", purchase.Source.ID))
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Server) processPurchaseLines(tx *sql.Tx, companyID, purchaseID int, form *UpdatePurchaseForm) error {
	lines := s.filterInvoiceLines(form.Lines, ADDED, UPDATED, DELETED)
	if len(lines) == 0 {
		return nil
	}

	itemIDs := make([]int, 0, len(lines))
	for _, l := range lines {
		itemIDs = append(itemIDs, l.ID)
	}

	taxIDs, err := resolveItemTaxIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}
	if err := resolveVariantsForLines(tx, companyID, lines); err != nil {
		return err
	}

	for _, l := range lines {
		variantID := l.VariantID

		// Guard: never reduce a PO line below the already-received qty.
		if form.Kind == PurchaseTransactionKinds.PurchaseOrder &&
			(l.Action == UPDATED || l.Action == DELETED) {
			var alreadyReceived float64
			err := tx.QueryRow(
				`SELECT COALESCE(SUM(pi.qty), 0)
				   FROM purchase_items pi
				   JOIN purchases p ON pi.purchase_id = p.id
				  WHERE p.company_id = $1
				    AND p.source->>'id' = (SELECT uuid FROM purchases WHERE id = $2 LIMIT 1)
				    AND p.transaction_kind = 'purchase_receipt'
				    AND pi.variant_id = $3
				    AND pi.deleted_at IS NULL`,
				companyID, purchaseID, variantID,
			).Scan(&alreadyReceived)
			if err != nil {
				return err
			}
			if l.Action == DELETED && alreadyReceived > 0 {
				return fmt.Errorf("line with item_id=%d cannot be removed: %g unit(s) have already been received", l.ID, alreadyReceived)
			}
			if l.Action == UPDATED && float64(l.Qty) < alreadyReceived {
				return fmt.Errorf("line with item_id=%d cannot be reduced below %g (already received)", l.ID, alreadyReceived)
			}
		}

		switch l.Action {
		case ADDED:
			stmt := `
        INSERT INTO purchase_items (company_id, purchase_id, variant_id, qty, unit_price, line_total, unit_id, discount, tax_id, tax_amount)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
      `
			if _, err := tx.Exec(stmt, companyID, purchaseID, variantID, l.Qty, l.Price, l.total, l.Unit, l.discount, taxIDs[l.ID], l.tax); err != nil {
				return err
			}
		case UPDATED:
			stmt := `
        UPDATE purchase_items
        SET qty = $4,
            unit_id = $5,
            unit_price = $6,
            line_total = $7,
            discount = $8,
            tax_id = $9,
            tax_amount = $10,
            updated_at = NOW(),
            deleted_at = NULL
        WHERE company_id = $1 AND purchase_id = $2 AND variant_id = $3
      `
			if _, err := tx.Exec(stmt, companyID, purchaseID, variantID, l.Qty, l.Unit, l.Price, l.total, l.discount, taxIDs[l.ID], l.tax); err != nil {
				return err
			}
		case DELETED:
			stmt := `
        UPDATE purchase_items
        SET deleted_at = NOW(), updated_at = NOW()
        WHERE company_id = $1 AND purchase_id = $2 AND variant_id = $3
      `
			if _, err := tx.Exec(stmt, companyID, purchaseID, variantID); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func (s *Server) destroyPurchase(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID
	purchase, err := s.findPurchaseByUUID(ctx, companyID, uuid)
	if err != nil {
		return err
	}

	// Block deletion of purchase orders that have already been received or paid.
	if purchase.Kind == PurchaseTransactionKinds.PurchaseOrder &&
		lockedPurchaseStatuses[PurchaseStatus(purchase.Status)] {
		return fmt.Errorf("purchase order %s cannot be deleted in status %q", purchase.Number, purchase.Status)
	}

	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// Soft delete through Update, not Delete: Builder.Delete stamps deleted_at
		// only, and the statement it replaced bumped updated_at too. purchaseRead
		// maps updated_at, so playsql stamps it.
		if _, err := ptx.Model(&purchaseRead{}).
			WhereEq("company_id", companyID).
			WhereEq("id", purchase.ID).
			Update(context.Background(), map[string]any{"deleted_at": time.Now()}); err != nil {
			return err
		}

		// Reverse inventory movements recorded for this purchase. Read outside the
		// transaction, and errors ignored, exactly as before.
		var movementRecorded bool
		if pdb, perr := s.play(); perr == nil {
			var row purchaseRead
			if rerr := pdb.Model(&purchaseRead{}).
				Select("movement_recorded").
				WithTrashed().
				WhereEq("company_id", companyID).
				WhereEq("id", purchase.ID).
				First(ctx, &row); rerr == nil {
				movementRecorded = row.MovementRecorded
			}
		}
		if movementRecorded {
			if err := s.reverseMovements(tx, companyID, "purchase", purchase.ID, InventoryMovementKinds.PurchaseReturn); err != nil {
				return err
			}
		}

		// Cancel the linked AP entry when a vendor bill is deleted.
		if purchase.Kind == PurchaseTransactionKinds.VendorBill {
			var ap accountsPayableRead
			err := ptx.Model(&accountsPayableRead{}).
				Select("id", "amount_paid").
				WhereEq("company_id", companyID).
				WhereEq("purchase_id", purchase.ID).
				First(context.Background(), &ap)
			if err != nil && !errors.Is(err, playsql.ErrNotFound) {
				return err
			}
			if ap.ID > 0 {
				if _, err = ptx.Model(&accountsPayableInsert{}).
					WhereEq("company_id", companyID).
					WhereEq("id", ap.ID).
					Update(context.Background(), map[string]any{
						"status":     string(PayableStatuses.Void),
						"updated_at": time.Now(),
					}); err != nil {
					return err
				}
				// Reverse only the unpaid portion from the vendor balance.
				remaining := purchase.Total - ap.AmountPaid
				if remaining > 0 {
					if err = s.updateVendorAmountPayable(tx, companyID, purchase.Vendor.ID, -remaining); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:purchase:%s", uuid)
	_ = c.Delete(ctx, key)
	return nil
}

// resolveReceivedStatus determines whether a purchase order should be marked
// as 'received' or 'partially_received' by comparing the total quantity of its
// original line items against the total quantity received across all linked
// purchase receipts (including any just inserted in the current transaction).
func resolveReceivedStatus(tx *sql.Tx, companyID int, sourceUUID string) (string, error) {
	var totalOrdered, totalReceived float64

	err := tx.QueryRow(
		"SELECT COALESCE(SUM(pi.qty), 0) FROM purchase_items pi "+
			"JOIN purchases p ON pi.purchase_id = p.id "+
			"WHERE p.uuid = $1 AND p.company_id = $2",
		sourceUUID, companyID,
	).Scan(&totalOrdered)
	if err != nil {
		return "", err
	}

	err = tx.QueryRow(
		"SELECT COALESCE(SUM(pi.qty), 0) FROM purchase_items pi "+
			"JOIN purchases p ON pi.purchase_id = p.id "+
			"WHERE p.source->>'id' = $1 AND p.transaction_kind = 'purchase_receipt' AND p.company_id = $2",
		sourceUUID, companyID,
	).Scan(&totalReceived)
	if err != nil {
		return "", err
	}

	if totalReceived >= totalOrdered {
		return "received", nil
	}
	return "partially_received", nil
}

// resolvePOPaymentStatus calculates the payment_status for a PO based on the sum
// of amount_paid vs amount_payable across all linked accounts_payable records.
func resolvePOPaymentStatus(tx *sql.Tx, companyID int, poUUID string) (PaidStatus, error) {
	var totalPayable, totalPaid float64
	err := tx.QueryRow(
		`SELECT COALESCE(SUM(ap.amount_payable), 0), COALESCE(SUM(ap.amount_paid), 0)
		   FROM accounts_payable ap
		   JOIN purchases p ON ap.purchase_id = p.id
		  WHERE p.source->>'id' = $1
		    AND p.transaction_kind = 'vendor_bill'
		    AND ap.company_id = $2
		    AND ap.status != $3`,
		poUUID, companyID, string(PayableStatuses.Void),
	).Scan(&totalPayable, &totalPaid)
	if err != nil {
		return PaidStatuses.UnPaid, err
	}

	if totalPayable == 0 {
		return PaidStatuses.UnPaid, nil
	}
	if totalPaid >= totalPayable {
		return PaidStatuses.Paid, nil
	}
	if totalPaid > 0 {
		return PaidStatuses.Partial, nil
	}
	return PaidStatuses.UnPaid, nil
}

// updatePOPaymentStatus recalculates and writes payment_status on the source PO.
// When fully paid it also advances purchase_status to 'closed'.
//
// The three branches differ only in what they SET and in one extra predicate, so
// they collapse into one builder. Each is deliberately allowed to match no row —
// a closed PO is not demoted to partially_paid, and a paid one is not demoted at
// all — so no zero-row guard belongs here.
//
// playsql stamps updated_at because purchaseRead maps it, and purchaseRead's
// softdelete tag adds `deleted_at IS NULL`. The raw statements had no such
// predicate; a soft-deleted PO should not have its status rewritten.
func updatePOPaymentStatus(tx *sql.Tx, companyID int, poUUID string) error {
	newPaymentStatus, err := resolvePOPaymentStatus(tx, companyID, poUUID)
	if err != nil {
		return err
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	paid := newPaymentStatus == PaidStatuses.Paid
	partial := newPaymentStatus == PaidStatuses.Partial

	changes := map[string]any{"payment_status": string(newPaymentStatus)}
	switch {
	case paid:
		changes["purchase_status"] = string(PurchaseStatuses.Closed)
	case partial:
		changes["purchase_status"] = string(PurchaseStatuses.PartiallyPaid)
	}

	_, err = ptx.Model(&purchaseRead{}).
		WhereEq("company_id", companyID).
		WhereEq("uuid", poUUID).
		// A closed PO keeps its status; an already-paid PO is never demoted.
		When(partial, func(q *playsql.Builder) {
			q.Where("purchase_status", "!=", string(PurchaseStatuses.Closed))
		}).
		When(!paid && !partial, func(q *playsql.Builder) {
			q.Where("payment_status", "!=", string(PaidStatuses.Paid))
		}).
		Update(context.Background(), changes)
	return err
}

// recordPurchaseMovements reads the purchase_items for the given purchase and
// records an IN movement for each line that has track_inventory enabled.
// warehouseID is taken from the purchase header; kind distinguishes receipt vs vendor_bill.
func (s *Server) recordPurchaseMovements(tx *sql.Tx, companyID, purchaseID, warehouseID int, kind InventoryMovementKind) error {
	rows, err := tx.Query(
		`SELECT variant_id, qty, unit_id, unit_price
		   FROM purchase_items
		  WHERE company_id = $1 AND purchase_id = $2 AND deleted_at IS NULL`,
		companyID, purchaseID,
	)
	if err != nil {
		return fmt.Errorf("recordPurchaseMovements: query lines: %w", err)
	}

	type line struct {
		variantID, unitID int
		qty, unitPrice    float64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.variantID, &l.qty, &l.unitID, &l.unitPrice); err != nil {
			rows.Close()
			return fmt.Errorf("recordPurchaseMovements: scan: %w", err)
		}
		lines = append(lines, l)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("recordPurchaseMovements: rows close: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("recordPurchaseMovements: rows: %w", err)
	}

	for _, l := range lines {
		if err := s.recordMovement(
			tx, companyID,
			l.variantID, warehouseID, l.unitID,
			l.qty, l.unitPrice,
			kind,
			"purchase", purchaseID,
		); err != nil {
			return err
		}
	}
	return nil
}

// confirmPurchase transitions a purchase receipt or vendor bill from 'draft' to
// its confirmed state ('received' for receipts, 'posted' for vendor bills) and
// atomically records the inventory IN movements. This is the single place where
// stock is committed for incoming goods — movements are NEVER recorded at
// creation time.
//
// Rules:
//   - Only purchase_receipt and vendor_bill documents can be confirmed.
//   - The document must currently be in 'draft' status.
//   - For a vendor_bill linked to a purchase_receipt, the receipt already moved
//     the stock at confirmation time; only the status is advanced here.
//   - For a vendor_bill with no linked receipt (or linked to a PO), movements
//     are recorded now.
func (s *Server) confirmPurchase(ctx context.Context, uuid string) error {
	companyID := CurrentCompany(ctx).ID

	purchase, err := s.findPurchaseByUUID(ctx, companyID, uuid)
	if err != nil {
		return err
	}

	if purchase.Kind != PurchaseTransactionKinds.PurchaseReceipt &&
		purchase.Kind != PurchaseTransactionKinds.VendorBill {
		return fmt.Errorf("only purchase receipts and vendor bills can be confirmed")
	}

	if PurchaseStatus(purchase.Status) != PurchaseStatuses.Draft &&
		purchase.Status != "" {
		return fmt.Errorf("purchase %s is not in draft status", purchase.Number)
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var targetStatus PurchaseStatus
		if purchase.Kind == PurchaseTransactionKinds.PurchaseReceipt {
			targetStatus = PurchaseStatuses.Received
		} else {
			targetStatus = PurchaseStatuses.Posted
		}

		// Determine whether this document should record inventory movements.
		// A vendor bill converted from a purchase receipt skips movement recording
		// because the receipt already committed those movements.
		linkedToReceipt := purchase.Source != nil &&
			purchase.Source.Type == PurchaseTransactionKinds.PurchaseReceipt
		needsMovements := purchase.Kind == PurchaseTransactionKinds.PurchaseReceipt ||
			(purchase.Kind == PurchaseTransactionKinds.VendorBill && !linkedToReceipt)

		if needsMovements {
			var movementKind InventoryMovementKind
			if purchase.Kind == PurchaseTransactionKinds.PurchaseReceipt {
				movementKind = InventoryMovementKinds.PurchaseReceipt
			} else {
				movementKind = InventoryMovementKinds.VendorBill
			}

			if err := s.recordPurchaseMovements(tx, companyID, purchase.ID, purchase.WarehouseID, movementKind); err != nil {
				return err
			}
		}

		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// purchase.ID came from a company-scoped read, so this must match a row.
		affected, err := ptx.Model(&purchaseRead{}).
			WhereEq("company_id", companyID).
			WhereEq("id", purchase.ID).
			Update(context.Background(), map[string]any{
				"purchase_status":   string(targetStatus),
				"movement_recorded": needsMovements,
			})
		return mustAffectRows(affected, err, "purchase")
	})
}
