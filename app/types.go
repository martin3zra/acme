package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/mailer"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
	"github.com/martin3zra/acme/pkg/validator"
)

type LoginFormRequest struct {
	support.FormRequest
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (f LoginFormRequest) Rules() map[string]any {
	return map[string]any{
		"email":    "required|email|max:100|lowercase",
		"password": "required",
	}
}

type CreatePasswordForm struct {
	support.FormRequest
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (f CreatePasswordForm) Rules() map[string]any {
	return map[string]any{
		"password": "required|confirmed",
	}
}

type StoreCustomerForm struct {
	support.FormRequest
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	CustomerType    string    `json:"customer_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (StoreCustomerForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("customers", "email"),
		},
		"phone":              "sometimes|min:3|max:120",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limited":     "required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

func (form StoreCustomerForm) Authorize() bool {
	return Can(form.User(), "create:customer")
}

type UpdateCustomerForm struct {
	support.FormRequest
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	CustomerType    string    `json:"customer_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (form UpdateCustomerForm) Authorize() bool {
	return Can(form.User(), "update:customer")
}

func (form UpdateCustomerForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("customers", "email").Ignore(form.ID, "id"),
		},
		"phone":              "sometimes|min:3|max:120",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

type ConfirmsPasswords struct {
	support.FormRequest
	Password string `json:"current_password"`
}

func (ConfirmsPasswords) Rules() map[string]any {
	return map[string]any{
		"current_password": "required|current_password",
	}
}

type ItemIdentifiers struct {
	Reference       *string `json:"reference,omitempty"`
	Code            *string `json:"code,omitempty"`
	SKU             *string `json:"sku,omitempty"`
	Barcode         *string `json:"barcode,omitempty"`
	VendorReference *string `json:"vendor_reference,omitempty"`
}

func (d *ItemIdentifiers) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *ItemIdentifiers) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type StoreItemForm struct {
	support.FormRequest
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

func (form StoreItemForm) Authorize() bool {
	return Can(form.User(), "create:item")
}

func (StoreItemForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			"max:120",
			validator.Rule{}.Unique("items", "name"),
		},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       "bail|required|exists:taxes,id",
		"unit_id":                      "bail|required|exists:units,id",
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type UpdateItemForm struct {
	support.FormRequest
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

func (form UpdateItemForm) Authorize() bool {
	return Can(form.User(), "update:item")
}

func (form UpdateItemForm) Rules() map[string]any {
	return map[string]any{
		"name":                         []any{"required", "min:3", "max:120", validator.Rule{}.Unique("items", "name").Ignore(form.ID, "id")},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       "required|exists:taxes,id",
		"unit_id":                      "required|exists:units,id",
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type TermType string

const (
	_CASH    TermType = "cash"
	_CREDIT  TermType = "credit"
	_OPENING TermType = "opening"
)

var InvoiceTermType = struct {
	Cash    TermType
	Credit  TermType
	Opening TermType
}{
	Cash:    _CASH,
	Credit:  _CREDIT,
	Opening: _OPENING,
}

type Role string

const (
	_DEVELOPER  Role = "developer"  //  Access to APIs, integrations, and developer tools
	_OWNER      Role = "owner"      // Full control of billing, settings, and organization
	_ADMIN      Role = "admin"      //  Manages users, roles, and global settings
	_SUPERVISOR Role = "supervisor" // Manages team data, limited settings access
	_STANDARD   Role = "standard"   // Regular user with core feature access
)

var Roles = struct {
	Developer  Role
	Owner      Role
	Admin      Role
	Supervisor Role
	Standard   Role
}{
	Developer:  _DEVELOPER,
	Owner:      _OWNER,
	Admin:      _ADMIN,
	Supervisor: _SUPERVISOR,
	Standard:   _STANDARD,
}

var RoleMap = []map[string]any{
	// {"id": string(Roles.Developer), "label": Roles.Developer},
	// {"id": string(Roles.Owner), "label": Roles.Owner},
	{"id": string(Roles.Admin), "label": Roles.Admin, "description": "Manages users, roles, and global settings"},
	{"id": string(Roles.Supervisor), "label": Roles.Supervisor, "description": "Manages team data, limited settings access"},
	{"id": string(Roles.Standard), "label": Roles.Standard, "description": "Regular user with core feature access"},
}

type PaidStatus string

const (
	_PAID_UNPAID  PaidStatus = "unpaid"
	_PAID_PARTIAL PaidStatus = "partial"
	_PAID_PAID    PaidStatus = "paid"
	_PAID_REMOVED PaidStatus = "removed"
)

var PaidStatuses = struct {
	UnPaid  PaidStatus
	Partial PaidStatus
	Paid    PaidStatus
	Removed PaidStatus
}{
	UnPaid:  _PAID_UNPAID,
	Partial: _PAID_PARTIAL,
	Paid:    _PAID_PAID,
	Removed: _PAID_REMOVED,
}

type InvoiceStatus string

const (
	_INVOICE_DRAFT     InvoiceStatus = "draft"
	_INVOICE_OPEN      InvoiceStatus = "open"
	_INVOICE_SENT      InvoiceStatus = "sent"
	_INVOICE_VIEWED    InvoiceStatus = "viewed"
	_INVOICE_OVERDUE   InvoiceStatus = "overdue"
	_INVOICE_COMPLETED InvoiceStatus = "completed"
	_INVOICE_VOID      InvoiceStatus = "void"
	_INVOICE_PARTIAL   InvoiceStatus = "partial"
)

var InvoiceStatuses = struct {
	Open      InvoiceStatus
	Draft     InvoiceStatus
	Sent      InvoiceStatus
	Viewed    InvoiceStatus
	Overdue   InvoiceStatus
	Completed InvoiceStatus
	Void      InvoiceStatus
	Partial   InvoiceStatus
}{
	Open:      _INVOICE_OPEN,
	Draft:     _INVOICE_DRAFT,
	Sent:      _INVOICE_SENT,
	Viewed:    _INVOICE_VIEWED,
	Overdue:   _INVOICE_OVERDUE,
	Completed: _INVOICE_COMPLETED,
	Void:      _INVOICE_VOID,
	Partial:   _INVOICE_PARTIAL,
}

type PaymentStatus string

const (
	_PAYMENT_VOID      PaymentStatus = "void"
	_PAYMENT_PENDING   PaymentStatus = "pending"
	_PAYMENT_COMPLETED PaymentStatus = "completed"
	_PAYMENT_FAILED    PaymentStatus = "failed"
)

var PaymentStatuses = struct {
	Void      PaymentStatus
	Pending   PaymentStatus
	Completed PaymentStatus
	Failed    PaymentStatus
}{
	Void:      _PAYMENT_VOID,
	Pending:   _PAYMENT_PENDING,
	Completed: _PAYMENT_COMPLETED,
	Failed:    _PAYMENT_FAILED,
}

type RedirectPreferences struct {
	Invoice  string `json:"invoice"`
	Customer string `json:"customer"`
	Product  string `json:"product"`
	Payment  string `json:"payment"`
}

type LineAction string

const (
	ADDED     LineAction = "added"
	UPDATED   LineAction = "updated"
	DELETED   LineAction = "deleted"
	UNCHANGED LineAction = "unchanged"
)

var LineActions = struct {
	Added     LineAction
	Updated   LineAction
	Deleted   LineAction
	Unchanged LineAction
}{
	Added:     ADDED,
	Updated:   UPDATED,
	Deleted:   DELETED,
	Unchanged: UNCHANGED,
}

type Discount struct {
	Val  float64 `json:"value"`
	Type string  `json:"type"`
}

func (d *Discount) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Discount) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type PaymentAmount struct {
	Amount float64 `json:"amount"`
}

type PaymentBase struct {
	PaymentAmount
	Reference string `json:"reference"`
}

type Cash struct {
	PaymentAmount
}

type Check struct {
	PaymentBase
}

type Card struct {
	PaymentBase
	Last4 int    `json:"last4"`
	Brand string `json:"brand"`
}

type Bt struct {
	PaymentBase
}

type Payment struct {
	Cash  Cash  `json:"cash"`
	Check Check `json:"ck"`
	Card  Card  `json:"card"`
	Bt    Bt    `json:"bt"`
}

func (d *Payment) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Payment) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type Line struct {
	ID       int        `json:"id"`
	Unit     int        `json:"unit"`
	Qty      int        `json:"qty"`
	Price    float64    `json:"price"`
	Rate     float64    `json:"rate"`
	Action   LineAction `json:"action"`
	tax      float64
	amount   float64
	discount float64
	total    float64
}

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int       `json:"customer_id"`
	Date       time.Time `json:"date"`
	Terms      string    `json:"terms"`
	TaxReceipt int       `json:"tax_receipt"`
	Discount   Discount  `json:"discount"`
	Notes      string    `json:"notes"`
	Lines      []*Line   `json:"lines"`
	Payment    Payment   `json:"payment"`
	// considere these fields as protected
	amount     float64
	amountDue  float64
	tax        float64
	total      float64
	paidStatus PaidStatus
	dueOn      *time.Time
	termType   TermType
}

func (form StoreInvoiceForm) Authorize() bool {
	return Can(form.User(), "create:invoice")
}

func (form StoreInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":    "bail|required|exists:customers,id",
		"date":           "bail|required|date|after:yesterday",
		"terms":          "bail|required|min:1",
		"tax_receipt":    "bail|required|exists:tax_receipts,id",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1",
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added",
		"discount":       "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

func (StoreInvoiceForm) Messages() map[string]string {
	return map[string]string{
		"customer_id.required": "You must specify the customer you want to invoice.",
		"lines.min":            "You must specify at least one item to invoice.",
	}
}

func (form *StoreInvoiceForm) PassedValidation() {
	// compute tax for each line
	form.computeTax()

	form.dueOn = nil
	form.paidStatus = PaidStatuses.Paid
	form.termType = InvoiceTermType.Cash
	termInDays := getNetDays(form.Terms)
	if termInDays > 1 {
		form.amountDue = form.total
		form.paidStatus = PaidStatuses.UnPaid
		form.termType = InvoiceTermType.Credit

		dueDate := form.Date.AddDate(0, 0, termInDays)
		form.dueOn = &dueDate
	}
}

func (form *StoreInvoiceForm) paymentTotalAmount() float64 {
	return form.Payment.Cash.Amount + form.Payment.Card.Amount + form.Payment.Check.Amount + form.Payment.Bt.Amount
}

func (form *StoreInvoiceForm) computeTax() {

	discountPercentage := form.Discount.Val
	if form.Discount.Type == "fixed" {
		totalAmount := float64(0)
		for _, line := range form.Lines {
			totalAmount += (line.Price * float64(line.Qty))
		}

		discountPercentage = float64(discountPercentage/totalAmount) * 100
	}

	for _, line := range form.Lines {
		if line.Action == LineActions.Deleted {
			continue
		}
		// We can store the line discoun on the database
		// We can add a discount value amount to the invoice.
		line.amount = round(line.Price*float64(line.Qty), 2)
		line.discount = round(line.amount*(discountPercentage/100), 2)
		line.tax = round((line.amount-line.discount)*(line.Rate/100), 2)
		line.total = round(line.amount-line.discount+line.tax, 2)

		form.tax += line.tax
		form.amount += line.amount
		form.total += line.total
	}

}

type UpdateInvoiceForm struct {
	StoreInvoiceForm
}

func (form UpdateInvoiceForm) Authorize() bool {
	return Can(form.User(), "update:invoice")
}

func (form UpdateInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":    "bail|required|exists:customers,id",
		"date":           "bail|required|date",
		"terms":          "bail|required|min:1",
		"tax_receipt":    "bail|required|exists:tax_receipts,id",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1", // ADD when rule here, only validate when is the action is added or updated
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added,updated,deleted,unchanged",
		"discount":       "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

type PaymentLine struct {
	ID        int        `json:"id"`
	Uuid      string     `json:"uuid"`
	AmountDue float64    `json:"amount_due"`
	Payment   float64    `json:"payment"`
	Discount  float64    `json:"discount"`
	Action    LineAction `json:"action"`
}

type StorePaymentForm struct {
	support.FormRequest
	CustomerID string         `json:"customer_id"`
	Date       time.Time      `json:"date"`
	Notes      string         `json:"notes"`
	Lines      []*PaymentLine `json:"lines"`
	Payment    Payment        `json:"payment"`
	Amount     float64        `json:"amount"`
}

func (form StorePaymentForm) Authorize() bool {
	return Can(form.User(), "create:payment")
}

func (form StorePaymentForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":        "bail|required|exists:customers,uuid",
		"date":               "bail|required|date|after:yesterday",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.uuid":       "required|exists:invoices,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		// "lines.*.action": "required|in:added",
	}
}

type UpdatePaymentForm struct {
	support.FormRequest
	CustomerID string         `json:"customer_id"`
	Date       time.Time      `json:"date"`
	Notes      string         `json:"notes"`
	Lines      []*PaymentLine `json:"lines"`
	Payment    Payment        `json:"payment"`
	Amount     float64        `json:"amount"`
}

func (form UpdatePaymentForm) Authorize() bool {
	return Can(form.User(), "update:payment")
}

func (form UpdatePaymentForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":        "bail|required|exists:customers,uuid",
		"date":               "bail|required|date",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.id":         "required|exists:receivables_income_items,id",
		"lines.*.uuid":       "required|exists:invoices,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		"lines.*.action":     "required|in:added,updated,deleted,unchanged",
	}
}

type MustVerifyAccount interface {
	HasVerifiedAccount() bool
	MarkAccountAsVerified(*sql.DB) bool
	SendAccountVerificationNotification(mailer.Mailer, map[string]string)
	GetEmailAddressForAccountVerification() string
}

type EmailVerificationForm struct {
	support.FormRequest
	Email string `json:"email"`
}

func (EmailVerificationForm) Rules() map[string]any {
	return map[string]any{
		"email": "required|email",
	}
}

type StoreCompanyForm struct {
	support.FormRequest
	Name    string `json:"name"`
	RNC     string `json:"rnc"`
	City    string `json:"city"`
	Address string `json:"address"`
}

func (form StoreCompanyForm) Authorize() bool {
	db := form.Context().Value(database.ConnectionKey{}).(*sql.DB)
	u := UserFromFoundationUser(form.User())
	if u.IsOwner(db) {
		return true
	}
	return Can(form.User(), "create:company") // OR OWNER
}

func (StoreCompanyForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			validator.Rule{}.Unique("companies", "name"),
		},
		"rnc":     "required|min:9|max:11:unique:companies,rnc",
		"city":    "required",
		"address": "required",
	}
}

