package validator

var defaultRules = []string{
	"required",
	"max",
	"min",
	"gt",
	"gte",
	"lte",
	"lt",
	"between",
	"different",
	"email",
	"exists",
	"unique",
	"current_password",
	"in",
	"uppercase",
	"lowercase",
}

var arrayRules = []string{
	"in",
}

var databaseRules = []string{
	"exists",
	"unique",
}

type Rule struct{}

type RuleConstraints interface {
	Constraints() string
}

type Errors map[string][]string

type RuleContract interface {
	Rules() map[string]any
}

type Validator struct {
	ValidatesAttributes
	errors Errors
}
