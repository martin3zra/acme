package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (v *Validator) Validate(ctx context.Context, object any, rules map[string]any, beforeValidation ...func()) bool {

	for _, cb := range beforeValidation {
		cb()
	}

	v.ctx = ctx

	v.validateAttributes(object, rules)

	return len(v.errors) == 0
}

// Validate a given struct/object/attribute against a rule set
func (v *Validator) validateAttributes(object any, rules map[string]any) {
	val := reflect.ValueOf(object)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i := range val.NumField() {
		v.currentPosition = i
		key := v.resolveKeyBasedOnJsonTag(val.Type(), i)
		f := val.Field(i)
		switch f.Kind() {
		case reflect.Struct:
			if f.Type() == reflect.TypeOf(time.Time{}) {
				v.resetParentKey()
				fieldRule, ok := rules[key]
				if !ok {
					continue
				}

				v.compileRuleSet(key, val.Field(i), v.resolveRuleComponents(fieldRule))
				continue
			}

			v.setParentKey(key)
			v.setKeySeparator(".")
			v.validateAttributes(f.Interface(), rules)
		case reflect.Slice:
			// If the Slice is empty compile any specified rule.
			if f.Len() == 0 {
				v.resetParentKey()
				fieldRule, ok := rules[key]
				if !ok {
					continue
				}
				v.compileRuleSet(key, val.Field(i), v.resolveRuleComponents(fieldRule))
				continue
			}

			v.setParentKey(key)
			v.setKeySeparator(".*.")
			for j := range f.Len() {
				v.validateAttributes(f.Index(j).Interface(), rules)
			}
			v.resetParentKey()
		default:
			ruleIdx := key
			if v.hasParentKey() {
				ruleIdx = fmt.Sprintf("%s%s%s", v.parentKey, v.keySeparator, key)
			}
			fieldRule, ok := rules[ruleIdx]
			if !ok {
				continue
			}
			v.compileRuleSet(key, val.Field(i), v.resolveRuleComponents(fieldRule))
		}
	}
}

func (v *Validator) Errors() Errors {
	return v.errors
}

func (v *Validator) messages(attribute, rule, kind string, value ...any) string {
	messages := map[string]any{
		"required": "The %s field is required.",
		"max": map[string]any{
			"int":    "The %s field must not be greater than %v.",
			"string": "The %s field must not be greater than %v characters.",
			"slice":  "The %s field must not have more than %v items.",
		},
		"min": map[string]any{
			"int":    "The %s field must not be greater than %v.",
			"string": "The %s field must be at least %v characters.",
			"slice":  "The %s field must has at least %v items.",
		},
		"gte": map[string]any{
			"int":    "The %s field must be greater than or equal to %v.",
			"string": "The %s field must be greater than or equal to %v characters.",
		},
		"gt": map[string]any{
			"int":    "The %s field must be greater than %v.",
			"string": "The %s field must be greater than %v characters.",
		},
		"lte": map[string]any{
			"int":    "The %s field must be less than or equal to %v.",
			"string": "The %s field must be less than or equal to %v characters.",
		},
		"lt": map[string]any{
			"int":    "The %s field must be less than or equal to %v.",
			"string": "The %s field must be less than %v characters.",
		},
		"between": map[string]any{
			"int":    "The %s field must be between %v and %v.",
			"string": "The %s field must be between %v and %v characters.",
		},
		"different":        "The %s field and %v must be different.",
		"email":            "The %s field must be a valid email address.",
		"exists":           "The selected %s is invalid.",
		"unique":           "The %s has already been taken.",
		"current_password": "The password is incorrect.",
		"in":               "The selected %s is invalid.",
		"lowercase":        "The %s field must be lowercase.",
		"uppercase":        "The %s field must be uppercase.",
		"date":             "The %s field must be a valid date.",
		"after":            "The %s field must be a date after %v.",
		"digits":           "The %s field must be %v digits.",
		"digits_between":   "The %s field must be between %v and %v digits.",
		"max_digits":       "The %s field must not have more than %v digits.",
		"min_digits":       "The %s field must have at least %v digits.",
	}

	message, ok := messages[rule]
	if ok {
		switch message := message.(type) {
		case map[string]any:
			return v.composeMessage(message[kind].(string), attribute, value...)
		case string:
			return v.composeMessage(message, attribute, value...)
		default:
			return fmt.Sprintf("The %s fail the %s rule.", attribute, rule)
		}
	}

	return fmt.Sprintf("The %s fail the %s rule.", attribute, rule)
}

