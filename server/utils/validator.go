package utils

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

// NewValidator returns a validator instance with the project's custom rules registered.
func NewValidator() *validator.Validate {
	validate := validator.New()

	// "letter": field must contain at least one unicode letter,
	// so values that are only numeric or symbol+numeric are rejected.
	if err := validate.RegisterValidation("letter", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if unicode.IsLetter(r) {
				return true
			}
		}
		return false
	}); err != nil {
		panic(err)
	}

	// "printable": every rune must be printable, rejecting invisible or
	// non-renderable characters (control chars, BOM, zero-width, line separators).
	if err := validate.RegisterValidation("printable", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsPrint(r) {
				return false
			}
		}
		return true
	}); err != nil {
		panic(err)
	}

	return validate
}
