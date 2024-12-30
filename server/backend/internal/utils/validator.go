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

// Validator instance
var validate = validator.New()

// ValidateJSON parses and validates JSON data from the request body into the given struct.
func ValidateJSON(request any, r *http.Request) error {
	if err := ensurePointer(request); err != nil {
		panic(err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, request); err != nil {
		return fmt.Errorf("%s: %w", backendErrors.ErrInvalidJSON, err)
	}

	return validateStruct(request)
}

// ValidateGenericStruct validates a struct using the validator.
func ValidateGenericStruct(obj any) error {
	return validateStruct(obj)
}

// ValidateQueryParameters binds and validates query parameters against a struct with 'query' tags.
func ValidateQueryParameters(obj any, r *http.Request) error {
	if err := ensurePointer(obj); err != nil {
		return err
	}

	return bindAndValidate(obj, r.URL.Query(), "query")
}

// ensurePointer verifies that the object is a pointer.
func ensurePointer(obj any) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}
	return nil
}

// validateStruct validates the given struct and returns the first validation error if any.
func validateStruct(obj any) error {
	if err := validate.Struct(obj); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return fmt.Errorf("validation error: %s", validationErrors[0].Error())
		}
		return err
	}
	return nil
}

// bindAndValidate binds data from a map to a struct using reflection and validates it.
func bindAndValidate(obj any, data map[string][]string, tagKey string) error {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the tag
		tag := fieldType.Tag.Get(tagKey)
		if tag == "" || !field.CanSet() {
			continue
		}

		// Get the value from data
		values, exists := data[tag]
		if !exists || len(values) == 0 {
			continue
		}
		value := values[0]

		// Set the field value
		if err := setFieldValue(field, value, tag); err != nil {
			return err
		}
	}

	return validateStruct(obj)
}

// setFieldValue sets the value of a struct field based on its type.
func setFieldValue(field reflect.Value, value, tag string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		return parseAndSetInt(field, value, tag)
	case reflect.Uint, reflect.Uint64:
		return parseAndSetUint(field, value, tag)
	case reflect.Float32, reflect.Float64:
		return parseAndSetFloat(field, value, tag)
	case reflect.Bool:
		return parseAndSetBool(field, value, tag)
	default:
		fmt.Printf("Unhandled field type: %v\n", field.Kind())
	}
	return nil
}

// parseAndSetInt parses and sets an integer field.
func parseAndSetInt(field reflect.Value, value, tag string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid integer value for '%s'", tag)
	}
	field.SetInt(intValue)
	return nil
}

// parseAndSetUint parses and sets an unsigned integer field.
func parseAndSetUint(field reflect.Value, value, tag string) error {
	uintValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid unsigned integer value for '%s'", tag)
	}
	field.SetUint(uintValue)
	return nil
}

// parseAndSetFloat parses and sets a float field.
func parseAndSetFloat(field reflect.Value, value, tag string) error {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float value for '%s'", tag)
	}
	field.SetFloat(floatValue)
	return nil
}

// parseAndSetBool parses and sets a boolean field.
func parseAndSetBool(field reflect.Value, value, tag string) error {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value for '%s'", tag)
	}
	field.SetBool(boolValue)
	return nil
}