func (v *Validator) composeMessage(message, attribute string, value ...any) string {
	if v.hasParentKey() {
		if v.keySeparator == "." {
			attribute = fmt.Sprintf("%s.%s", v.parentKey, attribute)
		} else {
			attribute = fmt.Sprintf("%s %d %s", v.parentKey, v.currentPosition+1, attribute)
		}
	}

	re := regexp.MustCompile("%v")
	matches := re.FindAllStringIndex(message, -1)
	if len(matches) >= 1 {
		// join all values to prepare the message
		var args = []any{attribute}
		args = append(args, value...)

		return fmt.Sprintf(message, args...)
	}

	if strings.Contains(message, "%s") {
		return fmt.Sprintf(message, attribute)
	}

	return message
}

func (v *Validator) record(key, message string) {
	if v.errors == nil {
		v.errors = make(map[string][]string)
	}

	v.shouldStopOnFirstFailure(v.canBail)

	// If we're validating a nested object (Array|Slice) we'll pre-append the parent key
	// to the error message, and add the human position to the message, so it's more
	// clear to the user to understand the error.
	if v.hasParentKey() {
		nestedKey := fmt.Sprintf("%s.%d.%s", v.parentKey, v.currentPosition+1, key)
		if v.keySeparator == "." {
			nestedKey = fmt.Sprintf("%s.%s", v.parentKey, key)
		}
		v.errors[nestedKey] = append(v.errors[key], message)
		return
	}

	v.errors[key] = append(v.errors[key], message)
}

func (v *Validator) ensureRuleExists(rule string) bool {
	return slices.Contains(defaultRules, rule)
}

func (v *Validator) resolveKeyBasedOnJsonTag(field reflect.Type, index int) string {
	return strings.Split(field.Field(index).Tag.Get("json"), ",")[0]
}

func (v *Validator) resolveRuleComponents(data any) []string {
	ruleContractValue, ok := data.(RuleConstraints)
	if ok {
		return strings.Split(ruleContractValue.Constraints(), "|")
	}
	return strings.Split(data.(string), "|")
}

func (v *Validator) compileRuleSet(key string, value reflect.Value, rules []string) {
	v.sometimes = slices.Contains(rules, "sometimes")
	v.canBail = slices.Contains(rules, "bail")
	rules = slices.DeleteFunc(rules, func(cmp string) bool {
		return cmp == "sometimes" || cmp == "bail"
	})

	for _, rule := range rules {
		v.needsToIgnore = false
		ruleComponents := strings.Split(rule, ":")

		rule := ruleComponents[0]
		prepends := strings.Split(rule, ".")

		if len(prepends) > 1 {
			v.needsToIgnore = prepends[1] == "ignore"
			rule = prepends[0]
			ruleComponents[0] = rule
		}

		if v.stopOnFirstFailure {
			v.shouldStopOnFirstFailure(false)
			break
		}

		if v.ensureRuleExists(rule) {
			if v.sometimes && value.IsZero() {
				break
			}

			if len(ruleComponents) == 2 {
				v.evaluateRuleWithValues(key, ruleComponents, value)
				continue
			}

			v.evaluateSingleRule(key, ruleComponents[0], value)
		}
	}
}

