package app

import (
	"context"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type expenseCategory struct {
	ID          int     `json:"id"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	TotalAmount float64 `json:"total_amount,omitempty"`
	foundation.Timestamps
}

type expense struct {
	ID         int             `json:"id"`
	UUID       string          `json:"uuid"`
	Date       Date            `json:"date"`
	Amount     float64         `json:"amount"`
	Notes      string          `json:"notes"`
	ReceiptURL string          `json:"receipt_url"`
	Category   expenseCategory `json:"category"`
	foundation.Timestamps
}

type ExpenseFilter struct {
	FromDate       *time.Time
	ToDate         *time.Time
	CategoryID     *int64
	IncludeDeleted bool
}

type ExpenseOption func(*ExpenseFilter)

func WithDateRange(from, to time.Time) ExpenseOption {
	return func(f *ExpenseFilter) {
		f.FromDate = &from
		f.ToDate = &to
	}
}

func WithFromDate(from time.Time) ExpenseOption {
	return func(f *ExpenseFilter) {
		f.FromDate = &from
	}
}

func WithToDate(to time.Time) ExpenseOption {
	return func(f *ExpenseFilter) {
		f.ToDate = &to
	}
}

func WithCategory(categoryID int64) ExpenseOption {
	return func(f *ExpenseFilter) {
		f.CategoryID = &categoryID
	}
}

func IncludeDeleted() ExpenseOption {
	return func(f *ExpenseFilter) {
		f.IncludeDeleted = true
	}
}

// applyExpenseFilter narrows a query over the expenses table by the caller's
// options. It is shared by findExpenses and findExpensesByCategories, where it
// constrains both the aggregate subquery and its matching existence check so the
// two can never drift apart.
func applyExpenseFilter(companyID int, filter ExpenseFilter) func(*playsql.Builder) {
	return func(b *playsql.Builder) {
		b.WhereEq("company_id", companyID)
		if filter.FromDate != nil {
			b.Where("date", ">=", *filter.FromDate)
		}
		if filter.ToDate != nil {
			b.Where("date", "<=", *filter.ToDate)
		}
		if filter.CategoryID != nil {
			b.WhereEq("category_id", *filter.CategoryID)
		}
	}
}

func expenseFilterFrom(opts []ExpenseOption) ExpenseFilter {
	filter := ExpenseFilter{}
	for _, opt := range opts {
		opt(&filter)
	}
	return filter
}

// findExpenseByUUID reads one expense with its category. It uses WithTrashed
// because the old query had no deleted_at predicate: a soft-deleted expense is
// still previewable by uuid.
func (s *Server) findExpenseByUUID(ctx context.Context, uuid string) (*expense, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row expenseRead
	if err := pdb.Model(&expenseRead{}).
		With("Category").
		WithTrashed().
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}

	return row.toExpense(), nil
}

// findExpenses lists the company's expenses, newest first.
//
// IncludeDeleted now works. The old query hardcoded `AND expenses.deleted_at IS
// NULL` into its base string and then appended the same predicate again when the
// option was absent, so the option could never widen the result set. It maps onto
// WithTrashed here. No caller passes it, so no caller changes behaviour.
func (s *Server) findExpenses(ctx context.Context, opts ...ExpenseOption) ([]*expense, error) {
	companyID := CurrentCompany(ctx).ID
	filter := expenseFilterFrom(opts)

	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	q := pdb.Model(&expenseRead{}).With("Category")
	if filter.IncludeDeleted {
		q.WithTrashed()
	}
	applyExpenseFilter(companyID, filter)(q)

	var rows []expenseRead
	if err := q.OrderBy("id", playsql.Desc).Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*expense, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toExpense())
	}
	return data, nil
}

