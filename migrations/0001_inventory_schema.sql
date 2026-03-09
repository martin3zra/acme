-- Inventory Management Schema Migration
-- This migration creates the complete inventory system with warehouses, variants, attributes, and stock tracking

-- 1. Add has_variants column to items table
ALTER TABLE items ADD COLUMN has_variants BOOLEAN DEFAULT FALSE;

-- 2. Create warehouses table
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    uuid UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(120) NOT NULL,
    address TEXT,
    description TEXT,
    status entity_status DEFAULT 'enabled'::entity_status,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(company_id, code)
);

-- 3. Create items_variants table (product variants)
CREATE TABLE items_variants (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    item_id INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    uuid UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    price DECIMAL(15, 2),
    cost_price DECIMAL(15, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 4. Create attributes table
CREATE TABLE attributes (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    uuid UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    type VARCHAR(20) DEFAULT 'select',
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(company_id, name)
);

-- 5. Create attribute_values table
CREATE TABLE attribute_values (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    attribute_id INTEGER NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    uuid UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    value VARCHAR(120) NOT NULL,
    display_name VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(company_id, attribute_id, value)
);

-- 6. Create product_attributes table (linking attributes to items)
CREATE TABLE product_attributes (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    item_id INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    attribute_id INTEGER NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, item_id, attribute_id)
);

-- 7. Create variant_attribute_values table (mapping variants to attribute values)
CREATE TABLE variant_attribute_values (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    variant_id INTEGER NOT NULL REFERENCES items_variants(id) ON DELETE CASCADE,
    attribute_id INTEGER NOT NULL REFERENCES attributes(id),
    attribute_value_id INTEGER NOT NULL REFERENCES attribute_values(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, variant_id, attribute_id)
);

-- 8. Create stock_levels table
CREATE TABLE stock_levels (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    variant_id INTEGER NOT NULL REFERENCES items_variants(id),
    quantity INTEGER DEFAULT 0,
    reorder_level INTEGER DEFAULT 0,
    reorder_quantity INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, warehouse_id, variant_id)
);

-- 9. Create indexes for performance
CREATE INDEX idx_warehouses_company ON warehouses(company_id);
CREATE INDEX idx_warehouses_deleted ON warehouses(deleted_at);
CREATE INDEX idx_items_variants_company ON items_variants(company_id);
CREATE INDEX idx_items_variants_item ON items_variants(item_id);
CREATE INDEX idx_items_variants_deleted ON items_variants(deleted_at);
CREATE INDEX idx_attributes_company ON attributes(company_id);
CREATE INDEX idx_attributes_deleted ON attributes(deleted_at);
CREATE INDEX idx_attribute_values_company ON attribute_values(company_id);
CREATE INDEX idx_attribute_values_attribute ON attribute_values(attribute_id);
CREATE INDEX idx_product_attributes_company ON product_attributes(company_id);
CREATE INDEX idx_product_attributes_item ON product_attributes(item_id);
CREATE INDEX idx_variant_attribute_values_variant ON variant_attribute_values(variant_id);
CREATE INDEX idx_variant_attribute_values_attribute ON variant_attribute_values(attribute_id);
CREATE INDEX idx_stock_levels_company ON stock_levels(company_id);
CREATE INDEX idx_stock_levels_warehouse ON stock_levels(warehouse_id);
CREATE INDEX idx_stock_levels_variant ON stock_levels(variant_id);
CREATE INDEX idx_stock_levels_combo ON stock_levels(company_id, warehouse_id, variant_id);
