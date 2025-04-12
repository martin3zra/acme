package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func (v *Validator) Validate(ctx context.Context, object any, rules map[string]any, beforeValidation ...func()) bool {

	for _, cb := range beforeValidation {
		cb()
	}
	// formRequest, ok := object.(support.FormRequestContract)
	// if ok {
	// 	// Trigger authorization
	// 	formRequest.PrepareForValidation()
	// }

	v.ctx = ctx
	rValue := reflect.ValueOf(object).Elem()
	rType := rValue.Type()

	if rType.Kind() == reflect.Struct {
		for i := range rType.NumField() {
			key := v.resolveKeyBasedOnJsonTag(rType, i)
			fieldRule, ok := rules[key]
			if !ok {
				continue
			}

			v.compileRuleSet(key, rValue.Field(i), v.resolveRuleComponents(fieldRule))
		}
	}

	return len(v.errors) == 0
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
		},
		"min": map[string]any{
			"int":    "The %s field must not be greater than %v.",
			"string": "The %s field must be at least %v characters.",
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
	rules = slices.DeleteFunc(rules, func(cmp string) bool {
		return cmp == "sometimes"
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

	if hasMultipleAttributes {
		if !v.validateBetween(int(value.Int()), attributes) {
			v.record(key, v.messages(key, ruleComponents[0], value.Kind().String(), attributes[0], attributes[1]))
		}

		return
	}

	v.evaluateSingleValueRule(key, ruleComponents[0], attributes[0], value)
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
	default:
		fmt.Println("data type not supported yet!")
	}

}

func (v *Validator) evaluateSingleRule(key, rule string, value reflect.Value) {
	if !v.validateRuleWithoutAttributes(rule, value) {
		v.record(key, v.messages(key, rule, value.Kind().String(), value))
	}
}

func (v *Validator) evaluateIntRules(key, rule string, fieldValue, ruleValue int) {
	if !v.validateIntRules(rule, fieldValue, ruleValue) {
		v.record(key, v.messages(key, rule, "int", ruleValue))
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
