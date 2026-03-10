-- Migration: Enhance item_variants table with identifier columns
-- This migration adds columns that were previously stored in items.identifiers JSONB

-- Step 1: Add new columns to items_variants table
ALTER TABLE items_variants 
  ADD COLUMN barcode VARCHAR(100),
  ADD COLUMN reference VARCHAR(100),
  ADD COLUMN vendor_reference VARCHAR(100),
  ADD COLUMN active BOOLEAN DEFAULT TRUE;

-- Step 2: Create indexes for lookups on new columns
CREATE INDEX idx_items_variants_sku ON items_variants(sku);
CREATE INDEX idx_items_variants_barcode ON items_variants(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_items_variants_item_id ON items_variants(item_id);
CREATE INDEX idx_items_variants_reference ON items_variants(reference) WHERE reference IS NOT NULL;
CREATE INDEX idx_items_variants_active ON items_variants(active);

-- Step 3: Create unique constraint for barcode per company (allowing NULLs)
CREATE UNIQUE INDEX idx_items_variants_company_barcode 
  ON items_variants(company_id, barcode) 
  WHERE barcode IS NOT NULL AND deleted_at IS NULL;

-- Step 4: Make price NOT NULL (all variants must have a price)
-- First, update any NULL prices to 0.00 before adding constraint
UPDATE items_variants SET price = 0.00 WHERE price IS NULL;
ALTER TABLE items_variants ALTER COLUMN price SET NOT NULL;
ALTER TABLE items_variants ALTER COLUMN price SET DEFAULT 0.00;

-- Step 5: Add constraint to ensure SKU uniqueness is enforced
-- Note: SKU column already has UNIQUE constraint from 0001_inventory_schema.sql
-- We'll add a partial index for performance with company_id
CREATE UNIQUE INDEX idx_items_variants_company_sku 
  ON items_variants(company_id, sku) 
  WHERE deleted_at IS NULL;

-- Note: The existing UNIQUE constraint on sku should be dropped after this migration
-- ALTER TABLE items_variants DROP CONSTRAINT items_variants_sku_key;
