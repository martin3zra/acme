package app

import (
	"context"

	"github.com/martin3zra/forge/foundation"
)

type unit struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	BaseQty int    `json:"base_qty"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findUnitByBasedQty(companyID int) (int, error) {
	pdb, err := s.play()
	if err != nil {
		return 0, err
	}

	var row unitRead
	err = pdb.Model(&unitRead{}).
		WhereEq("company_id", companyID).
		WhereEq("base_qty", 1).
		First(context.Background(), &row)
	if err != nil {
		return 0, err
	}

	return int(row.ID), nil
}

func (s *Server) findUnits(ctx context.Context) ([]*unit, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []unitRead
	if err := pdb.Model(&unitRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*unit, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toUnit())
	}
	return data, nil
}

func (s *Server) storeUnit(ctx context.Context, form *StoreUnitForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&unitRead{}).Insert(ctx, map[string]any{
		"company_id": CurrentCompany(ctx).ID,
		"name":       form.Name,
		"base_qty":   form.BaseQty,
	})
	return err
}

func (s *Server) updateUnit(ctx context.Context, id int, form *StoreUnitForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&unitRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", id).
		Update(ctx, map[string]any{"name": form.Name, "base_qty": form.BaseQty})
	return mustAffectRows(affected, err, "unit")
}
