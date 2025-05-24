package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type ContextUserID struct{}

type Auth struct {
	db       *sql.DB
	Hashable foundation.Hash
}

func NewAuth(ctx context.Context) *Auth {
	db := ctx.Value(database.ConnectionKey{}).(*sql.DB)

	if db == nil {
		panic("database connection need to be set.")
	}
	return &Auth{
		db:       db,
		Hashable: foundation.NewHashable(),
	}
}

func (a *Auth) Authenticate(username, password string) (foundation.Authenticatable, error) {
	user, err := a.attempt("email", username)
	if err != nil {
		log.Printf("error authenticating user: %s\n", err)
		return nil, err
	}

	if !a.EnsureIsCurrentPassword(user.GetAuthPassword(), password) {
		log.Printf("error invalid password")
		return nil, errors.New("error invalid password")
	}

	return user, nil
}

func (a *Auth) LoginUsingId(id int) (foundation.Authenticatable, error) {
	user, err := a.attempt("id", id)
	if err != nil {
		log.Printf("error authenticating user: %s\n", err)
		return nil, err
	}

	return user, nil
}

func (a *Auth) EnsureIsCurrentPassword(hashed, password string) bool {
	return a.Hashable.Check(password, hashed)
}

func (a *Auth) attempt(column, value any) (foundation.Authenticatable, error) {
	user := new(foundation.User)
	err := a.db.QueryRow(fmt.Sprintf("SELECT * FROM users WHERE %s = $1", column), value).
		Scan(&user.Id, &user.CurrentCompanyId, &user.FirstName, &user.LastName, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *Auth) GetCurrentPassword(userId int) (string, error) {
	var password string
	err := a.db.QueryRow("SELECT password FROM users WHERE id = $1", userId).
		Scan(&password)
	if err != nil {
		return password, err
	}
	return password, nil
}

// Retrieve the currently authenticated user...
func User(ctx context.Context) *foundation.User {
	var user foundation.User
	userCtx := ctx.Value(ContextUserID{}).(map[string]any)
	if userCtx != nil {
		userJson, _ := json.Marshal(userCtx)

		json.Unmarshal([]byte(userJson), &user)
	}

	return &user
}

// Retrieve the currently authenticated user's ID...
func ID(ctx context.Context) int {
	var user foundation.User
	userCtx := ctx.Value(ContextUserID{}).(map[string]any)
	if userCtx != nil {
		userJson, _ := json.Marshal(userCtx)

		json.Unmarshal([]byte(userJson), &user)
	}

	return user.Id
}
