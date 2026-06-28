package app

import (
	"context"
	"database/sql"

	"github.com/martin3zra/forge/foundation"
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
	var w warehouse
	err := s.db.QueryRow(
		"SELECT w.id, w.uuid, w.name, COALESCE(w.location, ''), w.status, w.created_at, w.updated_at, w.deleted_at "+
			"FROM warehouses w "+
			"WHERE w.company_id = $1 AND w.id = $2 AND w.deleted_at IS NULL",
		CurrentCompany(ctx).ID,
		warehouseID,
	).Scan(&w.ID, &w.UUID, &w.Name, &w.Location, &w.Status, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// findWarehouses lists active warehouses for the current company. Pilot:
// served by playsql (no enum predicate, so the whole query runs through it).
func (s *Server) findWarehouses(ctx context.Context) ([]*warehouse, error) {
	models, err := listWarehousesByCompany(ctx, s.play, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}

	data := make([]*warehouse, 0, len(models))
	for _, m := range models {
		data = append(data, toWarehouse(m))
	}
	return data, nil
}

func (s *Server) storeWarehouse(ctx context.Context, form *StoreWarehouseForm) error {
	location := sql.NullString{String: form.Location, Valid: form.Location != ""}
	_, err := s.db.Exec(
		"INSERT INTO warehouses (company_id, name, location) VALUES ($1, $2, $3)",
		CurrentCompany(ctx).ID,
		form.Name,
		location,
	)
	return err
}

func (s *Server) updateWarehouse(ctx context.Context, warehouseID int, form *UpdateWarehouseForm) error {
	location := sql.NullString{String: form.Location, Valid: form.Location != ""}
	_, err := s.db.Exec(
		"UPDATE warehouses SET name = $1, location = $2, updated_at = now() WHERE company_id = $3 AND id = $4",
		form.Name,
		location,
		CurrentCompany(ctx).ID,
		warehouseID,
	)
	return err
}

func (s *Server) deleteWarehouse(ctx context.Context, warehouseID int) error {
	_, err := s.db.Exec(
		"UPDATE warehouses SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID,
		warehouseID,
	)
	return err
}

func (s *Server) toggleWarehouseStatus(ctx context.Context, w *warehouse) error {
	status := w.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	_, err := s.db.Exec(
		"UPDATE warehouses SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID,
		w.ID,
		status,
	)
	return err
}
