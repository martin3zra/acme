package app

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/mailer"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/playsql"
)

type account struct {
	ID     int    `json:"id"`
	UUID   string `json:"uuid"`
	Status string `json:"status"`
	Owner  struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"owner"`
	VerifiedAt *time.Time `json:"verified_at"`
	foundation.Timestamps
}

func (a account) HasVerifiedAccount() bool {
	return a.VerifiedAt != nil
}

// MarkAccountAsVerified enables the account and its owner in one transaction.
//
// The owner update was an `UPDATE users ... FROM accounts WHERE users.id =
// accounts.owner_id AND accounts.id = $1` join. playsql has no UPDATE-FROM, but the
// join was only resolving the owner id — which the caller already holds, since every
// account read loads it. Both statements are guarded: each must touch exactly one
// row, and a miss rolls the transaction back rather than half-verifying.
//
// Returns false on any error, as it always did; the caller reports a generic failure.
func (a account) MarkAccountAsVerified(db *sql.DB) bool {
	err := database.WithTransaction(db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}
		now := time.Now()

		affected, err := ptx.Model(&accountRead{}).
			WhereEq("id", a.ID).
			Update(context.Background(), map[string]any{
				"verified_at": now,
				"status":      "enabled",
			})
		if err := mustAffectRows(affected, err, "account"); err != nil {
			return err
		}

		affected, err = ptx.Model(&userModel{}).
			WhereEq("id", a.Owner.ID).
			Update(context.Background(), map[string]any{
				"email_verified_at": now,
				"status":            "enabled",
			})
		return mustAffectRows(affected, err, "account owner")
	})

	if err != nil {
		log.Println("MarkAccountAsVerified:", err)
	}
	return err == nil
}

func (a account) SendAccountVerificationNotification(notify mailer.Mailer, attributes map[string]string) {
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
		To(a.GetEmailAddressForAccountVerification(), a.Owner.Name).
		Send(mail.NewVerification(foundation.AsMap(a), url))
}

func (a account) GetEmailAddressForAccountVerification() string {
	return a.Owner.Email
}

// The three reads below were the same query with a different WHERE. Each drops the
// `INNER JOIN users` for a belongsTo eager load; the join only ever asserted the
// owner exists (owner_id is a NOT NULL FK) and supplied three columns.
//
// findAccountByOwnerEmailAddress filters on a column of the related table, which
// WhereRelation expresses as an EXISTS subquery. An account has exactly one owner, so
// that selects the same rows the join did.

// accountQuery builds the shared projection: the account plus its owner.
func accountQuery(sess playSession) *playsql.Builder {
	return sess.Model(&accountRead{}).WithConstraint("Owner", withAccountOwner)
}

func (s *Server) findAccountByIDUsingTx(tx *sql.Tx, accountID int) (*account, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return nil, err
	}

	var row accountRead
	if err := accountQuery(ptx).
		WhereEq("id", accountID).
		First(context.Background(), &row); err != nil {
		return nil, err
	}
	return row.toAccount(), nil
}

func (s *Server) findAccountByUUID(uuid string) (*account, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row accountRead
	if err := accountQuery(pdb).
		WhereEq("uuid", uuid).
		First(context.Background(), &row); err != nil {
		return nil, err
	}
	return row.toAccount(), nil
}

func (s *Server) findAccountByOwnerEmailAddress(email string) (*account, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row accountRead
	if err := accountQuery(pdb).
		WhereRelation("Owner", "email", "=", email).
		First(context.Background(), &row); err != nil {
		return nil, err
	}
	return row.toAccount(), nil
}
