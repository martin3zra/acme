package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/martin3zra/acme/pkg/cache"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type linkedPurchaseReceipt struct {
	ID     int       `json:"id"`
	UUID   string    `json:"uuid"`
	Number string    `json:"number"`
	Date   time.Time `json:"date"`
}

type purchase struct {
	CompanyID      int                       `json:"company_id"`
	ID             int                       `json:"id"`
	UUID           string                    `json:"uuid"`
	Number         string                    `json:"number"` // purchases.code
	Vendor         vendor                    `json:"vendor"`
	WarehouseID    int                       `json:"warehouse_id"`
	Date           time.Time                 `json:"date"`
	DueOn          *time.Time                `json:"due_on"` // purchases.due_date
	Terms          string                    `json:"terms"`  // computed from date/due_on
	Amount         float64                   `json:"amount"` // purchases.subtotal
	Discount       Discount                  `json:"discount"`
	Tax            float64                   `json:"tax"` // purchases.tax_amount
	Total          float64                   `json:"total"`
	AmountDue      float64                   `json:"amount_due"`    // computed
	InvoiceNumber  string                    `json:"invoice_number,omitempty"` // vendor-supplied reference
	Status         string                    `json:"status"`        // purchases.purchase_status
	PaymentStatus  PaidStatus                `json:"payment_status"`
	Notes          string                    `json:"notes"` // purchases.notes
	Kind           PurchaseTransactionKind   `json:"transaction_kind"`
	Source         *PurchaseSource           `json:"source,omitempty"`
	LinkedReceipts []*linkedPurchaseReceipt  `json:"linked_receipts,omitempty"`

	EntityStatus foundation.Status `json:"-"` // purchases.status
}

func (s *Server) findPurchases(ctx context.Context, kind PurchaseTransactionKind) ([]*purchase, error) {
	rows, err := s.db.Query(
		"SELECT p.id, p.uuid, p.code, p.warehouse_id, p.date, p.due_date, p.subtotal, p.discount_amount, p.tax_amount, p.total, p.status, p.purchase_status, p.payment_status, COALESCE(p.notes, ''), p.transaction_kind, p.source, COALESCE(p.invoice_number, ''), "+
			"v.id, v.uuid, v.name, v.email, v.phone "+
			"FROM purchases p "+
			"INNER JOIN companies ON (p.company_id = companies.id) "+
			"INNER JOIN vendors v ON (p.company_id = v.company_id AND p.vendor_id = v.id) "+
			"WHERE p.company_id = $1 AND p.transaction_kind = $2 AND p.deleted_at IS NULL ORDER BY p.id DESC",
		CurrentCompany(ctx).ID, kind,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*purchase, 0)
	for rows.Next() {
		p := new(purchase)
		var discountAmount float64
		if err = rows.Scan(
			&p.ID,
			&p.UUID,
			&p.Number,
			&p.WarehouseID,
			&p.Date,
			&p.DueOn,
			&p.Amount,
			&discountAmount,
			&p.Tax,
			&p.Total,
			&p.EntityStatus,
			&p.Status,
			&p.PaymentStatus,
			&p.Notes,
			&p.Kind,
			&p.Source,
			&p.InvoiceNumber,
			&p.Vendor.ID,
			&p.Vendor.UUID,
			&p.Vendor.Name,
			&p.Vendor.Email,
			&p.Vendor.Phone,
		); err != nil {
			return nil, err
		}

		p.Discount = Discount{Val: discountAmount, Type: "fixed"}

		p.AmountDue = p.Total
		if p.PaymentStatus == PaidStatuses.Paid {
			p.AmountDue = 0
		}

		p.Terms = "pia"
		if p.DueOn != nil {
			difference := p.DueOn.Sub(p.Date)
			p.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
		}
		p.Vendor.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
		data = append(data, p)
	}

	return data, nil
}

