package support

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

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
			return foundation.Unauthorized{}
		}

		// In case we want to ignore more than one key/value we can do it here.
		ignoreKey := resolveKeyToIgnore(formRequest.Rules())
		if ignoreKey != nil && r.Method == "PUT" {
			id, _ := strconv.Atoi(r.PathValue("id"))
			formRequest.Ignore(id)
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

func resolveKeyToIgnore(rules map[string]any) *string {
	for k, rule := range rules {
		parts := strings.Split(rule.(string), "|")
		for _, part := range parts {
			components := strings.Split(part, ":")
			hasIgnore := slices.Contains(components, "unique.ignore")
			if hasIgnore {
				return &k
			}
		}
	}
	return nil
}
