package app

import "github.com/martin3zra/acme/pkg/foundation"

type Company struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Identifier *string `json:"identifier"`
	City       string  `json:"city"`
	Address    string  `json:"address"`
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
