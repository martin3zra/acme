package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type UserLinkedCompany struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
	Role string `json:"role"`
}

// findUser resolves a single user by one column. playsql selects and scans by
// column name, so a migration that appends a column to users can no longer shift
// the scan — the failure mode that broke login in 53c9d58.
func (s *Server) findUser(column string, value any) (*User, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row userModel
	if err := pdb.Model(&userModel{}).WhereEq(column, value).First(context.Background(), &row); err != nil {
		return nil, err
	}
	return row.toUser(), nil
}

// findUsers lists the account's users with the number of companies each is linked
// to. The old EXISTS(accounts_users …) predicate becomes WhereRelation over the
// accounts_users pivot; the correlated COUNT that joined companies_users to
// companies becomes WithCount over the companies_users pivot, constrained to this
// account. Both express the pivot as belongsToMany rather than a hand-written join.
func (s *Server) findUsers(ctx context.Context) ([]*User, error) {
	accountID := CurrentAccount(ctx)

	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []userModel
	if err := pdb.Model(&userModel{}).
		// Same projection as before — notably no password hash.
		Select("id", "uuid", "name", "email", "email_verified_at", "status", "created_at", "updated_at").
		WithCount("Companies", playsql.As("linked"), playsql.Constrain(func(b *playsql.Builder) {
			b.WhereEq("account_id", accountID)
		})).
		WhereRelation("Accounts", "id", "=", accountID).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*User, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toUser())
	}
	return data, nil
}

func (s *Server) findUserByEmail(email string) (*User, error) {
	return s.findUser("email", email)
}

func (s *Server) findUserByUUID(uuid string) (*User, error) {
	return s.findUser("uuid", uuid)
}

// findUserLinkedCompanies reads the link rows directly and pulls each company
// through a belongsTo. The users join is dropped: it only asserted the user exists,
// which the user_id filter already implies.
func (s *Server) findUserLinkedCompanies(ctx context.Context, id int) ([]*UserLinkedCompany, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []companyUserModel
	if err := pdb.Model(&companyUserModel{}).
		With("Company").
		WhereRelation("Company", "account_id", "=", CurrentAccount(ctx)).
		WhereEq("user_id", id).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*UserLinkedCompany, 0, len(rows))
	for _, r := range rows {
		l := &UserLinkedCompany{Role: r.Role}
		if c := r.Company; c != nil {
			l.ID = c.ID
			l.UUID = c.UUID
		}
		data = append(data, l)
	}
	return data, nil
}

// Stays raw: a CASE assignment guarded by a scalar subquery in the WHERE clause.
// playsql's Update replaces a column with a value and has no subquery predicate.
func (s *Server) updateProfile(uuid string, form *StoreProfileForm) error {
	// The scalar subquery yields NULL for an unknown account uuid, so `id = NULL`
	// matches nothing and the write silently succeeded. Guard it.
	res, err := s.db.Exec(`
    UPDATE users
    SET name = $2, pending_email = CASE WHEN LOWER(email) <> LOWER($3) AND pending_email IS NULL THEN $3 ELSE pending_email END
    WHERE id = (SELECT owner_id FROM accounts WHERE uuid = $1)
  `, uuid, form.Name, form.Email)
	return mustAffectRow(res, err, "account owner")
}

// findUserByAccountUUID returns the account's owner. The old `id = (SELECT owner_id
// FROM accounts WHERE uuid = $1)` scalar subquery becomes two reads; playsql has no
// subquery predicate, and an unknown uuid still yields a not-found error either way.
func (s *Server) findUserByAccountUUID(uuid string) (*User, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var acc accountRead
	if err := pdb.Model(&accountRead{}).
		WhereEq("uuid", uuid).
		First(context.Background(), &acc); err != nil {
		return nil, err
	}

	return s.findUser("id", acc.OwnerID)
}

// Stays raw: `email = pending_email` is a column-to-column assignment, which
// playsql's Update cannot express (it binds values, not column references).
func (s *Server) updatePendingEmail(user *User) error {
	// Matches nothing when pending_email changed underneath us (or was cleared), in
	// which case the caller must not report the address as verified.
	res, err := s.db.Exec("UPDATE users SET email = pending_email, pending_email = NULL WHERE id = $1 AND pending_email = $2", user.Id, user.PendingEmail)
	return mustAffectRow(res, err, "pending email")
}

