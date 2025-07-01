package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
)

func (s *Server) storeCompanyHandler(ctx *routing.Context) {
	var form StoreCompanyForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	// TODO: An owner and admin can create companies.
	user := UserFromFoundationUser(auth.User(ctx.Request.Context()))
	account := user.Account(s.db)
	if account == nil {
		account, err = user.OwnedBy(s.db)
		if err != nil {
			ctx.BackWithError(err)
			return
		}
	}

	err = s.storeCompany(account.ID, user.Id, form)
	if err != nil {
		log.Printf("Error creating company: %v", err)
		ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.company"}))
		return
	}

	attrs := map[string]any{"current_company": nil}
	company, err := user.currentCompany(s.db)
	if err == nil {
		attrs["current_company"] = company
	}
	err = s.sessionManager.ReGenerate(ctx.Request, user, attrs)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.BackWithQuery(map[string]any{"status": "success"})
}

func (s *Server) companyHandler(ctx *routing.Context) {

	companies, err := s.findCompanies(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Render("Settings/Companies/Index", map[string]any{
		"translations": trans("companies"),
		"companies":    companies,
	})
}

func (s *Server) companyUpdateSequences() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *SequenceForm) {

		if err := s.updateSequences(ctx.Request.Context(), ctx.Param("id"), form); err != nil {
			ctx.Error(err)
			return
		}

		ctx.Flash("success", "Sequences updated successfully!")

		ctx.Redirect(fmt.Sprintf("/settings/%s/profile", ctx.Param("account")))
	})
}
