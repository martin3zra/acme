package app

import (
	"context"

	"github.com/martin3zra/acme/pkg/foundation"
)

type unit struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	BaseQty int    `json:"base_qty"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findUnitByBasedQty(companyID int) (int, error) {
	var id int
	err := s.db.QueryRow(`SELECT id FROM units WHERE company_id = $1 AND base_qty = 1`, companyID).Scan(&id)

	return id, err
}

func (s *Server) findUnits(ctx context.Context) ([]*unit, error) {
	rows, err := s.db.Query("SELECT id, name, base_qty, created_at, updated_at, deleted_at FROM units WHERE company_id = $1", CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*unit, 0)
	for rows.Next() {
		u := new(unit)
		if err = rows.Scan(
			&u.ID,
			&u.Name,
			&u.BaseQty,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, u)

	}
	return data, nil
}

func (s *Server) storeUnit(ctx context.Context, form *StoreUnitForm) error {
	_, err := s.db.Exec("INSERT INTO units (company_id, name, base_qty) VALUES($1, $2, $3)",
		CurrentCompany(ctx).ID, form.Name, form.BaseQty)
	return err
}

func (s *Server) updateUnit(ctx context.Context, id int, form *StoreUnitForm) error {
	_, err := s.db.Exec("UPDATE units SET name = $3, base_qty = $4, updated_at = NOW() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, id, form.Name, form.BaseQty)
	return err
}