type StoreProfileForm struct {
	support.FormRequest
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (form StoreProfileForm) Rules() map[string]any {

	db := form.Context().Value(database.ConnectionKey{}).(*sql.DB)
	if db == nil {
		panic("database connection need to be set.")
	}

	var userId int
	if err := db.QueryRow("SELECT owner_id FROM accounts WHERE uuid = $1", form.Param("account")).Scan(&userId); err != nil {
		panic("unable to find the account owner.")
	}

	return map[string]any{
		"name": "required|min:3",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.
				Unique("users", "email").
				Ignore(userId, "id"),
		},
	}
}

type CompanyRole struct {
	Company string `json:"company"`
	Role    string `json:"role"`
}

type StoreUserForm struct {
	support.FormRequest
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	Companies []CompanyRole `json:"companies"`
}

func (form StoreUserForm) Authorize() bool {
	return Can(form.User(), "create:user")
}

func (form StoreUserForm) Rules() map[string]any {
	return map[string]any{
		"name": "required|min:3",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("users", "email").Ignore(form.Param("id"), "uuid"),
		},
		"companies":           "required|min:1",
		"companies.*.company": "required|exists:companies,uuid",
		"companies.*.role":    "required|in:admin,supervisor,standard",
	}
}

type User struct {
	foundation.User
	account *account
	Linked  int `json:"linked"`
}

