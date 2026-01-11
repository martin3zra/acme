package app

import (
	"context"
	"time"
)

type stats struct {
	TotalDueAmount float64 `json:"total_due_amount"`
	TotalCustomers int     `json:"total_customers"`
	TotalInvoices  int     `json:"total_invoices"`
}

type dueInvoice struct {
	UUID     string    `json:"uuid"`
	DueOn    time.Time `json:"due_on"`
	Customer struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"customer"`
	AmountDue float64 `json:"amount_due"`
}

func (s *Server) findStats(ctx context.Context) (*stats, error) {
	var data stats
	err := s.db.QueryRow(`SELECT
    (SELECT COALESCE(SUM(amount_due), 0) FROM invoices WHERE company_id = $1 AND due_on <= CURRENT_DATE AND paid_status IN ('unpaid','partial') ) AS total_due_amount,
    (SELECT COUNT(*) FROM customers WHERE company_id = $1) AS total_customers,
    (SELECT COUNT(*) FROM invoices WHERE company_id = $1) AS total_invoices;
  `, CurrentCompany(ctx).ID).Scan(&data.TotalDueAmount, &data.TotalCustomers, &data.TotalInvoices)

	return &data, err
}

func (s *Server) findLatestDueInvoices(ctx context.Context) ([]*dueInvoice, error) {
	rows, err := s.db.Query(`SELECT i.uuid, i.due_on, c.uuid as customer_uuid, c.name, i.amount_due
    FROM invoices i
    JOIN customers c ON c.id = i.customer_id
    WHERE i.company_id = $1
    AND i.due_on <= CURRENT_DATE
    AND i.paid_status IN ('partial', 'unpaid')
    AND i.status IN ('open', 'partial')
    ORDER BY i.due_on DESC
    LIMIT 10;`, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*dueInvoice, 0)
	for rows.Next() {
		row := new(dueInvoice)
		if err = rows.Scan(&row.UUID, &row.DueOn, &row.Customer.UUID, &row.Customer.Name, &row.AmountDue); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	return data, nil
}
