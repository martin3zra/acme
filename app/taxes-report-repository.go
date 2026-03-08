package app

import (
	"context"
	"time"
)

type TaxReportRow struct {
	TaxID    int     `json:"tax_id"`
	TaxName  string  `json:"tax_name"`
	TotalTax float64 `json:"total_tax"`
}

func (s *Server) findTaxesWiseSales(ctx context.Context, From, To time.Time) ([]TaxReportRow, error) {
	rows, err := s.db.Query(`
		SELECT  t.id AS tax_id, t.name AS tax_name, SUM(ii.qty * ii.price * t.rate / 100) AS total_tax
    FROM invoices i
    JOIN invoices_items ii ON i.id = ii.invoice_id
	    JOIN items_variants iv ON ii.variant_id = iv.id
	    JOIN items it ON iv.item_id = it.id
    JOIN taxes t ON it.tax_id = t.id
    WHERE i.company_id = $1
    AND i.paid_status = 'paid'
    AND i.date BETWEEN $2 AND $3
    GROUP BY t.id, t.name
    ORDER BY t.name;
	`, CurrentCompany(ctx).ID, From, To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []TaxReportRow
	for rows.Next() {
		var r TaxReportRow
		if err := rows.Scan(&r.TaxID, &r.TaxName, &r.TotalTax); err != nil {
			return nil, err
		}
		report = append(report, r)
	}
	return report, nil
}
