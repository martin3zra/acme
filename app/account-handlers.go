package app

import (
	"net/http"
	"net/url"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) verifyAccountHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		renderWith := func(data map[string]any) {
			data["translations"] = mergeTranslations(r.Context(), loadTranslations("verify"))
			err := i.Render(w, r, "Verify/Index", data)
			if err != nil {
				s.handleError(w, err)
				return
			}
		}

		if r.URL.Query().Has("status") {
			renderWith(map[string]any{"status": r.URL.Query().Get("status")})
			return
		}

		if !routing.VerifyRequest(r, string(s.config.secretKey)) {
			renderWith(map[string]any{"status": "signatured-is-not-valid"})
			return
		}

		accountUUID := r.PathValue("uuid")
		hash := r.PathValue("hash")

		if !ensureUUIDIsValid(accountUUID) {
			renderWith(map[string]any{"status": "uuid-is-not-valid"})
			return
		}

		account, err := s.findAccountByUUID(accountUUID)
		if err != nil {
			renderWith(map[string]any{"status": "not-found"})
			return
		}
		// create a new user instance that belongs to the account extending the foundation one
		// TODO: check the user is the owner by matching the owner ID with a user

		if account.HasVerifiedAccount() {
			renderWith(map[string]any{"status": "already-verified"})
			return
		}

		if !foundation.NewHashable().Sha1Equals(account.GetEmailAddressForAccountVerification(), hash) {
			renderWith(map[string]any{"status": "hash-do-not-match"})
			return
		}

		if !account.MarkAccountAsVerified(s.db) {
			s.handleError(w, err)
			return
		}
		// trigger event
		// create any default setting for this account.

		renderWith(map[string]any{"status": "account-verified"})
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

func back(w http.ResponseWriter, r *http.Request, attributes map[string]string) {
	// Get the referer (previous page URL)
	referer := r.Referer()
	if referer == "" {
		// Default fallback if referer is not present
		referer = "/"
	}

	// Parse the referer URL
	parsedURL, err := url.Parse(referer)
	if err != nil {
		http.Error(w, "Invalid referer", http.StatusBadRequest)
		return
	}

	// Add or update query parameters
	q := parsedURL.Query()
	for k, v := range attributes {
		q.Set(k, v)
	}
	parsedURL.RawQuery = q.Encode()
	http.Redirect(w, r, parsedURL.String(), http.StatusFound)
}
