package support

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/validator"
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
		parts := resolveRuleParts(rule)
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

func resolveRuleParts(data any) []string {
	mixedData, ok := data.([]any)
	if ok {
		rules := make([]string, 0)
		for index := range mixedData {
			switch attributes := mixedData[index].(type) {
			case validator.ConditionalRules:
				rules = append(rules, strings.Split(attributes.Constraints(), "|")...)
			case validator.RuleConstraints:
				rules = append(rules, strings.Split(attributes.Constraints(), "|")...)
			case string:
				rules = append(rules, strings.Split(attributes, "|")...)
			default:
				fmt.Println("Field rules not supported!")
			}
		}

		return rules
	}

	ruleContractValue, ok := data.(validator.RuleConstraints)
	if ok {
		return strings.Split(ruleContractValue.Constraints(), "|")
	}

	return strings.Split(data.(string), "|")
}
