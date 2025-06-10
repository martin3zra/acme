package app

import (
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
			log.Println("Unable to create company, we can't find an account to associated it.")
			ctx.Back()
			return
		}
	}

	err = s.storeCompany(account.ID, user.Id, form)
	if err != nil {
		log.Printf("Error creating company: %v", err)
		s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.company"}))
		ctx.Back()
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

	ctx.BackWith(map[string]any{"status": "success"})
}

func (s *Server) companyHandler(ctx *routing.Context) {

	companies, err := s.findCompanies(ctx.Request.Context())
	if err != nil {
		log.Println("something wrong occurred fetching companies:", err)
		ctx.Error(err)
		return
	}
	ctx.Render("Settings/Companies/Index", map[string]any{
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("companies")),
		"companies":    companies,
	})
}
