package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type item struct {
	ID           int               `json:"id"`
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	ItemType     string            `json:"item_type"`
	VariantSetup *itemVariantSetup `json:"variant_setup,omitempty"`
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

type itemVariantSummary struct {
	ID                   int         `json:"id"`
	UUID                 string      `json:"uuid"`
	CombinationSignature string      `json:"combination_signature"`
	SKU                  string      `json:"sku"`
	Name                 string      `json:"name"`
	Barcode              *string     `json:"barcode,omitempty"`
	Reference            *string     `json:"reference,omitempty"`
	VendorReference      *string     `json:"vendor_reference,omitempty"`
	Price                *float64    `json:"price"`
	CostPrice            *float64    `json:"cost_price,omitempty"`
	TrackInventory       bool        `json:"track_inventory"`
	StockByWarehouse     map[int]int `json:"stock_by_warehouse,omitempty"`
	IsDefault            bool        `json:"is_default"`
	Active               bool        `json:"active"`
}

type itemVariantSetup struct {
	HasVariants               bool                  `json:"has_variants"`
	AttributeIDs              []int                 `json:"attribute_ids"`
	SelectedValuesByAttribute map[int][]int         `json:"selected_values_by_attribute"`
	ExistingSignatures        []string              `json:"existing_signatures"`
	Variants                  []*itemVariantSummary `json:"variants"`
}

type invoiceItemVariant struct {
	ID          int             `json:"id"`
	VariantID   int             `json:"variant_id"`
	UUID        string          `json:"uuid"`
	Name        string          `json:"name"`
	VariantName string          `json:"variant_name"`
	Description string          `json:"description"`
	SKU         string          `json:"sku"`
	Barcode     *string         `json:"barcode,omitempty"`
	IsDefault   bool            `json:"is_default"`
	Price       float64         `json:"price"`
	StockQty    int64           `json:"stock_qty"`
	Identifier  ItemIdentifiers `json:"identifiers"`
	Tax         tax             `json:"tax"`
	Unit        struct {
		ID   *int    `json:"id"`
		Name *string `json:"name"`
	} `json:"unit"`
}

func (s *Server) findItemByID(ctx context.Context, itemID int) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id)"+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.id = $2 AND i.deleted_at IS NULL", CurrentCompany(ctx).ID, itemID).Scan(
		&i.ID,
		&i.UUID,
		&i.Name,
		&i.Description,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Status,
		&i.ItemType,
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

func (s *Server) findItemByUUID(ctx context.Context, itemUUID string) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id)"+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.uuid = $2 AND i.deleted_at IS NULL", CurrentCompany(ctx).ID, itemUUID).Scan(
		&i.ID,
		&i.UUID,
		&i.Name,
		&i.Description,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Status,
		&i.ItemType,
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

	is, err := s.db.Query("SELECT i.id, i.uuid, i.name, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
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
			&i.Description,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Status,
			&i.ItemType,
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

func (s *Server) findItemsByCriteria(ctx context.Context, term string) ([]*invoiceItemVariant, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the item you're looking for")
	}
	is, err := s.db.Query(`
		SELECT
			iv.id,
			iv.id AS variant_id,
			iv.uuid,
			i.name,
			iv.name AS variant_name,
			i.description,
			iv.sku,
			iv.barcode,
			iv.is_default,
			COALESCE(iv.price, 0),
			CASE WHEN iv.track_inventory THEN COALESCE(SUM(sl.quantity), 0) ELSE 0 END AS stock_qty,
			jsonb_build_object(
				'reference', iv.reference,
				'sku', iv.sku,
				'barcode', iv.barcode,
				'vendor_reference', iv.vendor_reference
			) AS identifiers,
			t.id,
			t.name,
			t.rate,
			iu.unit_id,
			iu.name AS unit_name
		FROM items_variants iv
		INNER JOIN items i ON (iv.company_id = i.company_id AND iv.item_id = i.id)
		INNER JOIN taxes t ON (i.company_id = t.company_id AND i.tax_id = t.id)
		LEFT JOIN LATERAL (
			SELECT iu.unit_id, u.name
			FROM items_units iu
			INNER JOIN units u ON (iu.unit_id = u.id)
			WHERE iu.item_id = i.id
			LIMIT 1
		) iu ON true
		LEFT JOIN stock_levels sl ON (sl.company_id = iv.company_id AND sl.variant_id = iv.id)
		WHERE iv.company_id = $1
			AND iv.deleted_at IS NULL
			AND iv.active = true
			AND i.deleted_at IS NULL
			AND (
				i.name ILIKE $2 OR
				iv.name ILIKE $2 OR
				iv.sku ILIKE $2 OR
				COALESCE(iv.barcode, '') ILIKE $2 OR
				COALESCE(iv.reference, '') ILIKE $2 OR
				COALESCE(iv.vendor_reference, '') ILIKE $2
			)
		GROUP BY iv.id, iv.uuid, i.name, iv.name, i.description, iv.sku, iv.barcode, iv.reference, iv.vendor_reference, iv.is_default,
			iv.price, iv.track_inventory, t.id, t.name, t.rate, iu.unit_id, iu.name
		ORDER BY i.name, iv.is_default DESC, iv.name
		LIMIT 50
	`, CurrentCompany(ctx).ID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer is.Close()

	data := make([]*invoiceItemVariant, 0)
	for is.Next() {
		i := new(invoiceItemVariant)
		if err = is.Scan(
			&i.ID,
			&i.VariantID,
			&i.UUID,
			&i.Name,
			&i.VariantName,
			&i.Description,
			&i.SKU,
			&i.Barcode,
			&i.IsDefault,
			&i.Price,
			&i.StockQty,
			&i.Identifier,
			&i.Tax.ID,
			&i.Tax.Name,
			&i.Tax.Rate,
			&i.Unit.ID,
			&i.Unit.Name,
		); err != nil {
			return nil, err
		}
		data = append(data, i)
	}

	return data, is.Err()
}

func (s *Server) findItemsByReference(ctx context.Context, term string) (*invoiceItemVariant, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the item you're looking for")
	}

	i := new(invoiceItemVariant)
	err := s.db.QueryRow(`
		SELECT
			iv.id,
			iv.id AS variant_id,
			iv.uuid,
			i.name,
			iv.name AS variant_name,
			i.description,
			iv.sku,
			iv.barcode,
			iv.is_default,
			COALESCE(iv.price, 0),
			CASE WHEN iv.track_inventory THEN COALESCE(st.qty, 0) ELSE 0 END AS stock_qty,
			jsonb_build_object(
				'reference', iv.reference,
				'sku', iv.sku,
				'barcode', iv.barcode,
				'vendor_reference', iv.vendor_reference
			) AS identifiers,
			t.id,
			t.name,
			t.rate,
			iu.unit_id,
			iu.name AS unit_name
		FROM items_variants iv
		INNER JOIN items i ON (iv.company_id = i.company_id AND iv.item_id = i.id)
		INNER JOIN taxes t ON (i.company_id = t.company_id AND i.tax_id = t.id)
		LEFT JOIN LATERAL (
			SELECT iu.unit_id, u.name
			FROM items_units iu
			INNER JOIN units u ON (iu.unit_id = u.id)
			WHERE iu.item_id = i.id
			LIMIT 1
		) iu ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(quantity), 0)::bigint AS qty
			FROM stock_levels sl
			WHERE sl.company_id = iv.company_id AND sl.variant_id = iv.id
		) st ON true
		WHERE iv.company_id = $1
			AND iv.deleted_at IS NULL
			AND iv.active = true
			AND i.deleted_at IS NULL
			AND (
				iv.sku = $2 OR
				COALESCE(iv.reference, '') = $2 OR
				COALESCE(iv.barcode, '') = $2
			)
		ORDER BY iv.is_default DESC, iv.id
		LIMIT 1
	`, CurrentCompany(ctx).ID, term).Scan(
		&i.ID,
		&i.VariantID,
		&i.UUID,
		&i.Name,
		&i.VariantName,
		&i.Description,
		&i.SKU,
		&i.Barcode,
		&i.IsDefault,
		&i.Price,
		&i.StockQty,
		&i.Identifier,
		&i.Tax.ID,
		&i.Tax.Name,
		&i.Tax.Rate,
		&i.Unit.ID,
		&i.Unit.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return i, err
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
	stmt, err := tx.Prepare("INSERT INTO items (name, description, tax_id, item_type, company_id) " +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		return err
	}

	var itemID int
	err = stmt.QueryRow(
		&form.Name,
		form.Description,
		form.TaxID,
		form.ItemType,
		companyID,
	).Scan(&itemID)

	if err != nil {
		return err
	}

	if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return err
	}

	if form.ItemType == "product" {
		if len(form.AttributeIDs) > 0 && len(form.VariantCombos) > 0 {
			if err = s.storeConfiguredVariants(tx, companyID, itemID, form); err != nil {
				return err
			}
		} else if len(form.VariantCombos) > 0 {
			if err = s.storeDefaultVariantFromCombo(tx, companyID, itemID, form.Name, form.VariantCombos[0]); err != nil {
				return err
			}
		} else {
			if err = s.storeDefaultVariant(tx, companyID, itemID, form.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) storeConfiguredVariants(tx *sql.Tx, companyID, itemID int, form *StoreItemForm) error {
	return s.addConfiguredVariants(tx, companyID, itemID, form.Name, form.AttributeIDs, form.VariantCombos)
}

func buildVariantSignature(selection map[int]int) string {
	if len(selection) == 0 {
		return ""
	}

	attributeIDs := make([]int, 0, len(selection))
	for attributeID := range selection {
		attributeIDs = append(attributeIDs, attributeID)
	}
	sort.Ints(attributeIDs)

	parts := make([]string, 0, len(attributeIDs))
	for _, attributeID := range attributeIDs {
		parts = append(parts, fmt.Sprintf("%d:%d", attributeID, selection[attributeID]))
	}

	return strings.Join(parts, "|")
}

func (s *Server) findExistingVariantSignatures(tx *sql.Tx, companyID, itemID int) (map[string]bool, error) {
	rows, err := tx.Query(
		`SELECT COALESCE(string_agg(vav.attribute_id::text || ':' || vav.attribute_value_id::text, '|' ORDER BY vav.attribute_id), '')
		 FROM items_variants iv
		 LEFT JOIN variant_attribute_values vav ON (vav.company_id = iv.company_id AND vav.variant_id = iv.id)
		 WHERE iv.company_id = $1 AND iv.item_id = $2 AND iv.deleted_at IS NULL
		 GROUP BY iv.id`,
		companyID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signatures := make(map[string]bool)
	for rows.Next() {
		var signature string
		if err = rows.Scan(&signature); err != nil {
			return nil, err
		}
		signatures[signature] = true
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return signatures, nil
}

// findVariantBySignature finds a variant by its combination signature (including soft-deleted)
func (s *Server) findVariantBySignature(tx *sql.Tx, companyID, itemID int, signature string) (*itemVariant, error) {
	v := new(itemVariant)
	err := tx.QueryRow(
		`SELECT id, uuid, item_id, sku, name, barcode, reference, vendor_reference,
		        combination_signature, is_default, price, cost_price, track_inventory, active, created_at, updated_at, deleted_at
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND combination_signature = $3
		 LIMIT 1`,
		companyID, itemID, signature,
	).Scan(
		&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.Barcode, &v.Reference, &v.VendorReference,
		&v.CombinationSignature, &v.IsDefault, &v.Price, &v.CostPrice, &v.TrackInventory, &v.Active, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return v, err
}

// reviveVariant reactivates a soft-deleted variant
func (s *Server) reviveVariant(tx *sql.Tx, companyID int, variantID int, sku, name string, barcode, reference, vendorRef *string, price, costPrice *float64, trackInventory, active bool) error {
	_, err := tx.Exec(
		`UPDATE items_variants
		 SET deleted_at = NULL,
		     active = $1,
		     sku = $2,
		     name = $3,
		     barcode = $4,
		     reference = $5,
		     vendor_reference = $6,
		     price = $7,
		     cost_price = $8,
		     track_inventory = $9,
		     updated_at = NOW()
		 WHERE company_id = $10 AND id = $11`,
		active, sku, name, barcode, reference, vendorRef, price, costPrice, trackInventory, companyID, variantID,
	)
	return err
}

func (s *Server) reactivateVariant(tx *sql.Tx, companyID int, variantID int) error {
	_, err := tx.Exec(
		`UPDATE items_variants
		 SET active = true,
		     updated_at = NOW()
		 WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL`,
		companyID, variantID,
	)

	return err
}

func (s *Server) updateVariant(tx *sql.Tx, companyID int, variantID int, sku, name string, barcode, reference, vendorRef *string, price, costPrice *float64, trackInventory, active bool) error {
	_, err := tx.Exec(
		`UPDATE items_variants
		 SET sku = $1,
		     name = $2,
		     barcode = $3,
		     reference = $4,
		     vendor_reference = $5,
		     price = $6,
		     cost_price = $7,
		     track_inventory = $8,
		     active = $9,
		     updated_at = NOW()
		 WHERE company_id = $10 AND id = $11 AND deleted_at IS NULL`,
		sku, name, barcode, reference, vendorRef, price, costPrice, trackInventory, active, companyID, variantID,
	)

	return err
}

func (s *Server) ensureUniqueVariantSKU(tx *sql.Tx, companyID int, variantID int, sku string) error {
	if sku == "" {
		return nil
	}

	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM items_variants
		WHERE company_id = $1 AND sku = $2 AND deleted_at IS NULL`
	args := []any{companyID, sku}

	if variantID > 0 {
		query += ` AND id <> $3`
		args = append(args, variantID)
	}

	query += `)`
	if err := tx.QueryRow(query, args...).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("sku already exists")
	}

	return nil
}

func (s *Server) ensureUniqueVariantBarcode(tx *sql.Tx, companyID int, variantID int, barcode string) error {
	if barcode == "" {
		return nil
	}

	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM items_variants
		WHERE company_id = $1 AND barcode = $2 AND deleted_at IS NULL`
	args := []any{companyID, barcode}

	if variantID > 0 {
		query += ` AND id <> $3`
		args = append(args, variantID)
	}

	query += `)`
	if err := tx.QueryRow(query, args...).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("barcode already exists")
	}

	return nil
}

func (s *Server) validateVariantComboUniqueness(tx *sql.Tx, companyID int, variantCombos []VariantCombo) error {
	seenSKUs := make(map[string]bool)
	seenBarcodes := make(map[string]bool)

	for _, combo := range variantCombos {
		sku := strings.TrimSpace(combo.SKU)
		barcode := strings.TrimSpace(combo.Barcode)

		if sku != "" {
			if seenSKUs[sku] {
				return fmt.Errorf("sku must be unique")
			}
			seenSKUs[sku] = true

			if err := s.ensureUniqueVariantSKU(tx, companyID, combo.VariantID, sku); err != nil {
				return err
			}
		}

		if barcode != "" {
			if seenBarcodes[barcode] {
				return fmt.Errorf("barcode must be unique")
			}
			seenBarcodes[barcode] = true

			if err := s.ensureUniqueVariantBarcode(tx, companyID, combo.VariantID, barcode); err != nil {
				return err
			}
		}
	}

	return nil
}

// isVariantReferenced checks if a variant is referenced by invoices or stock
func (s *Server) isVariantReferenced(tx *sql.Tx, companyID, variantID int) (bool, error) {
	var referenced bool
	err := tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM invoices_items WHERE company_id = $1 AND variant_id = $2
		) OR EXISTS(
			SELECT 1 FROM stock_levels WHERE company_id = $1 AND variant_id = $2 AND quantity != 0
		)`,
		companyID, variantID,
	).Scan(&referenced)

	return referenced, err
}

func (s *Server) addConfiguredVariants(tx *sql.Tx, companyID, itemID int, itemName string, attributeIDs []int, variantCombos []VariantCombo) error {
	if len(attributeIDs) == 0 {
		return errors.New("at least one attribute is required when variants are enabled")
	}

	if len(variantCombos) == 0 {
		return errors.New("at least one variant combination is required when variants are enabled")
	}

	// Attach product attributes with proper sort order
	for idx, attributeID := range attributeIDs {
		if err := s.attachProductAttribute(tx, companyID, itemID, attributeID, idx); err != nil {
			return err
		}
	}

	// Remove product attributes that are no longer in the list
	if len(attributeIDs) > 0 {
		placeholders := make([]string, len(attributeIDs))
		args := []interface{}{companyID, itemID}
		for i, attrID := range attributeIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, attrID)
		}
		_, err := tx.Exec(
			fmt.Sprintf(`DELETE FROM product_attributes
				WHERE company_id = $1 AND item_id = $2 AND attribute_id NOT IN (%s)`,
				strings.Join(placeholders, ", ")),
			args...,
		)
		if err != nil {
			return err
		}
	}

	existingSignatures, err := s.findExistingVariantSignatures(tx, companyID, itemID)
	if err != nil {
		return err
	}

	// Track desired signatures for regeneration reconciliation
	desiredSignatures := make(map[string]bool)

	nextIndex := len(existingSignatures) + 1
	for _, combo := range variantCombos {
		signature := buildVariantSignature(combo.AttributeValueIDs)
		desiredSignatures[signature] = true

		// Check if variant already exists (active or soft-deleted)
		existing, err := s.findVariantBySignature(tx, companyID, itemID, signature)
		if err != nil {
			return err
		}

		variantName, err := s.buildVariantNameFromSelection(tx, companyID, attributeIDs, combo.AttributeValueIDs)
		if err != nil {
			return err
		}
		if variantName == "" {
			variantName = fmt.Sprintf("%s %d", itemName, nextIndex)
		}

		sku := strings.TrimSpace(combo.SKU)
		if sku == "" {
			if existing != nil && existing.SKU != "" {
				sku = existing.SKU // Preserve existing SKU
			} else {
				sku = fmt.Sprintf("SKU-%s-%d", generateHashCode(itemName, 6), nextIndex)
			}
		}

		// Price is now required on variants, default to 0 if not provided
		variantPrice := combo.Price
		if variantPrice == nil {
			price := 0.0
			variantPrice = &price
		}

		var barcode, reference, vendorRef *string
		if combo.Barcode != "" {
			barcode = &combo.Barcode
		}
		if combo.Reference != "" {
			reference = &combo.Reference
		}
		if combo.VendorReference != "" {
			vendorRef = &combo.VendorReference
		}

		active := true
		if combo.Active != nil {
			active = *combo.Active
		}

		trackInventory := true
		if combo.TrackInventory != nil {
			trackInventory = *combo.TrackInventory
		}

		if existing != nil {
			// Variant exists - revive if soft-deleted, otherwise update inline editable fields.
			if existing.DeletedAt != nil {
				// Revive soft-deleted variant
				if err = s.reviveVariant(tx, companyID, existing.ID, sku, variantName, barcode, reference, vendorRef, variantPrice, combo.CostPrice, trackInventory, active); err != nil {
					return err
				}
				// Ensure variant attribute values are present
				if err = s.storeVariantAttributeValues(tx, companyID, existing.ID, combo.AttributeValueIDs); err != nil {
					return err
				}
			} else {
				if err = s.updateVariant(tx, companyID, existing.ID, sku, variantName, barcode, reference, vendorRef, variantPrice, combo.CostPrice, trackInventory, active); err != nil {
					return err
				}

				if err = s.storeVariantAttributeValues(tx, companyID, existing.ID, combo.AttributeValueIDs); err != nil {
					return err
				}
			}

			if err = s.syncVariantStockLevels(tx, companyID, existing.ID, trackInventory, combo.StockByWarehouse); err != nil {
				return err
			}

			continue
		}

		// Create new variant
		variant := &itemVariant{
			SKU:                  sku,
			Name:                 variantName,
			Barcode:              barcode,
			Reference:            reference,
			VendorReference:      vendorRef,
			CombinationSignature: signature,
			IsDefault:            len(existingSignatures) == 0 && nextIndex == 1,
			Price:                variantPrice,
			CostPrice:            combo.CostPrice,
			TrackInventory:       trackInventory,
			Active:               active,
		}

		if err = s.storeItemVariant(tx, companyID, itemID, variant); err != nil {
			return err
		}

		if err = s.storeVariantAttributeValues(tx, companyID, variant.ID, combo.AttributeValueIDs); err != nil {
			return err
		}

		if err = s.syncVariantStockLevels(tx, companyID, variant.ID, trackInventory, combo.StockByWarehouse); err != nil {
			return err
		}

		existingSignatures[signature] = true
		nextIndex++
	}

	// Reconcile obsolete variants (soft-delete if not referenced, deactivate if referenced)
	return s.reconcileObsoleteVariants(tx, companyID, itemID, desiredSignatures)
}

// reconcileObsoleteVariants soft-deletes variants not in desired set (unless referenced)
func (s *Server) reconcileObsoleteVariants(tx *sql.Tx, companyID, itemID int, desiredSignatures map[string]bool) error {
	// Find all active variants for the item
	rows, err := tx.Query(
		`SELECT id, combination_signature
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND deleted_at IS NULL`,
		companyID, itemID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type variantInfo struct {
		ID        int
		Signature string
	}
	variants := make([]variantInfo, 0)
	for rows.Next() {
		var v variantInfo
		if err = rows.Scan(&v.ID, &v.Signature); err != nil {
			return err
		}
		variants = append(variants, v)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Process each variant
	for _, v := range variants {
		if desiredSignatures[v.Signature] {
			continue // This variant is still desired
		}

		// Check if variant is referenced
		isReferenced, err := s.isVariantReferenced(tx, companyID, v.ID)
		if err != nil {
			return err
		}

		if isReferenced {
			// Keep variant but deactivate it
			_, err = tx.Exec(
				`UPDATE items_variants
				 SET active = false, updated_at = NOW()
				 WHERE company_id = $1 AND id = $2`,
				companyID, v.ID,
			)
			if err != nil {
				return err
			}
		} else {
			// Soft-delete unreferenced variant
			_, err = tx.Exec(
				`UPDATE items_variants
				 SET deleted_at = NOW(), active = false, updated_at = NOW()
				 WHERE company_id = $1 AND id = $2`,
				companyID, v.ID,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) buildVariantNameFromSelection(tx *sql.Tx, companyID int, attributeIDs []int, selected map[int]int) (string, error) {
	if len(selected) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(attributeIDs))
	for _, attributeID := range attributeIDs {
		attributeValueID, ok := selected[attributeID]
		if !ok {
			continue
		}

		var displayName string
		err := tx.QueryRow(
			`SELECT display_name
			 FROM attribute_values
			 WHERE company_id = $1 AND attribute_id = $2 AND id = $3 AND deleted_at IS NULL`,
			companyID, attributeID, attributeValueID,
		).Scan(&displayName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return "", err
		}

		if strings.TrimSpace(displayName) != "" {
			parts = append(parts, displayName)
		}
	}

	return strings.Join(parts, " / "), nil
}

func (s *Server) storeVariantAttributeValues(tx *sql.Tx, companyID, variantID int, attributeValueIDs map[int]int) error {
	for attributeID, attributeValueID := range attributeValueIDs {
		_, err := tx.Exec(
			`INSERT INTO variant_attribute_values (company_id, variant_id, attribute_id, attribute_value_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (company_id, variant_id, attribute_id) DO NOTHING`,
			companyID, variantID, attributeID, attributeValueID,
		)
		if err != nil {
			return err
		}
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
			"UPDATE items SET name = $1, description = $2, tax_id = $3, item_type = $4 WHERE company_id = $5 AND id = $6",
			form.Name, form.Description, form.TaxID, form.ItemType, companyID, itemID,
		)

		if err != nil {
			return err
		}

		if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
			return err
		}

		if form.ItemType == "product" {
			if len(form.VariantCombos) > 0 {
				if err = s.validateVariantComboUniqueness(tx, companyID, form.VariantCombos); err != nil {
					return err
				}
			}

			if len(form.AttributeIDs) > 0 && len(form.VariantCombos) > 0 {
				if err = s.addConfiguredVariants(tx, companyID, itemID, form.Name, form.AttributeIDs, form.VariantCombos); err != nil {
					return err
				}
			} else if len(form.VariantCombos) > 0 {
				if err = s.ensureDefaultVariantFromCombo(tx, companyID, itemID, form.Name, form.VariantCombos[0]); err != nil {
					return err
				}
			} else {
				if err = s.ensureDefaultVariant(tx, companyID, itemID); err != nil {
					return err
				}
			}
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
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.barcode, iv.reference, iv.vendor_reference,
		        iv.combination_signature, iv.is_default, iv.price, iv.cost_price, iv.track_inventory, iv.active, iv.created_at, iv.updated_at, iv.deleted_at
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
			&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.Barcode, &v.Reference, &v.VendorReference,
			&v.CombinationSignature, &v.IsDefault, &v.Price, &v.CostPrice, &v.TrackInventory, &v.Active, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
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
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.barcode, iv.reference, iv.vendor_reference,
		        iv.combination_signature, iv.is_default, iv.price, iv.cost_price, iv.track_inventory, iv.active, iv.created_at, iv.updated_at, iv.deleted_at
		 FROM items_variants iv
		 WHERE iv.company_id = $1 AND iv.id = $2 AND iv.deleted_at IS NULL`,
		CurrentCompany(ctx).ID, id,
	).Scan(
		&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.Barcode, &v.Reference, &v.VendorReference,
		&v.CombinationSignature, &v.IsDefault, &v.Price, &v.CostPrice, &v.TrackInventory, &v.Active, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
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
		`SELECT iv.id, iv.uuid, iv.item_id, iv.sku, iv.name, iv.barcode, iv.reference, iv.vendor_reference,
		        iv.combination_signature, iv.is_default, iv.price, iv.cost_price, iv.track_inventory, iv.active, iv.created_at, iv.updated_at, iv.deleted_at
		 FROM items_variants iv
		 WHERE iv.sku = $1 AND iv.deleted_at IS NULL`,
		sku,
	).Scan(
		&v.ID, &v.UUID, &v.ItemID, &v.SKU, &v.Name, &v.Barcode, &v.Reference, &v.VendorReference,
		&v.CombinationSignature, &v.IsDefault, &v.Price, &v.CostPrice, &v.TrackInventory, &v.Active, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("variant not found")
	}

	return v, err
}

func (s *Server) findItemVariantSetup(ctx context.Context, itemID int) (*itemVariantSetup, error) {
	setup := &itemVariantSetup{
		AttributeIDs:              make([]int, 0),
		SelectedValuesByAttribute: make(map[int][]int),
		ExistingSignatures:        make([]string, 0),
		Variants:                  make([]*itemVariantSummary, 0),
	}

	attributeRows, err := s.db.QueryContext(
		ctx,
		`SELECT attribute_id
		 FROM product_attributes
		 WHERE company_id = $1 AND item_id = $2
		 ORDER BY sort_order, attribute_id`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer attributeRows.Close()

	for attributeRows.Next() {
		var attributeID int
		if err = attributeRows.Scan(&attributeID); err != nil {
			return nil, err
		}
		setup.AttributeIDs = append(setup.AttributeIDs, attributeID)
	}
	if err = attributeRows.Err(); err != nil {
		return nil, err
	}

	selectionRows, err := s.db.QueryContext(
		ctx,
		`SELECT vav.attribute_id, vav.attribute_value_id
		 FROM variant_attribute_values vav
		 JOIN items_variants iv ON iv.id = vav.variant_id AND iv.company_id = vav.company_id
		 WHERE vav.company_id = $1 AND iv.item_id = $2 AND iv.deleted_at IS NULL
		 ORDER BY vav.attribute_id, vav.attribute_value_id`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer selectionRows.Close()

	seen := make(map[int]map[int]bool)
	for selectionRows.Next() {
		var attributeID, attributeValueID int
		if err = selectionRows.Scan(&attributeID, &attributeValueID); err != nil {
			return nil, err
		}

		if seen[attributeID] == nil {
			seen[attributeID] = make(map[int]bool)
		}
		if seen[attributeID][attributeValueID] {
			continue
		}

		setup.SelectedValuesByAttribute[attributeID] = append(setup.SelectedValuesByAttribute[attributeID], attributeValueID)
		seen[attributeID][attributeValueID] = true
	}
	if err = selectionRows.Err(); err != nil {
		return nil, err
	}

	variantRows, err := s.db.QueryContext(
		ctx,
		`SELECT iv.id, iv.uuid, iv.combination_signature, iv.sku, iv.name, iv.barcode, iv.reference, iv.vendor_reference,
		        iv.price, iv.cost_price, iv.track_inventory, iv.is_default, iv.active,
		        COALESCE((
		          SELECT jsonb_object_agg(sl.warehouse_id::text, sl.quantity)
		          FROM stock_levels sl
		          WHERE sl.company_id = iv.company_id AND sl.variant_id = iv.id
		        ), '{}'::jsonb) AS stock_by_warehouse
		 FROM items_variants iv
		 WHERE company_id = $1 AND item_id = $2 AND deleted_at IS NULL
		 ORDER BY is_default DESC, name`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer variantRows.Close()

	for variantRows.Next() {
		variant := &itemVariantSummary{}
		var rawStockByWarehouse []byte
		if err = variantRows.Scan(
			&variant.ID,
			&variant.UUID,
			&variant.CombinationSignature,
			&variant.SKU,
			&variant.Name,
			&variant.Barcode,
			&variant.Reference,
			&variant.VendorReference,
			&variant.Price,
			&variant.CostPrice,
			&variant.TrackInventory,
			&variant.IsDefault,
			&variant.Active,
			&rawStockByWarehouse,
		); err != nil {
			return nil, err
		}

		variant.StockByWarehouse = parseStockByWarehouse(rawStockByWarehouse)

		setup.Variants = append(setup.Variants, variant)
	}

	if err = variantRows.Err(); err != nil {
		return nil, err
	}

	// Get existing signatures directly from variants table (more efficient)
	signatureRows, err := s.db.QueryContext(
		ctx,
		`SELECT combination_signature
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND deleted_at IS NULL AND combination_signature != ''`,
		CurrentCompany(ctx).ID, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer signatureRows.Close()

	for signatureRows.Next() {
		var signature string
		if err = signatureRows.Scan(&signature); err != nil {
			return nil, err
		}

		setup.ExistingSignatures = append(setup.ExistingSignatures, signature)
	}

	if err = signatureRows.Err(); err != nil {
		return nil, err
	}

	setup.HasVariants = len(setup.Variants) > 1 || len(setup.AttributeIDs) > 0

	return setup, nil
}

func parseStockByWarehouse(raw []byte) map[int]int {
	if len(raw) == 0 {
		return map[int]int{}
	}

	parsed := make(map[string]int)
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return map[int]int{}
	}

	stockByWarehouse := make(map[int]int, len(parsed))
	for warehouseID, quantity := range parsed {
		id, err := strconv.Atoi(warehouseID)
		if err != nil {
			continue
		}
		stockByWarehouse[id] = quantity
	}

	return stockByWarehouse
}

// storeItemVariant creates a new item variant
func (s *Server) storeItemVariant(tx *sql.Tx, companyID, itemID int, variant *itemVariant) error {
	stmt, err := tx.Prepare(
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, barcode, reference, vendor_reference, combination_signature, is_default, price, cost_price, track_inventory, active, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(companyID, itemID, variant.SKU, variant.Name, variant.Barcode, variant.Reference, variant.VendorReference, variant.CombinationSignature, variant.IsDefault, variant.Price, variant.CostPrice, variant.TrackInventory, variant.Active).Scan(&variant.ID)
}

// storeDefaultVariant creates the default variant for an item
func (s *Server) storeDefaultVariant(tx *sql.Tx, companyID, itemID int, itemName string) error {
	stmt, err := tx.Prepare(
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, combination_signature, is_default, price, track_inventory, active, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), (SELECT 'SKU-' || gen_random_uuid()::text), $3, '', true, 0.00, true, true, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var variantID int
	return stmt.QueryRow(companyID, itemID, "Default").Scan(&variantID)
}

func (s *Server) storeDefaultVariantFromCombo(tx *sql.Tx, companyID, itemID int, itemName string, combo VariantCombo) error {
	sku := strings.TrimSpace(combo.SKU)
	if sku == "" {
		sku = fmt.Sprintf("SKU-%s", generateHashCode(itemName, 8))
	}

	price := 0.0
	if combo.Price != nil {
		price = *combo.Price
	}

	var costPrice *float64
	if combo.CostPrice != nil {
		costPrice = combo.CostPrice
	}

	trackInventory := true
	if combo.TrackInventory != nil {
		trackInventory = *combo.TrackInventory
	}

	var barcode, reference, vendorRef *string
	if combo.Barcode != "" {
		barcode = &combo.Barcode
	}
	if combo.Reference != "" {
		reference = &combo.Reference
	}
	if combo.VendorReference != "" {
		vendorRef = &combo.VendorReference
	}

	variant := &itemVariant{
		SKU:             sku,
		Name:            "Default",
		Barcode:         barcode,
		Reference:       reference,
		VendorReference: vendorRef,
		IsDefault:       true,
		Price:           &price,
		CostPrice:       costPrice,
		TrackInventory:  trackInventory,
		Active:          true,
	}

	if err := s.storeItemVariant(tx, companyID, itemID, variant); err != nil {
		return err
	}

	return s.syncVariantStockLevels(tx, companyID, variant.ID, trackInventory, combo.StockByWarehouse)
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

func (s *Server) ensureDefaultVariantFromCombo(tx *sql.Tx, companyID, itemID int, itemName string, combo VariantCombo) error {
	var variantID int
	err := tx.QueryRow(
		`SELECT id
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND is_default = true AND deleted_at IS NULL
		 LIMIT 1`,
		companyID, itemID,
	).Scan(&variantID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.storeDefaultVariantFromCombo(tx, companyID, itemID, itemName, combo)
		}
		return err
	}

	price := 0.0
	if combo.Price != nil {
		price = *combo.Price
	}

	var costPrice *float64
	if combo.CostPrice != nil {
		costPrice = combo.CostPrice
	}

	trackInventory := true
	if combo.TrackInventory != nil {
		trackInventory = *combo.TrackInventory
	}

	_, err = tx.Exec(
		`UPDATE items_variants
		 SET sku = COALESCE(NULLIF($1, ''), sku),
		     barcode = NULLIF($2, ''),
		     reference = NULLIF($3, ''),
		     vendor_reference = NULLIF($4, ''),
		     price = $5,
		     cost_price = $6,
		     track_inventory = $7,
		     active = true,
		     updated_at = NOW()
		 WHERE company_id = $8 AND id = $9`,
		combo.SKU, combo.Barcode, combo.Reference, combo.VendorReference, price, costPrice, trackInventory, companyID, variantID,
	)

	if err != nil {
		return err
	}

	return s.syncVariantStockLevels(tx, companyID, variantID, trackInventory, combo.StockByWarehouse)
}

func (s *Server) ensureWarehouseBelongsToCompany(tx *sql.Tx, companyID, warehouseID int) error {
	var exists bool
	err := tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM warehouses
			WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL
		)`,
		companyID, warehouseID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("warehouse %d does not belong to company", warehouseID)
	}

	return nil
}

func (s *Server) syncVariantStockLevels(tx *sql.Tx, companyID, variantID int, trackInventory bool, stockByWarehouse map[int]int) error {
	if !trackInventory {
		_, err := tx.Exec(
			`DELETE FROM stock_levels WHERE company_id = $1 AND variant_id = $2`,
			companyID, variantID,
		)
		return err
	}

	if stockByWarehouse == nil {
		return nil
	}

	warehouseIDs := make([]int, 0, len(stockByWarehouse))
	for warehouseID, quantity := range stockByWarehouse {
		if quantity < 0 {
			return fmt.Errorf("stock quantity cannot be negative")
		}

		if err := s.ensureWarehouseBelongsToCompany(tx, companyID, warehouseID); err != nil {
			return err
		}

		_, err := tx.Exec(
			`INSERT INTO stock_levels (company_id, warehouse_id, variant_id, quantity, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (company_id, warehouse_id, variant_id)
			 DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()`,
			companyID, warehouseID, variantID, quantity,
		)
		if err != nil {
			return err
		}

		warehouseIDs = append(warehouseIDs, warehouseID)
	}

	if len(warehouseIDs) == 0 {
		_, err := tx.Exec(
			`DELETE FROM stock_levels WHERE company_id = $1 AND variant_id = $2`,
			companyID, variantID,
		)
		return err
	}

	sort.Ints(warehouseIDs)
	placeholders := make([]string, len(warehouseIDs))
	args := make([]any, 0, len(warehouseIDs)+2)
	args = append(args, companyID, variantID)
	for idx, warehouseID := range warehouseIDs {
		placeholders[idx] = fmt.Sprintf("$%d", idx+3)
		args = append(args, warehouseID)
	}

	_, err := tx.Exec(
		fmt.Sprintf(
			`DELETE FROM stock_levels
			 WHERE company_id = $1
			   AND variant_id = $2
			   AND warehouse_id NOT IN (%s)`,
			strings.Join(placeholders, ", "),
		),
		args...,
	)

	return err
}

// attachProductAttribute links an attribute to an item (product)
func (s *Server) attachProductAttribute(tx *sql.Tx, companyID, itemID, attributeID, sortOrder int) error {
	stmt, err := tx.Prepare(
		`INSERT INTO product_attributes (company_id, item_id, attribute_id, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (company_id, item_id, attribute_id) 
		 DO UPDATE SET sort_order = EXCLUDED.sort_order, updated_at = NOW()`,
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
