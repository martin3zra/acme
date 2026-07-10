package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

// Phase 2 of the playsql adoption: production writes. This file is the first
// production code to link playsql (and, transitively, its SQL drivers). Write
// paths convert one at a time; everything not yet migrated stays raw database/sql.

// playTx wraps an in-flight *sql.Tx with playsql under the Postgres grammar, so a
// production write can use typed models on the *same* transaction the caller
// already opened (via forge database.WithTransaction). The caller still owns the
// tx and its commit/rollback.
func playTx(tx *sql.Tx) (*playsql.Tx, error) {
	return playsql.UseTx(tx, "postgres")
}

// play returns the process-wide playsql read executor built once in Boot
// (configurePlan). It keeps its (*playsql.DB, error) signature so the ~100 read
// call sites are untouched; the error is always nil now that construction moved
// to Boot. Read paths that are not inside a transaction (list/detail queries) use
// dedicated *Read models below (the JSON response structs can't double as playsql
// models: they embed foundation.Timestamps, which the parser skips, and hold
// pointer-to-struct fields it would misread as relations).
func (s *Server) play() (*playsql.DB, error) {
	return s.plan, nil
}

// playOn wraps a request-scoped *sql.DB with playsql. Unlike (*Server).play, which
// returns the process-wide plan built once in Boot, these callers are handed a db
// they do not own: the forge auth resolvers (NewAuth reads the connection from the
// request context), the validator (StoreProfileForm.Rules), and the *User/*AuthUser
// methods invoked with the caller's db. In tests that handle is the per-test
// transaction, so this must wrap the exact db passed — not the cached plan, which
// would pin these paths to one Server's connection and break test isolation.
func playOn(db *sql.DB) (*playsql.DB, error) {
	return playsql.Use(db, "postgres")
}

// withTrashedRelation opts an eager-loaded relation into soft-deleted rows. Use it
// wherever the raw SQL used a plain INNER JOIN with no deleted_at predicate but the
// related read model is softdelete-tagged (customerModel, vendorRead): without it the
// parent row comes back with a nil relation once the related row is soft-deleted.
func withTrashedRelation(b *playsql.Builder) { b.WithTrashed() }

// customerModel is the playsql read model for the customers table. Only real
// columns are mapped (db tags); deleted_at carries play:"softdelete" so queries
// exclude soft-deleted rows automatically, matching the old "deleted_at IS NULL".
type customerModel struct {
	ID            int        `db:"id" play:"pk,incrementing"`
	CompanyID     int        `db:"company_id"`
	UUID          string     `db:"uuid" play:"guarded"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	ContactName   string     `db:"contact_name"`
	Phone         string     `db:"phone"`
	Email         string     `db:"email"`
	Status        string     `db:"status"`
	AmountDue     float64    `db:"amount_due"`
	Address       string     `db:"address"`
	CustomerType  string     `db:"customer_type"`
	PaymentMethod string     `db:"payment_method"`
	CreditLimited bool       `db:"credit_limited"`
	CreditLimit   float64    `db:"credit_limit"`
	PaymentTerms  string     `db:"payment_terms"`
	TaxReceiptID  *int       `db:"tax_receipt_id"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" play:"softdelete"`

	// OpeningInvoice is the customer's opening-balance invoice. The raw read got it
	// from a LEFT JOIN onto a subquery filtered to type = 'opening'; the filter now
	// lives in withOpeningInvoice. Neither scopes by company or filters deleted_at.
	OpeningInvoice *invoiceRead `play:"hasOne,fk=customer_id"`
}

func (customerModel) TableName() string { return "customers" }

// withOpeningInvoice constrains the eager load to the opening-balance invoice, the
// way the old LEFT JOIN's subquery did.
// withOpeningInvoice narrows the eager load to the opening invoice, and to the four
// columns the subquery it replaces selected.
//
// customer_id is in the projection because it is load-bearing, not decorative: the
// eager loader groups the children it fetches by their foreign-key *field*, so a
// customer_id left out of the SELECT scans as 0 and no child ever matches its parent.
// The relation would come back nil, silently, for every customer.
func withOpeningInvoice(b *playsql.Builder) {
	b.Select("id", "customer_id", "date", "amount").WhereEq("type", "opening")
}

// openBalanceOfInvoice builds the customer's opening balance. A customer without an
// opening invoice gets a struct of nil pointers, exactly as the LEFT JOIN's NULLs did.
func openBalanceOfInvoice(inv *invoiceRead) *OpenBalance {
	ob := new(OpenBalance)
	if inv == nil {
		return ob
	}
	id, date, amount := inv.ID, inv.Date, inv.Amount
	ob.InvoiceID, ob.Date, ob.Amount = &id, &date, &amount
	return ob
}

// toCustomer maps the read model onto the JSON response struct.
func (r customerModel) toCustomer() *customer {
	c := &customer{
		ID:            r.ID,
		UUID:          r.UUID,
		Code:          r.Code,
		Name:          r.Name,
		ContactName:   r.ContactName,
		Phone:         r.Phone,
		Email:         r.Email,
		AmountDue:     r.AmountDue,
		Address:       r.Address,
		CustomerType:  r.CustomerType,
		PaymentMethod: r.PaymentMethod,
		CreditLimited: r.CreditLimited,
		CreditLimit:   r.CreditLimit,
		PaymentTerms:  r.PaymentTerms,
		TaxReceipt:    r.TaxReceiptID,
		Status:        foundation.Status(r.Status),
	}
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt
	c.DeletedAt = r.DeletedAt
	return c
}

// vendorRead is the playsql read model for the vendors table — same pattern as
// customerModel (real columns only, softdelete on deleted_at).
type vendorRead struct {
	ID            int        `db:"id" play:"pk,incrementing"`
	UUID          string     `db:"uuid"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	ContactName   string     `db:"contact_name"`
	Phone         string     `db:"phone"`
	Email         string     `db:"email"`
	Status        string     `db:"status"`
	AmountPayable float64    `db:"amount_payable"`
	PurchaseNote  string     `db:"purchase_note"`
	LeadTimeDays  int        `db:"lead_time_days"`
	Address       string     `db:"address"`
	VendorType    string     `db:"vendor_type"`
	PaymentMethod string     `db:"payment_method"`
	PaymentTerms  string     `db:"payment_terms"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" play:"softdelete"`

	// OpeningPayable is the vendor's opening-balance AP entry. The raw read got it
	// from a LEFT JOIN onto a subquery filtered to status = 'draft'.
	OpeningPayable *accountsPayableRead `play:"hasOne,fk=vendor_id"`
}

func (vendorRead) TableName() string { return "vendors" }

// withOpeningPayable constrains the eager load to the draft AP entry, the way the old
// LEFT JOIN's subquery did.
// withOpeningPayable narrows the eager load to the draft payable. vendor_id is in the
// projection for the same reason customer_id is in withOpeningInvoice's: it is the key
// the loader matches children to parents on.
func withOpeningPayable(b *playsql.Builder) {
	b.Select("id", "vendor_id", "invoice_date", "amount_total").WhereEq("status", "draft")
}

// openBalanceOfPayable builds the vendor's opening balance. Nil pointers when there is
// no draft entry, matching the LEFT JOIN's NULLs.
func openBalanceOfPayable(ap *accountsPayableRead) *OpenBalance {
	ob := new(OpenBalance)
	if ap == nil {
		return ob
	}
	id := int(ap.ID)
	date, amount := ap.InvoiceDate, ap.AmountTotal
	ob.InvoiceID, ob.Date, ob.Amount = &id, &date, &amount
	return ob
}

// receivableRead is the playsql read model for the receivables table, the
// cross-reference between a credit invoice and its customer. softdelete matches the
// `receivables.deleted_at IS NULL` the list read carried.
type receivableRead struct {
	ID         int        `db:"id" play:"pk,incrementing"`
	CompanyID  int        `db:"company_id"`
	UUID       string     `db:"uuid"`
	InvoiceID  int        `db:"invoice_id"`
	CustomerID int        `db:"customer_id"`
	DeletedAt  *time.Time `db:"deleted_at" play:"softdelete"`

	Invoice *invoiceRead `play:"belongsTo,fk=invoice_id"`
}

func (receivableRead) TableName() string { return "receivables" }

// toReceivable maps a receivables row plus its invoice onto the response struct.
// receivable.ID/UUID identify the receivables row; receivable.Invoice.ID/UUID
// identify the invoice. Notes and Payment were never selected and stay zero.
//
// Invoice cannot be nil in practice — the read filters on it through WhereRelation,
// which is an EXISTS — but a nil relation must not panic the way the old Scan into
// &row.Invoice.UUID could not.
func (r receivableRead) toReceivable() *receivable {
	out := new(receivable)
	out.ID = r.ID
	out.UUID = r.UUID
	if r.Invoice == nil {
		return out
	}

	out.Invoice.ID = r.Invoice.ID
	out.Invoice.UUID = r.Invoice.UUID
	out.Invoice.Number = r.Invoice.Code
	out.Invoice.NCF = r.Invoice.TaxNumber
	out.Invoice.Date = r.Invoice.Date
	out.Invoice.DueOn = r.Invoice.DueOn
	out.Invoice.Total = r.Invoice.Total
	out.Invoice.AmountDue = r.Invoice.AmountDue
	out.Invoice.PaidStatus = r.Invoice.PaidStatus
	return out
}

// toVendor maps the read model onto the JSON response struct.
func (r vendorRead) toVendor() *vendor {
	v := &vendor{
		ID:            r.ID,
		UUID:          r.UUID,
		Code:          r.Code,
		Name:          r.Name,
		ContactName:   r.ContactName,
		Phone:         r.Phone,
		Email:         r.Email,
		PurchaseNote:  r.PurchaseNote,
		LeadTimeDays:  r.LeadTimeDays,
		AmountPayable: r.AmountPayable,
		Address:       r.Address,
		VendorType:    r.VendorType,
		PaymentMethod: r.PaymentMethod,
		PaymentTerms:  r.PaymentTerms,
		Status:        foundation.Status(r.Status),
	}
	v.CreatedAt = r.CreatedAt
	v.UpdatedAt = r.UpdatedAt
	v.DeletedAt = r.DeletedAt
	return v
}

