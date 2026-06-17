package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
)

// AuthUser is the application's user identity. It owns the users-table schema and
// satisfies foundation.Authenticatable / foundation.MustVerifyPassword so the
// framework (auth/routing/support) can stay schema-agnostic.
type AuthUser struct {
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
	foundation.Timestamps
	Role string `json:"role"`
}

func (u *AuthUser) IsEmpty() bool {
	return *u == (AuthUser{})
}

func (u *AuthUser) GetAuthIdentifier() int {
	return u.Id
}

func (u *AuthUser) GetAuthIdentifierName() string {
	return "email"
}

func (u *AuthUser) GetRole() string {
	return u.Role
}

func (u *AuthUser) SetRole(role string) {
	u.Role = role
}

func (u *AuthUser) GetAuthPassword() string {
	return u.Password
}

func (u *AuthUser) HasNotChangedPassword() bool {
	return u.MustChangePassword
}

func (u *AuthUser) MarkPasswordAsChanged(db *sql.DB, password string) error {
	_, err := db.Exec("UPDATE users SET password = $2, must_change_password = FALSE, last_password_reset = NOW() WHERE id = $1",
		u.Id, foundation.NewHashable().Make(password),
	)
	return err
}

// AuthUserFromContext rebuilds the authenticated identity from the session value
// stored in the request context.
func AuthUserFromContext(ctx context.Context) *AuthUser {
	var user AuthUser
	userCtx := ctx.Value(auth.ContextUserID{})
	if userCtx != nil {
		userJson, _ := json.Marshal(userCtx.(map[string]any))
		_ = json.Unmarshal(userJson, &user)
	}

	return &user
}

func init() {
	auth.SetCredentialResolver(func(db *sql.DB, column string, value any) (foundation.Authenticatable, error) {
		user := new(AuthUser)
		err := db.QueryRow(fmt.Sprintf("SELECT * FROM users WHERE %s = $1", column), value).
			Scan(&user.Id, &user.Name, &user.Email,
				&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
				&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
		if err != nil {
			return nil, err
		}
		return user, nil
	})

	auth.SetPasswordResolver(func(db *sql.DB, userID int) (string, error) {
		var password string
		err := db.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&password)
		if err != nil {
			return password, err
		}
		return password, nil
	})

	auth.SetUserDecoder(func(ctx context.Context) foundation.Authenticatable {
		return AuthUserFromContext(ctx)
	})
}
