package app

import (
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) storeCompanyHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form StoreCompanyForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}
		user := UserFromFoundationUser(auth.User(r.Context()))
		err = s.storeCompany(auth.ID(r.Context()), form)
		if err != nil {
			log.Printf("Error creating company: %v", err)
			s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.company"}))
			i.Back(w, r)
			return
		}

		company := user.currentCompany(s.db)
		err = s.sessionManager.ReGenerate(r, user, map[string]any{
			"current_company": company,
		})
		if err != nil {
			s.handleError(w, err)
			return
		}

		back(w, r, map[string]string{"status": "success"})
	}

	return http.HandlerFunc(fn)
}
