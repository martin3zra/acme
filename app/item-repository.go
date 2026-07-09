package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
)

type item struct {
	ID          int             `json:"id"`
	UUID        string          `json:"uuid"`
	VariantID   int             `json:"variant_id"`
	SKU         string          `json:"sku,omitempty"`
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	ItemType    string          `json:"item_type"`
	Identifiers ItemIdentifiers `json:"identifiers"`
	// Units       []*UnitResponse `json:"units"`
	Tax  tax `json:"tax"`
	Unit struct {
		ID   *int    `json:"id"`
		Name *string `json:"name"`
	} `json:"unit"`
	Status foundation.Status `json:"status,omitempty"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findItemByID(ctx context.Context, itemID int) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.identifiers, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id)"+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.id = $2 AND i.deleted_at IS NULL", CurrentCompany(ctx).ID, itemID).Scan(
		&i.ID,
		&i.UUID,
		&i.Name,
		&i.Price,
		&i.Description,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Status,
		&i.ItemType,
		&i.Identifiers,
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

func (s *Server) findItems(ctx context.Context, itemType ItemType) ([]*item, error) {

	is, err := s.db.Query("SELECT i.id, i.uuid, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.identifiers, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.deleted_at IS NULL  AND ($2 = 'all' OR i.item_type = $2::item_type) ORDER BY i.name", CurrentCompany(ctx).ID, itemType)
	if err != nil {
		return nil, err
	}
	data := make([]*item, 0)
	for is.Next() {
		i := new(item)
		if err = is.Scan(
			&i.ID,
			&i.UUID,
			&i.Name,
			&i.Price,
			&i.Description,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Status,
			&i.ItemType,
			&i.Identifiers,
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

func (s *Server) findItemsByCriteria(ctx context.Context, term string) ([]*item, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the item you're looking for")
	}
	// One row per sellable variant: a variant item expands to a row per variant
	// ("Item — Variant"), a plain item to a single default-variant row. The term
	// matches item fields or the variant's own sku/reference/barcode/name, so a
	// search can land directly on a specific variant.
	is, err := s.db.Query("SELECT i.id, i.uuid, iv.id as variant_id, COALESCE(iv.sku, ''), "+
		"CASE WHEN iv.is_default THEN i.name ELSE i.name || ' — ' || iv.name END as name, "+
		"iv.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.identifiers,i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN items_variants iv ON (iv.item_id = i.id AND iv.company_id = i.company_id AND iv.deleted_at IS NULL) "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND ("+
		"i.name ILIKE $2 OR "+
		"identifiers->>'reference' ILIKE $2 OR "+
		"identifiers->>'code' ILIKE $2 OR "+
		"identifiers->>'sku' ILIKE $2 OR "+
		"identifiers->>'barcode' ILIKE $2 OR "+
		"identifiers->>'vendor_reference' ILIKE $2 OR "+
		"iv.name ILIKE $2 OR iv.sku ILIKE $2 OR iv.reference ILIKE $2 OR iv.barcode ILIKE $2"+
		") "+
		"AND i.deleted_at IS NULL ORDER BY i.name, iv.is_default DESC, iv.name",
		CurrentCompany(ctx).ID, "%"+term+"%",
	)
	if err != nil {
		return nil, err
	}
	data := make([]*item, 0)
	for is.Next() {
		i := new(item)
		if err = is.Scan(
			&i.ID,
			&i.UUID,
			&i.VariantID,
			&i.SKU,
			&i.Name,
			&i.Price,
			&i.Description,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Status,
			&i.ItemType,
			&i.Identifiers,
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

func (s *Server) findItemsByReference(ctx context.Context, term string) (*item, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the item you're looking for")
	}

	// Exact match resolves to one variant: a variant's own sku/reference/barcode
	// selects that variant directly; an item-level reference falls to the item's
	// default variant (is_default first).
	result := s.db.QueryRow("SELECT i.id, i.uuid, iv.id as variant_id, COALESCE(iv.sku, ''), "+
		"CASE WHEN iv.is_default THEN i.name ELSE i.name || ' — ' || iv.name END as name, "+
		"iv.price, i.description, i.item_type, i.identifiers, i.tax_id, t.name, t.rate, "+
		"iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN items_variants iv ON (iv.item_id = i.id AND iv.company_id = i.company_id AND iv.deleted_at IS NULL) "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND ("+
		"i.identifiers->>'reference' = $2 OR iv.sku = $2 OR iv.reference = $2 OR iv.barcode = $2"+
		") AND i.deleted_at IS NULL "+
		"ORDER BY (iv.sku = $2 OR iv.reference = $2 OR iv.barcode = $2) DESC, iv.is_default DESC LIMIT 1",
		CurrentCompany(ctx).ID, term)
	if result.Err() != nil {
		return nil, result.Err()
	}

	i := new(item)
	if err := result.Scan(
		&i.ID,
		&i.UUID,
		&i.VariantID,
		&i.SKU,
		&i.Name,
		&i.Price,
		&i.Description,
		&i.ItemType,
		&i.Identifiers,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Unit.ID,
		&i.Unit.Name,
	); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Server) storeItem(ctx context.Context, form *StoreItemForm) error {
	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return s.storeItemInternal(tx, companyID, form)
	})
}

func (s *Server) storeItemBackground(tx *sql.Tx, companyID int, form *StoreItemForm) error {
	return s.storeItemInternal(tx, companyID, form)
}

func (s *Server) storeItemInternal(tx *sql.Tx, companyID int, form *StoreItemForm) error {
	stmt, err := tx.Prepare("INSERT INTO items (name, price, description, tax_id, item_type, identifiers, company_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id")
	if err != nil {
		return err
	}

	var itemID int
	err = stmt.QueryRow(
		&form.Name,
		form.Price,
		form.Description,
		form.TaxID,
		form.ItemType,
		foundation.ToJSON(form.Identifiers),
		companyID,
	).Scan(&itemID)

	if err != nil {
		return err
	}

	if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return err
	}

	// Every item gets a default, inventory-tracked variant. Without it, the
	// item has nothing to stock against and inventory recording silently no-ops.
	return s.storeDefaultVariant(tx, companyID, itemID, form.Name, form.Price)
}

// storeDefaultVariant creates the single default variant for an item (no
// attributes). combination_signature 'default' + is_default=true.
func (s *Server) storeDefaultVariant(tx *sql.Tx, companyID, itemID int, name string, price float64) error {
	_, err := tx.Exec(
		`INSERT INTO items_variants
		    (company_id, item_id, name, sku, combination_signature, is_default, price, cost_price, track_inventory)
		 VALUES ($1, $2, $3, 'SKU-' || substr(gen_random_uuid()::text, 1, 8), 'default', true, $4, 0, true)`,
		companyID, itemID, name, price,
	)
	return err
}

// attachItemUnit sets the item's unit, creating the link row or replacing the unit
// on the existing one.
//
// The statement it replaced conflicted on `id`, the serial primary key, which the
// insert never supplies — so the ON CONFLICT branch could never fire and there was
// no unique index for it to target either. Every call appended a row, and because
// both item reads pick the unit through `LEFT JOIN LATERAL (... LIMIT 1)` with no
// ORDER BY, editing an item's unit silently kept the old one. The migration in
// 20260709120000 dedupes the table and adds the (company_id, item_id) constraint
// this now conflicts on.
func (s *Server) attachItemUnit(tx *sql.Tx, companyID, itemID, unitID int) error {
	_, err := tx.Exec(
		`INSERT INTO items_units (company_id, item_id, unit_id) VALUES ($1, $2, $3)
		 ON CONFLICT (company_id, item_id)
		 DO UPDATE SET unit_id = EXCLUDED.unit_id, updated_at = now()`,
		companyID, itemID, unitID,
	)
	return err
}

// insertVariantItem inserts the item row + unit for a variant-bearing item,
// deliberately WITHOUT a default variant (the variant matrix supplies the
// variants). It mirrors storeItemInternal's insert on purpose so the plain-item
// path stays untouched.
func (s *Server) insertVariantItem(tx *sql.Tx, companyID int, form *StoreItemWithAttributesForm) (int, error) {
	var itemID int
	err := tx.QueryRow(
		"INSERT INTO items (name, price, description, tax_id, item_type, identifiers, company_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		form.Name, form.Price, form.Description, form.TaxID, form.ItemType, foundation.ToJSON(form.Identifiers), companyID,
	).Scan(&itemID)
	if err != nil {
		return 0, err
	}
	if err := s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return 0, err
	}
	return itemID, nil
}

// storeItemWithVariants creates an item together with its attribute/variant
// matrix in one transaction. Unlike storeItem it does not create a default
// variant when combinations are supplied — the matrix owns the variants.
func (s *Server) storeItemWithVariants(ctx context.Context, form *StoreItemWithAttributesForm) error {
	companyID := CurrentCompany(ctx).ID
	svc := NewCreateProductWithVariantsService(s.db)
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		itemID, err := s.insertVariantItem(tx, companyID, form)
		if err != nil {
			return err
		}
		return svc.run(ctx, tx, companyID, form, itemID)
	})
}

func (s *Server) updateItem(ctx context.Context, itemID int, form *UpdateItemForm) error {
	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {

		res, err := tx.Exec(
			"UPDATE items SET name = $1, description = $2, price = $3, tax_id = $4, item_type = $5, identifiers = $6 WHERE company_id = $7 AND id = $8",
			form.Name, form.Description, form.Price, form.TaxID, form.ItemType, foundation.ToJSON(form.Identifiers), companyID, itemID,
		)

		if err := mustAffectRow(res, err, "item"); err != nil {
			return err
		}

		if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
			return err
		}

		return nil
	})
}

func (s *Server) deleteItem(ctx context.Context, itemID int) error {

	res, err := s.db.Exec(
		"UPDATE items SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, itemID,
	)

	return mustAffectRow(res, err, "item")
}

func (s *Server) toggleItemStatus(ctx context.Context, item *item) error {
	status := item.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	res, err := s.db.Exec(
		"UPDATE items SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, item.ID, status,
	)
	return mustAffectRow(res, err, "item")
}
