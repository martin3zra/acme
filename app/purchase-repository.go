package app

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/martin3zra/acme/pkg/cache"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type purchase struct {
	CompanyID     int                     `json:"company_id"`
	ID            int                     `json:"id"`
	UUID          string                  `json:"uuid"`   // computed from ID (table has no uuid)
	Number        string                  `json:"number"` // purchases.code
	Vendor        vendor                  `json:"vendor"`
	WarehouseID   int                     `json:"warehouse_id"`
	Date          time.Time               `json:"date"`
	DueOn         *time.Time              `json:"due_on"` // purchases.due_date
	Terms         string                  `json:"terms"`  // computed from date/due_on
	Amount        float64                 `json:"amount"` // purchases.subtotal
	Discount      Discount                `json:"discount"`
	Tax           float64                 `json:"tax"` // purchases.tax_amount
	Total         float64                 `json:"total"`
	AmountDue     float64                 `json:"amount_due"` // computed
	Status        string                  `json:"status"`     // purchases.purchase_status
	PaymentStatus PaidStatus              `json:"payment_status"`
	Notes         string                  `json:"notes"` // purchases.notes
	Kind          PurchaseTransactionKind `json:"transaction_kind"`
	Source        *PurchaseSource         `json:"source,omitempty"`

	EntityStatus foundation.Status `json:"-"` // purchases.status
}

func (s *Server) findPurchases(ctx context.Context, kind PurchaseTransactionKind) ([]*purchase, error) {
	rows, err := s.db.Query(
		"SELECT p.id, p.code, p.warehouse_id, p.date, p.due_date, p.subtotal, p.discount_amount, p.tax_amount, p.total, p.status, p.purchase_status, p.payment_status, COALESCE(p.notes, ''), p.transaction_kind, p.source, "+
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
			&p.Vendor.ID,
			&p.Vendor.UUID,
			&p.Vendor.Name,
			&p.Vendor.Email,
			&p.Vendor.Phone,
		); err != nil {
			return nil, err
		}

		p.UUID = strconv.Itoa(p.ID)
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

func (s *Server) findPurchaseByID(ctx context.Context, companyID int, purchaseID int) (*purchase, error) {
	p := new(purchase)
	var discountAmount float64
	err := s.db.QueryRow(
		"SELECT p.company_id, p.id, p.code, p.warehouse_id, p.date, p.due_date, p.subtotal, p.discount_amount, p.tax_amount, p.total, p.status, p.purchase_status, p.payment_status, COALESCE(p.notes, ''), p.transaction_kind, p.source, "+
			"v.id, v.uuid, v.name, v.email, v.phone "+
			"FROM purchases p "+
			"INNER JOIN companies ON (p.company_id = companies.id) "+
			"INNER JOIN vendors v ON (p.company_id = v.company_id AND p.vendor_id = v.id) "+
			"WHERE p.company_id = $1 AND p.id = $2 AND p.deleted_at IS NULL",
		companyID, purchaseID,
	).Scan(
		&p.CompanyID,
		&p.ID,
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
		&p.Vendor.ID,
		&p.Vendor.UUID,
		&p.Vendor.Name,
		&p.Vendor.Email,
		&p.Vendor.Phone,
	)
	if err != nil {
		return nil, err
	}

	p.UUID = strconv.Itoa(p.ID)
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

	var purchaseID int
	err = tx.QueryRow(
		"INSERT INTO purchases (company_id, vendor_id, warehouse_id, transaction_kind, notes, subtotal, discount_amount, tax_amount, total, payment_status, code, source, date, due_date) "+
			"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id",
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
	).Scan(&purchaseID)
	if err != nil {
		return "", err
	}

	if err := s.attachPurchaseLines(tx, companyID, purchaseID, form); err != nil {
		return "", err
	}

	return strconv.Itoa(purchaseID), nil
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

func (s *Server) updatePurchase(ctx context.Context, purchaseID int, form *UpdatePurchaseForm) error {
	companyID := CurrentCompany(ctx).ID
	purchase, err := s.findPurchaseByID(ctx, companyID, purchaseID)
	if err != nil {
		return err
	}

	discountAmount := float64(0)
	for _, l := range form.Lines {
		if l.Action == LineActions.Deleted {
			continue
		}
		discountAmount += l.discount
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
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
		)
		if err != nil {
			return err
		}

		return s.processPurchaseLines(tx, companyID, purchase.ID, form)
	})
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

func (s *Server) destroyPurchase(ctx context.Context, purchaseID int) error {
	companyID := CurrentCompany(ctx).ID
	purchase, err := s.findPurchaseByID(ctx, companyID, purchaseID)
	if err != nil {
		return err
	}

	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE purchases SET deleted_at = NOW(), updated_at = NOW() WHERE company_id = $1 AND id = $2", companyID, purchase.ID)
		return err
	})
	if err != nil {
		return err
	}

	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:purchase:%d", purchaseID)
	_ = c.Delete(ctx, key)
	return nil
}
