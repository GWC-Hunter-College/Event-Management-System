package validation_error

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ValidationErrorResponse struct {
	Errors []FieldError `json:"errors"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationError(errs validator.ValidationErrors) map[string]any {
	var verrs []FieldError

	for _, e := range errs {
		msg := ""
		switch e.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", e.Namespace())
		case "email":
			msg = fmt.Sprintf("%s must be a valid email address", e.Namespace())
		case "uuid":
			msg = fmt.Sprintf("%s must be a valid UUID", e.Namespace())
		case "url":
			msg = fmt.Sprintf("%s must be a valid URL", e.Namespace())
		case "oneof":
			msg = fmt.Sprintf("%s must be one of [%s]", e.Namespace(), e.Param())
		case "datetime":
			msg = fmt.Sprintf("%s must match datetime format %s", e.Namespace(), e.Param())
		case "timezone":
			msg = fmt.Sprintf("%s must be a valid IANA timezone string", e.Namespace())
		default:
			msg = fmt.Sprintf("%s is not valid (%s)", e.Namespace(), e.Tag())
		}

		verrs = append(verrs, FieldError{Field: e.Field(), Message: msg})
	}

	return map[string]any{"errors": verrs}
}
