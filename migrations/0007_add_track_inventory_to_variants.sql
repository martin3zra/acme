-- Migration: persist per-variant inventory tracking flag

ALTER TABLE items_variants
  ADD COLUMN IF NOT EXISTS track_inventory BOOLEAN;

UPDATE items_variants
SET track_inventory = TRUE
WHERE track_inventory IS NULL;

ALTER TABLE items_variants
  ALTER COLUMN track_inventory SET DEFAULT TRUE;

ALTER TABLE items_variants
  ALTER COLUMN track_inventory SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_items_variants_track_inventory
  ON items_variants(track_inventory);
