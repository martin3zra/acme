package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type UserLinkedCompany struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
	Role string `json:"role"`
}

func (s *Server) findUsers(ctx context.Context) ([]*User, error) {
	rows, err := s.db.Query(`
    SELECT id, uuid, name, email, email_verified_at, status, created_at, updated_at,
    (
      SELECT COUNT(*)
	    FROM companies_users
	    INNER JOIN companies ON companies_users.company_id = companies.id AND companies.account_id = $1
	    WHERE users.id = companies_users.user_id
    ) as linked
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
		if err = rows.Scan(&u.Id, &u.UUID, &u.Name, &u.Email, &u.EmailVerifiedAt, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.Linked); err != nil {
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

func (s *Server) findUserLinkedCompanies(ctx context.Context, id int) ([]*UserLinkedCompany, error) {
	rows, err := s.db.Query(`
    SELECT companies.id, companies.uuid, companies_users.role
    FROM companies
    INNER JOIN companies_users ON companies.id = companies_users.company_id
    INNER JOIN users ON companies_users.user_id = users.id
    WHERE companies.account_id = $1 AND users.id = $2
  `, CurrentAccount(ctx), id)

	if err != nil {
		return nil, err
	}

	data := make([]*UserLinkedCompany, 0)
	for rows.Next() {
		l := new(UserLinkedCompany)
		if err = rows.Scan(
			&l.ID,
			&l.UUID,
			&l.Role,
		); err != nil {
			return nil, err
		}

		data = append(data, l)

	}

	return data, err
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

func (s *Server) storeUser(ctx context.Context, form *StoreUserForm) (*User, error) {
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

func (s *Server) updateUser(ctx context.Context, form *StoreUserForm) error {
	accountID := CurrentAccount(ctx)
	var userId = 0

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare("UPDATE users SET name = $2 WHERE uuid = $1 RETURNING id")
		if err != nil {
			return err
		}
		err = stmt.QueryRow(form.Param("id"), form.Name).Scan(&userId)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM companies_users WHERE user_id = $1", userId)
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
			res, err := stmt.Exec(userId, cr.Role, current, accountID, cr.Company)
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

	return err
}
