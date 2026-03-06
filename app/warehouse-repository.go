package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"acme/pkg/database"
	"acme/pkg/foundation"
)

// findWarehouses returns all active warehouses for the current company
func (s *Server) findWarehouses(ctx context.Context) ([]*warehouse, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT w.id, w.uuid, w.code, w.name, w.address, w.description, w.status, 
		        w.created_at, w.updated_at, w.deleted_at 
		 FROM warehouses w 
		 WHERE w.company_id = $1 AND w.deleted_at IS NULL 
		 ORDER BY w.name`,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*warehouse, 0)
	for rows.Next() {
		w := new(warehouse)
		if err = rows.Scan(
			&w.ID, &w.UUID, &w.Code, &w.Name, &w.Address, &w.Description, &w.Status,
			&w.CreatedAt, &w.UpdatedAt, &w.DeletedAt,
		); err != nil {
			return data, err
		}
		data = append(data, w)
	}

	return data, rows.Err()
}

// findWarehouseByID returns a single warehouse by ID
func (s *Server) findWarehouseByID(ctx context.Context, id int) (*warehouse, error) {
	w := new(warehouse)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT w.id, w.uuid, w.code, w.name, w.address, w.description, w.status,
		        w.created_at, w.updated_at, w.deleted_at
		 FROM warehouses w
		 WHERE w.company_id = $1 AND w.id = $2 AND w.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&w.ID, &w.UUID, &w.Code, &w.Name, &w.Address, &w.Description, &w.Status,
		&w.CreatedAt, &w.UpdatedAt, &w.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("warehouse not found")
	}

	return w, err
}

// storeWarehouse creates a new warehouse
func (s *Server) storeWarehouse(ctx context.Context, form *StoreWarehouseForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return s.storeWarehouseInternal(ctx, tx, form)
	})
}

// storeWarehouseInternal handles the actual warehouse insertion
func (s *Server) storeWarehouseInternal(ctx context.Context, tx *sql.Tx, form *StoreWarehouseForm) error {
	companyID := CurrentCompany(ctx).ID

	// Generate warehouse code using sequence
	seqInfo, err := GetNextSequence(tx, companyID, "warehouse")
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO warehouses (company_id, uuid, code, name, address, description, status, created_at, updated_at)
		 VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, 'enabled'::entity_status, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	address := form.Address
	if address == "" {
		address = ""
	}
	description := form.Description
	if description == "" {
		description = ""
	}

	var warehouseID int
	err = stmt.QueryRowContext(ctx, companyID, seqInfo.Code, form.Name, address, description).Scan(&warehouseID)
	return err
}

// updateWarehouse updates an existing warehouse
func (s *Server) updateWarehouse(ctx context.Context, id int, form *UpdateWarehouseForm) error {
	companyID := CurrentCompany(ctx).ID

	_, err := s.db.ExecContext(
		ctx,
		`UPDATE warehouses 
		 SET name = $1, address = $2, description = $3, updated_at = NOW()
		 WHERE company_id = $4 AND id = $5 AND deleted_at IS NULL`,
		form.Name, form.Address, form.Description, companyID, id,
	)

	return err
}

// deleteWarehouse soft-deletes a warehouse
func (s *Server) deleteWarehouse(ctx context.Context, id int) error {
	companyID := CurrentCompany(ctx).ID

	_, err := s.db.ExecContext(
		ctx,
		`UPDATE warehouses 
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE company_id = $1 AND id = $2`,
		companyID, id,
	)

	return err
}

// toggleWarehouseStatus toggles warehouse between enabled/disabled
func (s *Server) toggleWarehouseStatus(ctx context.Context, id int) error {
	companyID := CurrentCompany(ctx).ID

	// Get current status
	var currentStatus foundation.Status
	err := s.db.QueryRowContext(
		ctx,
		`SELECT status FROM warehouses 
		 WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL`,
		companyID, id,
	).Scan(&currentStatus)

	if err != nil {
		return err
	}

	// Toggle status
	newStatus := foundation.StatusDisabled
	if currentStatus == foundation.StatusDisabled {
		newStatus = foundation.StatusEnabled
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE warehouses 
		 SET status = $1, updated_at = NOW()
		 WHERE company_id = $2 AND id = $3`,
		newStatus, companyID, id,
	)

	return err
}
