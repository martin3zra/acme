package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/support"
)

func (s *Server) verifyAccountHandler(ctx *routing.Context) {

	renderWithStatus := func(status string) {
		props := map[string]any{
			"translations": trans("verify"),
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

	attrs := map[string]any{"current_company": nil, "account": nil}
	attrs["account"] = map[string]any{
		"id":    account.ID,
		"uuid":  account.UUID,
		"owner": true,
	}

	err = s.sessionManager.ReGenerate(ctx.Request, user, attrs)
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
			"translations": trans("verify"),
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

	user, err := s.findUserByUUID(userUUID)
	if err != nil {
		renderWithStatus("not-found")
		return
	}

	// TODO : Check if the user isOrphan => abort(403, 'Sorry, but this user does not belongs to any account in our platform.');
	// TODO : Add owner relationship to the user (select the account that owns that user.)
	// TODO : Add new middle to ensure account is verified

	if ctx.Query("action") == "change" {
		if !foundation.NewHashable().Sha1Equals(fmt.Sprintf("%schange", *user.PendingEmail), hash) {
			renderWithStatus("hash-do-not-match")
			return
		}

		if err := s.updatePendingEmail(user); err != nil {
			log.Println("Something wrong happened updating the pending email: ", err)
			ctx.Error(err)
			return
		}

		user, err = s.findUserByUUID(userUUID)
		if err != nil {
			renderWithStatus("not-found")
			return
		}
	} else {

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
	}

	loggedUser, err := auth.NewAuth(ctx.Request.Context()).LoginUsingId(user.Id)
	if err != nil {
		ctx.Error(err)
		return
	}
	attrs := map[string]any{"current_company": nil, "account": nil}
	account, err := user.OwnedBy(s.db)
	if err == nil {
		attrs["account"] = map[string]any{
			"id":    account.ID,
			"uuid":  account.UUID,
			"owner": user.Account(s.db) != nil,
		}
	}
	company, err := user.currentCompany(s.db)
	if err == nil {
		attrs["current_company"] = company
	}

	err = s.sessionManager.ReGenerate(ctx.Request, loggedUser, attrs)
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
			"translations": trans("verify"),
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

		ctx.BackWithQuery(map[string]any{"status": "verification-link-sent"})
		return
	}

	account, err := s.findAccountByOwnerEmailAddress(form.Email)
	if err != nil {
		ctx.Back()
		return
	}

	s.sendAccountVerificationNotification(*account)

	ctx.BackWithQuery(map[string]any{"status": "verification-link-sent"})
}

func (s *Server) accountProfileHandler(ctx *routing.Context) {
	userUuid := ctx.Query("user_id")
	companyUuid := ctx.Query("company_id")
	companies, err := s.findCompanies(ctx.Request.Context())
	if err != nil {
		log.Println("something wrong occurred fetching companies:", err)
		ctx.Error(err)
		return
	}
	users, err := s.findUsers(ctx.Request.Context())
	if err != nil {
		log.Println("something wrong occurred fetching users:", err)
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"translations": trans("companies", "users", "profile"),
		"companies":    companies,
		"users":        users,
		"roles":        RoleMap,
	}

	if userUuid != "" {
		user, err := s.findUserByUUID(userUuid)
		if err != nil {
			ctx.Error(err)
			return
		}

		linked, err := s.findUserLinkedCompanies(ctx.Request.Context(), user.Id)
		if err != nil {
			ctx.Error(err)
			return
		}

		linkedMap := make([]map[string]any, 0)
		for _, link := range linked {
			linkedMap = append(linkedMap, map[string]any{
				"uuid": link.UUID,
				"role": link.Role,
			})
		}
		props["currentUser"] = map[string]any{
			"uuid":              user.UUID,
			"name":              user.Name,
			"email":             user.Email,
			"status":            user.Status,
			"email_verified_at": user.EmailVerifiedAt,
			"linkedCompanies":   linkedMap,
		}
		props["initialState"] = true
		props["subject"] = "user:view"
	}

	if companyUuid != "" {
		company, err := s.findCompanyByUUID(ctx.Request.Context(), companyUuid)
		if err != nil {
			ctx.Error(err)
			return
		}
		sequences, err := s.findSequences(ctx.Request.Context(), companyUuid)
		if err != nil {
			ctx.Error(err)
			return
		}

		taxes, err := s.findTaxes(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}

		preferences, err := s.findRedirectPreferences(ctx.Request.Context(), CurrentCompany(ctx.Request.Context()).UUID)
		if err != nil {
			log.Printf("Error fetching redirect preferences: %v", err)
			ctx.Back()
			return
		}

		taxReceipts, err := s.findTaxReceiptsForSetup(ctx.Request.Context())
		if err != nil {
			log.Printf("Error fetching tax receipts for setup: %v", err)
			ctx.Back()
			return
		}

		categories, err := s.findExpensesCategories(ctx.Request.Context())
		if err != nil {
			log.Printf("Error fetching tax receipts for setup: %v", err)
			ctx.Back()
			return
		}

		units, err := s.findUnits(ctx.Request.Context())
		if err != nil {
			log.Printf("Error fetching units: %v", err)
			ctx.Back()
			return
		}

		company.Sequences = &sequences.Sequence
		company.SeqLastUpdatedAt = &sequences.UpdatedAt
		company.Taxes = taxes
		company.RedirectPreferences = preferences.Redirect
		company.TaxReceipts = taxReceipts
		company.ExpenseCategories = categories
		company.Units = units
		props["company"] = company

		props["initialState"] = true
		props["subject"] = "company:view"
	}

	if ctx.QueryHas("open") {
		props["initialState"] = ctx.Query("open")
	}

	if ctx.QueryHas("subject") {
		props["subject"] = ctx.Query("subject")
	}

	ctx.Render("Settings/Account", props)
}

func (s *Server) updateAccountProfileHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreProfileForm) {

		uuid := ctx.Param("account")
		user, err := s.findUserByAccountUUID(uuid)
		if err != nil {
			s.session.Errors("name", s.trans("global.somethingWentWrong"))
			log.Println("errors", err.Error())
			ctx.Back()
			return
		}

		if err := s.updateProfile(uuid, form); err != nil {
			s.session.Errors("name", s.trans("global.somethingWentWrong"))
			log.Println("errors", err.Error())
			ctx.Back()
			return
		}

		if !strings.EqualFold(user.Email, form.Email) {
			if user.PendingEmail != nil {
				s.session.Errors("email", s.trans("global.pendingVerification", i18n.Replacements{"email": *user.PendingEmail}))
				ctx.Back()
				return
			}

			user.PendingEmail = &form.Email

			user.SendEmailVerificationChange(s.mailer, map[string]string{
				"url": fmt.Sprintf(
					"%s/verify-email/%s/%s?action=change",
					s.config.host,
					user.UUID,
					foundation.NewHashable().Sha1(fmt.Sprintf("%s%s", form.Email, "change")),
				),
				"secret": string(s.config.secretKey),
			})
		}

		attrs := map[string]any{"current_company": nil, "account": nil}
		account, err := user.OwnedBy(s.db)
		if err == nil {
			attrs["account"] = map[string]any{
				"id":    account.ID,
				"uuid":  account.UUID,
				"owner": user.Account(s.db) != nil,
			}
		}
		company, err := user.currentCompany(s.db)
		if err == nil {
			attrs["current_company"] = company
		}
		// re-generate session
		err = s.sessionManager.ReGenerate(ctx.Request, user, attrs)
		if err != nil {
			ctx.Error(err)
			return
		}

		s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.profile"}))

		ctx.Back()
	})
}
