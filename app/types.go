package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/mailer"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/support"
)

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

type MustVerifyAccount interface {
	HasVerifiedAccount() bool
	MarkAccountAsVerified(*sql.DB) bool
	SendAccountVerificationNotification(mailer.Mailer, map[string]string)
	GetEmailAddressForAccountVerification() string
}

type User struct {
	AuthUser
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

// MarkEmailAsVerified stamps the verification time. The raw statement discarded the
// affected-row count, so verifying a user id that matches no row returned true;
// mustAffectRows makes that false, as the name promises.
//
// updated_at is stamped by Update, since userModel maps the column.
func (u *User) MarkEmailAsVerified(db *sql.DB) bool {
	pdb, err := playOn(db)
	if err != nil {
		return false
	}

	affected, err := pdb.Model(&userModel{}).
		WhereEq("id", u.Id).
		Update(context.Background(), map[string]any{"email_verified_at": time.Now()})
	return mustAffectRows(affected, err, "user") == nil
}

// Account returns the account this user owns, or nil.
func (u *User) Account(db *sql.DB) *account {
	pdb, err := playOn(db)
	if err != nil {
		log.Println("An error occurred fetching the account using the ownerID:", err)
		return nil
	}

	var row accountRead
	if err := pdb.Model(&accountRead{}).
		WhereEq("owner_id", u.Id).
		First(context.Background(), &row); err != nil {
		log.Println("An error occurred fetching the account using the ownerID:", err)
		return nil
	}

	a := row.toAccountStruct()
	u.account = a

	return a
}

func (u *User) OwnedBy(db *sql.DB) (*account, error) {
	if u.account != nil {
		return u.account, nil
	}

	pdb, err := playOn(db)
	if err != nil {
		return nil, err
	}

	// Rooted on the pivot: the INNER JOIN becomes a belongsTo, and Has("Account") keeps
	// the join's semantics — a membership row whose account is gone matches nothing.
	var link accountUserModel
	if err := pdb.Model(&accountUserModel{}).
		With("Account").
		Has("Account").
		WhereEq("user_id", u.Id).
		First(context.Background(), &link); err != nil {
		return nil, err
	}
	if link.Account == nil {
		return nil, sql.ErrNoRows
	}

	a := link.Account.toAccountStruct()
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
	return UserFromFoundationUser(auth.User(ctx))
}

func UserFromFoundationUser(u foundation.Authenticatable) *User {
	au, ok := u.(*AuthUser)
	if !ok || au == nil {
		return &User{}
	}
	return &User{
		AuthUser: *au,
	}
}

// currentCompany reads the user's current company together with the role they hold in
// it. The role lives on the pivot, so the read is rooted there and the company arrives
// through a belongsTo; Has("Company") preserves the INNER JOIN.
func (u *User) currentCompany(db *sql.DB) (*Company, error) {
	pdb, err := playOn(db)
	if err != nil {
		return nil, err
	}

	var link companyUserModel
	if err := pdb.Model(&companyUserModel{}).
		With("Company").
		Has("Company").
		WhereEq("user_id", u.Id).
		WhereEq("current", true).
		First(context.Background(), &link); err != nil {
		return nil, err
	}
	if link.Company == nil {
		return nil, sql.ErrNoRows
	}

	company := link.Company.toCompany()
	company.UserRole = link.Role

	return company, nil
}

func CurrentCompany(ctx context.Context) *Company {
	cc := ctx.Value(CompanyKey{})
	if cc == nil {
		return nil
	}

	return cc.(*Company)
}

func CurrentAccount(ctx context.Context) int {
	ac := ctx.Value(AccountKey{})
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
	Template SequenceConfig  `json:"template"`
	Customer SequenceConfig  `json:"customer"`
	Vendor   SequenceConfig  `json:"vendor"`
	Estimate SequenceConfig  `json:"estimate"`
	Payment  SequenceConfig  `json:"payment"`
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
		"vendor":                  "required",
		"vendor.padding":          "required|min:3",
		"vendor.next":             "required|min:1",
		"estimate":                "required",
		"estimate.padding":        "required|min:3",
		"estimate.next":           "required|min:1",
		"payment":                 "required",
		"payment.padding":         "required|min:3",
		"payment.next":            "required|min:1",
		"template":                "sometimes",
		"template.padding":        "sometimes|min:3",
		"template.next":           "sometimes|min:1",
	}
}

func (form SequenceForm) Authorize() bool {
	return Can(form.User(), "update:company:sequence")
}

type Missing struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

type PrerequisiteResult struct {
	Resource string    `json:"resource"`
	Ok       bool      `json:"ok"`
	Missing  []Missing `json:"missing"`
}

type prereqCacheKeyType struct{}

var prereqCacheKey = prereqCacheKeyType{}

type prereqCache map[string]PrerequisiteResult

var (
	ErrPrerequisitesMissing = errors.New("resource prerequisites missing")
	ErrSettingsNotFound     = errors.New("company settings not found")
	ErrInvalidConfiguration = errors.New("invalid resource configuration")
)

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}

	formatted := d.Format("2006-01-02")
	return []byte(`"` + formatted + `"`), nil
}
