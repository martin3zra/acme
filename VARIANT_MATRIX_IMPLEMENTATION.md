# Variant Matrix Generator Implementation

## Overview

This implementation adds comprehensive variant matrix generation with automatic deduplication, reuse, and reconciliation capabilities. The system now:

- Generates all possible variant combinations from attribute values (cartesian product)
- Maintains consistent attribute ordering
- Submits the full desired variant matrix on edit/save
- Automatically reuses existing variants
- Revives soft-deleted variants when they reappear
- Reactivates inactive variants when they reappear
- Reconciles obsolete variants (soft-delete unreferenced, deactivate referenced)
- Prevents duplicate combinations via DB-level uniqueness constraint

## Changes Made

### 1. Database Migration (`migrations/0006_variant_combination_signature.sql`)

**New column**: `combination_signature` on `items_variants` table
- Stores canonical signature in format: `attr_id:value_id|attr_id:value_id` (sorted by attr_id)
- Backfills existing variants from `variant_attribute_values`
- Adds unique index to prevent duplicate combinations per item (where `deleted_at IS NULL`)

**To apply**:
```sql
-- Run the migration against your database
psql -d your_database -f migrations/0006_variant_combination_signature.sql
```

### 2. Backend Changes

#### Updated Models (`app/types.go`)
- Added `CombinationSignature` field to `itemVariant` struct

#### Enhanced Repository (`app/item-repository.go`)

**New helper methods**:
- `findVariantBySignature()` - Finds variant by signature (including soft-deleted)
- `reviveVariant()` - Reactivates soft-deleted variants with updated data
- `reactivateVariant()` - Reactivates inactive (non-deleted) variants when combination reappears
- `isVariantReferenced()` - Checks if variant is used in invoices or stock
- `reconcileObsoleteVariants()` - Handles soft-deletion/deactivation of obsolete variants

**Updated methods**:
- `storeItemVariant()` - Now persists combination_signature
- `storeDefaultVariant()` - Sets empty signature for default variants
- `addConfiguredVariants()` - Implements full reuse/revival/reconciliation logic
- `attachProductAttribute()` - Now updates sort_order on conflict (preserves attribute order)
- `findItemVariants()` - Selects combination_signature field
- `findVariantByID()` - Selects combination_signature field
- `findVariantBySKU()` - Selects combination_signature field
- `findItemVariantSetup()` - Uses stored signatures instead of computing from joins

**Regeneration flow**:
1. Attach/update product attributes with proper sort order
2. Remove product attributes no longer in submitted list
3. For each combination submitted:
   - Check if variant exists (by signature)
   - If exists and soft-deleted: revive it
   - If exists and inactive: reactivate it
   - If exists and active: keep as-is
   - If doesn't exist: create new variant
4. Reconcile obsolete variants:
   - If referenced by invoices/stock: deactivate (keep record)
   - If not referenced: soft-delete

### 3. Tests (`app/variant-matrix_test.go`)

**Unit tests**:
- `TestBuildVariantSignature` - Verifies signature generation is consistent and sorted
- `TestBuildVariantSignatureConsistency` - Ensures deterministic output
- `TestVariantSignatureOrdering` - Validates attribute order independence
- `TestVariantMatrixGeneration` - Documents cartesian product behavior

**Run tests**:
```bash
go test -v ./app -run TestVariant
```

## Usage Examples

### Frontend (Current Behavior)

The frontend generates all currently desired combinations and submits the full matrix in `CreateForm.tsx`:

```typescript
// User selects attributes and values:
// - Color: [Red, Blue]
// - Size: [Small, Medium]

// Frontend generates 4 combinations:
const combos = [
  { attribute_value_ids: {1: 101, 2: 201}, sku: "...", price: 10 },  // Red/Small
  { attribute_value_ids: {1: 101, 2: 202}, sku: "...", price: 10 },  // Red/Medium
  { attribute_value_ids: {1: 102, 2: 201}, sku: "...", price: 10 },  // Blue/Small
  { attribute_value_ids: {1: 102, 2: 202}, sku: "...", price: 10 },  // Blue/Medium
]

// Submit full desired matrix to backend
submit({
  attribute_ids: [1, 2],  // Color, Size (order matters!)
  variant_combos: combos
})
```