func (u *User) SendEmailVerification(notify mailer.Mailer, attributes map[string]string) {
	url, err := routing.TemporarySignedURL(
		attributes["url"],
		map[string]string{},
		attributes["secret"],
		60*time.Minute,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	notify.
		To(u.Email, u.Name).
		Send(mail.NewVerification(foundation.AsMap(u), url))
}

func (u *User) SendEmailVerificationChange(notify mailer.Mailer, attributes map[string]string) {
	url, err := routing.TemporarySignedURL(
		attributes["url"],
		map[string]string{},
		attributes["secret"],
		60*time.Minute,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	notify.
		To(*u.PendingEmail, u.Name).
		Send(mail.NewVerification(foundation.AsMap(u), url))
}

func (u *User) HasVerifiedEmail() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) MarkEmailAsVerified(db *sql.DB) bool {
	_, err := db.Exec("UPDATE users SET email_verified_at = now(), updated_at = now() WHERE id = $1", u.Id)
	return err == nil
}

func (u *User) Account(db *sql.DB) *account {
	var a = new(account)
	if err := db.QueryRow("SELECT id, uuid, owner_id, status, verified_at, created_at, updated_at, deleted_at FROM accounts WHERE owner_id = $1", u.Id).
		Scan(&a.ID, &a.UUID, &a.Owner.ID, &a.Status, &a.VerifiedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		log.Println("An error occurred fetching the account using the ownerID:", err)
		return nil
	}

	u.account = a

	return a
}

func (u *User) OwnedBy(db *sql.DB) (*account, error) {
	if u.account != nil {
		return u.account, nil
	}

	var a = new(account)
	if err := db.QueryRow(`
    SELECT accounts.id, accounts.uuid, accounts.owner_id, accounts.status, accounts.verified_at, accounts.created_at, accounts.updated_at, accounts.deleted_at
    FROM accounts
    INNER JOIN accounts_users on accounts.id = accounts_users.account_id
    WHERE accounts_users.user_id = $1
  `, u.Id).
		Scan(&a.ID, &a.UUID, &a.Owner.ID, &a.Status, &a.VerifiedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		return nil, err
	}

	u.account = a

	return a, nil
}

func (u *User) IsOwner(db *sql.DB) bool {
	return u.Account(db) != nil
}

func (u *User) IsNotOwner(db *sql.DB) bool {
	return !u.IsOwner(db)
}

func (u *User) IsOwned(db *sql.DB) bool {
	_, err := u.OwnedBy(db)
	return err == nil
}

func (u *User) IsNotOwned(db *sql.DB) bool {
	return !u.IsOwned(db)
}

func (u *User) IsOrphan(db *sql.DB) bool {
	return u.IsNotOwned(db) && u.IsNotOwner(db)
}

func UserFromContext(ctx context.Context) *User {
	u := auth.User(ctx)
	return &User{
		User: *u,
	}
}

func UserFromFoundationUser(u *foundation.User) *User {
	// TODO set the role here.
	return &User{
		User: *u,
	}
}

func (u *User) currentCompany(db *sql.DB) (*Company, error) {
	result := db.QueryRow(`
    SELECT companies.id, companies.uuid, companies.name, companies.identifier, companies.city,
    companies.address, companies.created_at, companies.updated_at, companies_users.role
    FROM companies
    JOIN companies_users ON companies.id = companies_users.company_id
    WHERE companies_users.user_id = $1 AND companies_users.current = true
  `, u.Id)
	var company Company
	err := result.Scan(
		&company.ID,
		&company.UUID,
		&company.Name,
		&company.Identifier,
		&company.City,
		&company.Address,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.UserRole,
	)
	if err != nil {
		return nil, err
	}

	return &company, err
}

func CurrentCompany(ctx context.Context) *Company {
	cc := ctx.Value(support.CompanyKey{})
	if cc == nil {
		return nil
	}

	return cc.(*Company)
}

func CurrentAccount(ctx context.Context) int {
	ac := ctx.Value(support.AccountKey{})
	if ac == nil {
		return 0
	}

	data, ok := ac.(map[string]any)
	if !ok {
		return 0
	}
	if val, ok := data["id"]; ok {
		if intVal, err := toInt(val); err == nil {
			return intVal
		}
	}
	return 0
}

type SequenceConfig struct {
	Prefix  string `json:"prefix"`
	Next    int    `json:"next"`
	Padding int    `json:"padding"`
}

type InvoiceSequence struct {
	Default SequenceConfig `json:"default"`
	Credit  SequenceConfig `json:"credit"`
	Cash    SequenceConfig `json:"cash"`
}

type CompanySequence struct {
	Invoice  InvoiceSequence `json:"invoice"`
	Customer SequenceConfig  `json:"customer"`
	// Estimate SequenceConfig  `json:"estimate"`
	Payment SequenceConfig `json:"payment"`
}

type SequenceForm struct {
	support.FormRequest
	CompanySequence
}

func (SequenceForm) Rules() map[string]any {
	return map[string]any{
		"invoice":                 "required",
		"invoice.default.padding": "required|min:3",
		"invoice.default.next":    "required|min:1",
		"invoice.cash.padding":    "required|min:3",
		"invoice.cash.next":       "required|min:1",
		"invoice.credit.padding":  "required|min:3",
		"invoice.credit.next":     "required|min:1",
		"customer":                "required",
		"customer.padding":        "required|min:3",
		"customer.next":           "required|min:1",
		// "estimate":                "required",
		// "estimate.padding":        "required|min:3",
		// "estimate.next":           "required|min:1",
		"payment":         "required",
		"payment.padding": "required|min:3",
		"payment.next":    "required|min:1",
	}
}

func (form SequenceForm) Authorize() bool {
	return Can(form.User(), "update:company:sequence")
}

type PrerequisiteResult struct {
	Resource string `json:"resource"`
	Ok       bool   `json:"ok"`
	Missing  []struct {
		Key     string `json:"key"`
		Message string `json:"message"`
	} `json:"missing"`
}

type prereqCacheKeyType struct{}

var prereqCacheKey = prereqCacheKeyType{}

type prereqCache map[string]PrerequisiteResult

var (
	ErrPrerequisitesMissing = errors.New("resource prerequisites missing")
	ErrSettingsNotFound     = errors.New("company settings not found")
	ErrInvalidConfiguration = errors.New("invalid resource configuration")
)

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type PresetRange struct {
	Label string `json:"label"`
	Key   string `json:"key"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
}

type ReportForm struct {
	support.FormRequest
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (ReportForm) Rules() map[string]any {
	return map[string]any{
		"from": "required|date",
		"to":   "required|date|before_or_equals:from",
	}
}

type ReportSalesForm struct {
	support.FormRequest
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	ReportType   string    `json:"reportType"`
	ShowInvoices bool      `json:"showInvoices"`
}

func (ReportSalesForm) Rules() map[string]any {
	return map[string]any{
		"from":       "required|date",
		"to":         "required|date|before_or_equals:from",
		"reportType": "required|in:sales_by_item,sales_by_customer,sales_by_date",
	}
}

type InvoiceType string

const (
	InvoiceTypeAll    InvoiceType = "all"
	InvoiceTypeCash   InvoiceType = "cash"
	InvoiceTypeCredit InvoiceType = "credit"
)

// Validate ensures the value is one of the allowed constants
func (t InvoiceType) Validate() error {
	switch t {
	case InvoiceTypeAll, InvoiceTypeCash, InvoiceTypeCredit:
		return nil
	default:
		return fmt.Errorf("invalid invoice type: %s", t)
	}
}

type CustomerType string

const (
	CustomerTypeAll        CustomerType = "all"
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeBusiness   CustomerType = "business"
)

// Validate ensures the value is one of the allowed constants
func (t CustomerType) Validate() error {
	switch t {
	case CustomerTypeAll, CustomerTypeIndividual, CustomerTypeBusiness:
		return nil
	default:
		return fmt.Errorf("invalid customer type: %s", t)
	}
}

type ItemType string

const (
	ItemTypeAll     ItemType = "all"
	ItemTypeProduct ItemType = "product"
	ItemTypeService ItemType = "service"
)

// Validate ensures the value is one of the allowed constants
func (t ItemType) Validate() error {
	switch t {
	case ItemTypeAll, ItemTypeProduct, ItemTypeService:
		return nil
	default:
		return fmt.Errorf("invalid item type: %s", t)
	}
}
