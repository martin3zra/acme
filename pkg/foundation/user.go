package foundation

import (
	"database/sql"
	"time"
)

type User struct {
	Id                 int        `json:"id"`
	UUID               string     `json:"uuid"`
	Status             string     `json:"status"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	PendingEmail       *string    `json:"pending_email"`
	Password           string     `json:"-"`
	EmailVerifiedAt    *time.Time `json:"email_verified_at"`
	LastPasswordReset  *time.Time `json:"last_password_reset"`
	MustChangePassword bool       `json:"must_change_password"`
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

func (u *User) HasNotChangedPassword() bool {
	return u.MustChangePassword
}

func (u *User) MarkPasswordAsChanged(db *sql.DB, password string) error {
	_, err := db.Exec("UPDATE users SET password = $2, must_change_password = FALSE, last_password_reset = NOW() WHERE id = $1",
		u.Id, NewHashable().Make(password),
	)
	return err
}

type Authenticatable interface {
	GetAuthIdentifier() int
	GetAuthIdentifierName() string
	GetAuthPassword() string
}

type MustVerifyPassword interface {
	HasNotChangedPassword() bool
	MarkPasswordAsChanged(db *sql.DB, password string) error
}
