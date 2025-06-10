package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

func (s *Server) findUsers(ctx context.Context) ([]*User, error) {
	rows, err := s.db.Query(`
    SELECT id, uuid, name, email, email_verified_at, status, created_at, updated_at
    FROM users
    WHERE EXISTS (
      SELECT 1 FROM accounts_users WHERE accounts_users.user_id = users.id AND accounts_users.account_id = $1
    )
  `, CurrentAccount(ctx))
	if err != nil {
		return nil, err
	}
	data := make([]*User, 0)
	for rows.Next() {
		u := new(User)
		if err = rows.Scan(&u.Id, &u.UUID, &u.Name, &u.Email, &u.EmailVerifiedAt, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		data = append(data, u)
	}
	return data, nil
}

func (s *Server) findUserByEmail(email string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where email = $1 ", email).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) findUserByUUID(uuid string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where uuid = $1 ", uuid).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) updateProfile(uuid string, form *StoreProfileForm) error {
	_, err := s.db.Exec(`
    UPDATE users
    SET name = $2, pending_email = CASE WHEN LOWER(email) <> LOWER($3) AND pending_email IS NULL THEN $3 ELSE pending_email END
    WHERE id = (SELECT owner_id FROM accounts WHERE uuid = $1)
  `, uuid, form.Name, form.Email)
	return err
}

func (s *Server) findUserByAccountUUID(uuid string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where id = (SELECT owner_id FROM accounts WHERE uuid = $1) ", uuid).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) updatePendingEmail(user *User) error {
	_, err := s.db.Exec("UPDATE users SET email = pending_email, pending_email = NULL WHERE id = $1 AND pending_email = $2", user.Id, user.PendingEmail)
	return err
}

func (s *Server) storeUser(ctx context.Context, form *StoreProfileForm) (*User, error) {
	accountID := CurrentAccount(ctx)
	var user User

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare("INSERT INTO users (name, email, password, status) VALUES($1,$2,$3,$4) RETURNING *")
		if err != nil {
			return err
		}
		err = stmt.QueryRow(form.Name, form.Email, foundation.NewHashable().Make("password"), "disabled").
			Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
				&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail,
			)
		if err != nil {
			return err
		}

		stmt, err = tx.Prepare("INSERT INTO accounts_users (account_id, user_id) VALUES($1, $2)")
		if err != nil {
			return err
		}

		_, err = stmt.Exec(accountID, user.Id)
		if err != nil {
			return err
		}

		stmt, err = tx.Prepare("INSERT INTO companies_users (company_id, user_id, role, current) SELECT id, $1, $2, $3 FROM companies WHERE account_id = $4 AND uuid = $5")
		if err != nil {
			return err
		}

		current := false
		for i, cr := range form.Companies {
			if i == 0 {
				current = true
			}
			res, err := stmt.Exec(user.Id, cr.Role, current, accountID, cr.Company)
			if err != nil {
				return err
			}
			current = false

			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return fmt.Errorf("company with UUID %s not found", cr.Company)
			}
		}

		return err
	})

	return &user, err
}