### Backend Processing

```go
// Backend canonicalizes signatures:
// {1: 101, 2: 201} → "1:101|2:201"
// {1: 101, 2: 202} → "1:101|2:202"
// etc.

// For each combo:
// 1. Check if signature exists in DB
// 2. If exists and active: keep as-is (no duplicate)
// 3. If exists and inactive: reactivate (reuse existing ID)
// 4. If exists and soft-deleted: revive (reuse existing ID)
// 5. If doesn't exist: create new variant
// 6. After all combos processed: reconcile obsolete variants
```

## Verification

### Manual Testing Scenarios

#### Scenario 1: Create item with variants
1. Navigate to Items → Create
2. Select "Product" type
3. Enable "Has Variants"
4. Select 2 attributes with 2 values each (e.g., Color: Red/Blue, Size: S/M)
5. Submit
6. **Expected**: 4 variants created with unique signatures

#### Scenario 2: Regeneration - Remove attribute value
1. Edit the item created above
2. Remove one attribute value (e.g., remove "Blue" from Color)
3. Submit
4. **Expected**: 
   - Red/Small and Red/Medium remain active
   - Blue/Small and Blue/Medium are reconciled:
     - referenced variants: `active = FALSE`, `deleted_at = NULL`
     - unreferenced variants: `deleted_at = NOW()` and `active = FALSE`
   - Check: `SELECT * FROM items_variants WHERE item_id = ? ORDER BY id`

#### Scenario 3: Revival - Re-add removed value
1. Edit the same item again
2. Re-add "Blue" to Color
3. Submit
4. **Expected**:
   - Blue/Small and Blue/Medium are restored as active (same IDs as before)
   - If soft-deleted: `deleted_at` becomes NULL
   - If inactive: `active` becomes TRUE
   - No new variants created (dedupe working)

#### Scenario 4: Referenced variants protection
1. Create item with variants
2. Create invoice using one variant
3. Edit item to remove that variant's combination
4. **Expected**:
   - Variant used in invoice: `active = FALSE`, `deleted_at = NULL` (kept)
   - Other obsolete variants: `deleted_at = NOW()` (soft-deleted)

#### Scenario 5: Attribute order preservation
1. Create item with attributes [Color, Size, Material]
2. Edit to reorder as [Size, Color, Material]
3. **Expected**:
   - `product_attributes.sort_order` updated to reflect new order
   - Variant names rebuilt with new attribute order

### Database Verification Queries

```sql
-- View all variants with signatures
SELECT 
  iv.id, 
  iv.sku, 
  iv.name, 
  iv.combination_signature,
  iv.active,
  iv.deleted_at,
  i.name as item_name
FROM items_variants iv
JOIN items i ON iv.item_id = i.id
WHERE i.company_id = 1
ORDER BY i.name, iv.combination_signature;

-- Check for duplicate signatures (should be 0)
SELECT 
  company_id, 
  item_id, 
  combination_signature, 
  COUNT(*) 
FROM items_variants 
WHERE deleted_at IS NULL 
GROUP BY company_id, item_id, combination_signature 
HAVING COUNT(*) > 1;

-- View product attributes order
SELECT 
  pa.item_id,
  i.name as item_name,
  pa.attribute_id,
  a.display_name,
  pa.sort_order
FROM product_attributes pa
JOIN items i ON pa.item_id = i.id
JOIN attributes a ON pa.attribute_id = a.id
WHERE pa.company_id = 1
ORDER BY pa.item_id, pa.sort_order;
```

## Rollback

If needed, rollback the migration:

```sql
-- Remove unique index
DROP INDEX IF EXISTS idx_items_variants_unique_combo;
DROP INDEX IF EXISTS idx_items_variants_signature;

-- Remove column
ALTER TABLE items_variants DROP COLUMN combination_signature;
```

## Performance Considerations

- Signature lookups are indexed for fast queries
- Unique constraint prevents duplicate inserts at DB level (no race conditions)
- Backfill may take time on large datasets (run during maintenance window)
- Frontend continues to generate combinations (no API overhead)

## Future Enhancements

- Bulk variant import/export with signature validation
- Variant templates for common attribute combinations
- Analytics on most/least popular combinations
- Automated SKU generation based on attribute values
- Variant cloning across items