// attachUserCompanies links a user to each named company, marking the first one
// current. The old statement was an INSERT ... SELECT over companies that relied on
// zero rows affected to detect an unknown uuid; playsql cannot express that, so the
// company is resolved first and a missing one raises the same error.
func attachUserCompanies(ptx *playsql.Tx, accountID, userID int, companies []CompanyRole) error {
	for i, cr := range companies {
		var company companyModel
		err := ptx.Model(&companyModel{}).
			WhereEq("account_id", accountID).
			WhereEq("uuid", cr.Company).
			First(context.Background(), &company)
		if err == playsql.ErrNotFound {
			return fmt.Errorf("company with UUID %s not found", cr.Company)
		}
		if err != nil {
			return err
		}

		if err := ptx.Insert(context.Background(), &companyUserModel{
			CompanyID: company.ID,
			UserID:    userID,
			Role:      cr.Role,
			Current:   i == 0,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) storeUser(ctx context.Context, form *StoreUserForm) (*User, error) {
	accountID := CurrentAccount(ctx)
	user := new(User)

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// Map insert so uuid stays unset for the DB default; the merged userModel
		// maps uuid, which a struct insert would write as empty.
		userID, err := ptx.Model(&userModel{}).Insert(context.Background(), map[string]any{
			"name":     form.Name,
			"email":    form.Email,
			"password": foundation.NewHashable().Make("password"),
			"status":   "disabled",
		})
		if err != nil {
			return err
		}

		// Insert only returns the pk, so the DB-generated uuid and defaults are read
		// back. This replaces the old INSERT ... RETURNING * projection.
		var stored userModel
		if err := ptx.Model(&userModel{}).
			WhereEq("id", userID).
			First(context.Background(), &stored); err != nil {
			return err
		}
		*user = *stored.toUser()

		if err := ptx.Insert(context.Background(), &accountUserModel{
			AccountID: accountID,
			UserID:    user.Id,
		}); err != nil {
			return err
		}

		return attachUserCompanies(ptx, accountID, user.Id, form.Companies)
	})

	return user, err
}

func (s *Server) updateUser(ctx context.Context, form *StoreUserForm) error {
	accountID := CurrentAccount(ctx)

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// The old statement was an UPDATE ... RETURNING id, so an unknown uuid failed
		// with no rows; resolving the id first preserves that.
		var stored userModel
		if err := ptx.Model(&userModel{}).
			Select("id").
			WhereEq("uuid", form.Param("id")).
			First(context.Background(), &stored); err != nil {
			return err
		}

		if _, err := ptx.Model(&userModel{}).
			WhereEq("id", stored.ID).
			Update(context.Background(), map[string]any{"name": form.Name}); err != nil {
			return err
		}

		// companies_users has a deleted_at column but no softdelete tag on its write
		// model, so this is a hard DELETE, matching the statement it replaces.
		if _, err := ptx.Model(&companyUserModel{}).
			WhereEq("user_id", stored.ID).
			Delete(context.Background()); err != nil {
			return err
		}

		return attachUserCompanies(ptx, accountID, stored.ID, form.Companies)
	})
}

// setRememberToken writes users.remember_token. The column is deliberately unmapped
// on userModel — nothing else should write it — but mass-assignment is unrestricted,
// so it is passed as an explicit key. A nil value writes NULL.
func (s *Server) setRememberToken(id int, hashed any) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	_, err = pdb.Model(&userModel{}).
		WhereEq("id", id).
		Update(context.Background(), map[string]any{"remember_token": hashed})
	return err
}

// storeRememberToken persists the hash of a user's remember token.
func (s *Server) storeRememberToken(id int, hashed string) error {
	return s.setRememberToken(id, hashed)
}

// clearRememberToken removes a user's remember token (logout / rotation cleanup).
func (s *Server) clearRememberToken(id int) error {
	return s.setRememberToken(id, nil)
}

// findUserIDByRememberToken returns the user id whose stored hash matches, or 0.
func (s *Server) findUserIDByRememberToken(hashed string) (int, error) {
	pdb, err := s.play()
	if err != nil {
		return 0, err
	}

	var row userModel
	err = pdb.Model(&userModel{}).
		Select("id").
		WhereEq("remember_token", hashed).
		WhereEq("status", "enabled").
		First(context.Background(), &row)
	if err == playsql.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}
