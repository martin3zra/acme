-- Migration: Add combination signature to variants for deduplication
-- This migration adds a canonical signature column to track unique attribute combinations
-- and prevent duplicate variants from being created

-- Step 1: Add combination_signature column (nullable during migration)
ALTER TABLE items_variants 
  ADD COLUMN combination_signature VARCHAR(500);

-- Step 2: Backfill existing variant signatures from variant_attribute_values
-- Build signature as sorted "attr_id:value_id|attr_id:value_id" format
UPDATE items_variants iv
SET combination_signature = (
  SELECT COALESCE(
    string_agg(
      vav.attribute_id::text || ':' || vav.attribute_value_id::text, 
      '|' ORDER BY vav.attribute_id
    ), 
    ''
  )
  FROM variant_attribute_values vav
  WHERE vav.company_id = iv.company_id 
    AND vav.variant_id = iv.id
);

-- Step 3: Set empty string for variants without attributes (default variants)
UPDATE items_variants 
SET combination_signature = '' 
WHERE combination_signature IS NULL;

-- Step 4: Make combination_signature NOT NULL
ALTER TABLE items_variants 
  ALTER COLUMN combination_signature SET NOT NULL;

-- Step 5: Create unique index to prevent duplicate combinations per item
-- Only enforce uniqueness for non-deleted variants
CREATE UNIQUE INDEX idx_items_variants_unique_combo 
  ON items_variants(company_id, item_id, combination_signature) 
  WHERE deleted_at IS NULL;

-- Step 6: Create index for signature lookups
CREATE INDEX idx_items_variants_signature 
  ON items_variants(combination_signature);

-- Verification query:
-- SELECT iv.id, iv.name, iv.combination_signature, i.name as item_name
-- FROM items_variants iv
-- JOIN items i ON iv.item_id = i.id
-- WHERE iv.company_id = 1
-- ORDER BY i.name, iv.combination_signature;