// unitRead is the playsql read model for the units table. Unlike customerModel/
// vendorRead it carries NO play:"softdelete" tag: the original unit reads never
// filtered deleted_at (units are not soft-deleted — the repo has no delete path),
// so the model maps deleted_at as a plain column to keep "return all rows".
type unitRead struct {
	ID        int64      `db:"id" play:"pk,incrementing"`
	CompanyID int        `db:"company_id"`
	Name      string     `db:"name"`
	BaseQty   int        `db:"base_qty"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (unitRead) TableName() string { return "units" }

// toUnit maps the read model onto the JSON response struct.
func (r unitRead) toUnit() *unit {
	u := &unit{
		ID:      r.ID,
		Name:    r.Name,
		BaseQty: r.BaseQty,
	}
	u.CreatedAt = r.CreatedAt
	u.UpdatedAt = r.UpdatedAt
	u.DeletedAt = r.DeletedAt
	return u
}

// userRead is the playsql read model for the users table.
//
// This model is the structural fix for the bug in 53c9d58: playsql selects and
// scans by column name, so a migration that appends a column (as remember_token
// did) can no longer shift a positional scan. remember_token is simply not mapped
// here, and therefore never selected.
//
// deleted_at carries no softdelete tag — none of the user reads filtered it.
// Linked is play:"readonly": it holds the WithCount aggregate that replaced the old
// correlated `(SELECT COUNT(*) ...) as linked` subquery.
type userRead struct {
	ID                 int        `db:"id" play:"pk,incrementing"`
	Name               string     `db:"name"`
	Email              string     `db:"email"`
	Password           string     `db:"password"`
	EmailVerifiedAt    *time.Time `db:"email_verified_at"`
	LastPasswordReset  *time.Time `db:"last_password_reset"`
	CreatedAt          *time.Time `db:"created_at"`
	UpdatedAt          *time.Time `db:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at"`
	UUID               string     `db:"uuid"`
	Status             string     `db:"status"`
	MustChangePassword bool       `db:"must_change_password"`
	PendingEmail       *string    `db:"pending_email"`
	Linked             int        `db:"linked" play:"readonly"`

	Companies []*companyRead `play:"belongsToMany,pivot=companies_users,foreignPivotKey=user_id,relatedPivotKey=company_id"`
	Accounts  []*accountRead `play:"belongsToMany,pivot=accounts_users,foreignPivotKey=user_id,relatedPivotKey=account_id"`
}

func (userRead) TableName() string { return "users" }

// authUserColumns is the projection the credential resolver used. It is spelled out
// rather than left to the default so the authentication read stays pinned to exactly
// these columns even if userRead grows.
var authUserColumns = []string{
	"id", "name", "email", "password", "email_verified_at", "last_password_reset",
	"created_at", "updated_at", "deleted_at", "uuid", "status", "must_change_password",
	"pending_email",
}

// toAuthUser maps the read model onto the authenticated identity. Role is not a users
// column; it is filled in from companies_users where the request needs it.
func (r userRead) toAuthUser() *AuthUser {
	u := &AuthUser{
		Id:                 r.ID,
		UUID:               r.UUID,
		Status:             r.Status,
		Name:               r.Name,
		Email:              r.Email,
		PendingEmail:       r.PendingEmail,
		Password:           r.Password,
		EmailVerifiedAt:    r.EmailVerifiedAt,
		LastPasswordReset:  r.LastPasswordReset,
		MustChangePassword: r.MustChangePassword,
	}
	u.CreatedAt = r.CreatedAt
	u.UpdatedAt = r.UpdatedAt
	u.DeletedAt = r.DeletedAt
	return u
}

// toUser maps the read model onto the response struct.
func (r userRead) toUser() *User {
	u := new(User)
	u.Id = r.ID
	u.Name = r.Name
	u.Email = r.Email
	u.Password = r.Password
	u.EmailVerifiedAt = r.EmailVerifiedAt
	u.LastPasswordReset = r.LastPasswordReset
	u.UUID = r.UUID
	u.Status = r.Status
	u.MustChangePassword = r.MustChangePassword
	u.PendingEmail = r.PendingEmail
	u.Linked = r.Linked
	u.CreatedAt = r.CreatedAt
	u.UpdatedAt = r.UpdatedAt
	u.DeletedAt = r.DeletedAt
	return u
}

