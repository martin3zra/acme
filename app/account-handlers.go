package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
)

func (s *Server) verifyAccountHandler(ctx *routing.Context) {

	renderWithStatus := func(status string) {
		props := map[string]any{
			"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("verify")),
			"status":       status,
		}
		ctx.Render("Verify/Index", props)
	}

	if ctx.QueryValues().Has("status") {
		renderWithStatus(ctx.Query("status"))
		return
	}

	if !routing.VerifyRequest(ctx.Request, string(s.config.secretKey)) {
		renderWithStatus("signatured-is-not-valid")
		return
	}

	accountUUID := ctx.Param("uuid")
	hash := ctx.Param("hash")

	if !ensureUUIDIsValid(accountUUID) {
		renderWithStatus("uuid-is-not-valid")
		return
	}

	account, err := s.findAccountByUUID(accountUUID)
	if err != nil {
		renderWithStatus("not-found")
		return
	}

	if account.HasVerifiedAccount() {
		renderWithStatus("already-verified")
		return
	}

	if !foundation.NewHashable().Sha1Equals(account.GetEmailAddressForAccountVerification(), hash) {
		renderWithStatus("hash-do-not-match")
		return
	}

	if !account.MarkAccountAsVerified(s.db) {
		ctx.Error(err)
		return
	}

	user, err := auth.NewAuth(ctx.Request.Context()).LoginUsingId(account.Owner.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.sessionManager.ReGenerate(ctx.Request, user, map[string]any{})
	if err != nil {
		ctx.Error(err)
		return
	}

	must, ok := user.(foundation.MustVerifyPassword)
	if ok && must.HasNotChangedPassword() {
		renderWithStatus("create-password")
		return
	}

	renderWithStatus("account-verified")
}

func (s *Server) verifyEmailHandler(ctx *routing.Context) {

	renderWithStatus := func(status string) {
		props := map[string]any{
			"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("verify")),
			"status":       status,
		}
		ctx.Render("Verify/Index", props)
	}

	if ctx.QueryValues().Has("status") {
		renderWithStatus(ctx.Query("status"))
		return
	}

	if !routing.VerifyRequest(ctx.Request, string(s.config.secretKey)) {
		renderWithStatus("signatured-is-not-valid")
		return
	}

	userUUID := ctx.Param("uuid")
	hash := ctx.Param("hash")

	if !ensureUUIDIsValid(userUUID) {
		renderWithStatus("uuid-is-not-valid")
		return
	}

	user, err := s.findUserByUUID(userUUID)
	if err != nil {
		renderWithStatus("not-found")
		return
	}

	// TODO : Check if the user isOrphan => abort(403, 'Sorry, but this user does not belongs to any account in our platform.');
	// TODO : Add owner relationship to the user (select the account that owns that user.)
	// TODO : Add new middle to ensure account is verified

	if user.HasVerifiedEmail() {
		renderWithStatus("already-verified")
		return
	}

	if !foundation.NewHashable().Sha1Equals(user.Email, hash) {
		renderWithStatus("hash-do-not-match")
		return
	}

	if !user.MarkEmailAsVerified(s.db) {
		ctx.Error(err)
		return
	}

	loggedUser, err := auth.NewAuth(ctx.Request.Context()).LoginUsingId(user.Id)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.sessionManager.ReGenerate(ctx.Request, loggedUser, map[string]any{})
	if err != nil {
		ctx.Error(err)
		return
	}

	must, ok := loggedUser.(foundation.MustVerifyPassword)
	if ok && must.HasNotChangedPassword() {
		renderWithStatus("create-password")
		return
	}

	renderWithStatus("account-verified")
}

func (s *Server) verifyEmailPromptHandler(ctx *routing.Context) {

	renderWithStatus := func(status string) {
		props := map[string]any{
			"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("verify")),
			"status":       status,
			"email":        true,
		}
		ctx.Render("Verify/Index", props)
	}

	if ctx.QueryValues().Has("status") {
		renderWithStatus(ctx.Query("status"))
		return
	}

	renderWithStatus("resend-verification-email")
}

func (s *Server) sendVerificationEmail(ctx *routing.Context) {
	var form EmailVerificationForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	if ctx.Query("kind") == "email" {
		user, err := s.findUserByEmail(form.Email)
		if err != nil {
			ctx.Error(err)
			return
		}

		user.SendEmailVerification(s.mailer, map[string]string{
			"url":    fmt.Sprintf("%s/verify-email/%s/%s", s.config.host, user.UUID, foundation.NewHashable().Sha1(user.Email)),
			"secret": string(s.config.secretKey),
		})

		ctx.BackWith(map[string]string{"status": "verification-link-sent"})
		return
	}

	account, err := s.findAccountByOwnerEmailAddress(form.Email)
	if err != nil {
		ctx.Back()
		return
	}

	s.sendAccountVerificationNotification(*account)

	ctx.BackWith(map[string]string{"status": "verification-link-sent"})
}

func (s *Server) updateAccountProfileHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreProfileForm) {

		uuid := ctx.Param("account")
		if err := s.updateProfile(uuid, form); err != nil {
			s.session.Errors("name", "Something wrong happened")
			log.Println("errors", err.Error())
			ctx.Back()
			return
		}

		user, err := s.findUserByAccountUUID(uuid)
		if err != nil {
			s.session.Errors("name", "Something wrong happened")
			log.Println("errors", err.Error())
			ctx.Back()
			return
		}

		account := user.OwnedBy(s.db)
		company := user.currentCompany(s.db)
		// re-generate session
		err = s.sessionManager.ReGenerate(ctx.Request, user, map[string]any{
			"current_company": company,
			"account": map[string]any{
				"uuid":  account.UUID,
				"owner": user.Account(s.db) != nil,
			},
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.profile"}))

		ctx.Back()
	})
}
