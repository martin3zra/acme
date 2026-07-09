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
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
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
	TaxReceipts         []*taxReceipt       `json:"tax_receipts"`
	ExpenseCategories   []*expenseCategory  `json:"expense_categories"`
	Units               []*unit             `json:"units"`
	HandlesVariants     bool                `json:"handles_variants"`
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

type importFile struct {
	ID            string
	UploadID      string
	UserID        string
	Source        *string
	Status        string
	Phase         *string
	TotalRows     *int
	ProcessedRows *int
	SuccessRows   *int
	FailedRows    *int
	WarningRows   *int
	ErrorMEssage  *string
	CreatedAt     *time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
}

func (s *Server) findCompanies(ctx context.Context) ([]*Company, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []companyRead
	if err := pdb.Model(&companyRead{}).
		WhereEq("account_id", CurrentAccount(ctx)).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*Company, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toCompany())
	}
	return data, nil
}

func (s *Server) findCompanyByUUID(ctx context.Context, uuid string) (*Company, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row companyRead
	if err := pdb.Model(&companyRead{}).
		WhereEq("account_id", CurrentAccount(ctx)).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}
	return row.toCompany(), nil
}

// findCompanyByID deliberately does not scope by account — it is looked up by
// primary key, as it always was.
func (s *Server) findCompanyByID(ctx context.Context, id int) (*Company, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row companyRead
	if err := pdb.Model(&companyRead{}).WhereEq("id", id).First(ctx, &row); err != nil {
		return nil, err
	}
	return row.toCompany(), nil
}

