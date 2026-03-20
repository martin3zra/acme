package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type vendorPayment struct {
	ID        int      `json:"id"`
	UUID      string   `json:"uuid"`
	CompanyID int      `json:"company_id"`
	Vendor    vendor   `json:"vendor"`
	Date      time.Time `json:"date"`
	Amount    float64  `json:"amount"`
	Notes     string   `json:"notes"`
	Payment   *Payment `json:"payment"`
	Status    string   `json:"status"`
	Code      string   `json:"code"`
	foundation.Timestamps
}

type vendorPaymentLine struct {
	ID         int       `json:"id"`
	APID       int64     `json:"ap_id"`
	APUUID     string    `json:"ap_uuid"`
	BillNumber string    `json:"bill_number"`
	BillDate   time.Time `json:"bill_date"`
	DueDate    time.Time `json:"due_date"`
	AmountDue  float64   `json:"amount_due"`
	Payment    float64   `json:"payment"`
	PaidStatus string    `json:"paid_status"`
}

func (s *Server) findPayables(ctx context.Context) ([]*Payable, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			p.id, p.uuid,
			ap.uuid, ap.id, ap.invoice_number,
			ap.invoice_date, ap.due_date,
			ap.amount_total, ap.amount_payable, ap.amount_paid,
			ap.status, ap.paid_status, COALESCE(ap.notes, '')
		FROM payables p
		INNER JOIN companies        ON (p.company_id = companies.id)
		INNER JOIN accounts_payable ap ON (p.company_id = ap.company_id AND p.vendor_id = ap.vendor_id AND p.accounts_payable_id = ap.id)
		INNER JOIN vendors          ON (p.company_id = vendors.company_id AND p.vendor_id = vendors.id)
		WHERE p.company_id = $1
		  AND ap.status != 'void'
		  AND ap.paid_status != 'paid'
		ORDER BY ap.due_date ASC`,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*Payable, 0)
	for rows.Next() {
		row := new(Payable)
		var notes string
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.InvoiceUUID,
			&row.InvoiceID,
			&row.InvoiceNumber,
			&row.InvoiceDate,
			&row.DueDate,
			&row.AmountTotal,
			&row.AmountPayable,
			&row.AmountPaid,
			&row.Status,
			&row.PaidStatus,
			&notes,
		); err != nil {
			return data, err
		}
		row.Notes = &notes
		data = append(data, row)
	}
	return data, nil
}

func (s *Server) storeVendorPayment(ctx context.Context, form *StoreVendorPaymentForm) error {
	companyID := CurrentCompany(ctx).ID
	v, err := s.findVendorByUUID(ctx, form.VendorID)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		seqInfo, err := GetNextSequence(tx, companyID, "vendor_payment")
		if err != nil {
			return err
		}

		var paymentID int
		err = tx.QueryRow(
			"INSERT INTO vendor_payments (company_id, vendor_id, date, amount, notes, payment, status, code) "+
				"VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id",
			companyID,
			v.ID,
			form.Date,
			form.Amount,
			form.Notes,
			foundation.ToJSON(form.Payment),
			"completed",
			seqInfo.Code,
		).Scan(&paymentID)
		if err != nil {
			return err
		}

		for _, line := range form.Lines {
			ap, err := s.findAPByUUID(tx, companyID, line.UUID)
			if err != nil {
				return err
			}
			if _, err = tx.Exec(
				"INSERT INTO vendor_payment_items (company_id, vendor_payment_id, accounts_payable_id, date, amount_due, payment_amount) "+
					"VALUES ($1,$2,$3,$4,$5,$6)",
				companyID, paymentID, ap.ID, form.Date, line.AmountDue, line.Payment,
			); err != nil {
				return err
			}
			if err = s.updateAPBalance(tx, companyID, ap.ID, line.Payment); err != nil {
				return err
			}
		}

		return s.updateVendorAmountPayable(tx, companyID, v.ID, -form.Amount)
	})
}

func (s *Server) findAPByUUID(tx *sql.Tx, companyID int, uuid string) (*Payable, error) {
	ap := new(Payable)
	err := tx.QueryRow(
		"SELECT id, uuid, amount_payable, amount_paid FROM accounts_payable WHERE company_id = $1 AND uuid = $2",
		companyID, uuid,
	).Scan(&ap.ID, &ap.UUID, &ap.AmountPayable, &ap.AmountPaid)
	return ap, err
}

func (s *Server) updateAPBalance(tx *sql.Tx, companyID int, apID int64, paymentAmount float64) error {
	var newPaid float64
	var amountPayable float64
	err := tx.QueryRow(
		"UPDATE accounts_payable SET amount_paid = amount_paid + $3, updated_at = NOW() WHERE company_id = $1 AND id = $2 RETURNING amount_paid, amount_payable",
		companyID, apID, paymentAmount,
	).Scan(&newPaid, &amountPayable)
	if err != nil {
		return err
	}

	var newStatus PaidStatus
	switch {
	case newPaid >= amountPayable:
		newStatus = PaidStatuses.Paid
	case newPaid > 0:
		newStatus = PaidStatuses.Partial
	default:
		newStatus = PaidStatuses.UnPaid
	}

	if newStatus == PaidStatuses.Paid {
		_, err = tx.Exec(
			"UPDATE accounts_payable SET paid_status = $3, paid_at = NOW(), updated_at = NOW() WHERE company_id = $1 AND id = $2",
			companyID, apID, newStatus,
		)
	} else {
		_, err = tx.Exec(
			"UPDATE accounts_payable SET paid_status = $3, updated_at = NOW() WHERE company_id = $1 AND id = $2",
			companyID, apID, newStatus,
		)
	}
	if err != nil {
		return err
	}

	// Propagate payment status to the source purchase order (if this AP entry
	// was created from a vendor bill that is linked to a PO via source).
	var poUUID string
	lookupErr := tx.QueryRow(
		`SELECT p_po.uuid
		   FROM accounts_payable ap
		   JOIN purchases p_bill ON ap.purchase_id = p_bill.id
		   JOIN purchases p_po   ON p_po.company_id = p_bill.company_id
		                        AND p_po.uuid = p_bill.source->>'id'
		                        AND p_po.transaction_kind = 'purchase_order'
		  WHERE ap.company_id = $1 AND ap.id = $2`,
		companyID, apID,
	).Scan(&poUUID)
	if lookupErr == nil && poUUID != "" {
		if err = updatePOPaymentStatus(tx, companyID, poUUID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) findVendorPaymentByUUID(ctx context.Context, uuid string) (*vendorPayment, error) {
	p := new(vendorPayment)
	var paymentJSON []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT vp.id, vp.uuid, vp.date, vp.amount, COALESCE(vp.notes,''), vp.payment, vp.status, vp.code,
		       v.id, v.uuid, v.name, v.email, v.phone
		FROM vendor_payments vp
		INNER JOIN vendors v ON (vp.company_id = v.company_id AND vp.vendor_id = v.id)
		WHERE vp.company_id = $1 AND vp.uuid = $2`,
		CurrentCompany(ctx).ID, uuid,
	).Scan(
		&p.ID, &p.UUID, &p.Date, &p.Amount, &p.Notes, &paymentJSON, &p.Status, &p.Code,
		&p.Vendor.ID, &p.Vendor.UUID, &p.Vendor.Name, &p.Vendor.Email, &p.Vendor.Phone,
	)
	if err != nil {
		return nil, err
	}
	if len(paymentJSON) > 0 {
		pm := new(Payment)
		if err = json.Unmarshal(paymentJSON, pm); err == nil {
			p.Payment = pm
		}
	}
	return p, nil
}

