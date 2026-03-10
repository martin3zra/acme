-- Migration: Clean up items table - remove columns migrated to variants
-- This migration removes obsolete columns from items table after successful data migration

-- IMPORTANT: Only run this migration after:
-- 1. Running migrations 0002, 0003, and 0004
-- 2. Verifying all data migrated successfully
-- 3. Updating application code to use new schema
-- 4. Testing the application thoroughly

-- Step 1: Drop the price column (now lives on variants)
ALTER TABLE items DROP COLUMN IF EXISTS price;

-- Step 2: Drop the identifiers JSONB column (migrated to variant columns)
ALTER TABLE items DROP COLUMN IF EXISTS identifiers;

-- Step 3: Drop the has_variants column (all items can have variants now)
ALTER TABLE items DROP COLUMN IF EXISTS has_variants;

-- Step 4: Drop old unique constraint on items_variants.sku (replaced by company-scoped index)
-- Note: The global unique constraint name may vary, check with:
-- SELECT conname FROM pg_constraint WHERE conrelid = 'items_variants'::regclass AND contype = 'u';
DO $$ 
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint 
    WHERE conname = 'items_variants_sku_key' 
    AND conrelid = 'items_variants'::regclass
  ) THEN
    ALTER TABLE items_variants DROP CONSTRAINT items_variants_sku_key;
  END IF;
END $$;

-- Step 5: Drop item_id from invoices_items once variant_id is fully populated
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'invoices_items'
      AND column_name = 'item_id'
  ) THEN
    IF EXISTS (SELECT 1 FROM invoices_items WHERE variant_id IS NULL) THEN
      RAISE EXCEPTION 'Cannot drop invoices_items.item_id because some rows still have NULL variant_id';
    END IF;

    ALTER TABLE invoices_items DROP COLUMN item_id;
  END IF;
END $$;

-- Verification: Check final schema
-- \d items
-- \d items_variants
-- \d invoices_items
