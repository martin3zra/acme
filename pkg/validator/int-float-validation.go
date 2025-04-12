package validator

func (va *ValidatesAttributes) validateIntRules(rule string, fieldValue, ruleValue int) bool {
	if rule == "max" {
		return fieldValue <= ruleValue
	}

	if rule == "min" {
		return fieldValue >= ruleValue
	}

	if rule == "gte" {
		return fieldValue >= ruleValue
	}

	if rule == "gt" {
		return fieldValue > ruleValue
	}

	if rule == "lte" {
		return fieldValue <= ruleValue
	}

	if rule == "lt" {
		return fieldValue < ruleValue
	}

	if rule == "different" {
		return fieldValue != ruleValue
	}

	return true
}

func (va *ValidatesAttributes) validateFloat64Rules(rule string, fieldValue, ruleValue float64) bool {
	if rule == "max" {
		return fieldValue <= ruleValue
	}

	if rule == "min" {
		return fieldValue >= ruleValue
	}

	if rule == "gte" {
		return fieldValue >= ruleValue
	}

	if rule == "gt" {
		return fieldValue > ruleValue
	}

	if rule == "lte" {
		return fieldValue <= ruleValue
	}

	if rule == "lt" {
		return fieldValue < ruleValue
	}

	if rule == "different" {
		return fieldValue != ruleValue
	}

	return true
}
