package app

import (
	"database/sql"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type Company struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	City       string `json:"city"`
	Address    string `json:"address"`
	foundation.Timestamps
}

func (s *Server) findCompanyById(id int) (*Company, error) {
	result := s.db.QueryRow(s.qs.Q("companies_find_by_id"), id)
	var company Company
	err := result.Scan(
		&company.ID,
		&company.Name,
		&company.Identifier,
		&company.City,
		&company.Address,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *Server) storeCompany(userID int, form StoreCompanyForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var companyID int
		stmt, err := tx.Prepare("INSERT INTO companies (name, identifier, city, address) VALUES($1, $2, $3, $4) RETURNING id")
		if err != nil {
			return err
		}

		if err = stmt.QueryRow(form.Name, form.RNC, form.City, form.Address).Scan(&companyID); err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO companies_users (company_id, user_id, current, role) VALUES($1, $2, $3, $4)",
			companyID, userID, true, "owner")
		return err
	})
}

func (s *Server) currentCompany(userID int) (*Company, error) {
	result := s.db.QueryRow(`
    SELECT id, name, identifier, city, address, created_at, updated_at
    FROM companies
    JOIN companies_users ON companies.id = companies_users.company_id
    WHERE companies_users.user_id = $1 AND companies_users.current = true
  `, userID)
	var company Company
	err := result.Scan(
		&company.ID,
		&company.Name,
		&company.Identifier,
		&company.City,
		&company.Address,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &company, nil
}