func (s *Server) findVendorPaymentLines(ctx context.Context, paymentID int) ([]*vendorPaymentLine, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT vpi.id,
		       ap.id, ap.uuid, ap.invoice_number,
		       ap.invoice_date, ap.due_date,
		       vpi.amount_due, vpi.payment_amount,
		       CASE WHEN (vpi.amount_due - vpi.payment_amount) <= 0 THEN 'paid' ELSE 'partial' END
		FROM vendor_payment_items vpi
		INNER JOIN accounts_payable ap ON (vpi.company_id = ap.company_id AND vpi.accounts_payable_id = ap.id)
		WHERE vpi.company_id = $1 AND vpi.vendor_payment_id = $2
		ORDER BY vpi.id`,
		CurrentCompany(ctx).ID, paymentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*vendorPaymentLine, 0)
	for rows.Next() {
		l := new(vendorPaymentLine)
		if err = rows.Scan(
			&l.ID,
			&l.APID, &l.APUUID, &l.BillNumber,
			&l.BillDate, &l.DueDate,
			&l.AmountDue, &l.Payment,
			&l.PaidStatus,
		); err != nil {
			return data, err
		}
		data = append(data, l)
	}
	return data, nil
}

func (s *Server) voidVendorPayment(ctx context.Context, uuid string) error {
	p, err := s.findVendorPaymentByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	if p.Status == "void" {
		return errors.New("vendor payment already voided")
	}

	lines, err := s.findVendorPaymentLines(ctx, p.ID)
	if err != nil {
		return err
	}

	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		for _, l := range lines {
			if err = s.updateAPBalance(tx, companyID, l.APID, -l.Payment); err != nil {
				return err
			}
			if _, err = tx.Exec(
				"UPDATE vendor_payment_items SET payment_amount = 0, updated_at = NOW() WHERE company_id = $1 AND vendor_payment_id = $2 AND accounts_payable_id = $3",
				companyID, p.ID, l.APID,
			); err != nil {
				return err
			}
		}

		if err = s.updateVendorAmountPayable(tx, companyID, p.Vendor.ID, p.Amount); err != nil {
			return err
		}

		_, err = tx.Exec(
			"UPDATE vendor_payments SET amount = 0, status = 'void', updated_at = NOW() WHERE company_id = $1 AND id = $2",
			companyID, p.ID,
		)
		return err
	})
}

func (s *Server) findVendorPayments(ctx context.Context) ([]*vendorPayment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT vp.id, vp.uuid, vp.date, vp.amount, COALESCE(vp.notes,''), vp.payment, vp.status, vp.code,
		       v.id, v.uuid, v.name, v.email, v.phone
		FROM vendor_payments vp
		INNER JOIN vendors v ON (vp.company_id = v.company_id AND vp.vendor_id = v.id)
		WHERE vp.company_id = $1
		ORDER BY vp.id DESC`,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*vendorPayment, 0)
	for rows.Next() {
		p := new(vendorPayment)
		var paymentJSON []byte
		if err = rows.Scan(
			&p.ID, &p.UUID, &p.Date, &p.Amount, &p.Notes, &paymentJSON, &p.Status, &p.Code,
			&p.Vendor.ID, &p.Vendor.UUID, &p.Vendor.Name, &p.Vendor.Email, &p.Vendor.Phone,
		); err != nil {
			return data, err
		}
		if len(paymentJSON) > 0 {
			pm := new(Payment)
			if err = json.Unmarshal(paymentJSON, pm); err == nil {
				p.Payment = pm
			}
		}
		data = append(data, p)
	}
	return data, nil
}
