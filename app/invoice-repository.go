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
	"github.com/martin3zra/playsql"
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
	VariantID    int64           `json:"variant_id"`
	VariantName  string          `json:"variant_name"`
	VariantSKU   string          `json:"variant_sku,omitempty"`
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

// deriveTerms labels the invoice from the gap between its date and due date. Only
// the two detail reads do this; the list read leaves Terms empty, as it always has.
func (i *invoice) deriveTerms() {
	i.Terms = "pia"
	if i.DueOn != nil {
		difference := i.DueOn.Sub(i.Date)
		// Difference in days
		i.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
	}
}

// The three header reads below dropped two joins each.
//
// INNER JOIN companies asserted existence only: company_id is a NOT NULL FK and the
// company_id predicate already scopes the query.
//
// LEFT JOIN tax_receipts contributed no columns and, being a LEFT join, could not
// filter — the NCF comes from invoices.tax_number, not from the receipt. tax_receipts.id
// is its primary key, so it could not duplicate rows either. It was dead weight.
//
// INNER JOIN customers becomes a belongsTo eager load. It needs WithTrashed: customerRead
// is softdelete-tagged, but the join never filtered customers.deleted_at, and an invoice
// to a since-deleted customer must still render.

func (s *Server) findInvoices(ctx context.Context, kind TransactionKind, invoiceType InvoiceType) ([]*invoice, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	// The old `($2 != 'invoice' OR $3 = 'all' OR invoices.type = $3::invoice_terms)`
	// predicate: the type filter only applies to invoices, and only when narrowed.
	narrowed := kind == TransactionKinds.Invoice && invoiceType != InvoiceTypeAll

	var rows []invoiceRead
	if err := pdb.Model(&invoiceRead{}).
		Select(invoiceListColumns...).
		WithConstraint("Customer", withTrashedRelation).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("transaction_kind", string(kind)).
		When(narrowed, func(q *playsql.Builder) {
			q.WhereEq("type", string(invoiceType))
		}).
		OrderBy("id", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*invoice, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toInvoice())
	}
	return data, nil
}

func (s *Server) findInvoicesByUUID(ctx context.Context, kind TransactionKind, companyID int, uuid string) (*invoice, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row invoiceRead
	if err := pdb.Model(&invoiceRead{}).
		WithConstraint("Customer", withTrashedRelation).
		WhereEq("company_id", companyID).
		WhereEq("transaction_kind", string(kind)).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}

	i := row.toInvoice()
	i.deriveTerms()
	return i, nil
}

// findInvoicesByID matches on the primary key alone — no transaction_kind filter, so
// templates/estimates/orders stay findable (the recurrence generator relies on this).
func (s *Server) findInvoicesByID(companyID, invoiceId int) (*invoice, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row invoiceRead
	if err := pdb.Model(&invoiceRead{}).
		WithConstraint("Customer", withTrashedRelation).
		WhereEq("company_id", companyID).
		WhereEq("id", invoiceId).
		First(context.Background(), &row); err != nil {
		return nil, err
	}

	i := row.toInvoice()
	i.deriveTerms()
	return i, nil
}

// playSession is satisfied by both *playsql.DB and *playsql.Tx.
type playSession interface {
	Model(model any) *playsql.Builder
}

