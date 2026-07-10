package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/martin3zra/forge/database"
)

// CreateProductWithVariantsService creates a product's variants from the
// attribute-value combinations requested on item creation. Ported from
// feature/inventory-management and adapted to main's items_variants schema
// (price/cost_price are NOT NULL here, so nil combo prices coalesce to 0).
type CreateProductWithVariantsService struct {
	db *sql.DB
}

// NewCreateProductWithVariantsService creates a new service instance.
func NewCreateProductWithVariantsService(db *sql.DB) *CreateProductWithVariantsService {
	return &CreateProductWithVariantsService{db: db}
}

// Execute creates the variants for an already-persisted item in its own
// transaction. Callers that also create the item row should use run so item and
// variants commit together.
func (s *CreateProductWithVariantsService) Execute(ctx context.Context, form *StoreItemWithAttributesForm, itemID int) error {
	if AuthUserFromContext(ctx).GetAuthIdentifier() == 0 {
		return errors.New("user context required")
	}

	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return s.run(ctx, tx, companyID, form, itemID)
	})
}

// run creates the variants within an existing transaction: either the matrix of
// requested attribute combinations, or a single default variant.
func (s *CreateProductWithVariantsService) run(ctx context.Context, tx *sql.Tx, companyID int, form *StoreItemWithAttributesForm, itemID int) error {
	// Attributes without concrete combinations are meaningless.
	if len(form.AttributeIDs) > 0 && len(form.VariantCombos) == 0 {
		return errors.New("variants required when attributes are specified")
	}

	if len(form.VariantCombos) == 0 {
		// Simple product: a single default variant.
		return s.createDefaultVariant(ctx, tx, companyID, itemID, form.Name)
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// The raw statement keyed on id alone. company_id is added here: itemID always
	// belongs to companyID on every live path (the item is inserted in this same
	// transaction), so this narrows nothing today and stops a stray id from flipping
	// another tenant's item tomorrow. mustAffectRows turns a zero-row update into an
	// error rather than a silent success. updated_at is stamped by Update, since
	// itemRead maps the column.
	affected, err := ptx.Model(&itemRead{}).
		WhereEq("company_id", companyID).
		WhereEq("id", itemID).
		Update(ctx, map[string]any{"has_variants": true})
	if err := mustAffectRows(affected, err, "item"); err != nil {
		return err
	}

	for i, attrID := range form.AttributeIDs {
		if err := s.attachProductAttribute(tx, companyID, itemID, attrID, i); err != nil {
			return err
		}
	}

	for _, combo := range form.VariantCombos {
		variant := &itemVariant{
			ItemID:    itemID,
			SKU:       combo.SKU,
			Price:     combo.Price,
			CostPrice: combo.CostPrice,
			IsDefault: false,
		}
		if combo.SKU == "" {
			variant.SKU = fmt.Sprintf("SKU-%s-%d", generateHashCode(form.Name, 6), itemID)
		}

		variantName, err := s.generateVariantName(ctx, tx, companyID, form.AttributeIDs, combo.AttributeValueIDs)
		if err != nil {
			return err
		}
		variant.Name = variantName

		if err := s.storeVariantWithAttributeMapping(ctx, tx, companyID, variant, combo.AttributeValueIDs); err != nil {
			return err
		}
	}

	return nil
}

// attachProductAttribute links an attribute to a product.
//
// The empty (but non-nil) update list is what makes this DO NOTHING: Upsert only fills
// in a default DO UPDATE list when the slice is nil, and the grammar emits DO NOTHING
// when it is empty. Re-attaching an attribute must not disturb its sort_order.
func (s *CreateProductWithVariantsService) attachProductAttribute(tx *sql.Tx, companyID, itemID, attributeID, sortOrder int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	_, err = ptx.Model(&productAttributeRead{}).Upsert(
		context.Background(),
		[]map[string]any{{
			"company_id":   companyID,
			"item_id":      itemID,
			"attribute_id": attributeID,
			"sort_order":   sortOrder,
		}},
		[]string{"company_id", "item_id", "attribute_id"},
		[]string{},
	)
	return err
}

// generateVariantName builds a human-readable name from the combo's display values,
// e.g. "Red Large".
//
// The order comes from attributeOrder, the attribute ids as the form listed them.
// This used to range over the combo's map directly, and Go randomises map iteration
// order, so the same combination could be named "Red Large" on one insert and
// "Large Red" on the next. Attribute ids absent from attributeOrder are appended in
// ascending id order rather than left to chance.
//
// The display names are fetched in one query instead of one per attribute.
// WithTrashed preserves the old behaviour: the per-value lookup did not filter
// deleted_at, so a soft-deleted value still contributes its name.
func (s *CreateProductWithVariantsService) generateVariantName(
	ctx context.Context, tx *sql.Tx, companyID int, attributeOrder []int, attributeValueIDs map[int]int,
) (string, error) {
	if len(attributeValueIDs) == 0 {
		return "Default", nil
	}

	ordered := orderedAttributeIDs(attributeOrder, attributeValueIDs)

	ptx, err := playTx(tx)
	if err != nil {
		return "", err
	}

	valueIDs := make([]any, 0, len(ordered))
	for _, attrID := range ordered {
		valueIDs = append(valueIDs, attributeValueIDs[attrID])
	}

	var values []attributeValueRead
	if err := ptx.Model(&attributeValueRead{}).
		Select("id", "attribute_id", "display_name").
		WithTrashed().
		WhereEq("company_id", companyID).
		WhereIn("id", valueIDs...).
		Get(ctx, &values); err != nil {
		return "", err
	}

	// Keyed on (attribute_id, id): the old query matched on both, so a value id paired
	// with the wrong attribute contributed nothing.
	type valueKey struct{ attrID, valueID int }
	byKey := make(map[valueKey]string, len(values))
	for _, v := range values {
		byKey[valueKey{v.AttributeID, v.ID}] = v.DisplayName
	}

	var names []string
	for _, attrID := range ordered {
		if name := byKey[valueKey{attrID, attributeValueIDs[attrID]}]; name != "" {
			names = append(names, name)
		}
	}

	return strings.Join(names, " "), nil
}

// orderedAttributeIDs returns the combo's attribute ids, those named in order first
// (in that order), then any remainder sorted ascending so the result is deterministic.
func orderedAttributeIDs(order []int, attributeValueIDs map[int]int) []int {
	ordered := make([]int, 0, len(attributeValueIDs))
	seen := make(map[int]bool, len(attributeValueIDs))

	for _, attrID := range order {
		if _, ok := attributeValueIDs[attrID]; ok && !seen[attrID] {
			ordered = append(ordered, attrID)
			seen[attrID] = true
		}
	}

	rest := make([]int, 0, len(attributeValueIDs))
	for attrID := range attributeValueIDs {
		if !seen[attrID] {
			rest = append(rest, attrID)
		}
	}
	sort.Ints(rest)

	return append(ordered, rest...)
}

// storeVariantWithAttributeMapping inserts the variant and its attribute-value
// links. price/cost_price coalesce to 0 because those columns are NOT NULL.
func (s *CreateProductWithVariantsService) storeVariantWithAttributeMapping(ctx context.Context, tx *sql.Tx, companyID int, variant *itemVariant, attributeValueIDs map[int]int) error {
	var price, cost float64
	if variant.Price != nil {
		price = *variant.Price
	}
	if variant.CostPrice != nil {
		cost = *variant.CostPrice
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// uuid is left to its column default (gen_random_uuid()), which is what the
	// statement's explicit call produced.
	id, err := ptx.Model(&itemVariantRead{}).Insert(ctx, map[string]any{
		"company_id": companyID,
		"item_id":    variant.ItemID,
		"sku":        variant.SKU,
		"name":       variant.Name,
		"is_default": variant.IsDefault,
		"price":      price,
		"cost_price": cost,
	})
	if err != nil {
		return err
	}
	variant.ID = int(id)

	if len(attributeValueIDs) == 0 {
		return nil
	}

	// One statement for the whole combo. The empty (non-nil) update list compiles to
	// DO NOTHING, as before.
	rows := make([]map[string]any, 0, len(attributeValueIDs))
	for _, attrID := range orderedAttributeIDs(nil, attributeValueIDs) {
		rows = append(rows, map[string]any{
			"company_id":         companyID,
			"variant_id":         variant.ID,
			"attribute_id":       attrID,
			"attribute_value_id": attributeValueIDs[attrID],
		})
	}

	_, err = ptx.Model(&variantAttributeValueRead{}).Upsert(
		ctx, rows,
		[]string{"company_id", "variant_id", "attribute_id"},
		[]string{},
	)
	return err
}

// createDefaultVariant creates the single default variant of a simple product.
func (s *CreateProductWithVariantsService) createDefaultVariant(ctx context.Context, tx *sql.Tx, companyID, itemID int, itemName string) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	sku := fmt.Sprintf("SKU-%s-%d", generateHashCode(itemName, 6), itemID)
	_, err = ptx.Model(&itemVariantRead{}).Insert(ctx, map[string]any{
		"company_id": companyID,
		"item_id":    itemID,
		"sku":        sku,
		"name":       "Default",
		"is_default": true,
	})
	return err
}

// generateHashCode derives a short deterministic code from a string.
func generateHashCode(s string, length int) string {
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = ((hash << 5) - hash) + int(s[i])
	}

	result := fmt.Sprintf("%d", hash)
	if len(result) > length {
		result = result[len(result)-length:]
	}
	return result
}
