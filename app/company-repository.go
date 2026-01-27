package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type Company struct {
	ID                  int                 `json:"id"`
	UUID                string              `json:"uuid"`
	Name                string              `json:"name"`
	Identifier          string              `json:"identifier"`
	City                string              `json:"city"`
	Address             string              `json:"address"`
	Taxes               []*tax              `json:"taxes"`
	Sequences           *CompanySequence    `json:"sequences"`
	SeqLastUpdatedAt    *time.Time          `json:"seq_last_updated_at"`
	RedirectPreferences RedirectPreferences `json:"redirect_preferences"`
	UserRole            string              `json:"_"`
	foundation.Timestamps
}

type CompanySeq struct {
	Sequence  CompanySequence `json:"sequence"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (d *CompanySequence) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *CompanySequence) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type CompanyRedirectPreferences struct {
	Redirect  RedirectPreferences `json:"redirect_preferences"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func (d *RedirectPreferences) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *RedirectPreferences) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

func (s *Server) findCompanies(ctx context.Context) ([]*Company, error) {
	rows, err := s.db.Query("SELECT id, uuid, name, identifier, city, address, created_at, updated_at, deleted_at FROM companies WHERE account_id = $1", CurrentAccount(ctx))
	if err != nil {
		return nil, err
	}
	data := make([]*Company, 0)
	for rows.Next() {
		c := new(Company)
		if err = rows.Scan(
			&c.ID,
			&c.UUID,
			&c.Name,
			&c.Identifier,
			&c.City,
			&c.Address,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, err
		}
		data = append(data, c)
	}
	return data, nil
}

func (s *Server) findCompanyByUUID(ctx context.Context, uuid string) (*Company, error) {
	c := new(Company)
	if err := s.db.QueryRow(`
  SELECT id, uuid, name, identifier, city, address, created_at, updated_at, deleted_at
  FROM companies
  WHERE account_id = $1
  AND uuid = $2
  `, CurrentAccount(ctx), uuid).Scan(
		&c.ID,
		&c.UUID,
		&c.Name,
		&c.Identifier,
		&c.City,
		&c.Address,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Server) storeCompany(accountID, userID int, form StoreCompanyForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var companyID int
		stmt, err := tx.Prepare("INSERT INTO companies (account_id, name, identifier, city, address) VALUES($1, $2, $3, $4, $5) RETURNING id")
		if err != nil {
			return err
		}

		if err = stmt.QueryRow(accountID, form.Name, form.RNC, form.City, form.Address).Scan(&companyID); err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO companies_users (company_id, user_id, current, role) VALUES($1, $2, $3, $4)",
			companyID, userID, true, "owner")
		if err != nil {
			return err
		}

		return s.linkCompanyDefaultSequences(tx, companyID)
	})
}

func (s *Server) linkCompanyDefaultSequences(tx *sql.Tx, companyID int) error {
	defaultSequences := map[string]any{
		"invoice": map[string]any{
			"cash": map[string]any{
				"prefix":  "INV-CO-",
				"next":    1,
				"padding": 4,
				"format":  "{prefix}-{year}-{seq}",
			},
			"credit": map[string]any{
				"prefix":  "INV-CRE-",
				"next":    1,
				"padding": 4,
				"format":  "{prefix}-{year}-{seq}",
			},
		},
		"payment": map[string]any{
			"prefix":  "PAY-",
			"next":    1,
			"padding": 4,
			"format":  "{prefix}-{year}-{seq}",
		},
		"customer": map[string]any{
			"prefix":  "CUST-",
			"next":    1,
			"padding": 6,
			"format":  "{prefix}-{year}-{seq}",
		},
		"estimate": map[string]any{
			"prefix":  "EST-",
			"next":    1,
			"padding": 6,
			"format":  "{prefix}-{year}-{seq}",
		},
	}

	defaultRedirectPreferences := RedirectPreferences{
		Invoice:  RedirectPreference.Stay,
		Estimate: RedirectPreference.Stay,
		Customer: RedirectPreference.List,
		Item:     RedirectPreference.List,
		Payment:  RedirectPreference.List,
		Order:    RedirectPreference.List,
	}

	_, err := tx.Exec(`
    INSERT INTO companies_settings(company_id, sequences, redirect_preferences)
    VALUES($1, $2, $3)
    ON CONFLICT(company_id) DO UPDATE SET sequences = $2, redirect_preferences = $3, updated_at = now()`,
		companyID, foundation.ToJSON(defaultSequences), foundation.ToJSON(defaultRedirectPreferences),
	)
	return err
}

func (s *Server) findSequences(ctx context.Context, uuid string) (*CompanySeq, error) {
	var seq CompanySeq
	err := s.db.QueryRow(`
    SELECT sequences, updated_at
    FROM companies_settings
    WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)
  `, CurrentAccount(ctx), uuid).
		Scan(&seq.Sequence, &seq.UpdatedAt)
	return &seq, err
}

func (s *Server) findRedirectPreferences(ctx context.Context, uuid string) (*CompanyRedirectPreferences, error) {
	var crp CompanyRedirectPreferences
	err := s.db.QueryRow(`
    SELECT redirect_preferences, updated_at
    FROM companies_settings
    WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)
  `, CurrentAccount(ctx), uuid).
		Scan(&crp.Redirect, &crp.UpdatedAt)
	return &crp, err
}

func (s *Server) updateSequences(ctx context.Context, uuid string, form *SequenceForm) error {
	_, err := s.db.Exec(`
    UPDATE companies_settings
    SET sequences = $3, updated_at = now()
    WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)`, CurrentAccount(ctx), uuid, foundation.ToJSON(form.CompanySequence))
	return err
}

func (s *Server) updateRedirectPreferences(ctx context.Context, uuid string, form *RedirectPreferencesForm) error {
	_, err := s.db.Exec(`
    UPDATE companies_settings
    SET redirect_preferences = $3, updated_at = now()
    WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)`, CurrentAccount(ctx), uuid, foundation.ToJSON(form))
	return err
}

func CheckResourcePrerequisites(ctx context.Context, resource string, companyID int) (context.Context, error) {
	cache, _ := ctx.Value(prereqCacheKey).(prereqCache)
	cacheKey := fmt.Sprintf("%s:%d", resource, companyID)

	if cache != nil {
		if cached, ok := cache[cacheKey]; ok {
			if cached.Ok {
				return ctx, nil
			}
			return ctx, ErrPrerequisitesMissing
		}
	}

	db := ctx.Value(database.ConnectionKey{}).(*sql.DB)
	var raw json.RawMessage

	err := db.QueryRowContext(ctx, `SELECT check_resource_prerequisites($1, $2)`, resource, companyID).Scan(&raw)

	if err != nil {
		// 3️⃣ Convert DB errors → domain errors
		if err == sql.ErrNoRows {
			return ctx, ErrSettingsNotFound
		}

		// PostgreSQL error handling
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23514": // check_violation
				return ctx, ErrInvalidConfiguration
			default:
				return ctx, fmt.Errorf("db error: %w", err)
			}
		}

		return ctx, err
	}

	var result PrerequisiteResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return ctx, err
	}

	if cache == nil {
		cache = make(prereqCache)
		ctx = context.WithValue(ctx, prereqCacheKey, cache)
	}
	cache[cacheKey] = result

	if !result.Ok {
		return ctx, ErrPrerequisitesMissing
	}

	return ctx, nil
}
