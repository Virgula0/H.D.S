package utils

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strconv"
)

// Initialize the validator
var validate = validator.New()

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
