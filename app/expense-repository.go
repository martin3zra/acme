package app

import (
	"context"
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
)

type category struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	foundation.Timestamps
}

type expense struct {
	ID         int       `json:"id"`
	UUID       string    `json:"uuid"`
	Date       time.Time `json:"date"`
	Amount     float64   `json:"amount"`
	Notes      string    `json:"notes"`
	ReceiptURL string    `json:"receipt_url"`
	Category   category  `json:"category"`
	foundation.Timestamps
}

func (s *Server) findExpenses(ctx context.Context) ([]*expense, error) {
	rows, err := s.db.Query(`
    select expenses.id, expenses.uuid, expenses.amount, expenses.notes, expenses.created_at, expenses.updated_at, expenses.deleted_at,
    expenses_categories.id, expenses_categories.uuid, expenses_categories.name
    from expenses
    inner join expenses_categories on (expenses.company_id = expenses_categories.company_id AND expenses.category_id = expenses_categories.id)
    where expenses.company_id = $1
    AND expenses.deleted_at IS NULL
    ORDER BY expenses.id DESC
  `, CurrentCompany(ctx).ID)
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

func (s *Server) findExpensesCategories(ctx context.Context) ([]*category, error) {
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
	data := make([]*category, 0)
	for rows.Next() {
		i := new(category)

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

func (s *Server) findExpenseCategory(ctx context.Context, uuid string) (*category, error) {
	var c category
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
