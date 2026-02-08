package app

import (
	"context"

	"github.com/martin3zra/acme/pkg/foundation"
)

type tax struct {
	ID   int64   `json:"id"`
	UUID string  `json:"uuid"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findTaxes(ctx context.Context) ([]*tax, error) {
	return s.findTaxesInternal(CurrentCompany(ctx).ID)
}

func (s *Server) findTaxesInternal(companyID int) ([]*tax, error) {
	rows, err := s.db.Query("SELECT id, uuid, name, rate, created_at, updated_at, deleted_at FROM taxes WHERE company_id = $1", companyID)
	if err != nil {
		return nil, err
	}

	data := make([]*tax, 0)
	for rows.Next() {
		i := new(tax)
		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Name,
			&i.Rate,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}

	return data, nil
}

func (s *Server) storeTax(ctx context.Context, form *StoreTaxForm) error {
	_, err := s.db.Exec("INSERT INTO taxes (company_id, name, rate) VALUES($1, $2, $3)",
		CurrentCompany(ctx).ID, form.Name, form.Rate)
	return err
}

func (s *Server) updateTax(ctx context.Context, uuid string, form *StoreTaxForm) error {
	_, err := s.db.Exec("UPDATE taxes SET name = $3, rate = $4, updated_at = NOW() WHERE company_id = $1 AND uuid = $2",
		CurrentCompany(ctx).ID, uuid, form.Name, form.Rate)
	return err
}
