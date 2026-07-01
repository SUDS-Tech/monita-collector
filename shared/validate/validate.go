package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/bastion-framework/bast"
	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator and satisfies bast.Validator.
type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return &Validator{v: v}
}

// Validate satisfies bast.Validator. Called by ctx.Bind after JSON decode.
// Returns *bast.ValidationError so Bast renders field-level 422 responses automatically.
func (val *Validator) Validate(in any) error {
	err := val.v.Struct(in)
	if err == nil {
		return nil
	}
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return bast.ErrBadRequest("VALIDATION_FAILED", err.Error())
	}
	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[fe.Field()] = fieldMessage(fe)
	}
	return &bast.ValidationError{Fields: fields}
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "email":
		return "must be a valid email address"
	case "min":
		if fe.Type().Kind() == reflect.String || fe.Type().Kind() == reflect.Slice {
			return "must have at least " + fe.Param() + " character(s)"
		}
		return "must be at least " + fe.Param()
	case "max":
		if fe.Type().Kind() == reflect.String || fe.Type().Kind() == reflect.Slice {
			return "must have at most " + fe.Param() + " character(s)"
		}
		return "must be at most " + fe.Param()
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "invalid value"
	}
}