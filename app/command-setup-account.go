package app

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/martin3zra/forge/console"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
)

func (s *Server) SetupAccount() {

	name := console.Ask("What's the account name?")
	email := console.Ask("What's the owner email?")

	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {

		stmt, err := tx.Prepare("INSERT INTO users (name, email, password, status) VALUES($1,$2,$3,$4) RETURNING id")
		if err != nil {
			return err
		}

		var userID int
		err = stmt.QueryRow(name, email, foundation.NewHashable().Make("password"), "disabled").Scan(&userID)
		if err != nil {
			return err
		}

		stmt, err = tx.Prepare("INSERT INTO accounts (owner_id) VALUES($1) RETURNING id")
		var accountID int
		stmt.QueryRow(userID).Scan(&accountID)
		if err != nil {
			return err
		}

		account, err := s.findAccountByIDUsingTx(tx, accountID)
		if err != nil {
			log.Fatal(err)
		}

		s.sendAccountVerificationNotification(*account)

		return nil
	}); err != nil {
		console.Info("The new account wasn't created. Something wrong happened.", err)
		log.Fatal(err)
		return
	}

	console.Info("The new account was created successfully!")
}

func (s *Server) ResendAccountVerificationEmail() {

	email := console.Ask("What's the owner email?")
	account, err := s.findAccountByOwnerEmailAddress(email)
	if err != nil {
		console.Info("It wasn't possible to resend the email verification. Something wrong happened.", err)
		log.Fatal(err)
		return
	}

	// If the account is already verified, inform the operator and skip sending an email.
	if v, ok := any(*account).(MustVerifyAccount); ok && v.HasVerifiedAccount() {
		console.Info("The account is already verified; no verification email was sent.")
		return
	}

	s.sendAccountVerificationNotification(*account)

	console.Info(fmt.Sprintf("Verification email resent successfully to %s.", email))
}

func (s *Server) sendAccountVerificationNotification(acc account) {

	var account any = acc
	if v, ok := account.(MustVerifyAccount); ok {
		if v.HasVerifiedAccount() {
			return
		}

		v.SendAccountVerificationNotification(s.mailer, map[string]string{
			"url":    fmt.Sprintf("%s/verify-account/%s/%s", s.config.host, acc.UUID, foundation.NewHashable().Sha1(acc.Owner.Email)),
			"secret": string(s.config.secretKey),
		})
	}

}
