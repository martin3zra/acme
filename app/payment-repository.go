package app

import (
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
)

type payment struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
	foundation.Timestamps
}

func (s *Server) findPayments(companyId int) ([]*payment, error) {
	rows, err := s.db.Query(`
    SELECT receivables_income.date, receivables_income.amount, receivables_income.created_at, receivables_income.updated_at
    FROM receivables_income
    WHERE company_id = $1
  `, companyId)
	if err != nil {
		return nil, err
	}
	data := make([]*payment, 0)
	for rows.Next() {
		i := new(payment)

		if err = rows.Scan(); err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	return data, nil
}