// invoiceMovementRecorded reports whether the invoice already moved stock. The two
// call sites read it from different sessions on purpose: updateInvoice reads inside
// its transaction, voidInvoice reads outside it.
func invoiceMovementRecorded(ctx context.Context, sess playSession, companyID, invoiceID int) (bool, error) {
	var row invoiceRead
	if err := sess.Model(&invoiceRead{}).
		Select("movement_recorded").
		WhereEq("company_id", companyID).
		WhereEq("id", invoiceID).
		First(ctx, &row); err != nil {
		return false, err
	}
	return row.MovementRecorded, nil
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
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}
		if _, err = ptx.Model(&invoiceInsert{}).
			WhereEq("company_id", companyID).
			WhereEq("id", invoice.ID).
			Update(context.Background(), map[string]any{
				"customer_id": form.CustomerID,
				"date":        form.Date,
				"due_on":      form.dueOn,
				"amount":      form.amount,
				"discount":    foundation.ToJSON(form.Discount),
				"tax":         form.tax,
				"total":       form.total,
				"amount_due":  form.amountDue,
				"note":        form.Notes,
				"payment":     foundation.ToJSON(form.Payment),
				"type":        termType,
				"paid_status": form.paidStatus,
			}); err != nil {
			return err
		}

		if err = s.processInvoiceLines(tx, companyID, invoice.ID, form); err != nil {
			return err
		}

		// Keep inventory in step with the edited lines. Only invoices/orders that
		// already moved stock get reconciled — drafts and estimates never did.
		if form.Kind == TransactionKinds.Invoice || form.Kind == TransactionKinds.Order {
			ptx, err := playTx(tx)
			if err != nil {
				return err
			}
			movementRecorded, err := invoiceMovementRecorded(context.Background(), ptx, companyID, invoice.ID)
			if err != nil {
				return err
			}
			if movementRecorded {
				if err = s.reconcileInvoiceStock(tx, companyID, invoice.ID); err != nil {
					return err
				}
			}
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
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}
		if _, err = ptx.Model(&invoiceInsert{}).
			WhereEq("company_id", companyID).
			WhereEq("id", invoice.ID).
			WhereEq("transaction_kind", kind).
			Update(context.Background(), map[string]any{
				"amount":      0,
				"discount":    nil,
				"tax":         0,
				"total":       0,
				"amount_due":  0,
				"payment":     nil,
				"status":      InvoiceStatuses.Void,
				"paid_status": PaidStatuses.Refunded,
			}); err != nil {
			return err
		}

		if _, err = ptx.Model(&InvoiceItem{}).
			WhereEq("company_id", companyID).
			WhereEq("invoice_id", invoice.ID).
			Update(context.Background(), map[string]any{
				"amount": 0,
				"qty":    0,
				"price":  0,
				"tax":    0,
				"total":  0,
			}); err != nil {
			return err
		}

		// The reverse side effects (receivable removal, and stock reversal when the
		// invoice moved stock) react to InvoiceVoided within this same transaction.
		// Read outside the transaction, and errors ignored, exactly as before: a
		// failure here just leaves movementRecorded false.
		var movementRecorded bool
		if pdb, perr := s.play(); perr == nil {
			movementRecorded, _ = invoiceMovementRecorded(ctx, pdb, companyID, invoice.ID)
		}

		return s.dispatcher().Dispatch(context.Background(), tx, InvoiceVoided{
			CompanyID:        companyID,
			InvoiceID:        invoice.ID,
			CustomerID:       invoice.Customer.ID,
			MovementRecorded: movementRecorded,
		})
	})
}

// lineWarehouseIDs collects the warehouse ids named by a set of lines.
func lineWarehouseIDs(lines []*Line) []int {
	ids := make([]int, 0, len(lines))
	for _, line := range lines {
		ids = append(ids, int(line.WarehouseID))
	}
	return ids
}

func (s *Server) attachInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form *StoreInvoiceForm) error {
	// Line warehouses come straight from the request. Checked here as well as in
	// recordMovement, because a draft writes the lines without moving stock.
	if err := assertWarehousesInCompany(tx, companyId, lineWarehouseIDs(form.Lines)...); err != nil {
		return err
	}

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
			"variant_id":   line.VariantID,
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
	if err := assertWarehousesInCompany(tx, companyId, lineWarehouseIDs(lines)...); err != nil {
		return err
	}
	if err := resolveVariantsForLines(tx, companyId, lines); err != nil {
		return err
	}
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
				"variant_id":   line.VariantID,
				"unit_id":      line.Unit,
				"qty":          line.Qty,
				"price":        line.Price,
				"tax":          line.Rate,
				"warehouse_id": line.WarehouseID,
			}}); err != nil {
				return err
			}
		case UPDATED:
			// Match on variant_id too: one item may appear on several lines as
			// different variants, so item_id alone no longer identifies a row.
			if _, err := ptx.Model(&InvoiceItem{}).
				WhereEq("company_id", companyId).
				WhereEq("invoice_id", invoiceId).
				WhereEq("item_id", line.ID).
				WhereEq("variant_id", line.VariantID).
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
				WhereEq("variant_id", line.VariantID).
				Delete(context.Background()); err != nil {
				return err
			}
		default:
			fmt.Println("Nothing to happen here.")
		}
	}
	return nil
}

