package app

import (
	"fmt"
	"testing"
	"time"
)

// Expense reads and writes converted from raw database/sql (and hand-concatenated
// fmt.Sprintf filter strings) to playsql. The repository had no test coverage at
// all, so these cover the filters, the category relation, the soft-delete rules
// and the WithSum rollup that replaced the GROUP BY.

func mkExpenseCategory(t *testing.T, f *fixture, name string) string {
	t.Helper()
	if err := f.s.storeExpenseCategory(f.ctx, &StoreExpenseCategoryForm{
		Name: name, Description: name + " description",
	}); err != nil {
		t.Fatalf("storeExpenseCategory(%q): %v", name, err)
	}
	return scalarString(t, f.s.db,
		`SELECT uuid::text FROM expenses_categories WHERE company_id = $1 AND name = $2`,
		f.company.ID, name)
}

func mkExpense(t *testing.T, f *fixture, categoryUUID string, amount float64, date time.Time) string {
	t.Helper()
	notes := fmt.Sprintf("%s/%.2f", date.Format("2006-01-02"), amount)
	if err := f.s.storeExpense(f.ctx, &StoreExpenseForm{
		Category: categoryUUID, Date: date, Amount: amount, Notes: notes,
	}); err != nil {
		t.Fatalf("storeExpense: %v", err)
	}
	return scalarString(t, f.s.db,
		`SELECT uuid::text FROM expenses WHERE company_id = $1 AND notes = $2`, f.company.ID, notes)
}

func day(offset int) time.Time {
	return time.Now().AddDate(0, 0, offset).Truncate(24 * time.Hour)
}

// TestFindExpenses_LoadsCategory: the old INNER JOIN becomes a belongsTo eager
// load, and only id/uuid/name are copied onto the response.
func TestFindExpenses_LoadsCategory(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Rent")
	mkExpense(t, f, cat, 500, day(-1))

	expenses, err := f.s.findExpenses(f.ctx)
	is.NoErr(err)
	is.Equal(len(expenses), 1)

	e := expenses[0]
	is.EqualFloat(e.Amount, 500)
	is.Equal(e.Category.Name, "Rent")
	is.Equal(e.Category.UUID, cat)
	is.True(e.Category.ID != 0, "category id should be populated")
	is.True(e.UUID != "", "expense uuid should come from the database default")
}

// TestFindExpenses_OrdersByIDDesc: newest first.
func TestFindExpenses_OrdersByIDDesc(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Fuel")
	mkExpense(t, f, cat, 10, day(-3))
	last := mkExpense(t, f, cat, 20, day(-2))

	expenses, err := f.s.findExpenses(f.ctx)
	is.NoErr(err)
	is.Equal(len(expenses), 2)
	is.Equal(expenses[0].UUID, last)
}

// TestFindExpenses_Filters: date bounds and category, previously appended as
// fmt.Sprintf placeholders.
func TestFindExpenses_Filters(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	rent := mkExpenseCategory(t, f, "Rent")
	fuel := mkExpenseCategory(t, f, "Fuel")
	rentID := scalarInt(t, s.db, `SELECT id FROM expenses_categories WHERE uuid = $1`, rent)

	old := mkExpense(t, f, rent, 100, day(-30))
	recent := mkExpense(t, f, rent, 200, day(-2))
	other := mkExpense(t, f, fuel, 300, day(-2))

	within, err := f.s.findExpenses(f.ctx, WithDateRange(day(-7), day(0)))
	is.NoErr(err)
	is.Equal(len(within), 2)

	from, err := f.s.findExpenses(f.ctx, WithFromDate(day(-40)))
	is.NoErr(err)
	is.Equal(len(from), 3)

	to, err := f.s.findExpenses(f.ctx, WithToDate(day(-20)))
	is.NoErr(err)
	is.Equal(len(to), 1)
	is.Equal(to[0].UUID, old)

	byCat, err := f.s.findExpenses(f.ctx, WithCategory(int64(rentID)))
	is.NoErr(err)
	is.Equal(len(byCat), 2)
	for _, e := range byCat {
		is.True(e.UUID != other, "fuel expense must be filtered out")
	}
	is.Equal(byCat[0].UUID, recent)
}

// TestFindExpenses_IncludeDeleted: the option was dead — the old query hardcoded
// `deleted_at IS NULL` into its base string, so it could never widen the result.
// It maps onto WithTrashed and now works.
func TestFindExpenses_IncludeDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Rent")
	kept := mkExpense(t, f, cat, 100, day(-1))
	gone := mkExpense(t, f, cat, 200, day(-1))

	is.NoErr(f.s.deleteExpense(f.ctx, gone))

	live, err := f.s.findExpenses(f.ctx)
	is.NoErr(err)
	is.Equal(len(live), 1)
	is.Equal(live[0].UUID, kept)

	all, err := f.s.findExpenses(f.ctx, IncludeDeleted())
	is.NoErr(err)
	is.Equal(len(all), 2)
}