func (v *Validator) evaluateRuleWithValues(key string, ruleComponents []string, value reflect.Value) {
	hasMultipleAttributes, attributes := v.hasMultipleAttributes(ruleComponents[1])
	if slices.Contains(databaseRules, ruleComponents[0]) {
		//set database handler for the validation.
		if !v.validateDatabaseRules(key, ruleComponents[0], attributes, value) {
			v.record(key, v.messages(key, ruleComponents[0], value.Kind().String(), attributes[0]))
		}
		return
	}

	if slices.Contains(arrayRules, ruleComponents[0]) {
		if !v.validateArrayRules(ruleComponents[0], attributes, value) {
			v.record(key, v.messages(key, ruleComponents[0], value.String(), attributes))
		}
		return
	}

	if slices.Contains(dateRules, ruleComponents[0]) {
		if !v.evaluateDateRule(ruleComponents[0], ruleComponents[1], value.Interface().(time.Time)) {
			v.record(key, v.messages(key, ruleComponents[0], value.String(), attributes))
		}
		return
	}

	if hasMultipleAttributes {
		if ruleComponents[0] == "digits_between" {
			if !v.validateBetween(digits(int(value.Int())), attributes) {
				v.record(key, v.messages(key, ruleComponents[0], value.Kind().String(), attributes[0], attributes[1]))
			}
			return
		}
		if !v.validateBetween(int(value.Int()), attributes) {
			v.record(key, v.messages(key, ruleComponents[0], value.Kind().String(), attributes[0], attributes[1]))
		}

		return
	}

	v.evaluateSingleValueRule(key, ruleComponents[0], attributes[0], value)
}

func (v *Validator) evaluateDateRule(rule string, ruleValue string, value time.Time) bool {
	if rule == "after" {
		if ruleValue == "yesterday" {
			return value.After(time.Now().AddDate(0, 0, -1))
		}
	}

	return true
}

func (v *Validator) evaluateSingleValueRule(key, rule string, ruleValue any, value reflect.Value) {
	castedRuleValue, _ := strconv.Atoi(ruleValue.(string))
	switch value.Kind() {
	case reflect.Int:
		v.evaluateIntRules(key, rule, int(value.Int()), castedRuleValue)
	case reflect.String:
		v.evaluateStringRules(key, rule, value.String(), castedRuleValue)
	case reflect.Float64:
		castedRuleValue, _ := strconv.ParseFloat(ruleValue.(string), 64)
		v.evaluateFloat64Rules(key, rule, value.Float(), castedRuleValue)
	case reflect.Slice:
		v.evaluateSliceRules(key, rule, value, castedRuleValue)
	default:
		fmt.Println("data type not supported yet!", value.Type())
	}

}

func (v *Validator) evaluateSingleRule(key, rule string, value reflect.Value) {
	if !v.validateRuleWithoutAttributes(rule, value) {
		v.record(key, v.messages(key, rule, value.Kind().String(), value))
	}
}

func (v *Validator) evaluateIntRules(key, rule string, fieldValue, ruleValue int) {
	if rule == "max_digits" || rule == "min_digits" {
		if !v.validateIntRules(rule, digits(fieldValue), ruleValue) {
			v.record(key, v.messages(key, rule, "int", ruleValue))
		}
		return
	}

	if !v.validateIntRules(rule, fieldValue, ruleValue) {
		v.record(key, v.messages(key, rule, "int", ruleValue))
	}
}

func (v *Validator) evaluateSliceRules(key, rule string, fieldValue reflect.Value, ruleValue int) {
	if !v.validateSliceRules(rule, fieldValue, ruleValue) {
		v.record(key, v.messages(key, rule, "slice", ruleValue))
	}
}

func (v *Validator) evaluateFloat64Rules(key, rule string, fieldValue, ruleValue float64) {
	if !v.validateFloat64Rules(rule, fieldValue, ruleValue) {
		v.record(key, v.messages(key, rule, "int", ruleValue))
	}
}

func (v *Validator) evaluateStringRules(key, rule string, fieldValue string, ruleValue int) {
	if !v.validateIntRules(rule, len(fieldValue), ruleValue) {
		v.record(key, v.messages(key, rule, "string", ruleValue))
	}
}
