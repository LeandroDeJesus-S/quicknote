package validation

import (
	"net/url"
)

type FormValidator interface {
	ValidateForm(form url.Values)                       // perform the validations and apply the error messages
	AddError(field, message string)                     // add a new error for a field
	AddValidator(field string, validator ValidatorFunc) // registers a new validator for a field
	Ok() bool                                           // returns true if there are no errors
	FieldErrors() map[string]string                     // returns the errors for each field
}

type formValidator struct {
	fieldErrors map[string]string
	validators  map[string]ValidatorFunc
}

func NewFormValidator() *formValidator {
	return &formValidator{
		fieldErrors: make(map[string]string),
		validators:  make(map[string]ValidatorFunc),
	}
}

func (fv *formValidator) ValidateForm(form url.Values) {
	for field, validatorFunc := range fv.validators {
		value := form.Get(field)
		if valid, errMsg := validatorFunc(value); !valid {
			fv.fieldErrors[field] = errMsg
		}
	}
}

func (fv *formValidator) AddError(field, message string) {
	fv.fieldErrors[field] = message
}

func (fv *formValidator) AddValidator(field string, validator ValidatorFunc) {
	fv.validators[field] = validator
}

func (fv *formValidator) Ok() bool {
	return len(fv.fieldErrors) == 0
}

func (fv *formValidator) FieldErrors() map[string]string {
	return fv.fieldErrors
}

// ValidatorFunc validates a form value and returns if it's valid and the error message
type ValidatorFunc func(v string) (bool, string)
