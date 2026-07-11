package app

import (
	"context"
	"database/sql"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

type LoginFormRequest struct {
	support.FormRequest
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (f LoginFormRequest) Rules() map[string]any {
	return map[string]any{
		"email":    "required|email|max:100|lowercase",
		"password": "required",
		"remember": "sometimes|boolean",
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

type ConfirmsPasswords struct {
	support.FormRequest
	Password string `json:"current_password"`
}

func (ConfirmsPasswords) Rules() map[string]any {
	return map[string]any{
		"current_password": "required|current_password",
	}
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

	pdb, err := playOn(db)
	if err != nil {
		panic("database connection need to be set.")
	}

	var acc accountRead
	if err := pdb.Model(&accountRead{}).
		Select("owner_id").
		WhereEq("uuid", form.Param("account")).
		First(context.Background(), &acc); err != nil {
		panic("unable to find the account owner.")
	}
	userId := acc.OwnerID

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
