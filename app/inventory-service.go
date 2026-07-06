package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// Execute creates the variants for an already-persisted item: either the matrix
// of requested attribute combinations, or a single default variant.
func (s *CreateProductWithVariantsService) Execute(ctx context.Context, form *StoreItemWithAttributesForm, itemID int) error {
	if AuthUserFromContext(ctx).GetAuthIdentifier() == 0 {
		return errors.New("user context required")
	}

	companyID := CurrentCompany(ctx).ID

	// Attributes without concrete combinations are meaningless.
	if len(form.AttributeIDs) > 0 && len(form.VariantCombos) == 0 {
		return errors.New("variants required when attributes are specified")
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		if len(form.VariantCombos) == 0 {
			// Simple product: a single default variant.
			return s.createDefaultVariant(ctx, tx, companyID, itemID, form.Name)
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE items SET has_variants = true, updated_at = NOW() WHERE id = $1`,
			itemID,
		); err != nil {
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

			variantName, err := s.generateVariantName(ctx, tx, companyID, combo.AttributeValueIDs)
			if err != nil {
				return err
			}
			variant.Name = variantName

			if err := s.storeVariantWithAttributeMapping(ctx, tx, companyID, variant, combo.AttributeValueIDs); err != nil {
				return err
			}
		}

		return nil
	})
}

// attachProductAttribute links an attribute to a product.
func (s *CreateProductWithVariantsService) attachProductAttribute(tx *sql.Tx, companyID, itemID, attributeID, sortOrder int) error {
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

// generateVariantName builds a human-readable name from the combo's display values.
func (s *CreateProductWithVariantsService) generateVariantName(ctx context.Context, tx *sql.Tx, companyID int, attributeValueIDs map[int]int) (string, error) {
	if len(attributeValueIDs) == 0 {
		return "Default", nil
	}

	var names []string
	for attrID, valueID := range attributeValueIDs {
		var displayName string
		err := tx.QueryRowContext(
			ctx,
			`SELECT display_name FROM attribute_values WHERE id = $1 AND attribute_id = $2 AND company_id = $3`,
			valueID, attrID, companyID,
		).Scan(&displayName)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}
		if displayName != "" {
			names = append(names, displayName)
		}
	}

	return strings.Join(names, " "), nil
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

	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, is_default, price, cost_price, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if err = stmt.QueryRowContext(
		ctx,
		companyID, variant.ItemID, variant.SKU, variant.Name, variant.IsDefault, price, cost,
	).Scan(&variant.ID); err != nil {
		return err
	}

	for attrID, valueID := range attributeValueIDs {
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO variant_attribute_values (company_id, variant_id, attribute_id, attribute_value_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (company_id, variant_id, attribute_id) DO NOTHING`,
			companyID, variant.ID, attrID, valueID,
		); err != nil {
			return err
		}
	}

	return nil
}

// createDefaultVariant creates the single default variant of a simple product.
func (s *CreateProductWithVariantsService) createDefaultVariant(ctx context.Context, tx *sql.Tx, companyID, itemID int, itemName string) error {
	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO items_variants (company_id, item_id, uuid, sku, name, is_default, created_at, updated_at)
		 VALUES ($1, $2, gen_random_uuid(), $3, $4, true, NOW(), NOW())`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	sku := fmt.Sprintf("SKU-%s-%d", generateHashCode(itemName, 6), itemID)
	_, err = stmt.ExecContext(ctx, companyID, itemID, sku, "Default")
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
