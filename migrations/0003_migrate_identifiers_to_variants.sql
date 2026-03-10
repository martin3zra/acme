-- Migration: Migrate identifier data from items.identifiers JSONB to items_variants columns
-- This migration extracts identifiers from the JSONB field and populates variant columns

-- Step 1: Ensure all items have at least one default variant
-- Create missing default variants for items that don't have any
INSERT INTO items_variants (company_id, item_id, sku, name, is_default, price, cost_price, active)
SELECT 
  i.company_id,
  i.id,
  COALESCE(
    i.identifiers->>'sku', 
    i.identifiers->>'code',
    CONCAT('SKU-', SUBSTRING(i.uuid::text, 1, 8))
  ),
  CONCAT(i.name, ' (Default)'),
  TRUE,
  COALESCE(i.price, 0.00),
  NULL,
  TRUE
FROM items i
WHERE i.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM items_variants iv 
    WHERE iv.item_id = i.id 
      AND iv.company_id = i.company_id
      AND iv.deleted_at IS NULL
  );

-- Step 2: Migrate identifiers from items.identifiers JSONB to default variants
UPDATE items_variants iv
SET 
  sku = COALESCE(
    NULLIF(iv.sku, ''),
    i.identifiers->>'sku',
    i.identifiers->>'code',
    CONCAT('SKU-', SUBSTRING(iv.uuid::text, 1, 8))
  ),
  barcode = i.identifiers->>'barcode',
  reference = i.identifiers->>'reference',
  vendor_reference = COALESCE(
    i.identifiers->>'vendor_reference',
    i.identifiers->>'vendorReference'
  )
FROM items i
WHERE iv.item_id = i.id
  AND iv.company_id = i.company_id
  AND iv.is_default = TRUE
  AND iv.deleted_at IS NULL
  AND i.deleted_at IS NULL;

-- Step 3: Ensure all non-default variants have unique SKUs
-- For variants without SKUs, generate one based on UUID
UPDATE items_variants
SET sku = CONCAT('SKU-', SUBSTRING(uuid::text, 1, 8))
WHERE sku IS NULL OR sku = '';

-- Step 4: Update variant prices from item prices where variant price is NULL
UPDATE items_variants iv
SET price = i.price
FROM items i
WHERE iv.item_id = i.id
  AND iv.company_id = i.company_id
  AND iv.price IS NULL
  AND i.price IS NOT NULL;

-- Step 5: Set default price of 0 for any remaining NULL prices
UPDATE items_variants 
SET price = 0.00 
WHERE price IS NULL;

-- Verification queries (run these manually to check migration success):
-- SELECT COUNT(*) FROM items i WHERE NOT EXISTS (SELECT 1 FROM items_variants WHERE item_id = i.id);
-- SELECT COUNT(*) FROM items_variants WHERE sku IS NULL OR sku = '';
-- SELECT COUNT(*) FROM items_variants WHERE price IS NULL;