// findInvoiceLines reads an invoice's lines with the item, variant, tax and unit each
// line displays.
//
// Five joins collapse into relations:
//
//   - INNER JOIN companies and INNER JOIN invoices asserted existence of NOT NULL FKs;
//     company_id and invoice_id already scope the read.
//   - INNER JOIN items and INNER JOIN items_variants become belongsTo. Neither filtered
//     deleted_at, which is why the item side uses invoiceLineItemRead rather than
//     itemRead: a line on a since-deleted item must still render its name.
//   - LEFT JOIN LATERAL (... LIMIT 1) resolved the item's unit. Since migration
//     20260709120000 added items_units_company_item_unique (company_id, item_id) there
//     is at most one link row per item, so it is a plain hasOne.
//
// The tax eager load is narrowed to (id, name): a line's rate and tax amount are
// frozen on invoices_items, not read off the tax row.
//
// The item's own two relations are nested inside one constraint rather than written as
// two `With("Item.…")` paths. Each with-clause reloads every segment it names and
// reassigns the field, so a second path through Item would overwrite the structs the
// first had populated and silently drop their ItemUnit.
func (s *Server) findInvoiceLines(ctx context.Context, companyID, invoiceID int) ([]*line, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []invoiceLineRead
	if err := pdb.Model(&invoiceLineRead{}).
		With("Variant").
		WithConstraint("Item", withInvoiceLineItem).
		WhereEq("company_id", companyID).
		WhereEq("invoice_id", invoiceID).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*line, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toLine())
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

