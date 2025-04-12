package app

import "github.com/martin3zra/acme/pkg/foundation"

type unit struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Add timestamps properties
	foundation.Timestamps
}

func (s *Server) findUnits(companyID int) ([]*unit, error) {
	rows, err := s.db.Query("SELECT id, name, created_at, updated_at, deleted_at FROM units WHERE company_id = $1", companyID)
	if err != nil {
		return nil, err
	}
	data := make([]*unit, 0)
	for rows.Next() {
		u := new(unit)
		if err = rows.Scan(
			&u.ID,
			&u.Name,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, u)

	}
	return data, nil
}
