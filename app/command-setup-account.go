package app

import (
	"context"
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
		accountID, err := createAccountOwner(tx, name, email)
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

// createAccountOwner inserts the owner and the account they own, returning the
// account id. The owner starts disabled with a placeholder password; the
// verification email is what lets them set a real one.
//
// The statements it replaces had three defects. The accounts Prepare assigned to
// `stmt` and then called stmt.QueryRow *before* checking its error, so a failed
// Prepare dereferenced a nil pointer. That QueryRow's own Scan error was discarded
// entirely, so a failed insert left accountID at 0 and surfaced later as a confusing
// "account not found" from the read below. And neither statement was ever closed:
// both were prepared for a single execution, and the first was leaked outright when
// `stmt` was reassigned.
//
// The user insert's error check is belt-and-braces: dropping it leaves userID at 0,
// and the accounts insert then fails its owner_id foreign key, so the call still
// returns an error. It is kept for the message, not for the control flow — no test
// pins it, because none can.
func createAccountOwner(tx *sql.Tx, name, email string) (int, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return 0, err
	}

	userID, err := ptx.Model(&userModel{}).Insert(context.Background(), map[string]any{
		"name":     name,
		"email":    email,
		"password": foundation.NewHashable().Make("password"),
		"status":   "disabled",
	})
	if err != nil {
		return 0, err
	}

	accountID, err := ptx.Model(&accountRead{}).Insert(context.Background(), map[string]any{
		"owner_id": userID,
	})
	if err != nil {
		return 0, err
	}

	return int(accountID), nil
}
