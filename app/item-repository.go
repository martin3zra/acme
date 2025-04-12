package app

import (
	"database/sql"
	"log"

	"github.com/martin3zra/acme/pkg/foundation"
)

type item struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	// Units       []*UnitResponse `json:"units"`
	Tax  tax `json:"tax"`
	Unit struct {
		ID   *int    `json:"id"`
		Name *string `json:"name"`
	} `json:"unit"`
	Status foundation.Status `json:"status"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findItemByID(companyID, itemID int) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id)"+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.id = $2 AND i.deleted_at IS NULL", companyID, itemID).Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Description,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Unit.ID,
		&i.Unit.Name,
	)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *Server) findItems(companyID int) ([]*item, error) {

	is, err := s.db.Query("SELECT i.id, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.deleted_at IS NULL", companyID)
	if err != nil {
		return nil, err
	}
	data := make([]*item, 0)
	for is.Next() {
		i := new(item)
		if err = is.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.Description,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.Unit.ID,
			&i.Unit.Name,
		); err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	return data, nil
}

func (s *Server) storeItem(companyID int, form StoreItemForm) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO items (name, price, description, tax_id, company_id) " +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		defer stmt.Close()
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error inserting new item: %v", txErr)
			return txErr
		}

		return err
	}

	var itemID int
	err = stmt.QueryRow(
		&form.Name,
		form.Price,
		form.Description,
		form.TaxID,
		companyID,
	).Scan(&itemID)

	if err != nil {
		return err
	}

	if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Server) attachItemUnit(tx *sql.Tx, companyID, itemID, unitID int) error {
	_, err := tx.Exec("INSERT INTO items_units (company_id, item_id, unit_id) VALUES($1, $2, $3) "+
		"ON CONFLICT (id) DO UPDATE SET updated_at = now()", companyID, itemID, unitID)

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error attaching new item unit: %v", txErr)
			return txErr
		}
	}
	return err
}

func (s *Server) updateItem(companyID, itemID int, form StoreItemForm) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE items SET name = $1, description = $2,  price = $3, tax_id = $4 WHERE company_id = $5 AND id = $6",
		form.Name, form.Description, form.Price, form.TaxID, companyID, itemID,
	)

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error attaching new item unit: %v", txErr)
			return txErr
		}
	}

	if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Server) deleteItem(companyID, itemID int) error {

	_, err := s.db.Exec(
		"UPDATE items SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		companyID, itemID,
	)

	return err
}

func (s *Server) toggleItemStatus(companyID int, item *item) error {
	status := item.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	_, err := s.db.Exec(
		"UPDATE items SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		companyID, item.ID, status,
	)
	return err
}