func (s *Server) findPurchaseByUUID(ctx context.Context, companyID int, uuid string) (*purchase, error) {
	p := new(purchase)
	var discountAmount float64
	err := s.db.QueryRow(
		"SELECT p.company_id, p.id, p.uuid, p.code, p.warehouse_id, p.date, p.due_date, p.subtotal, p.discount_amount, p.tax_amount, p.total, p.status, p.purchase_status, p.payment_status, COALESCE(p.notes, ''), p.transaction_kind, p.source, COALESCE(p.invoice_number, ''), "+
			"v.id, v.uuid, v.name, v.email, v.phone "+
			"FROM purchases p "+
			"INNER JOIN companies ON (p.company_id = companies.id) "+
			"INNER JOIN vendors v ON (p.company_id = v.company_id AND p.vendor_id = v.id) "+
			"WHERE p.company_id = $1 AND p.uuid = $2 AND p.deleted_at IS NULL",
		companyID, uuid,
	).Scan(
		&p.CompanyID,
		&p.ID,
		&p.UUID,
		&p.Number,
		&p.WarehouseID,
		&p.Date,
		&p.DueOn,
		&p.Amount,
		&discountAmount,
		&p.Tax,
		&p.Total,
		&p.EntityStatus,
		&p.Status,
		&p.PaymentStatus,
		&p.Notes,
		&p.Kind,
		&p.Source,
		&p.InvoiceNumber,
		&p.Vendor.ID,
		&p.Vendor.UUID,
		&p.Vendor.Name,
		&p.Vendor.Email,
		&p.Vendor.Phone,
	)
	if err != nil {
		return nil, err
	}

	p.Discount = Discount{Val: discountAmount, Type: "fixed"}

	p.AmountDue = p.Total
	if p.PaymentStatus == PaidStatuses.Paid {
		p.AmountDue = 0
	}

	p.Terms = "pia"
	if p.DueOn != nil {
		difference := p.DueOn.Sub(p.Date)
		p.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
	}
	p.Vendor.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	return p, nil
}

