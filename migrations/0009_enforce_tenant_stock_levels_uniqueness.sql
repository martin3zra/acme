-- Migration: enforce tenant-scoped uniqueness and lookup index for stock levels

-- Keep existing data safe by ensuring nullable quantities are normalized.
UPDATE stock_levels
SET quantity = 0
WHERE quantity IS NULL;

ALTER TABLE stock_levels
  ALTER COLUMN quantity SET DEFAULT 0;

ALTER TABLE stock_levels
  ALTER COLUMN quantity SET NOT NULL;

-- Ensure stock uniqueness always includes company_id (tenant).
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'stock_levels_company_id_warehouse_id_variant_id_key'
  ) THEN
    ALTER TABLE stock_levels
      RENAME CONSTRAINT stock_levels_company_id_warehouse_id_variant_id_key
      TO stock_levels_company_variant_warehouse_unique;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'stock_levels_company_variant_warehouse_unique'
  ) THEN
    ALTER TABLE stock_levels
      ADD CONSTRAINT stock_levels_company_variant_warehouse_unique
      UNIQUE (company_id, warehouse_id, variant_id);
  END IF;
END
$$;

DROP INDEX IF EXISTS idx_stock_levels_combo;
CREATE INDEX IF NOT EXISTS idx_stock_levels_combo
  ON stock_levels(company_id, warehouse_id, variant_id);
