package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	frontendErrors "github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"reflect"
	"strconv"
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

	if unmarshalErr := json.Unmarshal(body, request); unmarshalErr != nil {
		return fmt.Errorf("%s: %w", frontendErrors.ErrInvalidJson, unmarshalErr)
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

// ValidatePOSTFormRequest binds only form fields from the request body to a struct and validates them.
func ValidatePOSTFormRequest(obj any, r *http.Request) error {
	// Ensure obj is a pointer
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	// Parse only body form data
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Prefer PostForm for body parameters only
	formData := r.PostForm

	// Use reflection to bind form data to struct fields
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the 'form' tag
		tag := fieldType.Tag.Get("form")
		if tag == "" {
			continue
		}

		// Get value from body form data using the tag
		formValue := formData.Get(tag)
		if !field.CanSet() {
			continue
		}

		// Set the value based on the field type
		switch field.Kind() {
		case reflect.String:
			field.SetString(formValue)
		case reflect.Int, reflect.Int64:
			if intValue, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				field.SetInt(intValue)
			}
		case reflect.Float64, reflect.Float32:
			if floatValue, err := strconv.ParseFloat(formValue, 64); err == nil {
				field.SetFloat(floatValue)
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(formValue); err == nil {
				field.SetBool(boolValue)
			}
		default:
			log.Errorf("unhandled default case %v", field.Kind())
		}
	}

	// Validate the struct
	err := validate.Struct(obj)
	if err != nil {
		// Return the first validation error
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return fmt.Errorf("validation error: %s", validationErrors[0].Error())
		}
		return err
	}

	return nil
}
