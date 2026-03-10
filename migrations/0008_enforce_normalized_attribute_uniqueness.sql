-- Migration: enforce normalized uniqueness for attribute names and values
-- This enforces case-insensitive, trim-insensitive uniqueness for active records.

-- Step 1: fail early if legacy duplicates would violate the new constraints.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM attributes
    WHERE deleted_at IS NULL
    GROUP BY company_id, lower(btrim(name))
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION
      'Cannot enforce normalized uniqueness on attributes.name: duplicates exist for active attributes.';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM attribute_values
    WHERE deleted_at IS NULL
    GROUP BY company_id, attribute_id, lower(btrim(value))
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION
      'Cannot enforce normalized uniqueness on attribute_values.value: duplicates exist for active attribute values.';
  END IF;
END $$;

-- Step 2: enforce normalized uniqueness with partial unique indexes.
CREATE UNIQUE INDEX IF NOT EXISTS idx_attributes_company_name_norm_unique
  ON attributes (company_id, lower(btrim(name)))
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_attribute_values_company_attribute_value_norm_unique
  ON attribute_values (company_id, attribute_id, lower(btrim(value)))
  WHERE deleted_at IS NULL;

-- Optional diagnostics if migration fails:
-- SELECT company_id, lower(btrim(name)) AS normalized_name, COUNT(*)
-- FROM attributes
-- WHERE deleted_at IS NULL
-- GROUP BY company_id, lower(btrim(name))
-- HAVING COUNT(*) > 1;
--
-- SELECT company_id, attribute_id, lower(btrim(value)) AS normalized_value, COUNT(*)
-- FROM attribute_values
-- WHERE deleted_at IS NULL
-- GROUP BY company_id, attribute_id, lower(btrim(value))
-- HAVING COUNT(*) > 1;
