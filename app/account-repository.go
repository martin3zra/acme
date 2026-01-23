package app

import (
	"database/sql"
	"log"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/mailer"
	"github.com/martin3zra/acme/pkg/routing"
)

type account struct {
	ID     int    `json:"id"`
	UUID   string `json:"uuid"`
	Status string `json:"status"`
	Owner  struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	} `json:"owner"`
	VerifiedAt *time.Time `json:"verified_at"`
	foundation.Timestamps
}

func (a account) HasVerifiedAccount() bool {
	return a.VerifiedAt != nil
}

func (a account) MarkAccountAsVerified(db *sql.DB) bool {
	err := database.WithTransaction(db, func(tx *sql.Tx) error {
		if _, err := tx.Exec("UPDATE accounts SET verified_at = $2, status = 'enabled'::entity_status WHERE id = $1", a.ID, time.Now()); err != nil {
			return err
		}

		if _, err := tx.Exec("UPDATE users SET email_verified_at = $2, status = 'enabled'::user_status FROM accounts WHERE users.id = accounts.owner_id AND accounts.id = $1", a.ID, time.Now()); err != nil {
			return err
		}
		return nil
	})

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
		To(a.GetEmailAddressForAccountVerification(), a.Owner.Email).
		Send(mail.NewVerification(foundation.AsMap(a), url))
}

func (a account) GetEmailAddressForAccountVerification() string {
	return a.Owner.Email
}

func (s *Server) findAccountByIDUsingTx(tx *sql.Tx, accountID int) (*account, error) {
	var a account

	if err := tx.QueryRow(`
    SELECT accounts.id, accounts.uuid, accounts.verified_at, accounts.status, accounts.created_at, accounts.updated_at, accounts.deleted_at,
    users.id, users.email
    FROM accounts
    INNER JOIN users ON (accounts.owner_id = users.id)
    WHERE accounts.id = $1
  `, accountID).Scan(
		&a.ID, &a.UUID, &a.VerifiedAt, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt, &a.Owner.ID, &a.Owner.Email,
	); err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *Server) findAccountByUUID(uuid string) (*account, error) {
	var a account

	if err := s.db.QueryRow(`
    SELECT accounts.id, accounts.uuid, accounts.verified_at, accounts.status, accounts.created_at, accounts.updated_at, accounts.deleted_at,
    users.id, users.email
    FROM accounts
    INNER JOIN users ON (accounts.owner_id = users.id)
    WHERE accounts.uuid = $1
  `, uuid).Scan(
		&a.ID, &a.UUID, &a.VerifiedAt, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt, &a.Owner.ID, &a.Owner.Email,
	); err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *Server) findAccountByOwnerEmailAddress(email string) (*account, error) {
	var a account

	if err := s.db.QueryRow(`
    SELECT accounts.id, accounts.uuid, accounts.verified_at, accounts.status, accounts.created_at, accounts.updated_at, accounts.deleted_at,
    users.id, users.email
    FROM accounts
    INNER JOIN users ON (accounts.owner_id = users.id)
    WHERE users.email = $1
  `, email).Scan(
		&a.ID, &a.UUID, &a.VerifiedAt, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt, &a.Owner.ID, &a.Owner.Email,
	); err != nil {
		return nil, err
	}

	return &a, nil
}
