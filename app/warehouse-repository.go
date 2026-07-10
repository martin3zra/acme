package app

import (
	"context"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type warehouse struct {
	ID       int               `json:"id"`
	UUID     string            `json:"uuid"`
	Name     string            `json:"name"`
	Location string            `json:"location"`
	Status   foundation.Status `json:"status"`
	foundation.Timestamps
}

func (s *Server) findWarehouseByID(ctx context.Context, warehouseID int) (*warehouse, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row warehouseRead
	if err := pdb.Model(&warehouseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", warehouseID).
		First(ctx, &row); err != nil {
		return nil, err
	}
	return row.toWarehouse(), nil
}

func (s *Server) findWarehouses(ctx context.Context) ([]*warehouse, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []warehouseRead
	if err := pdb.Model(&warehouseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*warehouse, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toWarehouse())
	}
	return data, nil
}

func (s *Server) storeWarehouse(ctx context.Context, form *StoreWarehouseForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	// nullIfEmpty keeps the sql.NullString behaviour: a blank location stores NULL.
	_, err = pdb.Model(&warehouseRead{}).Insert(ctx, map[string]any{
		"company_id": CurrentCompany(ctx).ID,
		"name":       form.Name,
		"location":   nullIfEmpty(form.Location),
	})
	return err
}

func (s *Server) updateWarehouse(ctx context.Context, warehouseID int, form *UpdateWarehouseForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&warehouseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", warehouseID).
		Update(ctx, map[string]any{
			"name":     form.Name,
			"location": nullIfEmpty(form.Location),
		})
	return mustAffectRows(affected, err, "warehouse")
}

// deleteWarehouse soft-deletes through Update, not Delete: Builder.Delete stamps
// deleted_at only, and the statement it replaced bumped updated_at too. The softdelete
// tag also adds `deleted_at IS NULL`, so deleting an already-deleted warehouse is now
// a not-found rather than a silent second write.
func (s *Server) deleteWarehouse(ctx context.Context, warehouseID int) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&warehouseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", warehouseID).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return mustAffectRows(affected, err, "warehouse")
}

func (s *Server) toggleWarehouseStatus(ctx context.Context, w *warehouse) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	status := w.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}

	affected, err := pdb.Model(&warehouseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", w.ID).
		Update(ctx, map[string]any{"status": string(status)})
	return mustAffectRows(affected, err, "warehouse")
}
