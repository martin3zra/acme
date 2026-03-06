package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type item struct {
	ID          int             `json:"id"`
	UUID        string          `json:"uuid"`
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
	is, err := s.db.Query("SELECT i.id, i.uuid, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.identifiers,i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND ("+
		"i.name ILIKE $2 OR "+
		"identifiers->>'reference' ILIKE $2 OR "+
		"identifiers->>'code' ILIKE $2 OR "+
		"identifiers->>'sku' ILIKE $2 OR "+
		"identifiers->>'barcode' ILIKE $2 OR "+
		"identifiers->>'vendor_reference' ILIKE $2"+
		") "+
		"AND i.deleted_at IS NULL ORDER BY i.name",
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

	result := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.price, i.description, i.item_type, i.identifiers, i.tax_id, t.name, t.rate, "+
		"iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.identifiers->>'reference' = $2 AND i.deleted_at IS NULL", CurrentCompany(ctx).ID, term)
	if result.Err() != nil {
		return nil, result.Err()
	}

	i := new(item)
	if err := result.Scan(
		&i.ID,
		&i.UUID,
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

	return nil
}

func (s *Server) attachItemUnit(tx *sql.Tx, companyID, itemID, unitID int) error {
	_, err := tx.Exec("INSERT INTO items_units (company_id, item_id, unit_id) VALUES($1, $2, $3) "+
		"ON CONFLICT (id) DO UPDATE SET updated_at = now()", companyID, itemID, unitID)
	return err
}

func (s *Server) updateItem(ctx context.Context, itemID int, form *UpdateItemForm) error {
	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {

		_, err := tx.Exec(
			"UPDATE items SET name = $1, description = $2, price = $3, tax_id = $4, item_type = $5, identifiers = $6 WHERE company_id = $7 AND id = $8",
			form.Name, form.Description, form.Price, form.TaxID, form.ItemType, foundation.ToJSON(form.Identifiers), companyID, itemID,
		)

		if err != nil {
			return err
		}

		if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
			return err
		}

		return nil
	})
}

func (s *Server) deleteItem(ctx context.Context, itemID int) error {

	_, err := s.db.Exec(
		"UPDATE items SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, itemID,
	)

	return err
}

func (s *Server) toggleItemStatus(ctx context.Context, item *item) error {
	status := item.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	_, err := s.db.Exec(
		"UPDATE items SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, item.ID, status,
	)
	return err
}

// ============================================================================
// Item Variant Management Methods
// ============================================================================

// findItemVariants returns all active variants for an item
func (s *Server) findItemVariants(ctx context.Context, itemID int) ([]*itemVariant, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.is_default, iv.price, iv.cost_price,
		        iv.created_at, iv.updated_at, iv.deleted_at
		 FROM items_variants iv
		 WHERE iv.company_id = $1 AND iv.item_id = $2 AND iv.deleted_at IS NULL
		 ORDER BY iv.is_default DESC, iv.name`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*itemVariant, 0)
	for rows.Next() {
		v := new(itemVariant)
		if err = rows.Scan(
			&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.IsDefault, &v.Price, &v.CostPrice,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		); err != nil {
			return data, err
		}
		data = append(data, v)
	}

	return data, rows.Err()
}

// findVariantByID returns a single variant by ID
func (s *Server) findVariantByID(ctx context.Context, id int) (*itemVariant, error) {
	v := new(itemVariant)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.is_default, iv.price, iv.cost_price,
		        iv.created_at, iv.updated_at, iv.deleted_at
		 FROM items_variants iv
		 WHERE iv.company_id = $1 AND iv.id = $2 AND iv.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.IsDefault, &v.Price, &v.CostPrice,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("variant not found")
	}

	return v, err
}

// findVariantBySKU returns a variant by SKU (globally unique)
func (s *Server) findVariantBySKU(ctx context.Context, sku string) (*itemVariant, error) {
	v := new(itemVariant)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.is_default, iv.price, iv.cost_price,
		        iv.created_at, iv.updated_at, iv.deleted_at
		 FROM items_variants iv
		 WHERE iv.sku = $1 AND iv.deleted_at IS NULL`,
		sku,
	).Scan(
		&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.IsDefault, &v.Price, &v.CostPrice,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("variant not found")
	}

	return v, err
}

// storeItemVariant creates a new item variant
func (s *Server) storeItemVariant(tx *sql.Tx, companyID, itemID int, variant *itemVariant) error {
	stmt, err := tx.Prepare(
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, is_default, price, cost_price, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(companyID, itemID, variant.SKU, variant.Name, variant.IsDefault, variant.Price, variant.CostPrice).Scan(&variant.ID)
}

// storeDefaultVariant creates the default variant for an item
func (s *Server) storeDefaultVariant(tx *sql.Tx, companyID, itemID int, itemName string) error {
	stmt, err := tx.Prepare(
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, is_default, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), (SELECT 'SKU-' || gen_random_uuid()::text), $3, true, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var variantID int
	return stmt.QueryRow(companyID, itemID, "Default").Scan(&variantID)
}

// ensureDefaultVariant checks if a default variant exists, creates one if not
func (s *Server) ensureDefaultVariant(tx *sql.Tx, companyID, itemID int) error {
	var exists bool
	err := tx.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM items_variants WHERE company_id = $1 AND item_id = $2 AND is_default = true AND deleted_at IS NULL)`,
		companyID, itemID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return s.storeDefaultVariant(tx, companyID, itemID, "Default")
	}

	return nil
}

// attachProductAttribute links an attribute to an item (product)
func (s *Server) attachProductAttribute(tx *sql.Tx, companyID, itemID, attributeID, sortOrder int) error {
	stmt, err := tx.Prepare(
		`INSERT INTO product_attributes (company_id, item_id, attribute_id, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (company_id, item_id, attribute_id) DO NOTHING`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(companyID, itemID, attributeID, sortOrder)
	return err
}

// findProductAttributes returns all attributes linked to an item
func (s *Server) findProductAttributes(ctx context.Context, itemID int) ([]*productAttribute, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT pa.id, pa.item_id, pa.attribute_id, pa.sort_order,
		        a.id, a.uuid, a.name, a.type, a.display_name, a.description, a.created_at, a.updated_at, a.deleted_at
		 FROM product_attributes pa
		 JOIN attributes a ON pa.attribute_id = a.id
		 WHERE pa.company_id = $1 AND pa.item_id = $2
		 ORDER BY pa.sort_order, a.name`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]*productAttribute, 0)
	for rows.Next() {
		pa := new(productAttribute)
		pa.Attribute = new(attribute)
		if err = rows.Scan(
			&pa.ID, &pa.ItemID, &pa.AttributeID, &pa.SortOrder,
			&pa.Attribute.ID, &pa.Attribute.UUID, &pa.Attribute.Name, &pa.Attribute.Type,
			&pa.Attribute.DisplayName, &pa.Attribute.Description,
			&pa.Attribute.CreatedAt, &pa.Attribute.UpdatedAt, &pa.Attribute.DeletedAt,
		); err != nil {
			return data, err
		}
		data = append(data, pa)
	}

	return data, rows.Err()
}
