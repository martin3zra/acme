package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) verifyAccountHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		renderWith := func(data map[string]any) {
			err := i.Render(w, r, "Verify/Index", data)
			if err != nil {
				s.handleError(w, err)
				return
			}
		}

		if !routing.VerifyRequest(r, string(s.config.secretKey)) {
			// s.handleError(w, errors.New("signature is not valid")) // 403
			renderWith(map[string]any{"status": "signatured-is-not-valid"})
			return
		}

		accountUUID := r.PathValue("uuid")
		hash := r.PathValue("hash")

		if !ensureUUIDIsValid(accountUUID) {
			// s.handleError(w, errors.New("account UUID not found"))
			renderWith(map[string]any{"status": "uuid-is-not-valid"})
			return
		}

		account, err := s.findAccountByUUID(accountUUID)
		if err != nil {
			// s.handleError(w, errors.New("account not found"))
			renderWith(map[string]any{"status": "not-found."})
			return
		}
		// create a new user instance that belongs to the account extending the foundation one
		// TODO: check the user is the owner by matching the owner ID with a user

		if account.HasVerifiedAccount() {
			// s.handleError(w, errors.New("account verification has been made"))
			renderWith(map[string]any{"status": "already-verified."})
			return
		}

		if !foundation.NewHashable().Sha1Equals(account.GetEmailAddressForAccountVerification(), hash) {
			// s.handleError(w, errors.New("hash do not match"))
			renderWith(map[string]any{"status": "hash-do-not-match"})
			return
		}

		if !account.MarkAccountAsVerified(s.db) {
			// force owner status to enabled.
			s.handleError(w, err)
			return
		}
		// trigger event
		// create any default setting for this account.

		renderWith(map[string]any{"status": "account-verified"})
	}

	return http.HandlerFunc(fn)
}
