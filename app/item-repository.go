package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
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

// findItemByID reads one item with its tax and its unit.
//
// The unit came through `LEFT JOIN LATERAL (... LIMIT 1)`. That lateral existed
// because items_units could hold several rows per item — the bug migration
// 20260709120000 fixed. With `items_units_company_item_unique (company_id, item_id)`
// in place there is at most one link row, so the lateral is a plain hasOne and the
// LIMIT 1 has nothing left to disambiguate. An item with no unit still comes back,
// with item.Unit's fields nil, as the outer join produced.
//
// INNER JOIN taxes becomes a constrained belongsTo. It never filtered the tax's
// deleted_at and taxRead carries no softdelete tag, so a retired tax still resolves.
func (s *Server) findItemByID(ctx context.Context, itemID int) (*item, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row itemRead
	if err := pdb.Model(&itemRead{}).
		WithConstraint("Tax", withItemTax).
		With("ItemUnit.Unit").
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", itemID).
		First(ctx, &row); err != nil {
		return nil, err
	}

	return row.toItem(), nil
}

// findItems lists the company's items, ordered by name.
//
// `($2 = 'all' OR i.item_type = $2::item_type)` was a filter that disables itself for
// the sentinel value; Unless says that directly. The relations are loaded as in
// findItemByID.
func (s *Server) findItems(ctx context.Context, itemType ItemType) ([]*item, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []itemRead
	if err := pdb.Model(&itemRead{}).
		WithConstraint("Tax", withItemTax).
		With("ItemUnit.Unit").
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Unless(itemType == ItemTypeAll, func(q *playsql.Builder) {
			q.WhereEq("item_type", string(itemType))
		}).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*item, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toItem())
	}
	return data, nil
}

// findItemsByCriteria stays on raw database/sql: it expands one row per sellable
// variant with a CASE-built name, ORs an ILIKE across jsonb keys and joined variant
// columns, and orders by a boolean expression. None of that is expressible as a
// model read, and WhereRaw takes no bind parameters.
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

// findItemsByReference stays raw for the same reasons as findItemsByCriteria: the
// CASE name, the exact-match OR group across item and variant columns, and the
// `ORDER BY (iv.sku = $2 OR ...) DESC` that ranks a variant hit above the default.
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
	itemID, err := insertItemRow(tx, companyID, form.Name, form.Price, form.Description,
		form.TaxID, form.ItemType, form.Identifiers)
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

// insertItemRow inserts the items row and returns its id. storeItemInternal and
// insertVariantItem ran byte-for-byte identical INSERTs; they share this one now.
//
// The old plain-item path prepared the statement for a single execution, which costs
// an extra round trip and leaks the *sql.Stmt — it was never closed.
func insertItemRow(tx *sql.Tx, companyID int, name string, price float64, description string,
	taxID int, itemType string, identifiers ItemIdentifiers) (int, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return 0, err
	}

	id, err := ptx.Model(&itemRead{}).Insert(context.Background(), map[string]any{
		"company_id":  companyID,
		"name":        name,
		"price":       price,
		"description": description,
		"tax_id":      taxID,
		"item_type":   itemType,
		"identifiers": foundation.ToJSON(identifiers),
	})
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// storeDefaultVariant creates the single default variant for an item (no
// attributes). combination_signature 'default' + is_default=true.
//
// Stays raw: the sku is a SQL expression over gen_random_uuid(), which an Insert
// map cannot carry as a value.
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
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	_, err = ptx.Model(&itemUnitRead{}).
		Upsert(context.Background(),
			[]map[string]any{{"company_id": companyID, "item_id": itemID, "unit_id": unitID}},
			[]string{"company_id", "item_id"},
			[]string{"unit_id", "updated_at"},
		)
	return err
}

// insertVariantItem inserts the item row + unit for a variant-bearing item,
// deliberately WITHOUT a default variant (the variant matrix supplies the
// variants). It mirrors storeItemInternal's insert on purpose so the plain-item
// path stays untouched.
// insertVariantItem inserts the item row + unit for a variant-bearing item,
// deliberately WITHOUT a default variant (the variant matrix supplies the
// variants).
func (s *Server) insertVariantItem(tx *sql.Tx, companyID int, form *StoreItemWithAttributesForm) (int, error) {
	itemID, err := insertItemRow(tx, companyID, form.Name, form.Price, form.Description,
		form.TaxID, form.ItemType, form.Identifiers)
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

// updateItem now inherits `deleted_at IS NULL` from itemRead's softdelete tag, which
// the raw UPDATE lacked. Editing a soft-deleted item is a not-found.
//
// It also stamps items.updated_at, which the raw statement never set -- Builder.Update
// stamps the column because itemRead maps it, and the item reads need it mapped to
// populate the response's timestamps.
func (s *Server) updateItem(ctx context.Context, itemID int, form *UpdateItemForm) error {
	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		affected, err := ptx.Model(&itemRead{}).
			WhereEq("company_id", companyID).
			WhereEq("id", itemID).
			Update(ctx, map[string]any{
				"name":        form.Name,
				"description": form.Description,
				"price":       form.Price,
				"tax_id":      form.TaxID,
				"item_type":   form.ItemType,
				"identifiers": foundation.ToJSON(form.Identifiers),
			})
		if err := mustAffectRows(affected, err, "item"); err != nil {
			return err
		}

		return s.attachItemUnit(tx, companyID, itemID, form.UnitID)
	})
}

func (s *Server) deleteItem(ctx context.Context, itemID int) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&itemRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", itemID).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return mustAffectRows(affected, err, "item")
}

func (s *Server) toggleItemStatus(ctx context.Context, item *item) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	status := item.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}

	affected, err := pdb.Model(&itemRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", item.ID).
		Update(ctx, map[string]any{"status": string(status)})
	return mustAffectRows(affected, err, "item")
}