// findExpensesByCategories rolls each category's expenses up into a total.
//
// The old GROUP BY over an INNER JOIN becomes a correlated SUM subquery per
// category (WithSum) plus a matching WhereHas, which reproduces the inner join's
// "only categories with a matching expense" semantics — and keeps SUM from
// returning NULL over an empty set, since the old COALESCE is gone.
//
// One semantic narrowing: playsql always excludes soft-deleted rows from an
// aggregate subquery and from a relation-existence check, so IncludeDeleted no
// longer widens this rollup. Deleted expenses are now always out of the total.
// No caller passes IncludeDeleted, so no report changes.
func (s *Server) findExpensesByCategories(ctx context.Context, opts ...ExpenseOption) ([]*expenseCategory, error) {
	companyID := CurrentCompany(ctx).ID
	filter := expenseFilterFrom(opts)
	constrain := applyExpenseFilter(companyID, filter)

	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []expenseCategoryRead
	if err := pdb.Model(&expenseCategoryRead{}).
		Select("id", "name").
		WithSum("Expenses", "amount", playsql.As("total_amount"), playsql.Constrain(constrain)).
		WhereHas("Expenses", constrain).
		WhereEq("company_id", companyID).
		OrderBy("total_amount", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*expenseCategory, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toExpenseCategory())
	}
	return data, nil
}

// findExpensesCategories is the only category read that hides soft-deleted rows,
// so it filters explicitly — expenseCategoryRead carries no softdelete tag.
func (s *Server) findExpensesCategories(ctx context.Context) ([]*expenseCategory, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []expenseCategoryRead
	if err := pdb.Model(&expenseCategoryRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereNull("deleted_at").
		OrderBy("id", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*expenseCategory, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toExpenseCategory())
	}
	return data, nil
}

func (s *Server) findExpenseCategory(ctx context.Context, uuid string) (*expenseCategory, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row expenseCategoryRead
	if err := pdb.Model(&expenseCategoryRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}

	return row.toExpenseCategory(), nil
}

func (s *Server) storeExpense(ctx context.Context, form *StoreExpenseForm) error {
	c, err := s.findExpenseCategory(ctx, form.Category)
	if err != nil {
		return err
	}

	pdb, err := s.play()
	if err != nil {
		return err
	}

	return pdb.Insert(ctx, &expenseInsert{
		CompanyID:  CurrentCompany(ctx).ID,
		CategoryID: c.ID,
		Date:       form.Date,
		Amount:     form.Amount,
		Notes:      form.Notes,
	})
}

// deleteExpense soft-deletes via Update rather than Delete: Builder.Delete stamps
// deleted_at only, and the statement it replaced bumped updated_at too. Update
// stamps updated_at for free because expenseRead maps it, and expenseRead's
// softdelete tag adds the `deleted_at IS NULL` guard the raw statement lacked.
func (s *Server) deleteExpense(ctx context.Context, expenseID string) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&expenseRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", expenseID).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return err
}

func (s *Server) updateExpense(ctx context.Context, expenseID string, form *StoreExpenseForm) error {
	c, err := s.findExpenseCategory(ctx, form.Category)
	if err != nil {
		return err
	}

	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&expenseInsert{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", expenseID).
		Update(ctx, map[string]any{
			"date":        form.Date,
			"amount":      form.Amount,
			"notes":       form.Notes,
			"category_id": c.ID,
		})
	return err
}

func (s *Server) storeExpenseCategory(ctx context.Context, form *StoreExpenseCategoryForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	return pdb.Insert(ctx, &expenseCategoryInsert{
		CompanyID:   CurrentCompany(ctx).ID,
		Name:        form.Name,
		Description: form.Description,
	})
}

// updateExpenseCategory uses the read model so playsql stamps updated_at, which
// the statement it replaced set with NOW().
func (s *Server) updateExpenseCategory(ctx context.Context, uuid string, form *StoreExpenseCategoryForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&expenseCategoryRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		Update(ctx, map[string]any{
			"name":        form.Name,
			"description": form.Description,
		})
	return err
}