// updateInvoiceBalance stays raw: `amount_due = amount_due + $3` is a self-referencing
// increment, and both paid_status and status are CASE expressions over the new value.
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

	ptx, err := playTx(tx)
	if err != nil {
		return invoiceUUID, err
	}
	row := &invoiceInsert{
		CompanyID:          companyID,
		TaxReceiptID:       taxID,
		TaxReceiptSequence: taxSeq,
		TaxNumber:          taxNumber,
		Date:               form.Date,
		Type:               termType,
		DueOn:              form.dueOn,
		CustomerID:         form.CustomerID,
		Amount:             form.amount,
		Discount:           foundation.ToJSON(form.Discount),
		Tax:                form.tax,
		AmountDue:          form.amountDue,
		Total:              form.total,
		Note:               form.Notes,
		Status:             form.status,
		PaidStatus:         form.paidStatus,
		Payment:            foundation.ToJSON(form.Payment),
		Code:               seqInfo.Code,
		TransactionKind:    form.Kind,
		Source:             source,
		Recurrence:         recurrence,
	}
	if err = ptx.Insert(context.Background(), row); err != nil {
		return invoiceUUID, err
	}
	invoiceID := int(row.ID)

	// uuid is generated by the database default; read it back by the new id.
	invoiceUUID, err = playsql.RawScalar[string](ptx, context.Background(),
		"SELECT uuid FROM invoices WHERE company_id = $1 AND id = $2", companyID, row.ID)
	if err != nil {
		return invoiceUUID, err
	}

	// Set default warehouse to 1 for all lines if not specified
	for _, line := range form.Lines {
		if line.WarehouseID == 0 {
			line.WarehouseID = 1
		}
	}
	// Resolve each line's variant (explicit, or the item's default) before the
	// line is persisted, so invoices_items.variant_id is populated correctly.
	if err := resolveVariantsForLines(tx, companyID, form.Lines); err != nil {
		return invoiceUUID, err
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
			if _, err = ptx.Model(&invoiceInsert{}).
				WhereEq("company_id", companyID).
				WhereEq("id", invoiceID).
				Update(context.Background(), map[string]any{"movement_recorded": true}); err != nil {
				return invoiceUUID, err
			}
		}
	}

	if form.Source != nil && form.Source.ID != "" {
		// When we are duplicating an existing invoice, we set
		// a relationshipt bewteen both invoice using the
		// source column to keep track of them.
		if form.Source.Type == TransactionKinds.Invoice {
			if _, err := ptx.Model(&invoiceInsert{}).
				WhereEq("company_id", companyID).
				WhereEq("uuid", form.Source.ID).
				WhereEq("transaction_kind", form.Source.Type).
				Update(context.Background(), map[string]any{
					"source": foundation.AsJSON(map[string]any{
						"type": form.Kind,
						"id":   invoiceUUID,
						"code": seqInfo.Code,
					}),
				}); err != nil {
				return invoiceUUID, err
			}
		}

		if form.Source.Type == TransactionKinds.Estimate || form.Source.Type == TransactionKinds.Order {
			if _, err := ptx.Model(&invoiceInsert{}).
				WhereEq("company_id", companyID).
				WhereEq("uuid", form.Source.ID).
				WhereEq("transaction_kind", form.Source.Type).
				Update(context.Background(), map[string]any{
					"status": "closed",
					"source": foundation.ToJSON(map[string]any{
						"type": form.Kind,
						"id":   invoiceUUID,
						"code": seqInfo.Code,
					}),
				}); err != nil {
				return invoiceUUID, err
			}
		}

		// Delete the cache for the source of the transaction, so we
		// can display updated values such as status ...
		if err = s.purgeCacheByID(tx, form.Source.Type, form.Source.ID); err != nil {
			return invoiceUUID, err
		}
	}

	// A credit invoice raises InvoiceCreated; the receivable + customer-balance
	// side effects react in receivableListener, within this same transaction.
	if form.Kind == TransactionKinds.Invoice && form.Terms != "pia" {
		if err = s.dispatcher().Dispatch(context.Background(), tx, InvoiceCreated{
			CompanyID:  companyID,
			InvoiceID:  invoiceID,
			CustomerID: form.CustomerID,
			AmountDue:  form.amountDue,
		}); err != nil {
			return invoiceUUID, err
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

// recordInvoiceMovements records an OUT movement (negative qty) per line against
// the variant the line already resolved to (see resolveVariantForLine, run before
// the lines were persisted). Lines carry a concrete variant_id by this point; a
// zero here is a programming error, not something to paper over silently.
func (s *Server) recordInvoiceMovements(tx *sql.Tx, companyID, invoiceID int, lines []*Line) error {
	for _, l := range lines {
		if l.Action == LineActions.Deleted {
			continue
		}
		if l.WarehouseID == 0 {
			continue
		}
		if l.VariantID == 0 {
			return fmt.Errorf("recordInvoiceMovements: line for item_id=%d has no resolved variant", l.ID)
		}
		// OUT movement: qty is negative.
		if err := s.recordMovement(
			tx, companyID,
			l.VariantID, l.WarehouseID, l.Unit,
			-float64(l.Qty), l.Price,
			InventoryMovementKinds.Sale,
			"invoice", invoiceID,
		); err != nil {
			return err
		}
	}
	return nil
}

// reconcileInvoiceStock re-aligns an invoice's inventory footprint with its
// current lines after an edit. It first neutralises whatever the invoice has
// moved so far — one adjustment per (variant, warehouse) that nets the running
// total back to zero — then re-records fresh OUT movements from the surviving
// invoices_items rows through recordMovement, reusing its unit conversion and
// track_inventory rules. Reconciling an unchanged invoice nets to the same
// balance, so it is safe to run on every edit; callers gate it on
// movement_recorded so drafts and estimates (which never moved stock) are left
// alone. Handles qty changes, added/removed lines, and variant swaps (which
// arrive as delete+add) uniformly, since it works from the persisted rows.
func (s *Server) reconcileInvoiceStock(tx *sql.Tx, companyID, invoiceID int) error {
	// 1. Neutralise the invoice's current net contribution per variant+warehouse.
	// Stays raw: a GROUP BY with a HAVING over an aggregate is not a model read.
	rows, err := tx.Query(
		`SELECT variant_id, warehouse_id, SUM(qty)::float8
		   FROM inventory_movements
		  WHERE company_id = $1 AND reference_type = 'invoice' AND reference_id = $2
		  GROUP BY variant_id, warehouse_id
		 HAVING SUM(qty) <> 0`,
		companyID, invoiceID,
	)
	if err != nil {
		return fmt.Errorf("reconcileInvoiceStock: query recorded: %w", err)
	}
	type netMove struct {
		variantID, warehouseID int
		qty                    float64
	}
	var recorded []netMove
	for rows.Next() {
		var m netMove
		if err := rows.Scan(&m.variantID, &m.warehouseID, &m.qty); err != nil {
			rows.Close()
			return fmt.Errorf("reconcileInvoiceStock: scan recorded: %w", err)
		}
		recorded = append(recorded, m)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	for _, m := range recorded {
		neutral := -m.qty
		if err := ptx.Insert(context.Background(), &InventoryMovement{
			CompanyID:       companyID,
			VariantID:       m.variantID,
			WarehouseID:     m.warehouseID,
			TransactionKind: InventoryMovementKinds.Adjustment,
			Qty:             neutral,
			ReferenceType:   "invoice",
			ReferenceID:     invoiceID,
			CreatedAt:       time.Now().UTC(),
		}); err != nil {
			return fmt.Errorf("reconcileInvoiceStock: insert adjustment: %w", err)
		}
		// Stays raw: `quantity = quantity + $1` is a self-referencing increment.
		if _, err := tx.Exec(
			`UPDATE inventory_balances SET quantity = quantity + $1, updated_at = NOW()
			  WHERE company_id = $2 AND variant_id = $3 AND warehouse_id = $4`,
			neutral, companyID, m.variantID, m.warehouseID,
		); err != nil {
			return fmt.Errorf("reconcileInvoiceStock: neutralise balance: %w", err)
		}
	}

	// 2. Re-record OUT movements from the surviving lines. Collect first, then
	// record — recordMovement issues its own queries on tx, which cannot run
	// while a Rows from the same connection is still open.
	// invoiceLineRead carries no softdelete tag, so deleted_at is written out here as
	// the old statement had it. Get reads the whole result set before returning, so the
	// "collect first, then record" dance below is no longer forced by an open Rows —
	// but recordMovement still issues its own queries on tx, so the loop stays split.
	var lines []invoiceLineRead
	if err := ptx.Model(&invoiceLineRead{}).
		Select("variant_id", "warehouse_id", "unit_id", "qty", "price").
		WhereEq("company_id", companyID).
		WhereEq("invoice_id", invoiceID).
		WhereNull("deleted_at").
		Where("warehouse_id", "!=", 0).
		Get(context.Background(), &lines); err != nil {
		return fmt.Errorf("reconcileInvoiceStock: query lines: %w", err)
	}

	for _, l := range lines {
		// OUT movement: qty is negative.
		if err := s.recordMovement(
			tx, companyID,
			l.VariantID, int(l.WarehouseID), l.UnitID,
			-float64(l.Qty), l.Price,
			InventoryMovementKinds.Sale,
			"invoice", invoiceID,
		); err != nil {
			return err
		}
	}
	return nil
}
