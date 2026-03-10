-- Migration: Migrate invoices_items from item_id to variant_id
-- This migration changes invoices to reference variants instead of items directly

-- Step 1: Add variant_id column (nullable during migration)
ALTER TABLE invoices_items 
  ADD COLUMN variant_id INTEGER;

-- Step 2: Add foreign key constraint
ALTER TABLE invoices_items
  ADD CONSTRAINT fk_invoices_items_variant
  FOREIGN KEY (variant_id) 
  REFERENCES items_variants(id);

-- Step 3: Migrate existing invoice items to their default variants
UPDATE invoices_items ii
SET variant_id = (
  SELECT iv.id 
  FROM items_variants iv 
  WHERE iv.item_id = ii.item_id 
    AND iv.company_id = ii.company_id 
    AND iv.is_default = TRUE
    AND iv.deleted_at IS NULL
  LIMIT 1
)
WHERE ii.variant_id IS NULL;

-- Step 4: Verify all records have been migrated
-- If this query returns > 0, there are orphaned invoice items
-- SELECT COUNT(*) FROM invoices_items WHERE variant_id IS NULL;

-- Step 5: Make variant_id NOT NULL after migration
ALTER TABLE invoices_items 
  ALTER COLUMN variant_id SET NOT NULL;

-- Step 6: Create index for performance
CREATE INDEX idx_invoices_items_variant ON invoices_items(variant_id);
CREATE INDEX idx_invoices_items_company_variant ON invoices_items(company_id, variant_id);

-- Step 7: Drop the old item_id column
-- IMPORTANT: Only run after verifying all data migrated successfully
-- Uncomment when ready:
-- ALTER TABLE invoices_items DROP COLUMN item_id;

-- Verification query:
-- SELECT ii.*, iv.sku, iv.name as variant_name, i.name as item_name
-- FROM invoices_items ii
-- JOIN items_variants iv ON ii.variant_id = iv.id
-- JOIN items i ON iv.item_id = i.id
-- LIMIT 10;
