package app

import (
	"context"
	"time"
)

type SalesByCustomerWithInvoices struct {
	InvoiceID    int       `json:"invoice_id"`
	Date         time.Time `json:"date"`
	Total        float64   ` json:"total"`
	CustomerID   int       `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
}

type SalesByCustomer struct {
	CustomerID   int     `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	TotalSales   float64 `json:"total_sales"`
	InvoiceCount int     `json:"invoice_count"`
}

type CustomerGroup struct {
	CustomerID   int                           `json:"customer_id"`
	CustomerName string                        `json:"customer_name"`
	Invoices     []SalesByCustomerWithInvoices `json:"invoices"`
	Total        float64                       `json:"total"`
}

type SalesByDate struct {
	InvoiceDate  time.Time `json:"invoice_date"`
	TotalSales   float64   `json:"total_sales"`
	InvoiceCount int       `json:"invoice_count"`
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
	// data := make([]*SalesByCustomer, 0)
	// for rows.Next() {
	// 	var item SalesByCustomer
	// 	if err := rows.Scan(
	// 		&item.CustomerID,
	// 		&item.CustomerName,
	// 		&item.TotalSales,
	// 		&item.InvoiceCount,
	// 	); err != nil {
	// 		return nil, err
	// 	}
	// 	data = append(data, &item)
	// }
	// return data, nil

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
	// data := make([]*SalesByCustomerWithInvoices, 0)
	// for rows.Next() {
	// 	var item SalesByCustomerWithInvoices
	// 	if err := rows.Scan(
	// 		&item.InvoiceID,
	// 		&item.Date,
	// 		&item.Total,
	// 		&item.CustomerID,
	// 		&item.CustomerName,
	// 	); err != nil {
	// 		return nil, err
	// 	}
	// 	data = append(data, &item)
	// }
	// return data, nil
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

func (s *Server) groupByCustomer(invoices []*SalesByCustomerWithInvoices) []CustomerGroup {
	groups := make(map[string][]SalesByCustomerWithInvoices)
	for _, inv := range invoices {
		groups[inv.CustomerName] = append(groups[inv.CustomerName], *inv)
	}
	var result []CustomerGroup
	for customer, invs := range groups {
		result = append(result, CustomerGroup{CustomerName: customer, Invoices: invs})
	}
	return result
}
