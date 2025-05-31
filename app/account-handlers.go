package app

import (
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
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

func (s *Server) sendVerificationEmail(ctx *routing.Context) {
	var form EmailVerificationForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
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