// resolveCompanyID turns an account-scoped company uuid into its id.
//
// Every companies_settings statement used to inline this as
// `WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)`.
// playsql has no subquery predicate, so the lookup is hoisted — which also makes an
// unknown uuid an explicit error. The old UPDATEs compared company_id against a NULL
// subquery result, matched nothing, and reported success; see the tests.
func (s *Server) resolveCompanyID(ctx context.Context, uuid string) (int, error) {
	pdb, err := s.play()
	if err != nil {
		return 0, err
	}

	var row companyRead
	if err := pdb.Model(&companyRead{}).
		Select("id").
		WhereEq("account_id", CurrentAccount(ctx)).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *Server) storeCompany(accountID, userID int, form StoreCompanyForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		company := &companyInsert{
			AccountID:  accountID,
			Name:       form.Name,
			Identifier: form.RNC,
			City:       form.City,
			Address:    form.Address,
		}
		if err := ptx.Insert(context.Background(), company); err != nil {
			return err
		}
		companyID := int(company.ID)

		if err := ptx.Insert(context.Background(), &companyUserInsert{
			CompanyID: companyID,
			UserID:    userID,
			Current:   true,
			Role:      "owner",
		}); err != nil {
			return err
		}

		if err := s.copySharedData(tx, companyID); err != nil {
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
		"vendor": map[string]any{
			"prefix":  "VEND-",
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
		"order": map[string]any{
			"prefix":  "ORD-",
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
		Vendor:   RedirectPreference.List,
		Item:     RedirectPreference.List,
		Payment:  RedirectPreference.List,
		Order:    RedirectPreference.List,
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// A nil updateColumns means "update every inserted column except the conflict
	// columns and created_at", which is exactly the old DO UPDATE SET list. playsql
	// stamps updated_at because companySettingsRead maps it.
	_, err = ptx.Model(&companySettingsRead{}).Upsert(context.Background(),
		[]map[string]any{{
			"company_id":           companyID,
			"sequences":            foundation.ToJSON(defaultSequences),
			"redirect_preferences": foundation.ToJSON(defaultRedirectPreferences),
		}},
		[]string{"company_id"},
		nil,
	)
	return err
}

// Stays raw: a call to a stored procedure, not a statement playsql models.
func (s *Server) copySharedData(tx *sql.Tx, companyID int) error {
	_, err := tx.Exec(`SELECT copy_shared_data($1);`, companyID)
	return err
}

// findCompanySettings reads one company's settings row, projecting only the given
// columns. updated_at is always included — every caller reports it.
func (s *Server) findCompanySettings(ctx context.Context, uuid string, columns ...string) (*companySettingsRead, error) {
	companyID, err := s.resolveCompanyID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row companySettingsRead
	if err := pdb.Model(&companySettingsRead{}).
		Select(append(columns, "updated_at")...).
		WhereEq("company_id", companyID).
		First(ctx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

// updateCompanySettings writes one company's settings row. playsql stamps updated_at.
func (s *Server) updateCompanySettings(ctx context.Context, uuid string, changes map[string]any) error {
	companyID, err := s.resolveCompanyID(ctx, uuid)
	if err != nil {
		return err
	}

	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&companySettingsRead{}).
		WhereEq("company_id", companyID).
		Update(ctx, changes)
	return err
}

func (s *Server) findSequences(ctx context.Context, uuid string) (*CompanySeq, error) {
	row, err := s.findCompanySettings(ctx, uuid, "sequences")
	if err != nil {
		return nil, err
	}

	seq := &CompanySeq{UpdatedAt: row.UpdatedAt}
	if len(row.Sequences) > 0 {
		if err := json.Unmarshal(row.Sequences, &seq.Sequence); err != nil {
			return nil, err
		}
	}
	return seq, nil
}

func (s *Server) findRedirectPreferences(ctx context.Context, uuid string) (*CompanyRedirectPreferences, error) {
	row, err := s.findCompanySettings(ctx, uuid, "redirect_preferences")
	if err != nil {
		return nil, err
	}

	crp := &CompanyRedirectPreferences{UpdatedAt: row.UpdatedAt}
	if len(row.RedirectPreferences) > 0 {
		if err := json.Unmarshal(row.RedirectPreferences, &crp.Redirect); err != nil {
			return nil, err
		}
	}
	return crp, nil
}

func (s *Server) updateSequences(ctx context.Context, uuid string, form *SequenceForm) error {
	return s.updateCompanySettings(ctx, uuid, map[string]any{
		"sequences": foundation.ToJSON(form.CompanySequence),
	})
}

func (s *Server) updateRedirectPreferences(ctx context.Context, uuid string, form *RedirectPreferencesForm) error {
	return s.updateCompanySettings(ctx, uuid, map[string]any{
		"redirect_preferences": foundation.ToJSON(form),
	})
}

// companyHandlesVariants reports whether the current company manages product
// variants (the feature flag that gates the variant UI in the item editor).
func (s *Server) companyHandlesVariants(ctx context.Context) (bool, error) {
	return s.handlesVariantsByCompanyID(ctx, CurrentCompany(ctx).ID)
}

// handlesVariantsByCompanyID reads the flag by company id. Used by SharedProps,
// which runs before the CompanyKey context is populated and so cannot rely on
// CurrentCompany(ctx).
// A company with no settings row reports false rather than erroring, as before.
//
// The old query's COALESCE(handles_variants, false) was dead: the column is NOT NULL
// with a false default. What actually made a company without settings report false
// was the sql.ErrNoRows guard below, now playsql.ErrNotFound.
func (s *Server) handlesVariantsByCompanyID(ctx context.Context, companyID int) (bool, error) {
	pdb, err := s.play()
	if err != nil {
		return false, err
	}

	var row companySettingsRead
	err = pdb.Model(&companySettingsRead{}).
		Select("handles_variants").
		WhereEq("company_id", companyID).
		First(ctx, &row)
	if errors.Is(err, playsql.ErrNotFound) {
		return false, nil
	}
	return row.HandlesVariants, err
}

// findHandlesVariants reads the flag for a company (by account + uuid) for the
// settings screen.
func (s *Server) findHandlesVariants(ctx context.Context, uuid string) (bool, error) {
	row, err := s.findCompanySettings(ctx, uuid, "handles_variants")
	if err != nil {
		return false, err
	}
	return row.HandlesVariants, nil
}

// updateHandlesVariants toggles the flag for a company (by account + uuid).
func (s *Server) updateHandlesVariants(ctx context.Context, uuid string, enabled bool) error {
	return s.updateCompanySettings(ctx, uuid, map[string]any{"handles_variants": enabled})
}

// Stays raw: a call to a stored procedure, not a statement playsql models.
func (s *Server) upsertTaxReceipts(ctx context.Context, form *TaxReceiptsForm) error {
	_, err := s.db.Exec(`
      SELECT upsert_tax_receipts($1, $2::jsonb)
  `, CurrentCompany(ctx).ID, foundation.AsJSON(form.Receipts))
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

// updateUploadSession writes one upload session. playsql stamps updated_at, which
// every one of these statements set with NOW().
func (s *Server) updateUploadSession(id string, changes map[string]any, scope ...func(*playsql.Builder)) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	q := pdb.Model(&uploadSessionRead{}).WhereEq("id", id)
	for _, fn := range scope {
		fn(q)
	}

	_, err = q.Update(context.Background(), changes)
	return err
}

func (s *Server) storeUploadSession(form *UploadSession) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	// created_at/updated_at are stamped by playsql, replacing the literal NOW()s.
	_, err = pdb.Model(&uploadSessionRead{}).Insert(context.Background(), map[string]any{
		"id":        form.ID,
		"user_id":   form.UserID,
		"filename":  form.Filename,
		"file_size": form.FileSize,
		"delimiter": form.Delimiter,
		"encoding":  form.Encoding,
		"status":    form.Status,
	})
	return err
}

// findUploadSession returns (nil, nil) when the session does not exist, as before.
func (s *Server) findUploadSession(id string) (*UploadSession, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row uploadSessionRead
	err = pdb.Model(&uploadSessionRead{}).WhereEq("id", id).First(context.Background(), &row)
	if errors.Is(err, playsql.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row.toUploadSession(), nil
}

func (s *Server) updateUploadStatus(id string, status string) error {
	return s.updateUploadSession(id, map[string]any{"status": status})
}

// Stays raw: `uploaded_chunks = uploaded_chunks + 1` is a self-referencing
// increment, which playsql's Update cannot express.
func (s *Server) incrementUploadedChunks(id string) error {
	_, err := s.db.Exec(`
		UPDATE upload_sessions
		SET uploaded_chunks = uploaded_chunks + 1,
		    updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

// updateTotalChunks only ever sets the count once — the WhereNull guard is the old
// `AND total_chunks IS NULL`.
func (s *Server) updateTotalChunks(id string, total int) error {
	return s.updateUploadSession(id, map[string]any{"total_chunks": total},
		func(b *playsql.Builder) { b.WhereNull("total_chunks") })
}

// failUpload records why an upload failed.
//
// The statement it replaced passed its arguments in the wrong order — `message, id`
// against a query whose $1 was the id and $2 the error message. It therefore matched
// `WHERE id = <the error message>`, updated zero rows, and returned nil: an upload
// failure was never recorded and the session stayed stuck in its previous status.
func (s *Server) failUpload(id string, message string) error {
	return s.updateUploadSession(id, map[string]any{
		"status":        "failed",
		"error_message": message,
	})
}

func (s *Server) storeImport(id string, form *ImportForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&importRead{}).Insert(context.Background(), map[string]any{
		"id":        id,
		"upload_id": form.UploadID,
		"user_id":   UserFromFoundationUser(form.User()).UUID,
		"source":    form.Type,
		"status":    "queued",
	})
	return err
}

func (s *Server) findImportByID(id string) (*importFile, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row importRead
	if err := pdb.Model(&importRead{}).WhereEq("id", id).First(context.Background(), &row); err != nil {
		return nil, err
	}
	return row.toImportFile(), nil
}

func (s *Server) processRows(
	companyID int,
	importID string,
	source string,
	reader *csv.Reader,
	columnMap map[int]string,
	total int,
) error {
	rowNum := 1
	success := 0
	failed := 0
	var err error
	var taxes []*tax
	var unitID = 0
	warnings := []ImportIssue{}

	if source == "items" {
		unitID, err = s.findUnitByBasedQty(companyID)
		if err != nil {
			return err
		}

		taxes, err = s.findTaxesInternal(companyID)
		if err != nil {
			return err
		}
	}

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

		for k, v := range record {
			if str, ok := v.(string); ok && !utf8.ValidString(str) {
				failed++
				_ = database.WithTransaction(s.db, func(tx *sql.Tx) error {
					return s.saveRowIssue(tx, importID, ImportIssue{
						Row:     rowNum,
						Column:  k,
						Level:   IssueLevel.Error,
						Message: fmt.Sprintf("Invalid UTF-8 in %v", k),
						Value:   str,
					})
				})
				continue
			}
		}

		switch source {
		case "items":
			if err := s.storeItemFromRecord(companyID, unitID, taxes, importID, row, rowNum, record, &warnings); err != nil {
				failed++
				continue
			}
		case "vendors":
			if err := s.storeVendorFromRecord(companyID, importID, row, rowNum, record, &warnings); err != nil {
				failed++
				continue
			}
		default: // customers
			if err := s.storeCustomerFromRecord(companyID, importID, row, rowNum, record, &warnings); err != nil {
				failed++
				continue
			}
		}

		success++

		if rowNum%25 == 0 {
			warningCount := 0
			for _, w := range warnings {
				if w.Level == "warning" {
					warningCount++
				}
			}
			if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
				return s.updateProgress(tx, importID, rowNum-1, success, failed, warningCount)
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

	warningCount := 0
	for _, w := range warnings {
		if w.Level == "warning" {
			warningCount++
		}
	}
	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		if err := s.storeImportIssues(tx, importID, warnings); err != nil {
			return err
		}
		return s.updateProgress(tx, importID, rowNum-1, success, failed, warningCount)
	}); err != nil {
		log.Println("updating progress", err)
		return err
	}
	return nil
}

func (s *Server) storeItemFromRecord(companyID, unitID int, taxes []*tax, importID string, row []string, rowNum int, record map[string]any, warnings *[]ImportIssue) error {
	form, err := mapToStoreItemForm(rowNum, record, unitID, taxes, warnings)
	if form == nil {
		if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); err != nil {
			log.Println("saving row error", err)
			return err
		}
		return errors.New("Unable to map the record to the desired type")
	}
	// 🔥 ONE ROW = ONE TRANSACTION
	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return s.storeItemBackground(tx, companyID, form)
	}); err != nil {
		log.Println("storing record", err)
		if saveErr := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); saveErr != nil {
			log.Println("storing record", err)
			return err
		}
	}
	return nil
}

func (s *Server) storeCustomerFromRecord(companyID int, importID string, row []string, rowNum int, record map[string]any, warnings *[]ImportIssue) error {
	form, code, err := mapToStoreCustomerForm(rowNum, record, warnings)
	if form == nil || code == nil {
		if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); err != nil {
			log.Println("saving row error", err)
			return err
		}
		return errors.New("Unable to map the record to the desired type")
	}
	// 🔥 ONE ROW = ONE TRANSACTION
	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return s.storeCustomerInternal(tx, companyID, *code, form)
	}); err != nil {
		log.Println("storing record", err)
		if saveErr := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); saveErr != nil {
			log.Println("storing record", err)
			return err
		}
	}
	return nil
}

func (s *Server) storeVendorFromRecord(companyID int, importID string, row []string, rowNum int, record map[string]any, warnings *[]ImportIssue) error {
	form, code, err := mapToStoreVendorForm(rowNum, record, warnings)
	if form == nil || code == nil {
		if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); err != nil {
			log.Println("saving row error", err)
			return err
		}
		return errors.New("Unable to map the record to the desired type")
	}
	// 🔥 ONE ROW = ONE TRANSACTION — honor the CSV `code` column.
	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		// createdBy 0: the CSV import never maps an opening balance, so the
		// storeVendorOpenBalance path (the only user of createdBy) is not reached.
		return s.storeVendorInternal(tx, companyID, *code, 0, form)
	}); err != nil {
		log.Println("storing record", err)
		if saveErr := database.WithTransaction(s.db, func(tx *sql.Tx) error {
			return s.saveRowIssue(tx, importID, ImportIssue{
				Row:     rowNum,
				Column:  "all",
				Level:   IssueLevel.Error,
				Message: err.Error(),
				Value:   strings.Join(row, ","),
			})
		}); saveErr != nil {
			log.Println("storing record", err)
			return err
		}
	}
	return nil
}

// updateImport writes one import row. imports has no updated_at column, so playsql
// stamps nothing — matching the statements these replace.
func (s *Server) updateImport(id string, changes map[string]any) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&importRead{}).WhereEq("id", id).Update(context.Background(), changes)
	return err
}

func (s *Server) updateProgress(tx *sql.Tx, id string, processed, success, failed, warnings int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	_, err = ptx.Model(&importRead{}).WhereEq("id", id).Update(context.Background(), map[string]any{
		"processed_rows": processed,
		"success_rows":   success,
		"failed_rows":    failed,
		"warning_rows":   warnings,
	})
	return err
}

func issueRow(importID string, i ImportIssue) map[string]any {
	return map[string]any{
		"import_id":   importID,
		"row_number":  i.Row,
		"column_name": i.Column,
		"level":       string(i.Level),
		"message":     i.Message,
		"value":       i.Value,
	}
}

func (s *Server) saveRowIssue(tx *sql.Tx, importID string, i ImportIssue) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	_, err = ptx.Model(&importRowIssue{}).Insert(context.Background(), issueRow(importID, i))
	return err
}

// storeImportIssues bulk-inserts the row issues. InsertMany builds the multi-row
// INSERT, replacing the hand-rolled `($1,$2,...)` placeholder assembly.
func (s *Server) storeImportIssues(
	tx *sql.Tx,
	importID string,
	issues []ImportIssue,
) error {
	if len(issues) == 0 {
		return nil
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	rows := make([]map[string]any, 0, len(issues))
	for _, issue := range issues {
		rows = append(rows, issueRow(importID, issue))
	}

	_, err = ptx.Model(&importRowIssue{}).InsertMany(context.Background(), rows)
	return err
}

func (s *Server) completeImport(ifile *importFile) {
	// The error is ignored here, as it was before.
	_ = s.updateImport(ifile.ID, map[string]any{
		"status":      "completed",
		"finished_at": time.Now(),
	})

	emit(ifile.ID, ImportEvent{
		Type: "completed",
		Data: foundation.ToJSON(map[string]any{
			"total":     ifile.TotalRows,
			"processed": ifile.ProcessedRows,
			"success":   ifile.SuccessRows,
			"failed":    ifile.FailedRows,
			"warning":   ifile.WarningRows,
		}),
	})
}

func (s *Server) failImport(id, msg string) {
	_ = s.updateImport(id, map[string]any{
		"status":        "failed",
		"error_message": msg,
		"finished_at":   time.Now(),
	})

	emit(id, ImportEvent{"failed", msg})
}

func (s *Server) markStarted(importID string) error {
	return s.updateImport(importID, map[string]any{
		"status":     "processing",
		"started_at": time.Now(),
	})
}

func (s *Server) updateTotalRows(importID string, total int) error {
	return s.updateImport(importID, map[string]any{"total_rows": total})
}
