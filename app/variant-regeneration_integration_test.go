package app

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestVariantRegenerationLifecycleIntegration(t *testing.T) {
	db := openVariantIntegrationDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer tx.Rollback()

	companyID, taxID, _, ok := findVariantFixtureCompanyAndTax(t, tx)
	if !ok {
		t.Skip("no suitable fixture company/tax found in test database")
	}

	unique := time.Now().UnixNano()

	itemID := mustInsertTestItem(t, tx, companyID, taxID, unique)
	colorAttributeID := mustInsertTestAttribute(t, tx, companyID, fmt.Sprintf("color-%d", unique), "Color")
	sizeAttributeID := mustInsertTestAttribute(t, tx, companyID, fmt.Sprintf("size-%d", unique), "Size")

	redValueID := mustInsertTestAttributeValue(t, tx, companyID, colorAttributeID, "red", "Red", 1)
	blueValueID := mustInsertTestAttributeValue(t, tx, companyID, colorAttributeID, "blue", "Blue", 2)
	smallValueID := mustInsertTestAttributeValue(t, tx, companyID, sizeAttributeID, "small", "Small", 1)
	mediumValueID := mustInsertTestAttributeValue(t, tx, companyID, sizeAttributeID, "medium", "Medium", 2)

	s := &Server{}
	attributeIDs := []int{colorAttributeID, sizeAttributeID}

	initialCombos := []VariantCombo{
		{AttributeValueIDs: map[int]int{colorAttributeID: redValueID, sizeAttributeID: smallValueID}, Price: floatPtr(10)},
		{AttributeValueIDs: map[int]int{colorAttributeID: blueValueID, sizeAttributeID: smallValueID}, Price: floatPtr(10)},
	}

	if err = s.addConfiguredVariants(tx, companyID, itemID, "Variant Lifecycle Item", attributeIDs, initialCombos); err != nil {
		t.Fatalf("initial variant generation failed: %v", err)
	}

	redSmallSig := buildVariantSignature(map[int]int{colorAttributeID: redValueID, sizeAttributeID: smallValueID})
	blueSmallSig := buildVariantSignature(map[int]int{colorAttributeID: blueValueID, sizeAttributeID: smallValueID})
	redMediumSig := buildVariantSignature(map[int]int{colorAttributeID: redValueID, sizeAttributeID: mediumValueID})
	blueMediumSig := buildVariantSignature(map[int]int{colorAttributeID: blueValueID, sizeAttributeID: mediumValueID})

	redSmallInitial := mustFindVariantStateBySignature(t, tx, companyID, itemID, redSmallSig)
	blueSmallInitial := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueSmallSig)

	fullMatrixCombos := []VariantCombo{
		{AttributeValueIDs: map[int]int{colorAttributeID: redValueID, sizeAttributeID: smallValueID}, Price: floatPtr(10)},
		{AttributeValueIDs: map[int]int{colorAttributeID: redValueID, sizeAttributeID: mediumValueID}, Price: floatPtr(12)},
		{AttributeValueIDs: map[int]int{colorAttributeID: blueValueID, sizeAttributeID: smallValueID}, Price: floatPtr(10)},
		{AttributeValueIDs: map[int]int{colorAttributeID: blueValueID, sizeAttributeID: mediumValueID}, Price: floatPtr(12)},
	}

	if err = s.addConfiguredVariants(tx, companyID, itemID, "Variant Lifecycle Item", attributeIDs, fullMatrixCombos); err != nil {
		t.Fatalf("expanded variant generation failed: %v", err)
	}

	redSmallAfterExpand := mustFindVariantStateBySignature(t, tx, companyID, itemID, redSmallSig)
	blueSmallAfterExpand := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueSmallSig)
	redMediumAfterExpand := mustFindVariantStateBySignature(t, tx, companyID, itemID, redMediumSig)
	blueMediumAfterExpand := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueMediumSig)

	if redSmallAfterExpand.ID != redSmallInitial.ID {
		t.Fatalf("existing Red/Small variant was not preserved: initial=%d afterExpand=%d", redSmallInitial.ID, redSmallAfterExpand.ID)
	}
	if blueSmallAfterExpand.ID != blueSmallInitial.ID {
		t.Fatalf("existing Blue/Small variant was not preserved: initial=%d afterExpand=%d", blueSmallInitial.ID, blueSmallAfterExpand.ID)
	}
	if redMediumAfterExpand.ID == 0 || blueMediumAfterExpand.ID == 0 {
		t.Fatalf("new combinations were not created for medium size")
	}

	warehouseID := mustInsertTestWarehouse(t, tx, companyID, unique)
	mustInsertStockLevelReference(t, tx, companyID, warehouseID, blueSmallAfterExpand.ID)

	redOnlyCombos := []VariantCombo{
		{AttributeValueIDs: map[int]int{colorAttributeID: redValueID, sizeAttributeID: smallValueID}, Price: floatPtr(10)},
		{AttributeValueIDs: map[int]int{colorAttributeID: redValueID, sizeAttributeID: mediumValueID}, Price: floatPtr(12)},
	}

	if err = s.addConfiguredVariants(tx, companyID, itemID, "Variant Lifecycle Item", attributeIDs, redOnlyCombos); err != nil {
		t.Fatalf("remove blue variant regeneration failed: %v", err)
	}

	blueSmallAfterRemoval := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueSmallSig)
	blueMediumAfterRemoval := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueMediumSig)

	if blueSmallAfterRemoval.ID != blueSmallAfterExpand.ID {
		t.Fatalf("referenced Blue/Small variant id changed after removal: expected=%d actual=%d", blueSmallAfterExpand.ID, blueSmallAfterRemoval.ID)
	}
	if blueSmallAfterRemoval.Active {
		t.Fatalf("referenced Blue/Small variant should be inactive after removal")
	}
	if blueSmallAfterRemoval.DeletedAt.Valid {
		t.Fatalf("referenced Blue/Small variant should not be soft-deleted")
	}

	if blueMediumAfterRemoval.ID != blueMediumAfterExpand.ID {
		t.Fatalf("Blue/Medium variant id changed unexpectedly after removal: expected=%d actual=%d", blueMediumAfterExpand.ID, blueMediumAfterRemoval.ID)
	}
	if !blueMediumAfterRemoval.DeletedAt.Valid {
		t.Fatalf("unreferenced Blue/Medium variant should be soft-deleted after removal")
	}

	if err = s.addConfiguredVariants(tx, companyID, itemID, "Variant Lifecycle Item", attributeIDs, fullMatrixCombos); err != nil {
		t.Fatalf("re-add blue variant regeneration failed: %v", err)
	}

	blueSmallAfterReadd := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueSmallSig)
	blueMediumAfterReadd := mustFindVariantStateBySignature(t, tx, companyID, itemID, blueMediumSig)

	if blueSmallAfterReadd.ID != blueSmallAfterRemoval.ID {
		t.Fatalf("Blue/Small should reactivate in place: expected id=%d actual id=%d", blueSmallAfterRemoval.ID, blueSmallAfterReadd.ID)
	}
	if !blueSmallAfterReadd.Active || blueSmallAfterReadd.DeletedAt.Valid {
		t.Fatalf("Blue/Small should be active and not soft-deleted after re-add")
	}

	if blueMediumAfterReadd.ID != blueMediumAfterRemoval.ID {
		t.Fatalf("Blue/Medium should revive in place: expected id=%d actual id=%d", blueMediumAfterRemoval.ID, blueMediumAfterReadd.ID)
	}
	if !blueMediumAfterReadd.Active || blueMediumAfterReadd.DeletedAt.Valid {
		t.Fatalf("Blue/Medium should be revived and active after re-add")
	}

	mustAssertSingleVariantPerSignature(t, tx, companyID, itemID, redSmallSig)
	mustAssertSingleVariantPerSignature(t, tx, companyID, itemID, redMediumSig)
	mustAssertSingleVariantPerSignature(t, tx, companyID, itemID, blueSmallSig)
	mustAssertSingleVariantPerSignature(t, tx, companyID, itemID, blueMediumSig)
}

func openVariantIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("ACME_TEST_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 dbname=acme user=postgres password=secret sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration test: cannot open postgres connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		t.Skipf("skipping integration test: cannot ping postgres at %q: %v", dsn, err)
	}

	return db
}

func findVariantFixtureCompanyAndTax(t *testing.T, tx *sql.Tx) (int, int, int, bool) {
	t.Helper()

	var companyID, taxID, unitID int
	err := tx.QueryRow(
		`SELECT c.id, t.id, u.id
		 FROM companies c
		 INNER JOIN taxes t ON t.company_id = c.id AND t.deleted_at IS NULL
		 INNER JOIN units u ON u.company_id = c.id AND u.deleted_at IS NULL
		 WHERE c.deleted_at IS NULL
		 LIMIT 1`,
	).Scan(&companyID, &taxID, &unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, 0, false
		}
		t.Fatalf("query fixture company/tax/unit: %v", err)
	}

	return companyID, taxID, unitID, true
}

func mustInsertTestItem(t *testing.T, tx *sql.Tx, companyID, taxID int, unique int64) int {
	t.Helper()

	var itemID int
	err := tx.QueryRow(
		`INSERT INTO items (name, description, tax_id, item_type, company_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		fmt.Sprintf("Variant Lifecycle Item %d", unique),
		"integration test",
		taxID,
		"product",
		companyID,
	).Scan(&itemID)
	if err != nil {
		t.Fatalf("insert test item: %v", err)
	}

	return itemID
}

func mustInsertTestAttribute(t *testing.T, tx *sql.Tx, companyID int, name, displayName string) int {
	t.Helper()

	var attributeID int
	err := tx.QueryRow(
		`INSERT INTO attributes (company_id, name, display_name, type)
		 VALUES ($1, $2, $3, 'select')
		 RETURNING id`,
		companyID,
		name,
		displayName,
	).Scan(&attributeID)
	if err != nil {
		t.Fatalf("insert test attribute %s: %v", name, err)
	}

	return attributeID
}

func mustInsertTestAttributeValue(t *testing.T, tx *sql.Tx, companyID, attributeID int, value, displayName string, sortOrder int) int {
	t.Helper()

	var attributeValueID int
	err := tx.QueryRow(
		`INSERT INTO attribute_values (company_id, attribute_id, value, display_name, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		companyID,
		attributeID,
		value,
		displayName,
		sortOrder,
	).Scan(&attributeValueID)
	if err != nil {
		t.Fatalf("insert test attribute value %s: %v", value, err)
	}

	return attributeValueID
}

func mustInsertTestWarehouse(t *testing.T, tx *sql.Tx, companyID int, unique int64) int {
	t.Helper()

	var warehouseID int
	err := tx.QueryRow(
		`INSERT INTO warehouses (company_id, code, name)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		companyID,
		fmt.Sprintf("W%s", fmt.Sprintf("%d", unique)[len(fmt.Sprintf("%d", unique))-6:]),
		fmt.Sprintf("Warehouse %d", unique),
	).Scan(&warehouseID)
	if err != nil {
		t.Fatalf("insert test warehouse: %v", err)
	}

	return warehouseID
}

func mustInsertStockLevelReference(t *testing.T, tx *sql.Tx, companyID, warehouseID, variantID int) {
	t.Helper()

	_, err := tx.Exec(
		`INSERT INTO stock_levels (company_id, warehouse_id, variant_id, quantity)
		 VALUES ($1, $2, $3, $4)`,
		companyID,
		warehouseID,
		variantID,
		5,
	)
	if err != nil {
		t.Fatalf("insert stock level reference: %v", err)
	}
}

type variantState struct {
	ID        int
	Active    bool
	DeletedAt sql.NullTime
}

func mustFindVariantStateBySignature(t *testing.T, tx *sql.Tx, companyID, itemID int, signature string) variantState {
	t.Helper()

	state := variantState{}
	err := tx.QueryRow(
		`SELECT id, active, deleted_at
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND combination_signature = $3
		 LIMIT 1`,
		companyID,
		itemID,
		signature,
	).Scan(&state.ID, &state.Active, &state.DeletedAt)
	if err != nil {
		t.Fatalf("query variant state for %s: %v", signature, err)
	}

	return state
}

func mustAssertSingleVariantPerSignature(t *testing.T, tx *sql.Tx, companyID, itemID int, signature string) {
	t.Helper()

	var count int
	err := tx.QueryRow(
		`SELECT COUNT(*)
		 FROM items_variants
		 WHERE company_id = $1 AND item_id = $2 AND combination_signature = $3`,
		companyID,
		itemID,
		signature,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count variants for signature %s: %v", signature, err)
	}

	if count != 1 {
		t.Fatalf("expected exactly one variant for signature %s, got %d", signature, count)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}