# Inventory

Items, their variants, warehouses, and stock. Every quantity change is recorded
as a movement, so stock levels are always derivable from an audit trail.

## For users

### The pieces

- **Items** — the products/services you sell and buy. An item can be a simple
  product, a service, or a product **with variants**.
- **Variants** — concrete sellable units of an item. A simple product has one
  default variant; a product **with variants** has one per attribute combination
  (e.g. Color × Size). Stock is tracked per variant.
- **Attributes & values** — the dimensions (Color) and their options (Red, Blue)
  that generate the variant matrix.
- **Warehouses** — locations that hold stock. Balances are per variant, per
  warehouse.

### How stock moves

You never edit a stock number directly; instead the system records a **movement**
and updates the **balance**:

- **Sales** move stock **out** ([invoices.md](invoices.md)).
- **Purchase receipts** move stock **in** ([purchases.md](purchases.md)).
- **Transfers** move stock **between warehouses**, with a lifecycle: create →
  **dispatch** (leaves the source) → **receive** (arrives at the destination), or
  **cancel**.
- **Adjustments** correct a count (stock-take, breakage) with a reason.

### Where you look

Stocks (current levels), Movements (the history), Transfers, Adjustments — each
its own screen under Inventory.

## For developers

### Data model

- **`items`**, **`items_variants`**, **`items_units`** — item, its variants, and
  unit links. Models: `itemRead`/`lineItemRead`, `itemVariantRead`, `itemUnitRead`.
- **`warehouses`** — `warehouseRead`.
- **`inventory_movements`** — the append-only ledger of stock changes
  (`InventoryMovement` write model, `inventoryMovementRead`).
- **`inventory_balances`** — the current per-variant/per-warehouse quantity
  (`inventoryBalanceRead`); updated by an increment upsert (kept raw SQL — a
  replace-style upsert can't express `quantity += EXCLUDED.quantity`).
- **`inventory_transfers`**, **`inventory_transfer_lines`** — transfers and their
  lines (`inventoryTransferRead`, `inventoryTransferLine`).
- **`attributes`**, **`attribute_values`**, **`product_attributes`**,
  **`variant_attribute_values`** — the variant matrix.

### Routes (`app/route.go`)

Items: `GET/POST /items`, `POST /items/variants`, `PUT /items/:id`,
`PUT /items/:id/change-status`, `DELETE /items/:id`.
Attributes: `GET/POST/PUT/DELETE /attributes` and `/attributes/:id/values`,
`/attribute-values/:uuid` (gated by the variants feature flag).
Inventory: `GET /inventories/warehouses` (+ CRUD), `/stocks`, `/movements`,
`/transfers` (+ `create`, and `:id/dispatch|receive|cancel`, `:id`),
`/adjustments` (+ `create`).

### Key rules

- **Movement + balance** — every stock change writes an `inventory_movements`
  row and upserts `inventory_balances`. The balance upsert is raw SQL because it
  increments rather than replaces.
- **Transfer lifecycle** takes **row locks** (`LockForUpdate`) so concurrent
  dispatch/receive can't corrupt balances (`flow_transfer_lock_test.go`).
- **Variant matrix** — creating an item with attributes generates a variant per
  combination; a simple item gets one default, inventory-tracked variant.
- **Variants feature flag** — the attribute routes are gated per company
  (`RequiresVariants` middleware) and 404 when off.
- **Warehouse guard** — operations validate the warehouse belongs to the tenant
  (`flow_warehouse_guard_test.go`).

### Permissions

`viewAny:inventory`, `create:transfer`, `create:adjustment`, and the item /
attribute permissions (`:item`, `:attribute`). See [permissions.md](permissions.md).

### Entry points & tests

- Handlers: `app/item-handlers.go`, `app/inventory-handlers.go`,
  `app/warehouse-handlers.go`, `app/attribute-handlers.go`
- Repositories: `app/inventory-repository.go`, item/variant repositories
- Factories: `mkItem`, `mkVariantItem` (`app/factories_test.go`)
- Tests: `app/flow_inventory_test.go`, `flow_inventory_reads_playsql_test.go`,
  `flow_transfer_lifecycle_test.go`, `flow_transfer_lock_test.go`,
  `flow_transfer_variant_lines_test.go`, `flow_warehouse_guard_test.go`,
  `flow_item_variants_test.go`, `flow_variants_test.go`, `flow_attributes_test.go`,
  `flow_item_unit_test.go`, `flow_item_search_variants_test.go`

## Related

[invoices.md](invoices.md) · [purchases.md](purchases.md) ·
[taxes.md](taxes.md)
