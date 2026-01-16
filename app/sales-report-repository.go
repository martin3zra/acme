package app

import (
	"context"
	"fmt"
	"time"
)

type ReportGroup interface {
	Name() string
	TotalAmount() float64
}

type SalesByCustomerWithInvoices struct {
	InvoiceID    int       `json:"invoice_id"`
	Date         time.Time `json:"date"`
	Total        float64   ` json:"total"`
	CustomerID   int       `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
}

type CustomerGroup struct {
	CustomerID   int                           `json:"customer_id"`
	CustomerName string                        `json:"customer_name"`
	Invoices     []SalesByCustomerWithInvoices `json:"invoices"`
	Total        float64                       `json:"total"`
}

func (c CustomerGroup) Name() string {
	return c.CustomerName
}

func (c CustomerGroup) TotalAmount() float64 {
	return c.Total
}

type ItemGroup struct {
	ItemID   int
	ItemName string
	Total    float64
}

func (i ItemGroup) Name() string {
	return i.ItemName
}

func (i ItemGroup) TotalAmount() float64 {
	return i.Total
}

type SalesByDate struct {
	InvoiceDate  time.Time `json:"invoice_date"`
	TotalSales   float64   `json:"total_sales"`
	InvoiceCount int       `json:"invoice_count"`
}

func (s *Server) findItemWiseSales(ctx context.Context, From, To time.Time) ([]ItemGroup, error) {
	rows, err := s.db.Query(`
		SELECT i.id AS item_id, i.name AS item_name,  SUM(s.total) AS total_amount
		FROM items i
		JOIN invoices_items s ON (s.company_id = i.company_id AND s.item_id = i.id)
		JOIN invoices h ON (h.company_id = s.company_id AND h.id = s.invoice_id)
		WHERE s.company_id = $1
		AND h.date BETWEEN $2 AND $3
		GROUP BY i.id, i.name
		ORDER BY i.name;
	`, CurrentCompany(ctx).ID, From, To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []ItemGroup
	for rows.Next() {
		var g ItemGroup
		if err := rows.Scan(&g.ItemID, &g.ItemName, &g.Total); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (s *Server) findCustomerWiseSales(ctx context.Context, From, To time.Time) ([]CustomerGroup, error) {
	rows, err := s.db.Query(`
    SELECT c.id AS customer_id, c.name AS customer_name, SUM(i.total) AS total_sales, COUNT(*) AS invoice_count
    FROM invoices i
    JOIN customers c ON (i.company_id = c.company_id AND i.customer_id = c.id)
    WHERE i.company_id = $1
    AND i.date BETWEEN $2 AND $3
    GROUP BY c.id, c.name
    ORDER BY total_sales DESC;`, CurrentCompany(ctx).ID, From, To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var _count int // to scan COUNT(*) value | discard that value for now
	var groups []CustomerGroup
	for rows.Next() {
		var g CustomerGroup
		if err := rows.Scan(&g.CustomerID, &g.CustomerName, &g.Total, &_count); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil

}

func (s *Server) findCustomerWiseSalesWithInvoices(ctx context.Context, From, To time.Time) ([]CustomerGroup, error) {
	rows, err := s.db.Query(`
    SELECT i.id AS invoice_id, i.date, i.total, i.customer_id, c.name AS customer_name
    FROM invoices i
    JOIN customers c ON (i.company_id = c.company_id AND i.customer_id = c.id)
    WHERE i.company_id = $1
    AND i.date BETWEEN $2 AND $3
    ORDER BY customer_name DESC;`, CurrentCompany(ctx).ID, From, To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groupsMap := make(map[int]*CustomerGroup)
	for rows.Next() {
		var cid int
		var cname string
		var inv SalesByCustomerWithInvoices
		if err := rows.Scan(&inv.InvoiceID, &inv.Date, &inv.Total, &cid, &cname); err != nil {
			return nil, err
		}
		if groupsMap[cid] == nil {
			groupsMap[cid] = &CustomerGroup{CustomerID: cid, CustomerName: cname}
		}
		groupsMap[cid].Invoices = append(groupsMap[cid].Invoices, inv)
	}

	var groups []CustomerGroup
	for _, g := range groupsMap {
		groups = append(groups, *g)
	}
	return groups, nil
}

func (s *Server) fetchSalesGroups(ctx context.Context, form *ReportSalesForm) (any, error) {
	switch form.ReportType {
	case "sales_by_customer":
		if form.ShowInvoices {
			return s.findCustomerWiseSalesWithInvoices(ctx, form.From, form.To)
		}
		return s.findCustomerWiseSales(ctx, form.From, form.To)
	case "sales_by_item":
		return s.findItemWiseSales(ctx, form.From, form.To)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", form.ReportType)
	}
}
