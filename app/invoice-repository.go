package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
)

type invoice struct {
	CompanyID    int                `json:"company_id"`
	ID           int                `json:"id"`
	UUID         string             `json:"uuid"`
	Number       string             `json:"number"`
	NCF          *string            `json:"ncf"`
	Customer     customer           `json:"customer"`
	Date         time.Time          `json:"date"`
	DueOn        *time.Time         `json:"due_on"`
	Terms        string             `json:"terms"`
	TaxReceiptID *int               `json:"tax_receipt_id"`
	Amount       float64            `json:"amount"`
	Discount     Discount           `json:"discount"`
	Tax          float64            `json:"tax"`
	Total        float64            `json:"total"`
	AmountDue    float64            `json:"amount_due"`
	Status       InvoiceStatus      `json:"status"`
	PaidStatus   PaidStatus         `json:"paid_status"`
	Payment      Payment            `json:"payment"`
	Notes        string             `json:"notes"`
	Kind         TransactionKind    `json:"transaction_kind"`
	Source       *TransactionSource `json:"source,omitempty"`
}

type line struct {
	ID           int64           `json:"id"`
	VariantID    int64           `json:"-"`
	WarehouseID  int64           `json:"warehouse_id"`
	Qty          int64           `json:"qty"`
	RemainingQty *int64          `json:"remaining_qty,omitempty"`
	Price        float64         `json:"price"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Identifier   ItemIdentifiers `json:"identifiers"`
	Unit         struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"unit"`
	Tax struct {
		tax
		Amount float64 `json:"amount"`
	} `json:"tax"`
	Amount float64 `json:"amount"`
	// TaxAmount float64 `json:"tax_amount"`
	Total float64 `json:"total"`
	// Add timestamps properties
	foundation.Timestamps
	Action LineAction `json:"action"`
}

func (s *Server) findInvoices(ctx context.Context, kind TransactionKind, invoiceType InvoiceType) ([]*invoice, error) {
	rows, err := s.db.Query("SELECT invoices.id, invoices.uuid, invoices.code, invoices.date, invoices.due_on, invoices.amount, invoices.discount, invoices.tax, "+
		"invoices.total, invoices.amount_due, invoices.status, invoices.paid_status, invoices.payment, invoices.note, invoices.tax_receipt_id, invoices.transaction_kind, "+
		"invoices.source, invoices.tax_number, customers.id, customers.uuid, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		"LEFT JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id) "+
		"WHERE invoices.company_id = $1 "+
		"AND invoices.transaction_kind = $2 "+
		"AND ($2 != 'invoice' OR $3 = 'all' OR invoices.type = $3::invoice_terms) ORDER BY invoices.id DESC", CurrentCompany(ctx).ID, kind, invoiceType)
	if err != nil {
		return nil, err
	}
	data := make([]*invoice, 0)
	for rows.Next() {
		i := new(invoice)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Number,
			&i.Date,
			&i.DueOn,
			&i.Amount,
			&i.Discount,
			&i.Tax,
			&i.Total,
			&i.AmountDue,
			&i.Status,
			&i.PaidStatus,
			&i.Payment,
			&i.Notes,
			&i.TaxReceiptID,
			&i.Kind,
			&i.Source,
			&i.NCF,
			&i.Customer.ID,
			&i.Customer.UUID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone,
		); err != nil {
			return nil, err
		}

		i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findInvoicesByUUID(ctx context.Context, kind TransactionKind, companyID int, uuid string) (*invoice, error) {
	i := new(invoice)
	err := s.db.QueryRow("SELECT invoices.company_id, invoices.id, invoices.uuid, invoices.code, invoices.date, invoices.due_on, invoices.amount, invoices.amount_due, invoices.discount, invoices.tax, "+
		"invoices.total, invoices.status, invoices.paid_status, invoices.payment, invoices.note, invoices.tax_receipt_id, invoices.transaction_kind, "+
		"invoices.source, invoices.tax_number, invoices.note, customers.id, customers.uuid, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		"LEFT JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id) "+
		"WHERE invoices.company_id = $1 "+
		"AND invoices.transaction_kind = $2 "+
		"AND invoices.uuid = $3", companyID, kind, uuid).
		Scan(
			&i.CompanyID,
			&i.ID,
			&i.UUID,
			&i.Number,
			&i.Date,
			&i.DueOn,
			&i.Amount,
			&i.AmountDue,
			&i.Discount,
			&i.Tax,
			&i.Total,
			&i.Status,
			&i.PaidStatus,
			&i.Payment,
			&i.Notes,
			&i.TaxReceiptID,
			&i.Kind,
			&i.Source,
			&i.NCF,
			&i.Notes,
			&i.Customer.ID,
			&i.Customer.UUID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone)
	if err != nil {
		return nil, err
	}
	i.Terms = "pia"
	if i.DueOn != nil {
		difference := i.DueOn.Sub(i.Date)
		// Difference in days
		i.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
	}

	i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

	return i, nil
}

