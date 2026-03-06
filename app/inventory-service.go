package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"acme/pkg/database"
)

// CreateProductWithVariantsService handles the complex logic of creating products with variants
type CreateProductWithVariantsService struct {
	db *sql.DB
}

// NewCreateProductWithVariantsService creates a new service instance
func NewCreateProductWithVariantsService(db *sql.DB) *CreateProductWithVariantsService {
	return &CreateProductWithVariantsService{db: db}
}

// Execute orchestrates the creation of a product with variants
func (s *CreateProductWithVariantsService) Execute(ctx context.Context, form *StoreItemWithAttributesForm, itemID int) error {
	upCtx := ctx.Value("user")
	if upCtx == nil {
		return errors.New("user context required")
	}

	companyID := CurrentCompany(ctx).ID

	// Check if item has no variants but attributes are provided
	if len(form.AttributeIDs) > 0 && len(form.VariantCombos) == 0 {
		return errors.New("variants required when attributes are specified")
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		// Update item to mark it has variants if combos provided
		if len(form.VariantCombos) > 0 {
			_, err := tx.ExecContext(
				ctx,
				`UPDATE items SET has_variants = true, updated_at = NOW() WHERE id = $1`,
				itemID,
			)
			if err != nil {
				return err
			}

			// Attach product attributes
			for i, attrID := range form.AttributeIDs {
				err := s.attachProductAttribute(tx, companyID, itemID, attrID, i)
				if err != nil {
					return err
				}
			}

			// Create variants from combos
			for _, combo := range form.VariantCombos {
				variant := &itemVariant{
					ItemID:    itemID,
					SKU:       combo.SKU,
					Price:     combo.Price,
					CostPrice: combo.CostPrice,
				}

				// Generate variant name from attribute values if SKU not provided
				if combo.SKU == "" {
					variant.SKU = fmt.Sprintf("SKU-%s-%d", generateHashCode(form.Name, 6), itemID)
				}

				// Generate human-readable name from attribute display values
				variantName, err := s.generateVariantName(ctx, tx, companyID, combo.AttributeValueIDs)
				if err != nil {
					return err
				}
				variant.Name = variantName
				variant.IsDefault = false

				// Store variant
				err = s.storeVariantWithAttributeMapping(ctx, tx, companyID, variant, combo.AttributeValueIDs)
				if err != nil {
					return err
				}
			}
		} else {
			// No variants or attributes - create default variant
			err := s.createDefaultVariant(ctx, tx, companyID, itemID, form.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// attachProductAttribute links an attribute to a product
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

// generateVariantName creates a human-readable variant name from attribute values
func (s *CreateProductWithVariantsService) generateVariantName(ctx context.Context, tx *sql.Tx, companyID int, attributeValueIDs map[int]int) (string, error) {
	if len(attributeValueIDs) == 0 {
		return "Default", nil
	}

	// Fetch attribute values in order
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

// storeVariantWithAttributeMapping creates a variant and maps its attribute values
func (s *CreateProductWithVariantsService) storeVariantWithAttributeMapping(ctx context.Context, tx *sql.Tx, companyID int, variant *itemVariant, attributeValueIDs map[int]int) error {
	// Insert variant
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

	err = stmt.QueryRowContext(
		ctx,
		companyID, variant.ItemID, variant.SKU, variant.Name, variant.IsDefault, variant.Price, variant.CostPrice,
	).Scan(&variant.ID)
	if err != nil {
		return err
	}

	// Insert variant attribute value mappings
	for attrID, valueID := range attributeValueIDs {
		mapStmt, err := tx.PrepareContext(
			ctx,
			`INSERT INTO variant_attribute_values (company_id, variant_id, attribute_id, attribute_value_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (company_id, variant_id, attribute_id) DO NOTHING`,
		)
		if err != nil {
			return err
		}
		defer mapStmt.Close()

		_, err = mapStmt.ExecContext(ctx, companyID, variant.ID, attrID, valueID)
		if err != nil {
			return err
		}
	}

	return nil
}

// createDefaultVariant creates a default variant for simple products
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

	// Generate SKU from item name and ID
	sku := fmt.Sprintf("SKU-%s-%d", generateHashCode(itemName, 6), itemID)

	_, err = stmt.ExecContext(ctx, companyID, itemID, sku, "Default")
	return err
}

// generateHashCode creates a simple hash code from a string
func generateHashCode(s string, length int) string {
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = ((hash << 5) - hash) + int(s[i])
	}

	// Convert to positive and get last 'length' chars
	result := fmt.Sprintf("%d", hash)
	if len(result) > length {
		result = result[len(result)-length:]
	}
	return result
}
