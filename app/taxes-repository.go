package app

import (
	"context"

	"github.com/martin3zra/forge/foundation"
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

// findTaxesInternal deliberately does not filter deleted_at — the raw query never
// did, and nothing soft-deletes a tax. taxRead therefore carries no softdelete tag.
func (s *Server) findTaxesInternal(companyID int) ([]*tax, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []taxRead
	if err := pdb.Model(&taxRead{}).
		WhereEq("company_id", companyID).
		Get(context.Background(), &rows); err != nil {
		return nil, err
	}

	data := make([]*tax, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toTax())
	}
	return data, nil
}

func (s *Server) storeTax(ctx context.Context, form *StoreTaxForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&taxRead{}).Insert(ctx, map[string]any{
		"company_id": CurrentCompany(ctx).ID,
		"name":       form.Name,
		"rate":       form.Rate,
	})
	return err
}

func (s *Server) updateTax(ctx context.Context, uuid string, form *StoreTaxForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&taxRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		Update(ctx, map[string]any{"name": form.Name, "rate": form.Rate})
	return mustAffectRows(affected, err, "tax")
}