func (s *Server) findInvoicesByID(companyID, invoiceId int) (*invoice, error) {
	i := new(invoice)
	err := s.db.QueryRow("SELECT invoices.company_id,invoices.id, invoices.uuid, invoices.code, invoices.date, invoices.due_on, invoices.amount, invoices.amount_due, invoices.discount, invoices.tax, "+
		"invoices.total, invoices.status, invoices.paid_status, invoices.payment, invoices.note, invoices.tax_receipt_id, invoices.transaction_kind, "+
		"invoices.source, invoices.tax_number, invoices.note, customers.id, customers.uuid, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		// LEFT JOIN: templates/estimates/orders have no tax receipt, but must
		// still be findable (this is hit by the recurrence generator).
		"LEFT JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id) "+
		"WHERE invoices.company_id = $1 AND invoices.id = $2", companyID, invoiceId).
		Scan(
			&i.CompanyID,
			&i.ID,
			&i.UUID,
			&i.Number,
			&i.Date,
			&i.DueOn,
			&i.Amount,
			&i.AmountDue,
			&i.Discount,
			&i.Tax,
			&i.Total,
			&i.Status,
			&i.PaidStatus,
			&i.Payment,
			&i.Notes,
			&i.TaxReceiptID,
			&i.Kind,
			&i.Source,
			&i.NCF,
			&i.Notes,
			&i.Customer.ID,
			&i.Customer.UUID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone)
	if err != nil {
		return nil, err
	}
	i.Terms = "pia"
	if i.DueOn != nil {
		difference := i.DueOn.Sub(i.Date)
		// Difference in days
		i.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
	}

	i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

	return i, nil
}

func (s *Server) storeInvoice(ctx context.Context, form *StoreInvoiceForm) (string, error) {
	var invoiceUUID string
	companyID := CurrentCompany(ctx).ID

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var err error
		invoiceUUID, err = s.storeInvoiceInternal(tx, companyID, form)
		return err
	})

	return invoiceUUID, err
}