// TestDeleteExpense_StampsBothColumns: Builder.Delete would only set deleted_at,
// so the soft delete goes through Update to keep bumping updated_at.
func TestDeleteExpense_StampsBothColumns(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Rent")
	uuid := mkExpense(t, f, cat, 100, day(-1))

	before := scalarString(t, s.db, `SELECT updated_at::text FROM expenses WHERE uuid = $1`, uuid)
	time.Sleep(2 * time.Millisecond)
	is.NoErr(f.s.deleteExpense(f.ctx, uuid))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM expenses WHERE uuid = $1 AND deleted_at IS NOT NULL`, uuid), 1)
	after := scalarString(t, s.db, `SELECT updated_at::text FROM expenses WHERE uuid = $1`, uuid)
	is.True(after != before, "deleteExpense should bump updated_at")
}

// TestFindExpenseByUUID_TrashedStillReadable: the old detail query had no
// deleted_at predicate, so the read opts into trashed rows.
func TestFindExpenseByUUID_TrashedStillReadable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Rent")
	uuid := mkExpense(t, f, cat, 100, day(-1))
	is.NoErr(f.s.deleteExpense(f.ctx, uuid))

	e, err := f.s.findExpenseByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(e.UUID, uuid)
	is.Equal(e.Category.Name, "Rent")
	is.True(e.DeletedAt != nil, "the soft-deleted expense should still report deleted_at")
}

// TestUpdateExpense: reassigns the category and leaves updated_at alone, matching
// the statement it replaced (expenseInsert does not map updated_at).
func TestUpdateExpense(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	rent := mkExpenseCategory(t, f, "Rent")
	fuel := mkExpenseCategory(t, f, "Fuel")
	uuid := mkExpense(t, f, rent, 100, day(-1))
	before := scalarString(t, s.db, `SELECT updated_at::text FROM expenses WHERE uuid = $1`, uuid)

	time.Sleep(2 * time.Millisecond)
	is.NoErr(f.s.updateExpense(f.ctx, uuid, &StoreExpenseForm{
		Category: fuel, Date: day(-1), Amount: 250, Notes: "revised",
	}))

	e, err := f.s.findExpenseByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.EqualFloat(e.Amount, 250)
	is.Equal(e.Notes, "revised")
	is.Equal(e.Category.Name, "Fuel")
	is.Equal(scalarString(t, s.db, `SELECT updated_at::text FROM expenses WHERE uuid = $1`, uuid), before)
}

// TestUpdateExpenseCategory_BumpsUpdatedAt: the raw statement set updated_at =
// NOW(); playsql stamps it because expenseCategoryRead maps the column.
func TestUpdateExpenseCategory_BumpsUpdatedAt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	cat := mkExpenseCategory(t, f, "Rent")
	before := scalarString(t, s.db, `SELECT updated_at::text FROM expenses_categories WHERE uuid = $1`, cat)

	time.Sleep(2 * time.Millisecond)
	is.NoErr(f.s.updateExpenseCategory(f.ctx, cat, &StoreExpenseCategoryForm{
		Name: "Office rent", Description: "monthly",
	}))

	c, err := f.s.findExpenseCategory(f.ctx, cat)
	is.NoErr(err)
	is.Equal(c.Name, "Office rent")
	is.Equal(c.Description, "monthly")
	is.True(scalarString(t, s.db, `SELECT updated_at::text FROM expenses_categories WHERE uuid = $1`, cat) != before,
		"updateExpenseCategory should bump updated_at")
}

// TestFindExpensesCategories_HidesTrashed: the only category read that filters
// deleted_at does it explicitly, since expenseCategoryRead has no softdelete tag.
func TestFindExpensesCategories_HidesTrashed(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkExpenseCategory(t, f, "Rent")
	gone := mkExpenseCategory(t, f, "Fuel")
	_, err := s.db.Exec(`UPDATE expenses_categories SET deleted_at = now() WHERE uuid = $1`, gone)
	is.NoErr(err)

	cats, err := f.s.findExpensesCategories(f.ctx)
	is.NoErr(err)
	is.Equal(len(cats), 1)
	is.Equal(cats[0].Name, "Rent")

	// The detail read must still resolve it — storeExpense/updateExpense depend on it.
	c, err := f.s.findExpenseCategory(f.ctx, gone)
	is.NoErr(err)
	is.Equal(c.Name, "Fuel")
}

// TestFindExpensesByCategories_Rollup: WithSum + WhereHas reproduce the old
// GROUP BY over an INNER JOIN — totals per category, ordered high to low, and
// categories with no matching expense are omitted entirely.
func TestFindExpensesByCategories_Rollup(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	rent := mkExpenseCategory(t, f, "Rent")
	fuel := mkExpenseCategory(t, f, "Fuel")
	mkExpenseCategory(t, f, "Unused")

	mkExpense(t, f, rent, 500, day(-2))
	mkExpense(t, f, rent, 250, day(-1))
	mkExpense(t, f, fuel, 300, day(-1))

	rollup, err := f.s.findExpensesByCategories(f.ctx)
	is.NoErr(err)
	is.Equal(len(rollup), 2) // "Unused" has no expenses and is excluded

	is.Equal(rollup[0].Name, "Rent")
	is.EqualFloat(rollup[0].TotalAmount, 750)
	is.Equal(rollup[1].Name, "Fuel")
	is.EqualFloat(rollup[1].TotalAmount, 300)
}

// TestFindExpensesByCategories_DateRangeAndTrashed: the date filter constrains
// both the aggregate and its existence check, and soft-deleted expenses drop out
// of the total.
func TestFindExpensesByCategories_DateRangeAndTrashed(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	rent := mkExpenseCategory(t, f, "Rent")
	fuel := mkExpenseCategory(t, f, "Fuel")

	mkExpense(t, f, rent, 1000, day(-40)) // outside the window
	mkExpense(t, f, rent, 500, day(-2))
	voided := mkExpense(t, f, rent, 900, day(-2))
	mkExpense(t, f, fuel, 300, day(-30)) // outside the window

	is.NoErr(f.s.deleteExpense(f.ctx, voided))

	rollup, err := f.s.findExpensesByCategories(f.ctx, WithDateRange(day(-7), day(0)))
	is.NoErr(err)

	// Only Rent has a live expense inside the window; the 900 was soft-deleted and
	// the 1000/300 fall outside it.
	is.Equal(len(rollup), 1)
	is.Equal(rollup[0].Name, "Rent")
	is.EqualFloat(rollup[0].TotalAmount, 500)
}
