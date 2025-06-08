package app

import (
	"fmt"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) storeUserHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreProfileForm) {

		account, err := UserFromFoundationUser(ctx.User()).OwnedBy(s.db)
		if err != nil {
			ctx.Back()
			return
		}

		user, err := s.storeUser(account.ID, form)
		if err != nil {
			s.session.Errors("error", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.user"}))
			ctx.BackWith(map[string]any{
				"open":    true,
				"subject": "user:form",
			})
			return
		}

		user.SendEmailVerification(s.mailer, map[string]string{
			"url":    fmt.Sprintf("%s/verify-email/%s/%s", s.config.host, user.UUID, foundation.NewHashable().Sha1(user.Email)),
			"secret": string(s.config.secretKey),
		})

		s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.user"}))

		ctx.Redirect(fmt.Sprintf("/settings/%s/profile", UserFromFoundationUser(ctx.User()).Account(s.db).UUID))
	})
}
