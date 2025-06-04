package app

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/console"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

func (s *Server) SetupAccount() {

	name := "Coffee Cups"          //console.Ask("What's the account name?")
	email := "gcmassiel@gmail.com" //console.Ask("What's the owner email?")

	if err := database.WithTransaction(s.db, func(tx *sql.Tx) error {

		stmt, err := tx.Prepare("INSERT INTO users (first_name, last_name, email, password, status) VALUES($1,$2,$3,$4,$5) RETURNING id")
		if err != nil {
			return err
		}

		var userID int
		err = stmt.QueryRow(name, "", email, foundation.NewHashable().Make("password"), "disabled").Scan(&userID)
		if err != nil {
			return err
		}

		stmt, err = tx.Prepare("INSERT INTO accounts (name, owner_id) VALUES($1,$2) RETURNING id")
		var accountID int
		stmt.QueryRow(name, userID).Scan(&accountID)
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
		console.Info("The new account wasn't created. Something wrong happened.")
		log.Fatal(err)
		return
	}

	console.Info("The new account was created successfully!")
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
