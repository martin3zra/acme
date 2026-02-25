package app

import (
	"context"
	"time"
)

type stats struct {
	TotalDueAmount float64 `json:"total_due_amount"`
	TotalCustomers int     `json:"total_customers"`
	TotalInvoices  int     `json:"total_invoices"`
	TotalEstimates int     `json:"total_estimates"`
}

type dueInvoice struct {
	UUID     string    `json:"uuid"`
	Status   string    `json:"status"`
	DueOn    time.Time `json:"due_on"`
	Customer struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"customer"`
	AmountDue float64 `json:"amount_due"`
}

type ChartData struct {
	Month    string  `json:"month"`
	Sales    float64 `json:"sales"`
	Expenses float64 `json:"expenses"`
}

type Totals struct {
	TotalSales    float64 `json:"totalSales"`
	TotalReceipts float64 `json:"totalReceipts"`
	TotalExpenses float64 `json:"totalExpenses"`
	NetIncome     float64 `json:"netIncome"`
}

type DashboardData struct {
	ChartData []ChartData `json:"chartData"`
	Totals    Totals      `json:"totals"`
}

func (s *Server) findStats(ctx context.Context) (*stats, error) {
	var data stats
	err := s.db.QueryRow(`SELECT
    (SELECT COALESCE(SUM(amount_due), 0) FROM invoices WHERE company_id = $1 AND transaction_kind = 'invoice' AND due_on <= CURRENT_DATE AND paid_status IN ('unpaid','partial') ) AS total_due_amount,
    (SELECT COUNT(*) FROM customers WHERE company_id = $1) AS total_customers,
    (SELECT COUNT(*) FROM invoices WHERE company_id = $1 AND transaction_kind = 'invoice') AS total_invoices,
    (SELECT COUNT(*) FROM invoices WHERE company_id = $1 AND transaction_kind = 'estimate') AS total_estimates;
  `, CurrentCompany(ctx).ID).Scan(&data.TotalDueAmount, &data.TotalCustomers, &data.TotalInvoices, &data.TotalEstimates)

	return &data, err
}

func (s *Server) findLatestDueInvoices(ctx context.Context) ([]*dueInvoice, error) {
	rows, err := s.db.Query(`SELECT i.uuid, i.status, i.due_on, c.uuid as customer_uuid, c.name, i.amount_due
    FROM invoices i
    JOIN customers c ON c.id = i.customer_id
    WHERE i.company_id = $1
    AND i.transaction_kind = 'invoice'
    AND i.due_on <= CURRENT_DATE
    AND i.paid_status IN ('partial', 'unpaid')
    AND i.status = 'overdue'
    ORDER BY i.due_on DESC
    LIMIT 10;`, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*dueInvoice, 0)
	for rows.Next() {
		row := new(dueInvoice)
		if err = rows.Scan(&row.UUID, &row.Status, &row.DueOn, &row.Customer.UUID, &row.Customer.Name, &row.AmountDue); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	return data, nil
}

func (s *Server) findLatestEstimates(ctx context.Context) ([]*dueInvoice, error) {
	rows, err := s.db.Query(`SELECT i.uuid, i.status, i.date, c.uuid as customer_uuid, c.name, i.total as amount_due
    FROM invoices i
    JOIN customers c ON c.id = i.customer_id
    WHERE i.company_id = $1
    AND i.transaction_kind = 'estimate'
    AND i.status = 'sent'::invoice_status
    ORDER BY i.date DESC
    LIMIT 10;`, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*dueInvoice, 0)
	for rows.Next() {
		row := new(dueInvoice)
		if err = rows.Scan(&row.UUID, &row.Status, &row.DueOn, &row.Customer.UUID, &row.Customer.Name, &row.AmountDue); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	return data, nil
}

func (s *Server) findLastProfitOfLast12Months(ctx context.Context) ([]*ChartData, error) {
	rows, err := s.db.Query(`
    WITH months AS (
        SELECT generate_series(
            DATE_TRUNC('month', CURRENT_DATE - INTERVAL '11 months'),
            DATE_TRUNC('month', CURRENT_DATE),
            interval '1 month'
        ) AS month
    ),
    invoice_totals AS (
        SELECT 
            DATE_TRUNC('month', date) AS month,
            SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END) AS sales
        FROM invoices
        WHERE company_id = $1
          AND transaction_kind = 'invoice'
          AND date >= (CURRENT_DATE - INTERVAL '12 months')
        GROUP BY DATE_TRUNC('month', date)
    ),
    expense_totals AS (
        SELECT 
            DATE_TRUNC('month', date) AS month,
            SUM(amount) AS expenses
        FROM expenses
        WHERE company_id = $1
          AND deleted_at IS NULL
          AND date >= (CURRENT_DATE - INTERVAL '12 months')
        GROUP BY DATE_TRUNC('month', date)
    )
    SELECT 
        TO_CHAR(m.month, 'YYYY/Mon') AS year_month,
        COALESCE(i.sales, 0) AS sales,
        COALESCE(e.expenses, 0) AS expenses
    FROM months m
    LEFT JOIN invoice_totals i ON m.month = i.month
    LEFT JOIN expense_totals e ON m.month = e.month
    ORDER BY m.month;
  `, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*ChartData, 0)
	for rows.Next() {
		row := new(ChartData)
		if err = rows.Scan(&row.Month, &row.Sales, &row.Expenses); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	return data, nil
}

func (s *Server) findTotalsProfitOfLast12Months(ctx context.Context) (*Totals, error) {
	var data Totals
	err := s.db.QueryRow(`
    SELECT
      COALESCE(SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END), 0) AS total_sales,
      (SELECT COALESCE(SUM(r.amount), 0) FROM receivables_income r WHERE r.company_id = $1 AND r.date >= (CURRENT_DATE - INTERVAL '12 months') AND r.deleted_at IS NULL) AS total_receipts,
      (SELECT COALESCE(SUM(amount), 0) FROM expenses e WHERE e.company_id = $1 AND e.deleted_at IS NULL) AS total_expenses,
      COALESCE(SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END), 0) - COALESCE(SUM(0), 0) AS net_income
    FROM invoices 
    WHERE company_id = $1
    AND transaction_kind = 'invoice'
    AND date >= (CURRENT_DATE - INTERVAL '12 months');`, CurrentCompany(ctx).ID).
		Scan(&data.TotalSales, &data.TotalReceipts, &data.TotalExpenses, &data.NetIncome)
	return &data, err
}

func (s *Server) findLastProfitOfYear(ctx context.Context, year int) ([]*ChartData, error) {
	rows, err := s.db.Query(`
    WITH months AS (
        SELECT generate_series(
          make_date($2, 1, 1),
          make_date($2 + 1, 1, 1) - interval '1 month',
          interval '1 month'
      )::date AS month
    ),
    invoice_totals AS (
        SELECT 
            DATE_TRUNC('month', date) AS month,
            SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END) AS sales
        FROM invoices
        WHERE company_id = $1
          AND transaction_kind = 'invoice'
          AND EXTRACT(YEAR FROM date) = $2
        GROUP BY DATE_TRUNC('month', date)
    ),
    expense_totals AS (
        SELECT 
            DATE_TRUNC('month', date) AS month,
            SUM(amount) AS expenses
        FROM expenses
        WHERE company_id = $1
          AND deleted_at IS NULL
          AND EXTRACT(YEAR FROM date) = $2
        GROUP BY DATE_TRUNC('month', date)
    )
    SELECT 
        TO_CHAR(m.month, 'YYYY/Mon') AS year_month,
        COALESCE(i.sales, 0) AS sales,
        COALESCE(e.expenses, 0) AS expenses
    FROM months m
    LEFT JOIN invoice_totals i ON m.month = i.month
    LEFT JOIN expense_totals e ON m.month = e.month
    ORDER BY m.month;
  `, CurrentCompany(ctx).ID, year)
	if err != nil {
		return nil, err
	}
	data := make([]*ChartData, 0)
	for rows.Next() {
		row := new(ChartData)
		if err = rows.Scan(&row.Month, &row.Sales, &row.Expenses); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	return data, nil
}

func (s *Server) findTotalsProfitOfYear(ctx context.Context, year int) (*Totals, error) {
	var data Totals
	err := s.db.QueryRow(`
    SELECT
      COALESCE(SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END), 0) AS total_sales,
      (SELECT COALESCE(SUM(r.amount), 0) FROM receivables_income r WHERE r.company_id = $1 AND EXTRACT(YEAR FROM r.date) = $2 AND r.deleted_at IS NULL) AS total_receipts,
      (SELECT COALESCE(SUM(amount), 0) FROM expenses e WHERE e.company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND e.deleted_at IS NULL) AS total_expenses,
      COALESCE(SUM(CASE WHEN status = 'closed' THEN total ELSE 0 END), 0) - COALESCE(SUM(0), 0) AS net_income
    FROM invoices
    WHERE company_id = $1
    AND transaction_kind = 'invoice'
    AND EXTRACT(YEAR FROM date) = $2;
  `, CurrentCompany(ctx).ID, year).
		Scan(&data.TotalSales, &data.TotalReceipts, &data.TotalExpenses, &data.NetIncome)
	return &data, err
}
