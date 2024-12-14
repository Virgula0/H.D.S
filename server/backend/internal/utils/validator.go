package utils

import (
	"github.com/go-playground/validator/v10"
)

// Initialize the validator
var validate = validator.New()

func Validate(request any) error {
	// Validate the struct
	err := validate.Struct(request)

	if err != nil {
		// Return only the first occurence of the error.
		// If many we ignore the rest. One error is sufficient in failing.
		return err.(validator.ValidationErrors)[0]
	}

	return nil
}
