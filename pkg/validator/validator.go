package validator

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func ValidateStruct(s any) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var errors []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		errors = append(errors, FieldError{
			Field:   e.Field(),
			Tag:     e.Tag(),
			Value:   e.Param(),
			Message: msgForTag(e),
		})
	}
	return errors
}

type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param()
	case "max":
		return fe.Field() + " must be at most " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	case "email":
		return fe.Field() + " must be a valid email"
	default:
		return fe.Field() + " failed on " + fe.Tag()
	}
}

func DecodeAndValidate(r *http.Request, dst any) []FieldError {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return []FieldError{{Field: "body", Tag: "json", Message: "Invalid JSON: " + err.Error()}}
	}
	return ValidateStruct(dst)
}

// DecodeAndValidateGin binds JSON from the Gin context (used by HTTP handlers on Gin).
func DecodeAndValidateGin(c *gin.Context, dst any) []FieldError {
	if err := c.ShouldBindJSON(dst); err != nil {
		return []FieldError{{Field: "body", Tag: "json", Message: err.Error()}}
	}
	return ValidateStruct(dst)
}