func (s *Server) updateInvoice(ctx context.Context, uuid string, form *UpdateInvoiceForm) error {
	companyID := CurrentCompany(ctx).ID
	invoice, err := s.findInvoicesByUUID(ctx, form.Kind, companyID, uuid)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var termType *string
		if form.Kind == TransactionKinds.Invoice || form.Kind == TransactionKinds.Order {
			termType = (*string)(&form.termType)
		}
		_, err = tx.Exec(`
    UPDATE invoices
    SET customer_id = $3, date = $4, due_on = $5, amount = $6, discount = $7, tax = $8, total = $9,
    amount_due = $10, note = $11, payment = $12, type = $13, paid_status = $14
    WHERE company_id = $1 AND id = $2
    `,
			companyID, invoice.ID, form.CustomerID, form.Date, form.dueOn, form.amount,
			foundation.ToJSON(form.Discount), form.tax, form.total, form.amountDue,
			form.Notes, foundation.ToJSON(form.Payment), termType, form.paidStatus,
		)
		if err != nil {
			return err
		}

		if err = s.processInvoiceLines(tx, companyID, invoice.ID, form); err != nil {
			return err
		}

		if form.Kind != TransactionKinds.Invoice {
			return nil
		}

		// When the invoice terms is been updated from CASH to CREDIT
		if invoice.DueOn == nil && form.termType == InvoiceTermType.Credit {
			// Ensure to associated the new customer or current one the receivable
			customerID := invoice.Customer.ID
			if invoice.Customer.ID != form.CustomerID {
				customerID = form.CustomerID
			}

			if err = s.registerReceivable(tx, companyID, invoice.ID, customerID); err != nil {
				return err
			}
		}

		if invoice.DueOn != nil && form.termType == InvoiceTermType.Cash {
			if err = s.deleteInvoiceFromReceivables(tx, companyID, invoice.ID, invoice.Customer.ID); err != nil {
				return err
			}

			if err = s.updateCustomerAmountDue(tx, companyID, invoice.Customer.ID, -invoice.Total); err != nil {
				return err
			}
		}

		if invoice.Customer.ID != form.CustomerID {
			// Update customer balance. Logs this operations to keep track of it.
			if err = s.updateCustomerAmountDue(tx, companyID, invoice.Customer.ID, -invoice.Amount); err != nil {
				return err
			}

			if err = s.updateCustomerAmountDue(tx, companyID, form.CustomerID, form.total); err != nil {
				return err
			}

			if invoice.DueOn != nil {
				if err = s.changeCustomerFromReceivables(tx, companyID, invoice.ID, invoice.Customer.ID, form.CustomerID); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *Server) voidInvoice(ctx context.Context, kind TransactionKind, uuid string) error {
	companyID := CurrentCompany(ctx).ID
	invoice, err := s.findInvoicesByUUID(ctx, kind, companyID, uuid)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		_, err = tx.Exec(`
    UPDATE invoices
    SET amount = 0, discount = NULL, tax = 0, total = 0,
    amount_due = 0, payment = NULL, status = $4, paid_status = $5
    WHERE company_id = $1 AND id = $2 AND transaction_kind = $3
  `,
			companyID, invoice.ID, kind, InvoiceStatuses.Void, PaidStatuses.Refunded,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
    UPDATE invoices_items
    SET amount = 0, qty = 0, price = 0, tax = 0, total = 0
    WHERE company_id = $1 AND invoice_id = $2
  `,
			companyID, invoice.ID,
		)
		if err != nil {
			return err
		}

		if err = s.deleteInvoiceFromReceivables(tx, companyID, invoice.ID, invoice.Customer.ID); err != nil {
			return err
		}

		// Reverse any inventory movements that were recorded for this invoice.
		var movementRecorded bool
		_ = s.db.QueryRowContext(ctx,
			"SELECT movement_recorded FROM invoices WHERE company_id = $1 AND id = $2",
			companyID, invoice.ID,
		).Scan(&movementRecorded)
		if movementRecorded {
			if err = s.reverseMovements(tx, companyID, "invoice", invoice.ID, InventoryMovementKinds.SaleReturn); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Server) attachInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form *StoreInvoiceForm) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	rows := make([]map[string]any, 0, len(form.Lines))
	for _, line := range form.Lines {
		rows = append(rows, map[string]any{
			"company_id":   companyId,
			"invoice_id":   invoiceId,
			"item_id":      line.ID,
			"unit_id":      line.Unit,
			"qty":          line.Qty,
			"price":        line.Price,
			"rate":         line.Rate,
			"amount":       line.amount,
			"tax":          line.tax,
			"total":        line.total,
			"warehouse_id": line.WarehouseID,
		})
	}
	// InsertMany compiles a single multi-row INSERT, preserving the original
	// bulk-insert behaviour.
	_, err = ptx.Model(&InvoiceItem{}).InsertMany(context.Background(), rows)
	return err
}

func (s *Server) processInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form *UpdateInvoiceForm) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	lines := s.filterInvoiceLines(form.Lines, ADDED, UPDATED, DELETED)
	for _, line := range lines {
		switch line.Action {
		case ADDED:
			// Preserves the original column set exactly: only these eight columns,
			// with the tax column carrying the line rate (rate/amount/total are left
			// to their defaults on this path, unlike the create-time bulk insert).
			if _, err := ptx.Model(&InvoiceItem{}).InsertMany(context.Background(), []map[string]any{{
				"company_id":   companyId,
				"invoice_id":   invoiceId,
				"item_id":      line.ID,
				"unit_id":      line.Unit,
				"qty":          line.Qty,
				"price":        line.Price,
				"tax":          line.Rate,
				"warehouse_id": line.WarehouseID,
			}}); err != nil {
				return err
			}
		case UPDATED:
			if _, err := ptx.Model(&InvoiceItem{}).
				WhereEq("company_id", companyId).
				WhereEq("invoice_id", invoiceId).
				WhereEq("item_id", line.ID).
				Update(context.Background(), map[string]any{
					"qty":          line.Qty,
					"unit_id":      line.Unit,
					"warehouse_id": line.WarehouseID,
				}); err != nil {
				return err
			}
		case DELETED:
			if _, err := ptx.Model(&InvoiceItem{}).
				WhereEq("company_id", companyId).
				WhereEq("invoice_id", invoiceId).
				WhereEq("item_id", line.ID).
				Delete(context.Background()); err != nil {
				return err
			}
		default:
			fmt.Println("Nothing to happen here.")
		}
	}
	return nil
}

func (s *Server) findInvoiceLines(ctx context.Context, companyID, invoiceID int) ([]*line, error) {
	rows, err := s.db.Query(`
    SELECT ii.item_id, ii.qty, ii.price, items_units.unit_id, it.name, it.description, items_units.name,
    ii.created_at, ii.updated_at, ii.deleted_at, 'unchanged' as action, ii.amount, ii.total,
    taxes.id as tax_id, taxes.name as tax_name, ii.rate, ii.tax, it.identifiers, ii.warehouse_id
    FROM invoices_items AS ii
    INNER JOIN companies AS com ON (ii.company_id = com.id)
    INNER JOIN invoices AS i ON (ii.invoice_id = i.id AND ii.company_id = i.company_id)
    INNER JOIN items AS it ON(ii.item_id = it.id AND ii.company_id = it.company_id)
    LEFT JOIN LATERAL (
      SELECT items_units.unit_id, units.name
      FROM items_units
      INNER JOIN units ON (items_units.unit_id = units.id)
      WHERE items_units.item_id = it.id limit 1
    ) items_units ON true
    INNER JOIN taxes ON (it.company_id = taxes.company_id AND it.tax_id = taxes.id)
    WHERE ii.company_id = $1
    AND ii.invoice_id = $2`, companyID, invoiceID)
	if err != nil {
		return nil, err
	}
	data := make([]*line, 0)
	for rows.Next() {
		i := new(line)

		if err = rows.Scan(
			&i.ID,
			&i.Qty,
			&i.Price,
			&i.Unit.ID,
			&i.Name,
			&i.Description,
			&i.Unit.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.Action,
			&i.Amount,
			&i.Total,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Tax.Amount,
			&i.Identifier,
			&i.WarehouseID,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) registerReceivable(tx *sql.Tx, companyId, invoiceId, customerId int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	return ptx.Insert(context.Background(), &Receivable{
		CompanyID:  companyId,
		InvoiceID:  invoiceId,
		CustomerID: customerId,
	})
}

func (s *Server) deleteInvoiceFromReceivables(tx *sql.Tx, companyId, invoiceId, customerId int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	_, err = ptx.Model(&Receivable{}).
		WhereEq("company_id", companyId).
		WhereEq("invoice_id", invoiceId).
		WhereEq("customer_id", customerId).
		Delete(context.Background())
	return err
}

func (s *Server) changeCustomerFromReceivables(tx *sql.Tx, companyId, invoiceId, customerId, newCustomerId int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	_, err = ptx.Model(&Receivable{}).
		WhereEq("company_id", companyId).
		WhereEq("invoice_id", invoiceId).
		WhereEq("customer_id", customerId).
		Update(context.Background(), map[string]any{"customer_id": newCustomerId})
	return err
}

func (s *Server) filterInvoiceLines(lines []*Line, actions ...LineAction) []*Line {
	if len(actions) == 0 {
		return nil
	}

	// Convert actions to a lookup map for O(1) checks
	actionSet := make(map[string]struct{}, len(actions))
	for _, a := range actions {
		actionSet[string(a)] = struct{}{}
	}

	// Filter lines
	filtered := make([]*Line, 0, len(lines))
	for _, line := range lines {
		if _, ok := actionSet[string(line.Action)]; ok {
			filtered = append(filtered, line)
		}
	}

	return filtered
}

func (s *Server) updateInvoiceBalance(tx *sql.Tx, companyID, invoiceID int, balance float64) error {
	stmt := `
    UPDATE invoices
    SET amount_due = amount_due + $3, paid_status = CASE
      WHEN  amount_due + $3 = 0 THEN 'paid'::paid_status
      WHEN  amount_due + $3 = total THEN 'unpaid'::paid_status
      ELSE 'partial'::paid_status
    END,
    status = CASE
      WHEN amount_due + $3 = 0 THEN 'closed'::invoice_status
      ELSE 'sent'::invoice_status
    END
    WHERE company_id = $1 AND id = $2
  `
	_, err := tx.Exec(stmt, companyID, invoiceID, balance)

	return err
}

func (s *Server) storeInvoiceBackground(tx *sql.Tx, companyID int, form *StoreInvoiceForm) (string, error) {
	return s.storeInvoiceInternal(tx, companyID, form)
}

func (s *Server) storeInvoiceInternal(tx *sql.Tx, companyID int, form *StoreInvoiceForm) (string, error) {
	var invoiceUUID string
	var termType *string
	var taxReceiptSequence *taxReceiptSeq
	var err error

	if form.Kind == TransactionKinds.Invoice || form.Kind == TransactionKinds.Order {
		termType = (*string)(&form.termType)
	}

	if form.Kind == TransactionKinds.Invoice {
		taxReceiptSequence, err = s.grabTaxReceiptSequence(tx, companyID, form.TaxReceipt)
		if err != nil {
			return invoiceUUID, err
		}
	}

	stmt, err := tx.Prepare("INSERT INTO invoices (company_id, tax_receipt_id, tax_receipt_sequence, tax_number, date, type, due_on, customer_id, amount, discount, tax, amount_due, total, note, status, paid_status, payment, code, transaction_kind, source, recurrence) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) RETURNING id, uuid")
	if err != nil {
		return invoiceUUID, err
	}

	seqSource := string(form.Kind)
	if form.Kind == TransactionKinds.Invoice {
		seqSource = fmt.Sprintf("invoice.%v", form.termType)
	}
	seqInfo, err := GetNextSequence(tx, companyID, seqSource)
	if err != nil {
		return invoiceUUID, err
	}

	var taxID *int = nil
	var taxSeq *int64 = nil
	var taxNumber *string = nil
	if taxReceiptSequence != nil {
		taxID = &form.TaxReceipt
		taxSeq = &taxReceiptSequence.Seq
		taxNumber = &taxReceiptSequence.Number
	}

	// A recurring template stores which tax receipt to stamp on each generated
	// invoice. No sequence is consumed here — that happens when an invoice is
	// actually generated from the template (kind invoice grabs the sequence).
	if form.Kind == TransactionKinds.Template && form.TaxReceipt > 0 {
		taxID = &form.TaxReceipt
	}

	var source *[]byte
	if form.Source != nil && form.Source.ID != "" {
		j := foundation.AsJSON(form.Source)
		source = &j
	}

	var recurrence *[]byte
	if form.Recurrence != nil {
		// set next run until scheduler pick the template up
		form.Recurrence.NextRunAt = form.Recurrence.StartDate
		_recurrence := foundation.AsJSON(form.Recurrence)
		recurrence = &_recurrence
	}

	var invoiceID int
	err = stmt.QueryRow(
		companyID,
		taxID,
		taxSeq,
		taxNumber,
		form.Date,
		termType,
		&form.dueOn,
		form.CustomerID,
		form.amount,
		foundation.ToJSON(form.Discount),
		form.tax,
		form.amountDue,
		form.total,
		form.Notes,
		form.status,
		form.paidStatus,
		foundation.ToJSON(form.Payment),
		seqInfo.Code,
		form.Kind,
		source,
		recurrence,
	).Scan(&invoiceID, &invoiceUUID)

	if err != nil {
		return invoiceUUID, err
	}

	// Set default warehouse to 1 for all lines if not specified
	for _, line := range form.Lines {
		if line.WarehouseID == 0 {
			line.WarehouseID = 1
		}
	}
	if err = s.attachInvoiceLines(tx, companyID, invoiceID, form); err != nil {
		return invoiceUUID, err
	}

	// Record inventory OUT movements for invoices and orders that are live
	// (sent or closed). Estimates and templates never affect stock.
	if form.Kind == TransactionKinds.Invoice || form.Kind == TransactionKinds.Order {
		if form.status == InvoiceStatuses.Sent || form.status == InvoiceStatuses.Closed {
			if err = s.recordInvoiceMovements(tx, companyID, invoiceID, form.Lines); err != nil {
				return invoiceUUID, err
			}
			if _, err = tx.Exec(
				"UPDATE invoices SET movement_recorded = TRUE WHERE company_id = $1 AND id = $2",
				companyID, invoiceID,
			); err != nil {
				return invoiceUUID, err
			}
		}
	}

	if form.Source != nil && form.Source.ID != "" {
		// When we are duplicating an existing invoice, we set
		// a relationshipt bewteen both invoice using the
		// source column to keep track of them.
		if form.Source.Type == TransactionKinds.Invoice {
			_, err := tx.Exec(
				"UPDATE invoices SET source = $4 "+
					"WHERE company_id = $1 "+
					"AND uuid = $2 AND transaction_kind = $3",
				companyID, form.Source.ID, form.Source.Type, foundation.AsJSON(map[string]any{
					"type": form.Kind,
					"id":   invoiceUUID,
					"code": seqInfo.Code,
				}))
			if err != nil {
				return invoiceUUID, err
			}
		}

		if form.Source.Type == TransactionKinds.Estimate || form.Source.Type == TransactionKinds.Order {
			_, err := tx.Exec(
				"UPDATE invoices SET status = 'closed', source = $4 "+
					"WHERE company_id = $1 "+
					"AND uuid = $2 AND transaction_kind = $3",
				companyID, form.Source.ID, form.Source.Type, foundation.ToJSON(map[string]any{
					"type": form.Kind,
					"id":   invoiceUUID,
					"code": seqInfo.Code,
				}))
			if err != nil {
				return invoiceUUID, err
			}
		}

		// Delete the cache for the source of the transaction, so we
		// can display updated values such as status ...
		if err = s.purgeCacheByID(tx, form.Source.Type, form.Source.ID); err != nil {
			return invoiceUUID, err
		}
	}

	if form.Kind == TransactionKinds.Invoice {
		// trigger an event for this? Use pipe!!!
		if form.Terms != "pia" {
			if err = s.registerReceivable(tx, companyID, invoiceID, form.CustomerID); err != nil {
				return invoiceUUID, err
			}

			if err = s.updateCustomerAmountDue(tx, companyID, form.CustomerID, form.amountDue); err != nil {
				return invoiceUUID, err
			}
		}
	}

	return invoiceUUID, err
}

func (s *Server) purgeCacheByID(tx *sql.Tx, kind TransactionKind, uuid string) error {
	c := cache.NewPgCache(tx)
	key := fmt.Sprintf("preview:%s:%s", kind, uuid)
	if err := c.Delete(context.Background(), key); err != nil {
		log.Printf("Error deleting cache: %v", err)
		return err
	}
	return nil
}

// recordInvoiceMovements resolves the variant_id for each invoice line and
// records an OUT movement (negative qty) per line.
func (s *Server) recordInvoiceMovements(tx *sql.Tx, companyID, invoiceID int, lines []*Line) error {
	itemIDs := make([]int, 0, len(lines))
	for _, l := range lines {
		itemIDs = append(itemIDs, l.ID)
	}

	variantIDs, err := resolveItemVariantIDs(tx, companyID, itemIDs)
	if err != nil {
		return fmt.Errorf("recordInvoiceMovements: resolve variants: %w", err)
	}

	for _, l := range lines {
		if l.Action == LineActions.Deleted {
			continue
		}
		variantID, ok := variantIDs[l.ID]
		if !ok {
			// Item has no default variant; skip silently.
			continue
		}
		if l.WarehouseID == 0 {
			continue
		}
		// OUT movement: qty is negative.
		if err := s.recordMovement(
			tx, companyID,
			variantID, l.WarehouseID, l.Unit,
			-float64(l.Qty), l.Price,
			InventoryMovementKinds.Sale,
			"invoice", invoiceID,
		); err != nil {
			return err
		}
	}
	return nil
}