func (s *Server) findPurchaseLines(ctx context.Context, companyID, purchaseID int) ([]*line, error) {
	rows, err := s.db.Query(`
    SELECT it.id,
    pi.variant_id::bigint,
    pi.qty::bigint,
    pi.unit_price::float8,
    COALESCE(pi.unit_id, items_units.unit_id),
    it.name,
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

func (s *Server) findLinkedReceiptsForOrder(ctx context.Context, companyID int, purchaseOrderUUID string) ([]*linkedPurchaseReceipt, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, uuid, code, date FROM purchases "+
			"WHERE company_id = $1 AND source->>'id' = $2 AND transaction_kind = 'purchase_receipt' AND deleted_at IS NULL "+
			"ORDER BY id",
		companyID, purchaseOrderUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*linkedPurchaseReceipt, 0)
	for rows.Next() {
		r := new(linkedPurchaseReceipt)
		if err = rows.Scan(&r.ID, &r.UUID, &r.Number, &r.Date); err != nil {
			return nil, err
		}
		data = append(data, r)
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

func firstWarehouseID(tx *sql.Tx, companyID int) (int, error) {
	var warehouseID int
	err := tx.QueryRow(
		"SELECT id FROM warehouses WHERE company_id = $1 AND deleted_at IS NULL ORDER BY id LIMIT 1",
		companyID,
	).Scan(&warehouseID)
	if err != nil {
		return 0, err
	}
	return warehouseID, nil
}

func resolveItemVariantIDs(tx *sql.Tx, companyID int, itemIDs []int) (map[int]int, error) {
	rows, err := tx.Query(
		`SELECT DISTINCT ON (iv.item_id) iv.item_id, iv.id
     FROM items_variants iv
     WHERE iv.company_id = $1 AND iv.item_id = ANY($2)
     ORDER BY iv.item_id, iv.id`,
		companyID,
		pq.Array(itemIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int]int, len(itemIDs))
	for rows.Next() {
		var itemID int
		var variantID int
		if err := rows.Scan(&itemID, &variantID); err != nil {
			return nil, err
		}
		m[itemID] = variantID
	}
	return m, nil
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

	insertCols := "INSERT INTO purchases (company_id, vendor_id, warehouse_id, transaction_kind, notes, subtotal, discount_amount, tax_amount, total, payment_status, code, source, date, due_date"
	valuesCols := "VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14"
	args := []any{
		companyID,
		form.VendorID,
		warehouseID,
		form.Kind,
		form.Notes,
		form.amount,
		discountAmount,
		form.tax,
		form.total,
		form.paymentStatus,
		seqInfo.Code,
		source,
		form.Date,
		form.dueOn,
	}

	if isConversionReceipt {
		insertCols += ", purchase_status"
		valuesCols += fmt.Sprintf(", $%d", len(args)+1)
		args = append(args, "draft")
	}

	if form.InvoiceNumber != "" {
		insertCols += ", invoice_number"
		valuesCols += fmt.Sprintf(", $%d", len(args)+1)
		args = append(args, form.InvoiceNumber)
	}

	var purchaseID int
	var purchaseUUID string
	err = tx.QueryRow(
		insertCols+") "+valuesCols+") RETURNING id, uuid",
		args...,
	).Scan(&purchaseID, &purchaseUUID)
	if err != nil {
		return "", err
	}

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

		_, err = tx.Exec(
			"UPDATE purchases SET purchase_status = $3, source = $4, updated_at = NOW() WHERE company_id = $1 AND uuid = $2",
			companyID, form.Source.ID,
			poStatus,
			foundation.AsJSON(map[string]any{
				"type": string(form.Kind),
				"id":   purchaseUUID,
				"code": seqInfo.Code,
			}),
		)
		if err != nil {
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
		var rawSource []byte
		err := tx.QueryRow(
			"SELECT COALESCE(source, '{}'::jsonb) FROM purchases WHERE company_id = $1 AND uuid = $2",
			companyID, form.Source.ID,
		).Scan(&rawSource)
		if err != nil {
			return "", err
		}

		var existingSource map[string]any
		if jsonErr := json.Unmarshal(rawSource, &existingSource); jsonErr != nil {
			existingSource = map[string]any{}
		}
		existingSource["target"] = map[string]any{
			"type": string(PurchaseTransactionKinds.VendorBill),
			"id":   purchaseUUID,
			"code": seqInfo.Code,
		}

		_, err = tx.Exec(
			"UPDATE purchases SET source = $3, updated_at = NOW() WHERE company_id = $1 AND uuid = $2",
			companyID, form.Source.ID,
			foundation.AsJSON(existingSource),
		)
		if err != nil {
			return "", err
		}
	}

	return purchaseUUID, nil
}

func (s *Server) createAPForVendorBill(tx *sql.Tx, companyID, purchaseID, vendorID int, form *StorePurchaseForm) error {
	var apID int
	err := tx.QueryRow(
		"INSERT INTO accounts_payable "+
			"(company_id, vendor_id, purchase_id, invoice_number, invoice_date, due_date, "+
			"amount_total, tax_amount, discount_amount, amount_paid, "+
			"currency, payment_terms, status, created_by) "+
			"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id",
		companyID,
		vendorID,
		purchaseID,
		form.InvoiceNumber,
		form.Date,
		form.dueOn,
		form.total,
		form.tax,
		0,
		0,
		"DOP",
		form.Terms,
		PayableStatuses.Pending,
		form.User().Id,
	).Scan(&apID)
	if err != nil {
		return err
	}

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

	variantIDs, err := resolveItemVariantIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}
	taxIDs, err := resolveItemTaxIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}

	vals := []any{}
	for _, l := range form.Lines {
		variantID, ok := variantIDs[l.ID]
		if !ok {
			return fmt.Errorf("missing item variant for item_id=%d", l.ID)
		}
		vals = append(vals,
			companyID,
			purchaseID,
			variantID,
			l.Qty,
			l.Price,
			l.total,
			l.Unit,
			l.discount,
			taxIDs[l.ID],
			l.tax,
		)
	}

	stmt := "INSERT INTO purchase_items (company_id, purchase_id, variant_id, qty, unit_price, line_total, unit_id, discount, tax_id, tax_amount) VALUES "
	stmt += database.PrepareBulkInsert(10, len(form.Lines))
	_, err = tx.Exec(stmt, vals...)
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
		_, err := tx.Exec(`
      UPDATE purchases
      SET vendor_id = $3,
          date = $4,
          due_date = $5,
          subtotal = $6,
          discount_amount = $7,
          tax_amount = $8,
          total = $9,
          notes = $10,
          payment_status = $11,
          transaction_kind = $12,
          invoice_number = $13,
          updated_at = NOW()
      WHERE company_id = $1 AND id = $2
    `,
			companyID,
			purchase.ID,
			form.VendorID,
			form.Date,
			form.dueOn,
			form.amount,
			discountAmount,
			form.tax,
			form.total,
			form.Notes,
			form.paymentStatus,
			form.Kind,
			form.InvoiceNumber,
		)
		if err != nil {
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

	variantIDs, err := resolveItemVariantIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}
	taxIDs, err := resolveItemTaxIDs(tx, companyID, itemIDs)
	if err != nil {
		return err
	}

	for _, l := range lines {
		variantID, ok := variantIDs[l.ID]
		if !ok {
			return fmt.Errorf("missing item variant for item_id=%d", l.ID)
		}

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
		_, err := tx.Exec("UPDATE purchases SET deleted_at = NOW(), updated_at = NOW() WHERE company_id = $1 AND id = $2", companyID, purchase.ID)
		if err != nil {
			return err
		}

		// Cancel the linked AP entry when a vendor bill is deleted.
		if purchase.Kind == PurchaseTransactionKinds.VendorBill {
			var apID int
			var alreadyPaid float64
			err := tx.QueryRow(
				"SELECT id, amount_paid FROM accounts_payable WHERE company_id = $1 AND purchase_id = $2",
				companyID, purchase.ID,
			).Scan(&apID, &alreadyPaid)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if apID > 0 {
				if _, err = tx.Exec(
					"UPDATE accounts_payable SET status = $3, updated_at = NOW() WHERE company_id = $1 AND id = $2",
					companyID, apID, PayableStatuses.Cancelled,
				); err != nil {
					return err
				}
				// Reverse only the unpaid portion from the vendor balance.
				remaining := purchase.Total - alreadyPaid
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
		poUUID, companyID, string(PayableStatuses.Cancelled),
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
func updatePOPaymentStatus(tx *sql.Tx, companyID int, poUUID string) error {
	newPaymentStatus, err := resolvePOPaymentStatus(tx, companyID, poUUID)
	if err != nil {
		return err
	}

	if newPaymentStatus == PaidStatuses.Paid {
		_, err = tx.Exec(
			`UPDATE purchases
			    SET payment_status  = $3,
			        purchase_status = $4,
			        updated_at      = NOW()
			  WHERE company_id = $1 AND uuid = $2`,
			companyID, poUUID, newPaymentStatus, string(PurchaseStatuses.Closed),
		)
	} else if newPaymentStatus == PaidStatuses.Partial {
		_, err = tx.Exec(
			`UPDATE purchases
			    SET payment_status  = $3,
			        purchase_status = $4,
			        updated_at      = NOW()
			  WHERE company_id = $1 AND uuid = $2
			    AND purchase_status NOT IN ('closed')`,
			companyID, poUUID, newPaymentStatus, string(PurchaseStatuses.PartiallyPaid),
		)
	} else {
		_, err = tx.Exec(
			`UPDATE purchases
			    SET payment_status = $3,
			        updated_at     = NOW()
			  WHERE company_id = $1 AND uuid = $2
			    AND payment_status != $4`,
			companyID, poUUID, newPaymentStatus, string(PaidStatuses.Paid),
		)
	}
	return err
}
