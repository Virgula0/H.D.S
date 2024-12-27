package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

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

	if unmarshalErr := json.Unmarshal(body, request); unmarshalErr != nil {
		return fmt.Errorf("%s: %w", backendErrors.ErrInvalidJSON, unmarshalErr)
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

func ValidateGenericStruct(obj any) error {
	err := validate.Struct(obj)
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

// ValidateQueryParameters validates query parameters against a struct with 'query' tags.
func ValidateQueryParameters(obj any, r *http.Request) error {
	// Ensure obj is a pointer
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	// Get query parameters from URL
	queryParams := r.URL.Query()

	// Use reflection to bind query parameters to struct fields
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the 'query' tag
		tag := fieldType.Tag.Get("query")
		if tag == "" {
			continue
		}

		// Get value from query parameters using the tag
		queryValue := queryParams.Get(tag)
		if queryValue == "" {
			continue
		}

		// Ensure the field can be set
		if !field.CanSet() {
			continue
		}

		// Set the value based on the field type
		switch field.Kind() {
		case reflect.String:
			field.SetString(queryValue)
		case reflect.Int, reflect.Int64:
			if intValue, err := strconv.ParseInt(queryValue, 10, 64); err == nil {
				field.SetInt(intValue)
			} else {
				return fmt.Errorf("invalid integer value for '%s'", tag)
			}
		case reflect.Uint, reflect.Uint64:
			if uintValue, err := strconv.ParseUint(queryValue, 10, 64); err == nil {
				field.SetUint(uintValue)
			} else {
				return fmt.Errorf("invalid unsigned integer value for '%s'", tag)
			}
		case reflect.Float64, reflect.Float32:
			if floatValue, err := strconv.ParseFloat(queryValue, 64); err == nil {
				field.SetFloat(floatValue)
			} else {
				return fmt.Errorf("invalid float value for '%s'", tag)
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(queryValue); err == nil {
				field.SetBool(boolValue)
			} else {
				return fmt.Errorf("invalid boolean value for '%s'", tag)
			}
		default:
			// Logging unhandled types (optional)
			fmt.Printf("Unhandled field type: %v\n", field.Kind())
		}
	}

	// Validate the struct using validator
	validate := validator.New()
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
