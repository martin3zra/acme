package app

import (
	"context"
	"fmt"
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
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

func (s *Server) findExpenseByUUID(ctx context.Context, uuid string) (*expense, error) {
	var c expense
	err := s.db.QueryRow(`
   select expenses.id, expenses.uuid, expenses.amount, expenses.notes, expenses.date, expenses.created_at, expenses.updated_at, expenses.deleted_at,
    expenses_categories.id, expenses_categories.uuid, expenses_categories.name
    from expenses
    inner join expenses_categories on (expenses.company_id = expenses_categories.company_id AND expenses.category_id = expenses_categories.id)
    where expenses.company_id = $1
    and expenses.uuid = $2
  `, CurrentCompany(ctx).ID, uuid).Scan(
		&c.ID,
		&c.UUID,
		&c.Amount,
		&c.Notes,
		&c.Date.Time,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
		&c.Category.ID,
		&c.Category.UUID,
		&c.Category.Name,
	)
	return &c, err
}

func (s *Server) findExpenses(ctx context.Context, opts ...ExpenseOption) ([]*expense, error) {
	companyID := CurrentCompany(ctx).ID

	// Default filter
	filter := ExpenseFilter{}

	// Apply options
	for _, opt := range opts {
		opt(&filter)
	}

	query := `
    select expenses.id, expenses.uuid, expenses.amount, expenses.notes, expenses.date, expenses.created_at, expenses.updated_at, expenses.deleted_at,
    expenses_categories.id, expenses_categories.uuid, expenses_categories.name
    from expenses
    inner join expenses_categories on (expenses.company_id = expenses_categories.company_id AND expenses.category_id = expenses_categories.id)
    where expenses.company_id = $1
    AND expenses.deleted_at IS NULL
  `
	args := []any{companyID}
	argPos := 2

	// Soft delete handling
	if !filter.IncludeDeleted {
		query += " AND expenses.deleted_at IS NULL"
	}

	// Date filters
	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND expenses.date >= $%d", argPos)
		args = append(args, *filter.FromDate)
		argPos++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND expenses.date <= $%d", argPos)
		args = append(args, *filter.ToDate)
		argPos++
	}

	// Category filter
	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND expenses.category_id = $%d", argPos)
		args = append(args, *filter.CategoryID)
		argPos++
	}

	query += " ORDER BY expenses.id DESC"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	data := make([]*expense, 0)
	for rows.Next() {
		i := new(expense)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Amount,
			&i.Notes,
			&i.Date.Time,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.Category.ID,
			&i.Category.UUID,
			&i.Category.Name,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findExpensesByCategories(ctx context.Context, opts ...ExpenseOption) ([]*expenseCategory, error) {
	companyID := CurrentCompany(ctx).ID

	// Default filter
	filter := ExpenseFilter{}

	// Apply options
	for _, opt := range opts {
		opt(&filter)
	}

	query := `
    select expenses_categories.id,
		expenses_categories.name,
		COALESCE(SUM(expenses.amount), 0) as total_amount
    from expenses
    inner join expenses_categories on (expenses.company_id = expenses_categories.company_id AND expenses.category_id = expenses_categories.id)
    where expenses.company_id = $1
  `
	args := []any{companyID}
	argPos := 2

	// Soft delete handling
	if !filter.IncludeDeleted {
		query += " AND expenses.deleted_at IS NULL"
	}

	// Date filters
	if filter.FromDate != nil && filter.ToDate != nil {
		query += fmt.Sprintf(" AND expenses.date BETWEEN $%d AND $%d", argPos, argPos+1)
		args = append(args, *filter.FromDate)
		args = append(args, *filter.ToDate)
		argPos += 2
	} else {

		if filter.FromDate != nil {
			query += fmt.Sprintf(" AND expenses.date >= $%d::date", argPos)
			args = append(args, *filter.FromDate)
			argPos++
		}

		if filter.ToDate != nil {
			query += fmt.Sprintf(" AND expenses.date <= $%d::date", argPos)
			args = append(args, *filter.ToDate)
			argPos++
		}
	}

	// Category filter
	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND expenses.category_id = $%d", argPos)
		args = append(args, *filter.CategoryID)
		argPos++
	}

	query += `
    GROUP BY expenses_categories.id, expenses_categories.name
    ORDER BY total_amount DESC
`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	data := make([]*expenseCategory, 0)
	for rows.Next() {
		i := new(expenseCategory)

		if err = rows.Scan(
			&i.ID,
			&i.Name,
			&i.TotalAmount,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findExpensesCategories(ctx context.Context) ([]*expenseCategory, error) {
	rows, err := s.db.Query(`
    select id, uuid, name, description, created_at, updated_at, deleted_at
    from expenses_categories
    where company_id = $1
    AND deleted_at IS NULL
    ORDER BY id DESC
  `, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*expenseCategory, 0)
	for rows.Next() {
		i := new(expenseCategory)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Name,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findExpenseCategory(ctx context.Context, uuid string) (*expenseCategory, error) {
	var c expenseCategory
	err := s.db.QueryRow(`
   select id, uuid, name, description, created_at, updated_at, deleted_at
    from expenses_categories
    where company_id = $1
    AND uuid = $2
  `, CurrentCompany(ctx).ID, uuid).Scan(
		&c.ID,
		&c.UUID,
		&c.Name,
		&c.Description,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	return &c, err
}

func (s *Server) storeExpense(ctx context.Context, form *StoreExpenseForm) error {

	c, err := s.findExpenseCategory(ctx, form.Category)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO expenses (company_id, category_id, date, amount, notes) VALUES($1, $2, $3, $4, $5)",
		CurrentCompany(ctx).ID, c.ID, form.Date, form.Amount, form.Notes)
	return err
}

func (s *Server) deleteExpense(ctx context.Context, expenseID string) error {
	_, err := s.db.Exec(
		"UPDATE expenses SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND uuid = $2",
		CurrentCompany(ctx).ID, expenseID,
	)

	return err
}

func (s *Server) updateExpense(ctx context.Context, expenseID string, form *StoreExpenseForm) error {

	c, err := s.findExpenseCategory(ctx, form.Category)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"UPDATE expenses SET date = $1, amount = $2, notes = $3, category_id = $4 WHERE company_id = $5 AND uuid = $6",
		form.Date, form.Amount, form.Notes, c.ID, CurrentCompany(ctx).ID, expenseID,
	)

	return err
}

func (s *Server) storeExpenseCategory(ctx context.Context, form *StoreExpenseCategoryForm) error {
	_, err := s.db.Exec("INSERT INTO expenses_categories (company_id, name, description) VALUES($1, $2, $3)",
		CurrentCompany(ctx).ID, form.Name, form.Description)
	return err
}

func (s *Server) updateExpenseCategory(ctx context.Context, uuid string, form *StoreExpenseCategoryForm) error {
	_, err := s.db.Exec("UPDATE expenses_categories SET name = $3, description = $4, updated_at = NOW() WHERE company_id = $1 AND uuid = $2",
		CurrentCompany(ctx).ID, uuid, form.Name, form.Description)
	return err
}
