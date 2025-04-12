package foundation

import "time"

type User struct {
	Id                int        `json:"id"`
	CurrentCompanyId  *int       `json:"current_company_id"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	Email             string     `json:"email"`
	Password          string     `json:"-"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at"`
	LastPasswordReset *time.Time `json:"last_password_reset"`
	Timestamps
}

func (u *User) GetAuthIdentifier() int {
	return u.Id
}

func (u *User) GetAuthIdentifierName() string {
	return "email"
}

func (u *User) GetAuthPassword() string {
	return u.Password
}

type Authenticatable interface {
	GetAuthIdentifier() int
	GetAuthIdentifierName() string
	GetAuthPassword() string
}
