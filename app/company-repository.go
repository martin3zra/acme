package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

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

func (s *Server) findCompanyByID(ctx context.Context, id int) (*Company, error) {
	c := new(Company)
	if err := s.db.QueryRow(`
  SELECT id, uuid, name, identifier, city, address, created_at, updated_at, deleted_at
  FROM companies
  WHERE id = $1
  `, id).Scan(
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
		"template": map[string]any{
			"prefix":  "TPL",
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

func (s *Server) storeUploadSession(form *UploadSession) error {
	_, err := s.db.Exec(`
    INSERT INTO upload_sessions (id, user_id, filename, file_size, delimiter, encoding, status, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`, form.ID, form.UserID, form.Filename, form.FileSize, form.Delimiter, form.Encoding, form.Status)
	return err
}

func (s *Server) findUploadSession(id string) (*UploadSession, error) {

	var sess UploadSession

	err := s.db.QueryRow(`
		SELECT
			id, user_id, filename, file_size, delimiter, encoding, status,
			total_chunks, uploaded_chunks, error_message,
			created_at, updated_at
		FROM upload_sessions
		WHERE id = $1
	`, id).Scan(
		&sess.ID,
		&sess.UserID,
		&sess.Filename,
		&sess.FileSize,
		&sess.Delimiter,
		&sess.Encoding,
		&sess.Status,
		&sess.TotalChunks,
		&sess.UploadedChunks,
		&sess.ErrorMessage,
		&sess.CreatedAt,
		&sess.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &sess, nil
}

func (s *Server) updateUploadStatus(id string, status string) error {
	_, err := s.db.Exec(`
		UPDATE upload_sessions
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, status)
	return err
}

func (s *Server) incrementUploadedChunks(id string) error {
	_, err := s.db.Exec(`
		UPDATE upload_sessions
		SET uploaded_chunks = uploaded_chunks + 1,
		    updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (s *Server) updateTotalChunks(id string, total int) error {
	_, err := s.db.Exec(`
		UPDATE upload_sessions
		SET total_chunks = $2,
		    updated_at = NOW()
		WHERE id = $1 AND total_chunks IS NULL
	`, id, total)
	return err
}

func (s *Server) failUpload(id string, message string) error {
	_, err := s.db.Exec(`
		UPDATE upload_sessions
		SET status = 'failed',
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, message, id)
	return err
}

func (s *Server) storeImport(id string, form *ImportForm) error {
	_, err := s.db.Exec(`
    INSERT INTO imports (id, upload_id, user_id, status)
    VALUES ($1, $2, $3, 'queued')`, id, form.UploadID, form.User().UUID)
	return err
}

func (s *Server) processRows(
	companyID int,
	importID string,
	reader *csv.Reader,
	columnMap map[int]string,
	total int,
) error {
	rowNum := 1
	success := 0
	failed := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++

		record := map[string]any{}
		for i, field := range row {
			if col, ok := columnMap[i]; ok {
				record[col] = strings.TrimSpace(field)
			}
		}

		// Validation BEFORE transaction
		if record["name"] == "" {
			failed++
			if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
				return s.saveRowError(tx, importID, rowNum, "Name is required", row)
			}); err != nil {
				log.Println("saving row error", err)
			}
			continue
		}

		for k, v := range record {
			if str, ok := v.(string); ok && !utf8.ValidString(str) {
				failed++
				_ = database.WithTransaction(s.db, func(tx *sql.Tx) error {
					return s.saveRowError(tx, importID, rowNum, "Invalid UTF-8 in "+k, row)
				})
				continue
			}
		}

		form := mapToStoreItemForm(record)
		if form == nil {
			failed++
			if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
				return s.saveRowError(tx, importID, rowNum, "Data mismatch", row)
			}); err != nil {
				log.Println("saving row error", err)
			}
			continue
		}

		// 🔥 ONE ROW = ONE TRANSACTION
		if err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.storeItemBackground(tx, companyID, form)
		}); err != nil {
			log.Println("storing record", err)
			if saveErr := database.WithTransaction(s.db, func(tx *sql.Tx) error {
				return s.saveRowError(tx, importID, rowNum, err.Error(), row)
			}); saveErr != nil {
				log.Println("storing record", err)
				failed++
				continue
			}
		}

		success++

		if rowNum%25 == 0 {
			if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
				return s.updateProgress(tx, importID, rowNum-1, success, failed)
			}); err != nil {
				log.Println("updating progress", err)
			}

			emit(importID, ImportEvent{
				"progress",
				map[string]int{
					"processed": rowNum - 1,
					"total":     total,
				},
			})
		}
	}

	return nil
}

// func (s *Server) processRows(
// 	companyID int,
// 	importID string,
// 	reader *csv.Reader,
// 	columnMap map[int]string,
// 	total int,
// ) error {
// 	rowNum := 1
// 	success := 0
// 	failed := 0

// 	return database.WithTransaction(s.db, func(tx *sql.Tx) error {

// 		for {
// 			row, err := reader.Read()
// 			if err == io.EOF {
// 				break
// 			}

// 			rowNum++

// 			record := map[string]any{}
// 			for i, field := range row {
// 				if col, ok := columnMap[i]; ok {
// 					record[col] = strings.TrimSpace(field)
// 				}
// 			}

// 			if record["name"] == "" {
// 				failed++
// 				s.saveRowError(tx, importID, rowNum, "Name is required", row)
// 				continue
// 			}

// 			form := mapToStoreItemForm(record)
// 			if form == nil {
// 				failed++
// 				s.saveRowError(tx, importID, rowNum, "Data mismatch", row)
// 				continue
// 			}
// 			err = s.storeItemBackground(tx, companyID, form)
// 			if err != nil {
// 				log.Println("error storing record", err)
// 				failed++
// 				s.saveRowError(tx, importID, rowNum, err.Error(), row)
// 				continue
// 			}

// 			success++

// 			if rowNum%25 == 0 {
// 				s.updateProgress(tx, importID, rowNum-1, success, failed)
// 				emit(importID, ImportEvent{
// 					"progress",
// 					map[string]int{
// 						"processed": rowNum - 1,
// 						"total":     total,
// 					},
// 				})
// 			}
// 		}
// 		return nil
// 	})

// }

func (s *Server) updateProgress(tx *sql.Tx, id string, processed, success, failed int) error {
	_, err := tx.Exec(`
		UPDATE imports
		SET processed_rows=$2, success_rows=$3, failed_rows=$4
		WHERE id=$1
	`, id, processed, success, failed)
	return err
}

func (s *Server) saveRowError(tx *sql.Tx, importID string, row int, msg string, data []string) error {
	_, err := tx.Exec(`
		INSERT INTO import_row_errors (import_id, row_number, error, raw_data)
		VALUES ($1, $2, $3, $4)
	`, importID, row, msg, strings.Join(data, ","))
	return err
}

func (s *Server) completeImport(id string) {
	s.db.Exec(`
		UPDATE imports
		SET status='completed', finished_at=now()
		WHERE id=$1
	`, id)

	emit(id, ImportEvent{"completed", nil})
}

func (s *Server) failImport(id, msg string) {
	s.db.Exec(`
		UPDATE imports
		SET status='failed', error_message=$2, finished_at=now()
		WHERE id=$1
	`, id, msg)

	emit(id, ImportEvent{"failed", msg})
}

func (s *Server) markStarted(importID string) error {
	_, err := s.db.Exec(`
		UPDATE imports
		SET status = 'processing',
		    started_at = now()
		WHERE id = $1
	`, importID)
	return err
}

func (s *Server) updateTotalRows(importID string, total int) error {
	_, err := s.db.Exec(`
		UPDATE imports
		SET total_rows = $2
		WHERE id = $1
	`, importID, total)
	return err
}
