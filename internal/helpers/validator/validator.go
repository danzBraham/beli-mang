package validator_helper

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
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

func InitCustomValidation() {
	validate.RegisterValidation("imageurl", validateImageURL)
	validate.RegisterValidation("onestartingpoint", validateOneStartingPoint)
}

func ValidatePayload(payload interface{}) error {
	if err := validate.Struct(payload); err != nil {
		return fmt.Errorf(formatValidatorErrors(err))
	}
	return nil
}

func validateImageURL(fl validator.FieldLevel) bool {
	u, err := url.ParseRequestURI(fl.Field().String())
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	ext := path.Ext(u.Path)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return false
	}
	return true
}

func validateOneStartingPoint(fl validator.FieldLevel) bool {
	orders, ok := fl.Field().Interface().([]purchase_entity.Order)
	if !ok {
		return false
	}

	count := 0
	for _, order := range orders {
		if order.IsStartingPoint {
			count++
		}
	}

	return count == 1
}
