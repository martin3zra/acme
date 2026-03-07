package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type item struct {
	ID           int               `json:"id"`
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	Price        float64           `json:"price"`
	Description  string            `json:"description"`
	ItemType     string            `json:"item_type"`
	HasVariants  bool              `json:"has_variants"`
	Identifiers  ItemIdentifiers   `json:"identifiers"`
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
	ID        int    `json:"id"`
	UUID      string `json:"uuid"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type itemVariantSetup struct {
	HasVariants               bool                  `json:"has_variants"`
	AttributeIDs              []int                 `json:"attribute_ids"`
	SelectedValuesByAttribute map[int][]int         `json:"selected_values_by_attribute"`
	ExistingSignatures        []string              `json:"existing_signatures"`
	Variants                  []*itemVariantSummary `json:"variants"`
}

func (s *Server) findItemByID(ctx context.Context, itemID int) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.has_variants, i.identifiers, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
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
		&i.HasVariants,
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

func (s *Server) findItemByUUID(ctx context.Context, itemUUID string) (*item, error) {
	var i item
	err := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.price, i.description, i.tax_id, t.name, t.rate, i.status, "+
		"i.item_type, i.has_variants, i.identifiers, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name  "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id)"+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.uuid = $2 AND i.deleted_at IS NULL", CurrentCompany(ctx).ID, itemUUID).Scan(
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
		&i.HasVariants,
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
		"i.item_type, i.has_variants, i.identifiers, i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
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
			&i.HasVariants,
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
		"i.item_type, i.has_variants, i.identifiers,i.created_at, i.updated_at, i.deleted_at, iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND ("+
		"i.name ILIKE $2 OR "+
		"identifiers->>'reference' ILIKE $2 OR "+
		"identifiers->>'code' ILIKE $2 OR "+
		"identifiers->>'sku' ILIKE $2 OR "+
		"identifiers->>'barcode' ILIKE $2 OR "+
		"identifiers->>'vendor_reference' ILIKE $2 OR "+
		"EXISTS (SELECT 1 FROM items_variants iv WHERE iv.company_id = i.company_id AND iv.item_id = i.id AND iv.deleted_at IS NULL AND (iv.name ILIKE $2 OR iv.sku ILIKE $2))"+
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
			&i.HasVariants,
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

	result := s.db.QueryRow("SELECT i.id, i.uuid, i.name, i.price, i.description, i.item_type, i.has_variants, i.identifiers, i.tax_id, t.name, t.rate, "+
		"iu.unit_id, iu.name as unit_name "+
		"FROM items i "+
		"INNER JOIN taxes t ON(i.company_id = t.company_id AND i.tax_id = t.id) "+
		"LEFT JOIN LATERAL (SELECT iu.unit_id, u.name FROM items_units iu INNER JOIN units u ON (iu.unit_id = u.id) WHERE iu.item_id = i.id limit 1) iu ON true "+
		"WHERE i.company_id = $1 AND i.deleted_at IS NULL AND ("+
		"i.identifiers->>'reference' = $2 OR "+
		"EXISTS (SELECT 1 FROM items_variants iv WHERE iv.company_id = i.company_id AND iv.item_id = i.id AND iv.deleted_at IS NULL AND iv.sku = $2)"+
		")", CurrentCompany(ctx).ID, term)
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
		&i.HasVariants,
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
	stmt, err := tx.Prepare("INSERT INTO items (name, price, description, tax_id, item_type, has_variants, identifiers, company_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
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
		form.HasVariants,
		foundation.ToJSON(form.Identifiers),
		companyID,
	).Scan(&itemID)

	if err != nil {
		return err
	}

	if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
		return err
	}

	if form.ItemType == "product" {
		if form.HasVariants {
			if err = s.storeConfiguredVariants(tx, companyID, itemID, form); err != nil {
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
	return s.addConfiguredVariants(tx, companyID, itemID, form.Name, form.Price, form.AttributeIDs, form.VariantCombos)
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

func (s *Server) addConfiguredVariants(tx *sql.Tx, companyID, itemID int, itemName string, basePrice float64, attributeIDs []int, variantCombos []VariantCombo) error {
	if len(attributeIDs) == 0 {
		return errors.New("at least one attribute is required when variants are enabled")
	}

	if len(variantCombos) == 0 {
		return errors.New("at least one variant combination is required when variants are enabled")
	}

	for idx, attributeID := range attributeIDs {
		if err := s.attachProductAttribute(tx, companyID, itemID, attributeID, idx); err != nil {
			return err
		}
	}

	existingSignatures, err := s.findExistingVariantSignatures(tx, companyID, itemID)
	if err != nil {
		return err
	}

	nextIndex := len(existingSignatures) + 1
	for _, combo := range variantCombos {
		signature := buildVariantSignature(combo.AttributeValueIDs)
		if existingSignatures[signature] {
			continue
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
			sku = fmt.Sprintf("SKU-%s-%d", generateHashCode(itemName, 6), nextIndex)
		}

		variantPrice := combo.Price
		if variantPrice == nil {
			price := basePrice
			variantPrice = &price
		}

		variant := &itemVariant{
			SKU:       sku,
			Name:      variantName,
			IsDefault: len(existingSignatures) == 0 && nextIndex == 1,
			Price:     variantPrice,
			CostPrice: combo.CostPrice,
		}

		if err = s.storeItemVariant(tx, companyID, itemID, variant); err != nil {
			return err
		}

		if err = s.storeVariantAttributeValues(tx, companyID, variant.ID, combo.AttributeValueIDs); err != nil {
			return err
		}

		existingSignatures[signature] = true
		nextIndex++
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
		hasVariants := form.ItemType == "product" && form.HasVariants

		_, err := tx.Exec(
			"UPDATE items SET name = $1, description = $2, price = $3, tax_id = $4, item_type = $5, has_variants = $6, identifiers = $7 WHERE company_id = $8 AND id = $9",
			form.Name, form.Description, form.Price, form.TaxID, form.ItemType, hasVariants, foundation.ToJSON(form.Identifiers), companyID, itemID,
		)

		if err != nil {
			return err
		}

		if err = s.attachItemUnit(tx, companyID, itemID, form.UnitID); err != nil {
			return err
		}

		if form.ItemType == "product" {
			if form.HasVariants {
				if err = s.addConfiguredVariants(tx, companyID, itemID, form.Name, form.Price, form.AttributeIDs, form.VariantCombos); err != nil {
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

func (s *Server) findItemVariantSetup(ctx context.Context, itemID int) (*itemVariantSetup, error) {
	setup := &itemVariantSetup{
		AttributeIDs:              make([]int, 0),
		SelectedValuesByAttribute: make(map[int][]int),
		ExistingSignatures:        make([]string, 0),
		Variants:                  make([]*itemVariantSummary, 0),
	}

	err := s.db.QueryRowContext(
		ctx,
		`SELECT has_variants
		 FROM items
		 WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL`,
		CurrentCompany(ctx).ID, itemID,
	).Scan(&setup.HasVariants)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("item not found")
		}
		return nil, err
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
		`SELECT id, uuid, sku, name, is_default
		 FROM items_variants
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
		if err = variantRows.Scan(&variant.ID, &variant.UUID, &variant.SKU, &variant.Name, &variant.IsDefault); err != nil {
			return nil, err
		}

		setup.Variants = append(setup.Variants, variant)
	}

	if err = variantRows.Err(); err != nil {
		return nil, err
	}

	signatureRows, err := s.db.QueryContext(
		ctx,
		`SELECT COALESCE(string_agg(vav.attribute_id::text || ':' || vav.attribute_value_id::text, '|' ORDER BY vav.attribute_id), '')
		 FROM items_variants iv
		 LEFT JOIN variant_attribute_values vav ON (vav.company_id = iv.company_id AND vav.variant_id = iv.id)
		 WHERE iv.company_id = $1 AND iv.item_id = $2 AND iv.deleted_at IS NULL
		 GROUP BY iv.id`,
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

	if !setup.HasVariants && (len(setup.AttributeIDs) > 0 || len(setup.SelectedValuesByAttribute) > 0 || len(setup.Variants) > 1) {
		setup.HasVariants = true
	}

	return setup, nil
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