// companyRead is the playsql read model for the companies table. No softdelete tag:
// neither the `linked` count, the linked-companies join, nor any of the company
// reads ever filtered companies.deleted_at.
//
// identifier is nullable in the schema; playsql scans a SQL NULL as the zero value,
// which is what the old reads assumed all along.
type companyRead struct {
	ID         int        `db:"id" play:"pk,incrementing"`
	UUID       string     `db:"uuid"`
	Name       string     `db:"name"`
	Identifier string     `db:"identifier"`
	City       string     `db:"city"`
	Address    string     `db:"address"`
	AccountID  int        `db:"account_id"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

func (companyRead) TableName() string { return "companies" }

// toCompany maps the read model onto the response struct. The nested settings
// (sequences, redirect preferences, variants flag) are loaded separately, as before.
func (r companyRead) toCompany() *Company {
	c := &Company{
		ID:         r.ID,
		UUID:       r.UUID,
		Name:       r.Name,
		Identifier: r.Identifier,
		City:       r.City,
		Address:    r.Address,
	}
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt
	c.DeletedAt = r.DeletedAt
	return c
}

// companyInsert is the write model for the companies table. uuid and the timestamps
// are DB-generated and stay unmapped.
type companyInsert struct {
	ID         int64  `db:"id" play:"pk,incrementing"`
	AccountID  int    `db:"account_id"`
	Name       string `db:"name"`
	Identifier string `db:"identifier"`
	City       string `db:"city"`
	Address    string `db:"address"`
}

func (companyInsert) TableName() string { return "companies" }

// companySettingsRead is the playsql model for companies_settings. It backs the reads
// and the writes both: updated_at is mapped, so playsql stamps it on every Update and
// Upsert exactly where the raw statements said `updated_at = now()`.
//
// sequences and redirect_preferences are jsonb, mapped as []byte and decoded by the
// caller — CompanySequence and RedirectPreferences are structs, and a plain struct
// field would need a json cast to stay clear of playsql's relation heuristics.
//
// All three columns are NOT NULL with defaults. The old reads wrapped the flag in
// COALESCE(handles_variants, false), which never had anything to coalesce.
type companySettingsRead struct {
	ID                  int       `db:"id" play:"pk,incrementing"`
	CompanyID           int       `db:"company_id"`
	Sequences           []byte    `db:"sequences"`
	RedirectPreferences []byte    `db:"redirect_preferences"`
	HandlesVariants     bool      `db:"handles_variants"`
	UpdatedAt           time.Time `db:"updated_at"`
}

func (companySettingsRead) TableName() string { return "companies_settings" }

// uploadSessionRead is the playsql model for upload_sessions, backing its reads and
// its writes. The primary key is a string, so playsql treats it as non-incrementing
// and inserts the supplied value instead of asking the database for one.
//
// The nullable columns are pointers here and converted to sql.Null* in
// toUploadSession; a sql.NullInt64 field would be a plain struct, which playsql maps
// as a column but scans through an awkward double-pointer path.
type uploadSessionRead struct {
	ID             string    `db:"id" play:"pk"`
	UserID         int64     `db:"user_id"`
	Filename       string    `db:"filename"`
	FileSize       int64     `db:"file_size"`
	Delimiter      string    `db:"delimiter"`
	Encoding       string    `db:"encoding"`
	Status         string    `db:"status"`
	TotalChunks    *int64    `db:"total_chunks"`
	UploadedChunks int       `db:"uploaded_chunks"`
	ErrorMessage   *string   `db:"error_message"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func (uploadSessionRead) TableName() string { return "upload_sessions" }

// toUploadSession maps the read model onto the response struct.
func (r uploadSessionRead) toUploadSession() *UploadSession {
	u := &UploadSession{
		ID:             r.ID,
		UserID:         r.UserID,
		Filename:       r.Filename,
		FileSize:       r.FileSize,
		Delimiter:      r.Delimiter,
		Encoding:       r.Encoding,
		Status:         r.Status,
		UploadedChunks: r.UploadedChunks,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.TotalChunks != nil {
		u.TotalChunks = sql.NullInt64{Int64: *r.TotalChunks, Valid: true}
	}
	if r.ErrorMessage != nil {
		u.ErrorMessage = sql.NullString{String: *r.ErrorMessage, Valid: true}
	}
	return u
}

// importRead is the playsql model for the imports table, backing its reads and its
// writes. Like upload_sessions the primary key is a string (a uuid). The table has no
// updated_at, so playsql stamps nothing on update — matching the raw statements.
type importRead struct {
	ID            string     `db:"id" play:"pk"`
	UploadID      string     `db:"upload_id"`
	UserID        string     `db:"user_id"`
	Source        *string    `db:"source"`
	Status        string     `db:"status"`
	Phase         *string    `db:"phase"`
	TotalRows     *int       `db:"total_rows"`
	ProcessedRows *int       `db:"processed_rows"`
	SuccessRows   *int       `db:"success_rows"`
	FailedRows    *int       `db:"failed_rows"`
	WarningRows   *int       `db:"warning_rows"`
	ErrorMessage  *string    `db:"error_message"`
	CreatedAt     *time.Time `db:"created_at"`
	StartedAt     *time.Time `db:"started_at"`
	FinishedAt    *time.Time `db:"finished_at"`
}

func (importRead) TableName() string { return "imports" }

// toImportFile maps the read model onto the internal struct.
func (r importRead) toImportFile() *importFile {
	return &importFile{
		ID:            r.ID,
		UploadID:      r.UploadID,
		UserID:        r.UserID,
		Source:        r.Source,
		Status:        r.Status,
		Phase:         r.Phase,
		TotalRows:     r.TotalRows,
		ProcessedRows: r.ProcessedRows,
		SuccessRows:   r.SuccessRows,
		FailedRows:    r.FailedRows,
		WarningRows:   r.WarningRows,
		ErrorMEssage:  r.ErrorMessage,
		CreatedAt:     r.CreatedAt,
		StartedAt:     r.StartedAt,
		FinishedAt:    r.FinishedAt,
	}
}

// importRowIssue is the write model for import_row_issues. created_at is left to its
// database default.
type importRowIssue struct {
	ID         int64  `db:"id" play:"pk,incrementing"`
	ImportID   string `db:"import_id"`
	RowNumber  int    `db:"row_number"`
	ColumnName string `db:"column_name"`
	Level      string `db:"level"`
	Message    string `db:"message"`
	Value      string `db:"value"`
}

func (importRowIssue) TableName() string { return "import_row_issues" }

// accountRead is the playsql read model for the accounts table.
//
// No softdelete tag: none of the account reads filtered deleted_at. accounts.uuid is
// nullable in the schema; playsql scans a SQL NULL as the field's zero value, which
// is what the old scan into a string assumed.
//
// Owner is the account's user, joined by every account read. The eager load is
// constrained to three columns — see withAccountOwner — so a list of accounts does
// not drag every owner's password hash into memory.
type accountRead struct {
	ID         int        `db:"id" play:"pk,incrementing"`
	UUID       string     `db:"uuid"`
	OwnerID    int        `db:"owner_id"`
	Status     string     `db:"status"`
	VerifiedAt *time.Time `db:"verified_at"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`

	Owner *userRead `play:"belongsTo,fk=owner_id"`
}

func (accountRead) TableName() string { return "accounts" }

// withAccountOwner eager-loads the account's owner, projecting only the three columns
// the response struct carries. userRead maps the password hash; without this the
// default projection would select it.
func withAccountOwner(b *playsql.Builder) { b.Select("id", "email", "name") }

// toAccount maps the read model onto the response struct.
func (r accountRead) toAccount() *account {
	a := &account{
		ID:         r.ID,
		UUID:       r.UUID,
		Status:     r.Status,
		VerifiedAt: r.VerifiedAt,
	}
	a.CreatedAt = r.CreatedAt
	a.UpdatedAt = r.UpdatedAt
	a.DeletedAt = r.DeletedAt
	if o := r.Owner; o != nil {
		a.Owner.ID = o.ID
		a.Owner.Email = o.Email
		a.Owner.Name = o.Name
	}
	return a
}

// companyUserRead is the companies_users link row. It is a table in its own right
// (surrogate id, role, current, abilities), so the user reads model it directly
// rather than leaning on playsql's belongsToMany pivot bag.
type companyUserRead struct {
	ID        int    `db:"id" play:"pk,incrementing"`
	CompanyID int    `db:"company_id"`
	UserID    int    `db:"user_id"`
	Role      string `db:"role"`
	Current   bool   `db:"current"`

	Company *companyRead `play:"belongsTo,fk=company_id"`
}

func (companyUserRead) TableName() string { return "companies_users" }

// userInsert is the write model for the users table. uuid and the timestamps are
// DB-generated and stay unmapped; so does remember_token, which no write here owns.
type userInsert struct {
	ID       int64  `db:"id" play:"pk,incrementing"`
	Name     string `db:"name"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Status   string `db:"status"`
}

func (userInsert) TableName() string { return "users" }

// accountUserInsert is the write model for the accounts_users link row.
type accountUserInsert struct {
	ID        int64 `db:"id" play:"pk,incrementing"`
	AccountID int   `db:"account_id"`
	UserID    int   `db:"user_id"`
}

func (accountUserInsert) TableName() string { return "accounts_users" }

// accountUserRead is the read side of the accounts_users pivot. User.OwnedBy roots on
// it so the account arrives through a belongsTo instead of an INNER JOIN, mirroring how
// companyUserRead carries the role alongside its company.
type accountUserRead struct {
	ID        int `db:"id" play:"pk,incrementing"`
	AccountID int `db:"account_id"`
	UserID    int `db:"user_id"`

	Account *accountRead `play:"belongsTo,fk=account_id"`
}

func (accountUserRead) TableName() string { return "accounts_users" }

// toAccountStruct maps the read model onto the `account` response struct. Only
// owner_id is carried into Owner; the old queries selected no other owner column.
func (r accountRead) toAccountStruct() *account {
	a := new(account)
	a.ID = r.ID
	a.UUID = r.UUID
	a.Owner.ID = r.OwnerID
	a.Status = r.Status
	a.VerifiedAt = r.VerifiedAt
	a.CreatedAt = r.CreatedAt
	a.UpdatedAt = r.UpdatedAt
	a.DeletedAt = r.DeletedAt
	return a
}

// companyUserInsert is the write model for the companies_users link row.
type companyUserInsert struct {
	ID        int64  `db:"id" play:"pk,incrementing"`
	CompanyID int    `db:"company_id"`
	UserID    int    `db:"user_id"`
	Role      string `db:"role"`
	Current   bool   `db:"current"`
}

func (companyUserInsert) TableName() string { return "companies_users" }

// purchaseRead is the playsql read model for the purchases table (the header, not
// the lines — findPurchaseLines stays raw).
//
// deleted_at carries play:"softdelete": every purchase read filtered it.
//
// source is jsonb and maps as []byte, decoded in toPurchase; purchase.Source is a
// pointer-to-struct, which playsql would otherwise read as a relation.
//
// notes and invoice_number are nullable and the old reads wrapped them in COALESCE.
// purchase_status is NOT NULL, so its COALESCE(...::text, ”) never had anything to
// coalesce. playsql scans a SQL NULL as the zero value in every case.
type purchaseRead struct {
	ID               int                     `db:"id" play:"pk,incrementing"`
	CompanyID        int                     `db:"company_id"`
	VendorID         int                     `db:"vendor_id"`
	WarehouseID      int                     `db:"warehouse_id"`
	UUID             string                  `db:"uuid"`
	Code             string                  `db:"code"`
	Date             time.Time               `db:"date"`
	DueDate          *time.Time              `db:"due_date"`
	Subtotal         float64                 `db:"subtotal"`
	DiscountAmount   float64                 `db:"discount_amount"`
	TaxAmount        float64                 `db:"tax_amount"`
	Total            float64                 `db:"total"`
	Status           string                  `db:"status"`
	PurchaseStatus   string                  `db:"purchase_status"`
	PaymentStatus    PaidStatus              `db:"payment_status"`
	Notes            string                  `db:"notes"`
	InvoiceNumber    string                  `db:"invoice_number"`
	TransactionKind  PurchaseTransactionKind `db:"transaction_kind"`
	Source           []byte                  `db:"source"`
	MovementRecorded bool                    `db:"movement_recorded"`
	CreatedAt        *time.Time              `db:"created_at"`
	UpdatedAt        *time.Time              `db:"updated_at"`
	DeletedAt        *time.Time              `db:"deleted_at" play:"softdelete"`

	Vendor *vendorRead `play:"belongsTo,fk=vendor_id"`
}

func (purchaseRead) TableName() string { return "purchases" }

// purchaseListColumns is the projection the list read used: no company_id, no
// movement_recorded, no timestamps.
var purchaseListColumns = []string{
	"id", "uuid", "code", "warehouse_id", "date", "due_date", "subtotal",
	"discount_amount", "tax_amount", "total", "status", "purchase_status",
	"payment_status", "notes", "transaction_kind", "source", "invoice_number",
	"vendor_id",
}

// toPurchase maps the read model onto the response struct, deriving the three
// computed fields the old scan loops built by hand.
func (r purchaseRead) toPurchase() *purchase {
	p := &purchase{
		CompanyID:     r.CompanyID,
		ID:            r.ID,
		UUID:          r.UUID,
		Number:        r.Code,
		WarehouseID:   r.WarehouseID,
		Date:          r.Date,
		DueOn:         r.DueDate,
		Amount:        r.Subtotal,
		Discount:      Discount{Val: r.DiscountAmount, Type: "fixed"},
		Tax:           r.TaxAmount,
		Total:         r.Total,
		InvoiceNumber: r.InvoiceNumber,
		Status:        r.PurchaseStatus,
		PaymentStatus: r.PaymentStatus,
		Notes:         r.Notes,
		Kind:          r.TransactionKind,
		EntityStatus:  foundation.Status(r.Status),
	}

	p.AmountDue = p.Total
	if p.PaymentStatus == PaidStatuses.Paid {
		p.AmountDue = 0
	}

	p.Terms = "pia"
	if p.DueOn != nil {
		difference := p.DueOn.Sub(p.Date)
		p.Terms = fmt.Sprintf("net%d", int(difference.Hours())/24)
	}

	if len(r.Source) > 0 {
		src := new(PurchaseSource)
		if err := json.Unmarshal(r.Source, src); err == nil {
			p.Source = src
		}
	}
	if v := r.Vendor; v != nil {
		p.Vendor.ID = v.ID
		p.Vendor.UUID = v.UUID
		p.Vendor.Name = v.Name
		p.Vendor.Email = v.Email
		p.Vendor.Phone = v.Phone
		p.Vendor.Address = v.Address
	}
	return p
}

// inventoryTransferRead is the playsql model for inventory_transfers, backing the
// insert, the three status transitions, and the locking read in
// loadTransferForUpdate. updated_at is mapped, so playsql stamps it where the raw
// statements said `updated_at = NOW()`.
//
// playsql v0.3.0 added LockForUpdate, so that read no longer has to stay raw. The
// lock is what serialises concurrent transitions of the same transfer.
type inventoryTransferRead struct {
	ID              int        `db:"id" play:"pk,incrementing"`
	CompanyID       int        `db:"company_id"`
	UUID            string     `db:"uuid"`
	FromWarehouseID int        `db:"from_warehouse_id"`
	ToWarehouseID   int        `db:"to_warehouse_id"`
	Notes           *string    `db:"notes"`
	Status          string     `db:"status"`
	RequestedBy     *int       `db:"requested_by"`
	DispatchedAt    *time.Time `db:"dispatched_at"`
	ReceivedAt      *time.Time `db:"received_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

func (inventoryTransferRead) TableName() string { return "inventory_transfers" }

// inventoryTransferLine is the write model for a transfer's product lines.
type inventoryTransferLine struct {
	ID         int64   `db:"id" play:"pk,incrementing"`
	CompanyID  int     `db:"company_id"`
	TransferID int     `db:"transfer_id"`
	VariantID  int     `db:"variant_id"`
	Qty        float64 `db:"qty"`
	// unit_id is nullable, mapped as int64 rather than *int so the belongsTo below
	// matches: the loader compares a child's primary key against this field by Go
	// value, and a *int never equals an int64. A NULL scans to 0, no unit carries id 0,
	// so the relation resolves to nil — the LEFT JOIN's outer arm.
	UnitID      int64   `db:"unit_id"`
	UnitCost    float64 `db:"unit_cost"`
	Description *string `db:"description"`

	Variant *itemVariantRead `play:"belongsTo,fk=variant_id"`
	Unit    *unitRead        `play:"belongsTo,fk=unit_id"`
}

func (inventoryTransferLine) TableName() string { return "inventory_transfer_lines" }

// inventoryBalanceRead is the read side of inventory_balances, used by
// dispatchTransfer to lock a source balance before checking it.
//
// The table has no surrogate id — its key is (company_id, variant_id, warehouse_id) —
// so no field carries play:"pk". Nothing writes through this model: the balance
// upsert and the reversal increment are both raw, because playsql cannot express a
// conflict branch that accumulates. Only First/Get, which never touch the pk.
type inventoryBalanceRead struct {
	CompanyID   int     `db:"company_id"`
	VariantID   int     `db:"variant_id"`
	WarehouseID int     `db:"warehouse_id"`
	Quantity    float64 `db:"quantity"`
}

func (inventoryBalanceRead) TableName() string { return "inventory_balances" }

// itemVariantRead is a read model for items_variants. No softdelete tag —
// recordMovement's track_inventory lookup never filtered deleted_at, and neither did
// findInvoiceLines' join.
//
// recordMovement pins its projection with Select("track_inventory"), so the name/sku/
// is_default columns findInvoiceLines needs cost it nothing.
type itemVariantRead struct {
	ID             int     `db:"id" play:"pk,incrementing"`
	CompanyID      int     `db:"company_id"`
	ItemID         int     `db:"item_id"`
	TrackInventory bool    `db:"track_inventory"`
	Name           string  `db:"name"`
	SKU            *string `db:"sku"`
	IsDefault      bool    `db:"is_default"`

	// findPurchaseLines reaches the item through the variant (iv.item_id), not from a
	// column on purchase_items. Only that read loads it.
	Item *lineItemRead `play:"belongsTo,fk=item_id"`
}

func (itemVariantRead) TableName() string { return "items_variants" }

// inventoryMovementRead is the read side of inventory_movements. The write model is
// InventoryMovement; this one exists so reverseMovements can select its four columns.
type inventoryMovementRead struct {
	ID              int64                 `db:"id" play:"pk,incrementing"`
	CompanyID       int                   `db:"company_id"`
	VariantID       int                   `db:"variant_id"`
	WarehouseID     int                   `db:"warehouse_id"`
	Qty             float64               `db:"qty"`
	UnitCost        float64               `db:"unit_cost"`
	TransactionKind InventoryMovementKind `db:"transaction_kind"`
	// reference_type and reference_id are nullable; a NULL scans to the zero value,
	// which is what the old COALESCE(..., '') / COALESCE(..., 0) produced.
	ReferenceType string     `db:"reference_type"`
	ReferenceID   int        `db:"reference_id"`
	CreatedAt     *time.Time `db:"created_at"`

	Variant   *itemVariantRead `play:"belongsTo,fk=variant_id"`
	Warehouse *warehouseRead   `play:"belongsTo,fk=warehouse_id"`
}

func (inventoryMovementRead) TableName() string { return "inventory_movements" }

// warehouseRead is the playsql model for the warehouses table, backing its reads and
// its writes. softdelete matches the `deleted_at IS NULL` every warehouse read
// carries.
//
// location is nullable and the old reads wrapped it in COALESCE(location, ”). playsql
// scans a SQL NULL as the field's zero value, so the wrapper is unnecessary; writes
// pass nullIfEmpty so an empty form field stores NULL rather than ”.
type warehouseRead struct {
	ID        int        `db:"id" play:"pk,incrementing"`
	CompanyID int        `db:"company_id"`
	UUID      string     `db:"uuid"`
	Name      string     `db:"name"`
	Location  string     `db:"location"`
	Status    string     `db:"status"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" play:"softdelete"`
}

func (warehouseRead) TableName() string { return "warehouses" }

// toWarehouse maps the read model onto the JSON response struct.
func (r warehouseRead) toWarehouse() *warehouse {
	w := &warehouse{
		ID:       r.ID,
		UUID:     r.UUID,
		Name:     r.Name,
		Location: r.Location,
		Status:   foundation.Status(r.Status),
	}
	w.CreatedAt = r.CreatedAt
	w.UpdatedAt = r.UpdatedAt
	w.DeletedAt = r.DeletedAt
	return w
}

// taxRead is the playsql model for the taxes table. No softdelete tag: findTaxes
// never filtered deleted_at, and nothing soft-deletes a tax.
type taxRead struct {
	ID        int64      `db:"id" play:"pk,incrementing"`
	CompanyID int        `db:"company_id"`
	UUID      string     `db:"uuid"`
	Name      string     `db:"name"`
	Rate      float64    `db:"rate"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (taxRead) TableName() string { return "taxes" }

// toTax maps the read model onto the JSON response struct.
func (r taxRead) toTax() *tax {
	t := &tax{ID: r.ID, UUID: r.UUID, Name: r.Name, Rate: r.Rate}
	t.CreatedAt = r.CreatedAt
	t.UpdatedAt = r.UpdatedAt
	t.DeletedAt = r.DeletedAt
	return t
}

// taxReceiptRead is the playsql read model for tax_receipts.
//
// No softdelete tag: neither list read filtered deleted_at, so a retired receipt
// still appears in the settings list. grabTaxReceiptSequence filters it explicitly
// (see there) — that asymmetry is the existing behaviour, not something this model
// should quietly change.
//
// sequence_start, sequence_end and current are NOT NULL; the old reads wrapped them
// in COALESCE(..., 0), which never had anything to coalesce.
type taxReceiptRead struct {
	ID            int        `db:"id" play:"pk,incrementing"`
	CompanyID     int        `db:"company_id"`
	Name          string     `db:"name"`
	Serie         string     `db:"serie"`
	Type          string     `db:"type"`
	SequenceStart int        `db:"sequence_start"`
	SequenceEnd   int        `db:"sequence_end"`
	Current       int        `db:"current"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (taxReceiptRead) TableName() string { return "tax_receipts" }

// toTaxReceipt maps the read model onto the JSON response struct.
func (r taxReceiptRead) toTaxReceipt() *taxReceipt {
	t := &taxReceipt{
		ID:            r.ID,
		Name:          r.Name,
		Serie:         r.Serie,
		Type:          r.Type,
		SequenceStart: r.SequenceStart,
		SequenceEnd:   r.SequenceEnd,
		Current:       r.Current,
	}
	t.CreatedAt = r.CreatedAt
	t.UpdatedAt = r.UpdatedAt
	t.DeletedAt = r.DeletedAt
	return t
}

// attributeRead is the playsql model for the attributes table, backing its reads and
// its writes. Every attribute query filtered deleted_at, so deleted_at carries
// play:"softdelete".
//
// Values is a hasMany: findAttributesWithValues used to loop over the attributes and
// issue one query per attribute. With("Values") loads them all in a second query.
type attributeRead struct {
	ID          int        `db:"id" play:"pk,incrementing"`
	CompanyID   int        `db:"company_id"`
	UUID        string     `db:"uuid"`
	Name        string     `db:"name"`
	Type        string     `db:"type"`
	DisplayName string     `db:"display_name"`
	Description *string    `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" play:"softdelete"`

	Values []*attributeValueRead `play:"hasMany,fk=attribute_id"`
}

func (attributeRead) TableName() string { return "attributes" }

// toAttribute maps the read model onto the JSON response struct. Values are only
// populated when the caller eager-loaded them.
func (r attributeRead) toAttribute() *attribute {
	a := &attribute{
		ID:          r.ID,
		UUID:        r.UUID,
		Name:        r.Name,
		Type:        r.Type,
		DisplayName: r.DisplayName,
		Description: r.Description,
	}
	a.CreatedAt = r.CreatedAt
	a.UpdatedAt = r.UpdatedAt
	a.DeletedAt = r.DeletedAt
	if r.Values != nil {
		a.Values = make([]*attributeValue, 0, len(r.Values))
		for _, v := range r.Values {
			a.Values = append(a.Values, v.toAttributeValue())
		}
	}
	return a
}

// attributeValueRead is the playsql model for attribute_values, backing its reads and
// its writes. softdelete matches the `deleted_at IS NULL` every value query carried.
//
// display_name and sort_order are nullable in the schema; playsql scans a SQL NULL as
// the field's zero value, which is what the old scans into string/int assumed.
type attributeValueRead struct {
	ID          int        `db:"id" play:"pk,incrementing"`
	CompanyID   int        `db:"company_id"`
	AttributeID int        `db:"attribute_id"`
	UUID        string     `db:"uuid"`
	Value       string     `db:"value"`
	DisplayName string     `db:"display_name"`
	SortOrder   int        `db:"sort_order"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" play:"softdelete"`
}

func (attributeValueRead) TableName() string { return "attribute_values" }

// toAttributeValue maps the read model onto the JSON response struct.
func (r attributeValueRead) toAttributeValue() *attributeValue {
	v := &attributeValue{
		ID:          r.ID,
		UUID:        r.UUID,
		AttributeID: r.AttributeID,
		Value:       r.Value,
		DisplayName: r.DisplayName,
		SortOrder:   r.SortOrder,
	}
	v.CreatedAt = r.CreatedAt
	v.UpdatedAt = r.UpdatedAt
	v.DeletedAt = r.DeletedAt
	return v
}

// productAttributeRead is the items ↔ attributes link table. Only the columns the
// in-use checks need are mapped; the table has no deleted_at.
type productAttributeRead struct {
	ID          int `db:"id" play:"pk,incrementing"`
	CompanyID   int `db:"company_id"`
	ItemID      int `db:"item_id"`
	AttributeID int `db:"attribute_id"`
}

func (productAttributeRead) TableName() string { return "product_attributes" }

// variantAttributeValueRead is the variants ↔ attribute_values link table, used to
// refuse deleting a value that variants still reference.
type variantAttributeValueRead struct {
	ID               int `db:"id" play:"pk,incrementing"`
	CompanyID        int `db:"company_id"`
	VariantID        int `db:"variant_id"`
	AttributeID      int `db:"attribute_id"`
	AttributeValueID int `db:"attribute_value_id"`
}

func (variantAttributeValueRead) TableName() string { return "variant_attribute_values" }

// expenseModel is the playsql read model for the expenses table. receipt_url is
// deliberately unmapped: it is nullable, no read ever selected it, and mapping it
// would pull a NULL into the default projection.
type expenseModel struct {
	ID         int        `db:"id" play:"pk,incrementing"`
	CompanyID  int        `db:"company_id"`
	UUID       string     `db:"uuid" play:"guarded"`
	CategoryID int        `db:"category_id"`
	Date       time.Time  `db:"date"`
	Amount     float64    `db:"amount"`
	Notes      string     `db:"notes"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at" play:"softdelete"`

	Category *expenseCategoryModel `play:"belongsTo,fk=category_id"`
}

func (expenseModel) TableName() string { return "expenses" }

// toExpense maps the read model onto the JSON response struct. Only the three
// category columns the old INNER JOIN selected are copied across.
func (r expenseModel) toExpense() *expense {
	e := &expense{
		ID:     r.ID,
		UUID:   r.UUID,
		Date:   Date{Time: r.Date},
		Amount: r.Amount,
		Notes:  r.Notes,
	}
	e.CreatedAt = r.CreatedAt
	e.UpdatedAt = r.UpdatedAt
	e.DeletedAt = r.DeletedAt
	if r.Category != nil {
		e.Category = expenseCategory{
			ID:   r.Category.ID,
			UUID: r.Category.UUID,
			Name: r.Category.Name,
		}
	}
	return e
}

// expenseCategoryModel is the playsql read model for the expenses_categories table.
// It carries no play:"softdelete" tag even though the column exists: only
// findExpensesCategories filtered deleted_at, and the other three reads must keep
// resolving a soft-deleted category (storeExpense/updateExpense look one up by
// uuid). That one read filters explicitly with WhereNull.
//
// TotalAmount is play:"readonly" — it has no backing column and is excluded from
// the default projection, appearing only as the WithSum aggregate alias in
// findExpensesByCategories.
type expenseCategoryModel struct {
	ID          int        `db:"id" play:"pk,incrementing"`
	CompanyID   int        `db:"company_id"`
	UUID        string     `db:"uuid" play:"guarded"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	TotalAmount float64    `db:"total_amount" play:"readonly"`

	Expenses []*expenseModel `play:"hasMany,fk=category_id"`
}

func (expenseCategoryModel) TableName() string { return "expenses_categories" }

// toExpenseCategory maps the read model onto the JSON response struct.
func (r expenseCategoryModel) toExpenseCategory() *expenseCategory {
	c := &expenseCategory{
		ID:          r.ID,
		UUID:        r.UUID,
		Name:        r.Name,
		Description: r.Description,
		TotalAmount: r.TotalAmount,
	}
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt
	c.DeletedAt = r.DeletedAt
	return c
}

// paymentRead is the playsql read model for the receivables_income table (customer
// payments). payment is jsonb and maps as []byte: playsql classifies any pointer or
// slice-of-struct field as a relation, and a plain struct field would need a json
// cast, so the blob is decoded in toPayment — which also keeps the previous
// behaviour of tolerating a malformed blob instead of failing the read.
//
// Invoices is play:"readonly": it has no backing column and carries the WithCount
// aggregate that replaced the old correlated `(select count(*) …) as invoices`.
type paymentRead struct {
	ID         int           `db:"id" play:"pk,incrementing"`
	UUID       string        `db:"uuid"`
	Code       string        `db:"code"`
	CustomerID int           `db:"customer_id"`
	Date       time.Time     `db:"date"`
	Amount     float64       `db:"amount"`
	Notes      string        `db:"notes"`
	Payment    []byte        `db:"payment"`
	Status     PaymentStatus `db:"status"`
	CreatedAt  *time.Time    `db:"created_at"`
	UpdatedAt  *time.Time    `db:"updated_at"`
	DeletedAt  *time.Time    `db:"deleted_at" play:"softdelete"`
	Invoices   int           `db:"invoices" play:"readonly"`

	Customer *customerModel     `play:"belongsTo,fk=customer_id"`
	Items    []*paymentItemRead `play:"hasMany,fk=receivable_income_id"`
}

func (paymentRead) TableName() string { return "receivables_income" }

// toPayment maps the read model onto the JSON response struct.
//
// It fills every customer field the response struct has. The old list query only
// selected uuid/name/amount_due while the detail query also took id/email/address,
// so the list response now carries those three as well — a superset of the same
// tenant's own data, which no caller depends on being absent.
func (r paymentRead) toPayment() *payment {
	p := &payment{
		ID:       r.ID,
		UUID:     r.UUID,
		Code:     r.Code,
		Date:     r.Date,
		Amount:   r.Amount,
		Notes:    r.Notes,
		Invoices: r.Invoices,
		Status:   r.Status,
	}
	p.CreatedAt = r.CreatedAt
	p.UpdatedAt = r.UpdatedAt
	p.DeletedAt = r.DeletedAt
	if len(r.Payment) > 0 {
		pm := new(Payment)
		if err := json.Unmarshal(r.Payment, pm); err == nil {
			p.Payment = *pm
		}
	}
	if c := r.Customer; c != nil {
		p.Customer.ID = c.ID
		p.Customer.UUID = c.UUID
		p.Customer.Name = c.Name
		p.Customer.Email = c.Email
		p.Customer.AmountDue = c.AmountDue
		p.Customer.Address = c.Address
		p.Customer.Phone = c.Phone
	}
	return p
}

// paymentItemRead is the read side of receivables_income_items. It carries no
// play:"softdelete" tag even though deleted_at exists: neither the old count
// subquery nor findPaymentLines filtered it, and playsql would otherwise exclude
// trashed children from the WithCount aggregate too.
type paymentItemRead struct {
	ID                 int        `db:"id" play:"pk,incrementing"`
	ReceivableIncomeID int        `db:"receivable_income_id"`
	InvoiceID          int        `db:"invoice_id"`
	AmountDue          float64    `db:"amount_due"`
	PaymentAmount      float64    `db:"payment_amount"`
	CreatedAt          *time.Time `db:"created_at"`
	UpdatedAt          *time.Time `db:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at"`

	Invoice *paymentInvoiceRead `play:"belongsTo,fk=invoice_id"`
}

func (paymentItemRead) TableName() string { return "receivables_income_items" }

// invoiceRead is the playsql read model for the invoices table — the header, not the
// lines (findInvoiceLines stays raw; see there).
//
// discount, payment and source are jsonb. They map as []byte and are decoded in
// toInvoice: playsql classifies a pointer-to-struct field as a relation, which is
// what invoice.Source is, and decoding here keeps a malformed blob from failing the
// whole read.
//
// No softdelete tag: none of the three header reads filtered invoices.deleted_at.
type invoiceRead struct {
	ID               int             `db:"id" play:"pk,incrementing"`
	CompanyID        int             `db:"company_id"`
	CustomerID       int             `db:"customer_id"`
	UUID             string          `db:"uuid"`
	Code             string          `db:"code"`
	Date             time.Time       `db:"date"`
	DueOn            *time.Time      `db:"due_on"`
	Amount           float64         `db:"amount"`
	AmountDue        float64         `db:"amount_due"`
	Discount         []byte          `db:"discount"`
	Tax              float64         `db:"tax"`
	Total            float64         `db:"total"`
	Status           InvoiceStatus   `db:"status"`
	PaidStatus       PaidStatus      `db:"paid_status"`
	Payment          []byte          `db:"payment"`
	Note             string          `db:"note"`
	TaxReceiptID     *int            `db:"tax_receipt_id"`
	TaxNumber        *string         `db:"tax_number"`
	TransactionKind  TransactionKind `db:"transaction_kind"`
	Source           []byte          `db:"source"`
	MovementRecorded bool            `db:"movement_recorded"`

	Customer *customerModel `play:"belongsTo,fk=customer_id"`
}

func (invoiceRead) TableName() string { return "invoices" }

// invoiceListColumns is the projection the list read used. It omits company_id and
// movement_recorded, which are internal and were never sent to the client.
var invoiceListColumns = []string{
	"id", "uuid", "code", "date", "due_on", "amount", "discount", "tax", "total",
	"amount_due", "status", "paid_status", "payment", "note", "tax_receipt_id",
	"transaction_kind", "source", "tax_number", "customer_id",
}

// toInvoice maps the read model onto the response struct. Terms is not derived here:
// the list read never set it, only the two detail reads do (see deriveTerms).
func (r invoiceRead) toInvoice() *invoice {
	i := &invoice{
		CompanyID:    r.CompanyID,
		ID:           r.ID,
		UUID:         r.UUID,
		Number:       r.Code,
		NCF:          r.TaxNumber,
		Date:         r.Date,
		DueOn:        r.DueOn,
		TaxReceiptID: r.TaxReceiptID,
		Amount:       r.Amount,
		Tax:          r.Tax,
		Total:        r.Total,
		AmountDue:    r.AmountDue,
		Status:       r.Status,
		PaidStatus:   r.PaidStatus,
		Notes:        r.Note,
		Kind:         r.TransactionKind,
	}
	if len(r.Discount) > 0 {
		_ = json.Unmarshal(r.Discount, &i.Discount)
	}
	if len(r.Payment) > 0 {
		_ = json.Unmarshal(r.Payment, &i.Payment)
	}
	if len(r.Source) > 0 {
		src := new(TransactionSource)
		if err := json.Unmarshal(r.Source, src); err == nil {
			i.Source = src
		}
	}
	if c := r.Customer; c != nil {
		i.Customer.ID = c.ID
		i.Customer.UUID = c.UUID
		i.Customer.Name = c.Name
		i.Customer.Email = c.Email
		i.Customer.Phone = c.Phone
		i.Customer.Address = c.Address
	}
	return i
}

// paymentInvoiceRead is a narrow projection of the invoices table: only the columns
// a payment line reports. It stays separate from invoiceRead because the two answer
// different queries; both deliberately omit a softdelete tag, since neither the
// payment-line joins nor the invoice header reads filtered invoices.deleted_at.
//
// tax_number is nullable; playsql scans a SQL NULL as the field's zero value, so an
// invoice with no tax receipt reports an empty NCF rather than dropping the line.
type paymentInvoiceRead struct {
	ID        int        `db:"id" play:"pk,incrementing"`
	UUID      string     `db:"uuid"`
	Code      string     `db:"code"`
	Date      time.Time  `db:"date"`
	DueOn     *time.Time `db:"due_on"`
	Total     float64    `db:"total"`
	TaxNumber string     `db:"tax_number"`
	Note      string     `db:"note"`
}

func (paymentInvoiceRead) TableName() string { return "invoices" }

// accountsPayableRead is the playsql read model for the accounts_payable table.
// The table has no deleted_at, so there is no softdelete tag. amount_payable is a
// generated column: it is read here but must never be written (see
// accountsPayableInsert, which leaves it unmapped).
//
// Register is the one-to-one payables cross-reference row. findPayables reads the
// AP entry as the root (it owns the filters and the due_date ordering) and pulls
// p.id / p.uuid through this relation, which is the inverse of the old query's
// payables-rooted INNER JOIN.
type accountsPayableRead struct {
	ID            int64     `db:"id" play:"pk,incrementing"`
	UUID          string    `db:"uuid"`
	VendorID      int       `db:"vendor_id"`
	InvoiceNumber string    `db:"invoice_number"`
	InvoiceDate   time.Time `db:"invoice_date"`
	DueDate       time.Time `db:"due_date"`
	AmountTotal   float64   `db:"amount_total"`
	AmountPayable float64   `db:"amount_payable"`
	AmountPaid    float64   `db:"amount_paid"`
	Status        string    `db:"status"`
	PaidStatus    string    `db:"paid_status"`
	Notes         *string   `db:"notes"`

	Register *payableRegisterRead `play:"hasOne,fk=accounts_payable_id"`
}

func (accountsPayableRead) TableName() string { return "accounts_payable" }

// toPayable maps an AP entry plus its payables register row onto the response
// struct. Payable.ID/UUID identify the payables row; Payable.InvoiceID/InvoiceUUID
// identify the accounts_payable row.
func (r accountsPayableRead) toPayable() *Payable {
	notes := ""
	if r.Notes != nil {
		notes = *r.Notes
	}
	p := &Payable{
		InvoiceID:     r.ID,
		InvoiceUUID:   r.UUID,
		InvoiceNumber: r.InvoiceNumber,
		InvoiceDate:   r.InvoiceDate,
		DueDate:       r.DueDate,
		AmountTotal:   r.AmountTotal,
		AmountPayable: r.AmountPayable,
		AmountPaid:    r.AmountPaid,
		Status:        PayableStatus(r.Status),
		PaidStatus:    PaidStatus(r.PaidStatus),
		Notes:         &notes,
	}
	if r.Register != nil {
		p.ID = r.Register.ID
		p.UUID = r.Register.UUID
	}
	return p
}

// payableRegisterRead is the read side of the payables table. It is separate from
// the payableRegister write model because it maps uuid, which is DB-generated and
// must stay unmapped on insert.
type payableRegisterRead struct {
	ID                int64  `db:"id" play:"pk,incrementing"`
	UUID              string `db:"uuid"`
	AccountsPayableID int64  `db:"accounts_payable_id"`
	VendorID          int64  `db:"vendor_id"`
}

func (payableRegisterRead) TableName() string { return "payables" }

// vendorPaymentRead is the playsql read model for the vendor_payments table.
// company_id and the timestamps are deliberately unmapped: the old reads never
// selected them and vendorPayment leaves them zero.
//
// payment is jsonb; it is scanned raw and unmarshalled in toVendorPayment so a
// malformed blob leaves Payment nil instead of failing the whole read, matching
// the previous behaviour.
type vendorPaymentRead struct {
	ID   int64  `db:"id" play:"pk,incrementing"`
	UUID string `db:"uuid"`
	// int, not int64: eager loading matches this against vendorRead.ID by Go value,
	// so a widened type would silently never match and leave Vendor nil.
	VendorID int       `db:"vendor_id"`
	Date     time.Time `db:"date"`
	Amount   float64   `db:"amount"`
	Notes    *string   `db:"notes"`
	Payment  []byte    `db:"payment"`
	Status   string    `db:"status"`
	Code     string    `db:"code"`

	Vendor *vendorRead `play:"belongsTo,fk=vendor_id"`
}

func (vendorPaymentRead) TableName() string { return "vendor_payments" }

// toVendorPayment maps the read model onto the JSON response struct. Only the
// five vendor columns the old INNER JOIN selected are copied across.
func (r vendorPaymentRead) toVendorPayment() *vendorPayment {
	p := &vendorPayment{
		ID:     int(r.ID),
		UUID:   r.UUID,
		Date:   r.Date,
		Amount: r.Amount,
		Status: r.Status,
		Code:   r.Code,
	}
	if r.Notes != nil {
		p.Notes = *r.Notes
	}
	if len(r.Payment) > 0 {
		pm := new(Payment)
		if err := json.Unmarshal(r.Payment, pm); err == nil {
			p.Payment = pm
		}
	}
	if r.Vendor != nil {
		p.Vendor = vendor{
			ID:    r.Vendor.ID,
			UUID:  r.Vendor.UUID,
			Name:  r.Vendor.Name,
			Email: r.Vendor.Email,
			Phone: r.Vendor.Phone,
		}
	}
	return p
}

// vendorPaymentItemRead is the read side of vendor_payment_items, carrying the
// settled AP entry as a belongsTo so the line can report the bill's number and
// dates. The write model (vendorPaymentItem) stays separate.
type vendorPaymentItemRead struct {
	ID                int64   `db:"id" play:"pk,incrementing"`
	VendorPaymentID   int64   `db:"vendor_payment_id"`
	AccountsPayableID int64   `db:"accounts_payable_id"`
	AmountDue         float64 `db:"amount_due"`
	PaymentAmount     float64 `db:"payment_amount"`

	AccountsPayable *accountsPayableRead `play:"belongsTo,fk=accounts_payable_id"`
}

func (vendorPaymentItemRead) TableName() string { return "vendor_payment_items" }

// Receivable is the write model for the receivables table. The pk is DB-assigned
// (serial); playsql omits the zero id on insert and reads it back via RETURNING.
// Columns with database defaults (timestamps) are intentionally not mapped, so
// the INSERT lets the database fill them.
type Receivable struct {
	ID         int64 `db:"id" play:"pk,incrementing"`
	CompanyID  int   `db:"company_id"`
	InvoiceID  int   `db:"invoice_id"`
	CustomerID int   `db:"customer_id"`
}

func (Receivable) TableName() string { return "receivables" }

// InvoiceItem is the write model for invoice line rows (invoices_items). It backs
// the bulk line insert; the pk is DB-assigned, and timestamp columns are left to
// their database defaults (not mapped here).
type InvoiceItem struct {
	ID          int64   `db:"id" play:"pk,incrementing"`
	CompanyID   int     `db:"company_id"`
	InvoiceID   int     `db:"invoice_id"`
	ItemID      int     `db:"item_id"`
	VariantID   int     `db:"variant_id"`
	UnitID      int     `db:"unit_id"`
	Qty         int     `db:"qty"`
	Price       float64 `db:"price"`
	Rate        float64 `db:"rate"`
	Amount      float64 `db:"amount"`
	Tax         float64 `db:"tax"`
	Total       float64 `db:"total"`
	WarehouseID int     `db:"warehouse_id"`
}

func (InvoiceItem) TableName() string { return "invoices_items" }

// invoiceInsert is the write model for creating an invoice row. uuid is
// deliberately NOT mapped: it is DB-generated, and mapping it would make playsql
// insert an empty string over the default. After Insert sets ID, the caller reads
// the generated uuid back. JSON columns hold pre-encoded values (foundation
// ToJSON/AsJSON) exactly as the prior raw INSERT passed them, so encoding is
// unchanged. Timestamp columns are left to their database defaults.
type invoiceInsert struct {
	ID                 int64           `db:"id" play:"pk,incrementing"`
	CompanyID          int             `db:"company_id"`
	TaxReceiptID       *int            `db:"tax_receipt_id"`
	TaxReceiptSequence *int64          `db:"tax_receipt_sequence"`
	TaxNumber          *string         `db:"tax_number"`
	Date               time.Time       `db:"date"`
	Type               *string         `db:"type"`
	DueOn              *time.Time      `db:"due_on"`
	CustomerID         int             `db:"customer_id"`
	Amount             float64         `db:"amount"`
	Discount           string          `db:"discount"`
	Tax                float64         `db:"tax"`
	AmountDue          float64         `db:"amount_due"`
	Total              float64         `db:"total"`
	Note               string          `db:"note"`
	Status             InvoiceStatus   `db:"status"`
	PaidStatus         PaidStatus      `db:"paid_status"`
	Payment            string          `db:"payment"`
	Code               string          `db:"code"`
	TransactionKind    TransactionKind `db:"transaction_kind"`
	Source             *[]byte         `db:"source"`
	Recurrence         *[]byte         `db:"recurrence"`
}

func (invoiceInsert) TableName() string { return "invoices" }

// InventoryMovement is the write model for a stock movement row. created_at is
// set explicitly by the caller (no DB default is relied on). The balance upsert
// that follows a movement stays raw SQL: it increments (quantity +=
// EXCLUDED.quantity), which playsql's replace-style Upsert cannot express.
type InventoryMovement struct {
	ID              int64                 `db:"id" play:"pk,incrementing"`
	CompanyID       int                   `db:"company_id"`
	VariantID       int                   `db:"variant_id"`
	WarehouseID     int                   `db:"warehouse_id"`
	TransactionKind InventoryMovementKind `db:"transaction_kind"`
	Qty             float64               `db:"qty"`
	UnitCost        float64               `db:"unit_cost"`
	ReferenceType   string                `db:"reference_type"`
	ReferenceID     int                   `db:"reference_id"`
	CreatedAt       time.Time             `db:"created_at"`
}

func (InventoryMovement) TableName() string { return "inventory_movements" }

// vendorInsert is the write model for creating a vendor row. amount_payable seeds
// from the opening balance; the pk is DB-assigned.
type vendorInsert struct {
	ID            int64   `db:"id" play:"pk,incrementing"`
	CompanyID     int     `db:"company_id"`
	Name          string  `db:"name"`
	ContactName   string  `db:"contact_name"`
	Email         string  `db:"email"`
	Phone         string  `db:"phone"`
	PaymentMethod string  `db:"payment_method"`
	PaymentTerms  string  `db:"payment_terms"`
	PurchaseNote  string  `db:"purchase_note"`
	LeadTimeDays  int     `db:"lead_time_days"`
	AmountPayable float64 `db:"amount_payable"`
	VendorType    string  `db:"vendor_type"`
	Code          string  `db:"code"`
	Address       string  `db:"address"`
}

func (vendorInsert) TableName() string { return "vendors" }

// PurchaseItem is the write model for purchase line rows (purchase_items). Backs
// the bulk line insert; the pk is DB-assigned.
type PurchaseItem struct {
	ID         int64   `db:"id" play:"pk,incrementing"`
	CompanyID  int     `db:"company_id"`
	PurchaseID int     `db:"purchase_id"`
	VariantID  int     `db:"variant_id"`
	Qty        int     `db:"qty"`
	UnitPrice  float64 `db:"unit_price"`
	LineTotal  float64 `db:"line_total"`
	UnitID     int     `db:"unit_id"`
	Discount   float64 `db:"discount"`
	TaxID      int     `db:"tax_id"`
	TaxAmount  float64 `db:"tax_amount"`
}

func (PurchaseItem) TableName() string { return "purchase_items" }

// paymentInsert is the write model for a customer payment (receivables_income).
type paymentInsert struct {
	ID         int64         `db:"id" play:"pk,incrementing"`
	CompanyID  int           `db:"company_id"`
	CustomerID int           `db:"customer_id"`
	Date       time.Time     `db:"date"`
	Amount     float64       `db:"amount"`
	Notes      string        `db:"notes"`
	Payment    string        `db:"payment"`
	Status     PaymentStatus `db:"status"`
	Code       string        `db:"code"`
}

func (paymentInsert) TableName() string { return "receivables_income" }

// vendorPaymentInsert is the write model for a vendor payment (vendor_payments).
type vendorPaymentInsert struct {
	ID        int64     `db:"id" play:"pk,incrementing"`
	CompanyID int       `db:"company_id"`
	VendorID  int       `db:"vendor_id"`
	Date      time.Time `db:"date"`
	Amount    float64   `db:"amount"`
	Notes     string    `db:"notes"`
	Payment   string    `db:"payment"`
	Status    string    `db:"status"`
	Code      string    `db:"code"`
}

func (vendorPaymentInsert) TableName() string { return "vendor_payments" }

// paymentItem is the write model for customer payment allocation rows
// (receivables_income_items). Backs the bulk allocation insert.
type paymentItem struct {
	ID                 int64     `db:"id" play:"pk,incrementing"`
	CompanyID          int       `db:"company_id"`
	ReceivableIncomeID int       `db:"receivable_income_id"`
	Date               time.Time `db:"date"`
	InvoiceID          int       `db:"invoice_id"`
	AmountDue          float64   `db:"amount_due"`
	PaymentAmount      float64   `db:"payment_amount"`
}

func (paymentItem) TableName() string { return "receivables_income_items" }

// vendorPaymentItem is the write model for vendor payment allocation rows
// (vendor_payment_items). Backs the bulk allocation insert.
type vendorPaymentItem struct {
	ID                int64     `db:"id" play:"pk,incrementing"`
	CompanyID         int       `db:"company_id"`
	VendorPaymentID   int       `db:"vendor_payment_id"`
	AccountsPayableID int64     `db:"accounts_payable_id"`
	Date              time.Time `db:"date"`
	AmountDue         float64   `db:"amount_due"`
	PaymentAmount     float64   `db:"payment_amount"`
}

func (vendorPaymentItem) TableName() string { return "vendor_payment_items" }

// accountsPayableInsert is the write model for an AP entry (accounts_payable).
// amount_payable is a generated column and is intentionally not mapped.
type accountsPayableInsert struct {
	ID             int64         `db:"id" play:"pk,incrementing"`
	CompanyID      int           `db:"company_id"`
	VendorID       int           `db:"vendor_id"`
	PurchaseID     int           `db:"purchase_id"`
	InvoiceNumber  string        `db:"invoice_number"`
	InvoiceDate    time.Time     `db:"invoice_date"`
	DueDate        *time.Time    `db:"due_date"`
	AmountTotal    float64       `db:"amount_total"`
	TaxAmount      float64       `db:"tax_amount"`
	DiscountAmount float64       `db:"discount_amount"`
	AmountPaid     float64       `db:"amount_paid"`
	Currency       string        `db:"currency"`
	PaymentTerms   string        `db:"payment_terms"`
	Status         PayableStatus `db:"status"`
	PaidStatus     PaidStatus    `db:"paid_status"`
	CreatedBy      int           `db:"created_by"`
}

func (accountsPayableInsert) TableName() string { return "accounts_payable" }

// payableRegister is the write model for the payables cross-reference row.
type payableRegister struct {
	ID                int64 `db:"id" play:"pk,incrementing"`
	CompanyID         int   `db:"company_id"`
	AccountsPayableID int   `db:"accounts_payable_id"`
	VendorID          int   `db:"vendor_id"`
}

func (payableRegister) TableName() string { return "payables" }

// openingInvoiceInsert is the write model for a customer's opening-balance invoice
// — a partial-column insert (the rest of invoices' columns take DB defaults), so
// it is a dedicated model rather than the full invoiceInsert.
type openingInvoiceInsert struct {
	ID         int64         `db:"id" play:"pk,incrementing"`
	CompanyID  int           `db:"company_id"`
	Date       time.Time     `db:"date"`
	Type       TermType      `db:"type"`
	DueOn      time.Time     `db:"due_on"`
	CustomerID int           `db:"customer_id"`
	Amount     float64       `db:"amount"`
	AmountDue  float64       `db:"amount_due"`
	Total      float64       `db:"total"`
	Note       string        `db:"note"`
	Status     InvoiceStatus `db:"status"`
	PaidStatus PaidStatus    `db:"paid_status"`
	Code       string        `db:"code"`
}

func (openingInvoiceInsert) TableName() string { return "invoices" }

// openingPayableInsert is the write model for a vendor's opening-balance AP entry.
// Its column set differs from accountsPayableInsert (no purchase_id; carries
// payment_method + notes), so it is its own model. amount_payable is generated.
type openingPayableInsert struct {
	ID             int64         `db:"id" play:"pk,incrementing"`
	CompanyID      int           `db:"company_id"`
	VendorID       int           `db:"vendor_id"`
	InvoiceNumber  string        `db:"invoice_number"`
	InvoiceDate    time.Time     `db:"invoice_date"`
	DueDate        time.Time     `db:"due_date"`
	AmountTotal    float64       `db:"amount_total"`
	TaxAmount      float64       `db:"tax_amount"`
	DiscountAmount float64       `db:"discount_amount"`
	AmountPaid     float64       `db:"amount_paid"`
	Currency       string        `db:"currency"`
	PaymentTerms   string        `db:"payment_terms"`
	PaymentMethod  string        `db:"payment_method"`
	Status         PayableStatus `db:"status"`
	PaidStatus     PaidStatus    `db:"paid_status"`
	Notes          string        `db:"notes"`
	CreatedBy      int           `db:"created_by"`
}

func (openingPayableInsert) TableName() string { return "accounts_payable" }

// itemUnitRead is the playsql read model for the items_units link table.
//
// No softdelete tag: the LEFT JOIN LATERAL it replaces never filtered deleted_at,
// and the (company_id, item_id) unique constraint added in migration 20260709120000
// does not consider it either.
//
// ItemID is an int, not an int64, so it matches itemRead.ID by Go value — the eager
// loader groups children on that field, and a width mismatch silently yields no
// relation. UnitID is an int64 to match unitRead.ID for the same reason.
type itemUnitRead struct {
	ID        int64 `db:"id" play:"pk,incrementing"`
	CompanyID int   `db:"company_id"`
	ItemID    int   `db:"item_id"`
	UnitID    int64 `db:"unit_id"`

	// Mapped so the upsert stamps them. playsql only emits a timestamp column the
	// model knows about, filling it with Go's time.Now(); attachItemUnit's DO UPDATE
	// list names updated_at, which then compiles to `updated_at = EXCLUDED.updated_at`
	// and carries that stamp. Unmap it and EXCLUDED.updated_at falls back to the
	// column default, CURRENT_TIMESTAMP -- the transaction's start time, which under
	// the txdb test harness never advances.
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`

	Unit *unitRead `play:"belongsTo,fk=unit_id"`
}

func (itemUnitRead) TableName() string { return "items_units" }

// itemRead is the playsql read model for the items table.
//
// identifiers is jsonb. ItemIdentifiers is a struct, and playsql treats a struct
// field as a relation, so the column is scanned raw and unmarshalled in toItem —
// the same shape invoiceRead uses for discount and payment.
//
// TaxID is an int64 to match taxRead.ID.
type itemRead struct {
	ID          int               `db:"id" play:"pk,incrementing"`
	CompanyID   int               `db:"company_id"`
	UUID        string            `db:"uuid"`
	Name        string            `db:"name"`
	Price       float64           `db:"price"`
	Description string            `db:"description"`
	TaxID       int64             `db:"tax_id"`
	ItemType    string            `db:"item_type"`
	Identifiers []byte            `db:"identifiers"`
	Status      foundation.Status `db:"status"`
	CreatedAt   *time.Time        `db:"created_at"`
	UpdatedAt   *time.Time        `db:"updated_at"`
	DeletedAt   *time.Time        `db:"deleted_at" play:"softdelete"`

	Tax      *taxRead      `play:"belongsTo,fk=tax_id"`
	ItemUnit *itemUnitRead `play:"hasOne,fk=item_id"`
}

func (itemRead) TableName() string { return "items" }

// withItemTax narrows the tax eager load to the three columns the old projection
// selected (i.tax_id, t.name, t.rate). Loading the whole row would populate tax.UUID
// and its timestamps, which the item responses have never carried.
//
// id is in the projection because belongsTo matches children on their primary key;
// dropping it would leave every item's tax nil.
func withItemTax(b *playsql.Builder) {
	b.Select("id", "name", "rate")
}

// toItem maps the read model onto the JSON response struct.
//
// Tax came off an INNER JOIN, so it is always present in practice; a nil relation
// leaves the zero tax rather than panicking. Unit came off a LEFT JOIN, so nil is a
// normal result and item.Unit's two fields stay nil pointers.
//
// A malformed identifiers blob leaves Identifiers zero instead of failing the whole
// read, matching how the other jsonb columns are handled.
func (r itemRead) toItem() *item {
	i := &item{
		ID:          r.ID,
		UUID:        r.UUID,
		Name:        r.Name,
		Price:       r.Price,
		Description: r.Description,
		ItemType:    r.ItemType,
		Status:      r.Status,
	}
	i.CreatedAt = r.CreatedAt
	i.UpdatedAt = r.UpdatedAt
	i.DeletedAt = r.DeletedAt

	if len(r.Identifiers) > 0 {
		_ = json.Unmarshal(r.Identifiers, &i.Identifiers)
	}

	if r.Tax != nil {
		i.Tax = *r.Tax.toTax()
	} else {
		i.Tax.ID = r.TaxID
	}

	if r.ItemUnit != nil && r.ItemUnit.Unit != nil {
		id := int(r.ItemUnit.Unit.ID)
		name := r.ItemUnit.Unit.Name
		i.Unit.ID, i.Unit.Name = &id, &name
	}

	return i
}

// lineItemRead is a second read model over `items`, for invoice and purchase lines.
//
// It exists instead of reusing itemRead because itemRead carries a softdelete tag,
// and playsql excludes soft-deleted rows from an eager load. Neither
// findInvoiceLines' nor findPurchaseLines' `INNER JOIN items` filtered deleted_at: a
// line on a since-deleted item still rendered its name and description, and a
// historical document must keep doing so.
type lineItemRead struct {
	ID          int    `db:"id" play:"pk,incrementing"`
	CompanyID   int    `db:"company_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Identifiers []byte `db:"identifiers"`
	TaxID       int64  `db:"tax_id"`

	Tax      *taxRead      `play:"belongsTo,fk=tax_id"`
	ItemUnit *itemUnitRead `play:"hasOne,fk=item_id"`
}

func (lineItemRead) TableName() string { return "items" }

// withInvoiceLineTax narrows the tax eager load to the two columns the old query
// selected, taxes.id and taxes.name. Unlike withItemTax this is a projection
// narrowing only, not a guard: toLine copies those two fields by hand, so a wider
// SELECT would fetch more columns but change nothing in the response. id stays in the
// projection because belongsTo matches on the related primary key.
//
// The line's rate and tax amount come off invoices_items, not off the tax row.
func withInvoiceLineTax(b *playsql.Builder) {
	b.Select("id", "name")
}

// withInvoiceLineItem loads the item's tax and unit from inside the item's own eager
// load, so `Item` is traversed exactly once.
//
// Writing this as two paths on the root — With("Item.ItemUnit.Unit") plus
// WithConstraint("Item.Tax", ...) — does not work: each with-clause reloads every
// segment it names and reassigns the relation field, so the second pass replaces the
// item structs the first had filled in and their ItemUnit goes back to nil. The unit
// then reads as 0, which the recurring-invoice flow writes straight into
// invoices_items.unit_id and trips a foreign key.
func withInvoiceLineItem(b *playsql.Builder) {
	b.With("ItemUnit.Unit").WithConstraint("Tax", withInvoiceLineTax)
}

// invoiceLineRead is the playsql read model for invoices_items.
//
// No softdelete tag, matching InvoiceItem, its write counterpart: the DELETED branch
// of updateInvoiceLines calls Delete on a model without the tag, which is a hard
// DELETE, so no soft-deleted line rows exist. findInvoiceLines never filtered
// deleted_at either. reconcileInvoiceStock's `deleted_at IS NULL` is written out
// explicitly where it is needed.
type invoiceLineRead struct {
	ID          int64      `db:"id" play:"pk,incrementing"`
	CompanyID   int        `db:"company_id"`
	InvoiceID   int        `db:"invoice_id"`
	ItemID      int        `db:"item_id"`
	VariantID   int        `db:"variant_id"`
	UnitID      int        `db:"unit_id"`
	Qty         int64      `db:"qty"`
	Price       float64    `db:"price"`
	Rate        float64    `db:"rate"`
	Amount      float64    `db:"amount"`
	Tax         float64    `db:"tax"`
	Total       float64    `db:"total"`
	WarehouseID int64      `db:"warehouse_id"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`

	Item    *lineItemRead    `play:"belongsTo,fk=item_id"`
	Variant *itemVariantRead `play:"belongsTo,fk=variant_id"`
}

func (invoiceLineRead) TableName() string { return "invoices_items" }

// toLine maps an invoices_items row and its relations onto the response struct.
//
// line.ID carries the *item* id, not the invoices_items id. That is what the old
// query selected into it (`SELECT ii.item_id, ...` scanned into &i.ID), and
// updateInvoiceLines matches lines on (item_id, variant_id), so the client round-trips
// this value as the item id. Preserved deliberately.
//
// line.Tax.Rate and line.Tax.Amount come off the line (ii.rate, ii.tax), not off the
// tax row — a line freezes the rate it was billed at.
//
// Action is the constant 'unchanged' the old projection selected.
func (r invoiceLineRead) toLine() *line {
	l := &line{
		ID:          int64(r.ItemID),
		VariantID:   int64(r.VariantID),
		Qty:         r.Qty,
		Price:       r.Price,
		Amount:      r.Amount,
		Total:       r.Total,
		WarehouseID: r.WarehouseID,
		Action:      UNCHANGED,
	}
	l.CreatedAt = r.CreatedAt
	l.UpdatedAt = r.UpdatedAt
	l.DeletedAt = r.DeletedAt

	l.Tax.Rate = r.Rate
	l.Tax.Amount = r.Tax

	if r.Variant != nil {
		l.VariantName = r.Variant.Name
		if r.Variant.SKU != nil {
			l.VariantSKU = *r.Variant.SKU
		}
	}

	if r.Item != nil {
		l.Name = r.Item.Name
		l.Description = r.Item.Description
		if len(r.Item.Identifiers) > 0 {
			_ = json.Unmarshal(r.Item.Identifiers, &l.Identifier)
		}
		if r.Item.Tax != nil {
			l.Tax.ID = r.Item.Tax.ID
			l.Tax.Name = r.Item.Tax.Name
		}
		if r.Item.ItemUnit != nil {
			l.Unit.ID = r.Item.ItemUnit.UnitID
			if r.Item.ItemUnit.Unit != nil {
				l.Unit.Name = r.Item.ItemUnit.Unit.Name
			}
		}
	}

	// `CASE WHEN iv.is_default THEN it.name ELSE it.name || ' — ' || iv.name END`.
	if r.Item != nil && r.Variant != nil && !r.Variant.IsDefault {
		l.Name = r.Item.Name + " — " + r.Variant.Name
	}

	return l
}

// purchaseLineRead is the playsql read model for purchase_items.
//
// UnitID and TaxID are nullable columns mapped as plain int64 rather than pointers.
// A NULL scans to 0, no row carries id 0, so the belongsTo simply resolves to nil —
// which is exactly the fallback arm of the `COALESCE(pi.unit_id, ...)` and
// `COALESCE(pi.tax_id, it.tax_id)` the old query used. A pointer field would break the
// relation instead: the loader matches a child's primary key against the parent's
// foreign-key field by Go value, and a *int64 never equals an int64.
type purchaseLineRead struct {
	ID         int64      `db:"id" play:"pk,incrementing"`
	CompanyID  int        `db:"company_id"`
	PurchaseID int        `db:"purchase_id"`
	VariantID  int        `db:"variant_id"`
	Qty        float64    `db:"qty"`
	UnitPrice  float64    `db:"unit_price"`
	LineTotal  float64    `db:"line_total"`
	UnitID     int64      `db:"unit_id"`
	TaxID      int64      `db:"tax_id"`
	TaxAmount  float64    `db:"tax_amount"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at" play:"softdelete"`

	Variant *itemVariantRead `play:"belongsTo,fk=variant_id"`
	// SelectedUnit is the line's own unit override (LEFT JOIN units ON pi.unit_id).
	SelectedUnit *unitRead `play:"belongsTo,fk=unit_id"`
	// LineTax is the line's own tax override (the pi.tax_id arm of the COALESCE).
	LineTax *taxRead `play:"belongsTo,fk=tax_id"`
}

func (purchaseLineRead) TableName() string { return "purchase_items" }

// withPurchaseLineVariant loads the variant's item, and the item's tax and unit,
// inside the variant's own eager load. Everything below Variant is reached in one
// traversal for the reason withInvoiceLineItem documents: two `With` paths sharing a
// prefix reload that prefix and discard what the earlier path attached to it.
//
// Splitting this into With("Variant.Item.ItemUnit.Unit") + With("Variant.Item.Tax")
// leaves the item's unit nil. It shows up only on lines that have no unit override of
// their own, since a line with pi.unit_id set reads its unit off SelectedUnit and never
// consults the item's — which is why the fallback is the case worth testing.
func withPurchaseLineVariant(b *playsql.Builder) {
	b.WithConstraint("Item", withPurchaseLineItem)
}

// withPurchaseLineItem loads the item's unit and its full tax row.
//
// It cannot reuse withInvoiceLineItem: that one narrows the tax to (id, name) because
// an invoice line freezes its own rate on invoices_items. A purchase line has no rate
// column — the old query selected taxes.rate — so the rate must come off the tax row.
func withPurchaseLineItem(b *playsql.Builder) {
	b.With("ItemUnit.Unit").With("Tax")
}

// toPurchaseLine maps a purchase_items row and its relations onto the response struct.
//
// Like toLine, line.ID carries the *item* id — the old projection selected `it.id`
// into it. The item is reached through the variant (iv.item_id); purchase_items has
// no item_id column of its own.
//
// Two COALESCEs become nil checks:
//
//	COALESCE(pi.unit_id, items_units.unit_id)      -> SelectedUnit, else the item's unit
//	COALESCE(unit_selected.name, items_units.name) -> same, for the name
//	COALESCE(pi.tax_id, it.tax_id)                 -> LineTax, else the item's tax
//
// Unlike an invoice line, a purchase line's rate is read off the tax row, not frozen
// on the line: the old query selected taxes.rate. Only tax_amount comes off the line.
func (r purchaseLineRead) toPurchaseLine() *line {
	l := &line{
		ID:        0,
		VariantID: int64(r.VariantID),
		Qty:       int64(r.Qty),
		Price:     r.UnitPrice,
		Amount:    r.Qty * r.UnitPrice,
		Total:     r.LineTotal,
		Action:    UNCHANGED,
	}
	l.CreatedAt = r.CreatedAt
	l.UpdatedAt = r.UpdatedAt
	l.DeletedAt = r.DeletedAt
	l.Tax.Amount = r.TaxAmount

	var item *lineItemRead
	if r.Variant != nil {
		l.VariantName = r.Variant.Name
		if r.Variant.SKU != nil {
			l.VariantSKU = *r.Variant.SKU
		}
		item = r.Variant.Item
	}

	if item != nil {
		l.ID = int64(item.ID)
		l.Name = item.Name
		l.Description = item.Description
		if len(item.Identifiers) > 0 {
			_ = json.Unmarshal(item.Identifiers, &l.Identifier)
		}
	}

	// `CASE WHEN iv.is_default THEN it.name ELSE it.name || ' — ' || iv.name END`.
	if item != nil && r.Variant != nil && !r.Variant.IsDefault {
		l.Name = item.Name + " — " + r.Variant.Name
	}

	// COALESCE(pi.unit_id, items_units.unit_id) and its name.
	if r.SelectedUnit != nil {
		l.Unit.ID = r.SelectedUnit.ID
		l.Unit.Name = r.SelectedUnit.Name
	} else if item != nil && item.ItemUnit != nil {
		l.Unit.ID = item.ItemUnit.UnitID
		if item.ItemUnit.Unit != nil {
			l.Unit.Name = item.ItemUnit.Unit.Name
		}
	}

	// COALESCE(pi.tax_id, it.tax_id): the line's own tax wins, else the item's.
	tax := r.LineTax
	if tax == nil && item != nil {
		tax = item.Tax
	}
	if tax != nil {
		l.Tax.ID = tax.ID
		l.Tax.Name = tax.Name
		l.Tax.Rate = tax.Rate
	}

	return l
}

// movementTimeLayout is the shape TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') produced.
// created_at is `timestamp without time zone`, so formatting the scanned value
// reproduces the same wall clock the database rendered.
const movementTimeLayout = "2006-01-02 15:04"

func formatMovementTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(movementTimeLayout)
}

// withVariantItem loads the variant's item in the variant's own eager load, so the
// path below Variant is traversed once. See withInvoiceLineItem for why two `With`
// paths sharing a prefix cannot both survive.
func withVariantItem(b *playsql.Builder) { b.With("Item") }

// liveVariantAndItem reproduces `AND iv.deleted_at IS NULL AND i.deleted_at IS NULL`
// on the movement reads' INNER JOINs. Neither itemVariantRead nor lineItemRead carries
// a softdelete tag, so the predicates are written out.
//
// The dotted path puts the closure on the innermost segment, so the two calls filter
// the variant and the item respectively.
func liveVariantAndItem(b *playsql.Builder) *playsql.Builder {
	return b.
		WhereHas("Variant", func(q *playsql.Builder) { q.WhereNull("deleted_at") }).
		WhereHas("Variant.Item", func(q *playsql.Builder) { q.WhereNull("deleted_at") })
}

// itemNameOf and variantSKUOf unwrap the eager-loaded relation defensively: the reads
// that use them filter on the relation with WhereHas, so it is always present.
func itemNameOf(v *itemVariantRead) string {
	if v == nil || v.Item == nil {
		return ""
	}
	return v.Item.Name
}

func variantSKUOf(v *itemVariantRead) string {
	if v == nil || v.SKU == nil {
		return ""
	}
	return *v.SKU
}

func variantNameOf(v *itemVariantRead) string {
	if v == nil {
		return ""
	}
	return v.Name
}

// toMovementRow maps an inventory_movements row and its relations onto the response.
//
// The warehouse came off an INNER JOIN with no deleted_at predicate, so the relation is
// loaded WithTrashed and a retired warehouse still names the movement it recorded.
func (r inventoryMovementRead) toMovementRow() *inventoryMovementRow {
	row := &inventoryMovementRow{
		ID:            r.ID,
		VariantID:     int64(r.VariantID),
		VariantName:   variantNameOf(r.Variant),
		SKU:           variantSKUOf(r.Variant),
		ItemName:      itemNameOf(r.Variant),
		WarehouseID:   int64(r.WarehouseID),
		Kind:          string(r.TransactionKind),
		Qty:           r.Qty,
		UnitCost:      r.UnitCost,
		ReferenceType: r.ReferenceType,
		ReferenceID:   int64(r.ReferenceID),
		CreatedAt:     formatMovementTime(r.CreatedAt),
	}
	if r.Warehouse != nil {
		row.Warehouse = r.Warehouse.Name
	}
	return row
}

// toAdjustmentRow maps an adjustment movement onto its response struct. `notes` was a
// literal ” in the projection and stays empty; `reason` is the movement's
// reference_type.
func (r inventoryMovementRead) toAdjustmentRow() *adjustmentRow {
	row := &adjustmentRow{
		ID:          r.ID,
		VariantID:   int64(r.VariantID),
		VariantName: variantNameOf(r.Variant),
		SKU:         variantSKUOf(r.Variant),
		ItemName:    itemNameOf(r.Variant),
		WarehouseID: int64(r.WarehouseID),
		Qty:         r.Qty,
		Reason:      r.ReferenceType,
		Notes:       "",
		CreatedAt:   formatMovementTime(r.CreatedAt),
	}
	if r.Warehouse != nil {
		row.Warehouse = r.Warehouse.Name
	}
	return row
}

// toTransferLineRow maps a transfer line and its relations onto the response struct.
// line_total was computed in the projection; reference is pulled out of the item's
// jsonb identifiers.
func (r inventoryTransferLine) toTransferLineRow() *transferLineRow {
	row := &transferLineRow{
		ID:          r.ID,
		VariantID:   int64(r.VariantID),
		VariantName: variantNameOf(r.Variant),
		SKU:         variantSKUOf(r.Variant),
		ItemName:    itemNameOf(r.Variant),
		Qty:         r.Qty,
		UnitCost:    r.UnitCost,
		LineTotal:   r.Qty * r.UnitCost,
	}
	if r.Description != nil {
		row.Description = *r.Description
	}
	if r.Unit != nil {
		row.Unit = r.Unit.Name
	}
	if r.Variant != nil && r.Variant.Item != nil && len(r.Variant.Item.Identifiers) > 0 {
		var ids ItemIdentifiers
		if err := json.Unmarshal(r.Variant.Item.Identifiers, &ids); err == nil && ids.Reference != nil {
			row.Reference = *ids.Reference
		}
	}
	return row
}
