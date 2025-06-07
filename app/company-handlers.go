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
	user := UserFromFoundationUser(auth.User(ctx.Request.Context()))
	err = s.storeCompany(user.Id, form)
	if err != nil {
		log.Printf("Error creating company: %v", err)
		s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.company"}))
		ctx.Back()
		return
	}

	company := user.currentCompany(s.db)
	err = s.sessionManager.ReGenerate(ctx.Request, user, map[string]any{
		"current_company": company,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.BackWith(map[string]string{"status": "success"})
}

func (s *Server) companyHandler(ctx *routing.Context) {

	companies, err := s.findCompanies()
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
