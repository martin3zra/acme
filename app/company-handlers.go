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
