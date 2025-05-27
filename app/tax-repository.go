package app

import (
	"context"

	"github.com/martin3zra/acme/pkg/foundation"
)

type tax struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findTaxes(ctx context.Context) ([]*tax, error) {
	rows, err := s.db.Query("SELECT id, name, rate, created_at, updated_at, deleted_at FROM taxes WHERE company_id = $1", CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}

	data := make([]*tax, 0)
	for rows.Next() {
		i := new(tax)
		if err = rows.Scan(
			&i.ID,
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
