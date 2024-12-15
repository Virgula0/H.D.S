package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	backendErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/go-playground/validator/v10"
)

// Initialize the validator
var validate = validator.New()

// ValidateJSON Gets a struct with JSON annotations fields and validate it using validator
// http.Request is passed for checking if it is actually JSON
func ValidateJSON(request any, r *http.Request) error {

	// Check if request is a pointer.
	// The only way we have to do so is using reflection
	if reflect.TypeOf(request).Kind() != reflect.Ptr {
		panic("Request in validator is not a pointer")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, request); err != nil {
		return fmt.Errorf("%s: %w", backendErrors.ErrInvalidJSON, err)
	}

	// Validate the unmarshalled struct
	err = validate.Struct(request)
	if err != nil {
		// If validation fails, return the first validation error
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return validationErrors[0] // Return the first error
		}
		return err // Return the validation error
	}

	return nil
}
