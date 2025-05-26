package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) verifyAccountHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		renderWithStatus := func(status string) {
			props := map[string]any{
				"translations": mergeTranslations(r.Context(), loadTranslations("verify")),
				"status":       status,
			}
			err := i.Render(w, r, "Verify/Index", props)
			if err != nil {
				s.handleError(w, err)
				return
			}
		}

		if r.URL.Query().Has("status") {
			renderWithStatus(r.URL.Query().Get("status"))
			return
		}

		if !routing.VerifyRequest(r, string(s.config.secretKey)) {
			renderWithStatus("signatured-is-not-valid")
			return
		}

		accountUUID := r.PathValue("uuid")
		hash := r.PathValue("hash")

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
			s.handleError(w, err)
			return
		}

		user, err := auth.NewAuth(r.Context()).LoginUsingId(account.Owner.ID)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = s.sessionManager.ReGenerate(r, user)
		if err != nil {
			s.handleError(w, err)
			return
		}

		must, ok := user.(foundation.MustVerifyPassword)
		if ok && must.HasNotChangedPassword() {
			renderWithStatus("create-password")
			return
		}

		renderWithStatus("account-verified")
	}

	return http.HandlerFunc(fn)
}

func (s *Server) sendVerificationEmail(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form EmailVerificationForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		account, err := s.findAccountByOwnerEmailAddress(form.Email)
		if err != nil {
			i.Back(w, r)
			return
		}

		s.sendAccountVerificationNotification(*account)

		back(w, r, map[string]string{"status": "verification-link-sent"})
	}
	return http.HandlerFunc(fn)
}
