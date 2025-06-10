package support

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/session"
)

func ParseRequest(r *http.Request, params any) error {

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Println(err)
		return new(foundation.BadRequest)
	}

	formRequest, ok := params.(FormRequestContract)
	if ok {
		formRequest.SetContext(r.Context())

		if !formRequest.Authorize() {
			session.GetSession(r).Errors("status", "Unauthorized")
			return foundation.Unauthorized{}
		}

		formRequest.Validate(params, formRequest.Rules(), formRequest.PrepareForValidation)
		errorMesssages := formRequest.Errors()
		if len(errorMesssages) > 0 {
			session.GetSession(r).FormErrors(foundation.ErrorBag(errorMesssages))
			return errors.New(foundation.ToJSON(errorMesssages))
		}

		formRequest.PassedValidation()

		return nil
	}

	return new(foundation.UnprocessableEntity)
}
