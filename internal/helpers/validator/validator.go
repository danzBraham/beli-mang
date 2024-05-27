package validator_helper

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func formatValidatorErrors(err error) string {
	var sb strings.Builder
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			sb.WriteString(fmt.Sprintf("Field '%s' failed on the '%s' tag;", fieldError.Field(), fieldError.Tag()))
		}
		return sb.String()
	}
	return err.Error()
}

func ValidatePayload(payload interface{}) error {
	if err := validate.Struct(payload); err != nil {
		return fmt.Errorf(formatValidatorErrors(err))
	}
	return nil
}
