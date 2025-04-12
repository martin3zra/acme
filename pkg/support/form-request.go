package support

import (
	"context"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/validator"
)

type FormRequestContract interface {
	PrepareForValidation()
	Authorize() bool
	Validate(object any, rules map[string]any, prepareForValidation func()) validator.Validator
	Rules() map[string]any
	Ignore(ignore any, column ...string)
	Errors() validator.Errors
	User() *foundation.User
	SetContext(ctx context.Context)
}

type FormRequest struct {
	validator *validator.Validator
	ctx       context.Context
}

func (f *FormRequest) SetContext(ctx context.Context) {
	f.ctx = ctx
	f.getValidatorInstance()
}

func (f *FormRequest) User() *foundation.User {
	return auth.User(f.ctx)
}

func (f *FormRequest) setValidatorInstance(validator *validator.Validator) {
	f.validator = validator
}

func (f *FormRequest) getValidatorInstance() *validator.Validator {
	if f.validator != nil {
		return f.validator
	}

	validator := new(validator.Validator)
	f.setValidatorInstance(validator)
	return validator
}

func (f *FormRequest) Validate(object any, rules map[string]any, prepareForValidation func()) validator.Validator {

	if f.validator == nil {
		f.getValidatorInstance()
	}

	f.validator.Validate(f.ctx, object, rules, prepareForValidation)

	return *f.validator
}

func (f *FormRequest) Ignore(ignore any, column ...string) {
	f.validator.Ignore(ignore, column...)
}

func (f *FormRequest) Errors() validator.Errors {
	return f.validator.Errors()
}

func (f *FormRequest) Rules() map[string]any {
	return map[string]any{}
}

func (f *FormRequest) Authorize() bool       { return true }
func (f *FormRequest) PrepareForValidation() {}

func (f *FormRequest) PassesAuthorization() bool { return true }

func (f *FormRequest) FailedAuthorization() {
	// return here everything for a 403 status code
}
