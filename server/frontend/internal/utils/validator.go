package utils

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
)

// Validator instance
var validate = validator.New()

// ValidatePOSTFormRequest binds and validates form fields from the request body.
func ValidatePOSTFormRequest(obj any, r *http.Request) error {
	if err := ensurePointer(obj); err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	return bindAndValidate(obj, r.PostForm, "form")
}

// ValidatePOSTFieldsFromMultipartFormData binds and validates form fields from the request body.
func ValidatePOSTFieldsFromMultipartFormData(obj any, r *http.Request) error {
	if err := ensurePointer(obj); err != nil {
		return err
	}

	formData := FormData{
		Values: r.Form,
		Files:  r.MultipartForm.File,
	}

	return bindAndValidateMultipart(obj, formData, "form")
}

// ValidateQueryParameters validates query parameters against struct 'query' tags.
func ValidateQueryParameters(obj any, r *http.Request) error {
	if err := ensurePointer(obj); err != nil {
		return err
	}

	return bindAndValidate(obj, r.URL.Query(), "query")
}

// ensurePointer verifies that the given object is a pointer.
func ensurePointer(obj any) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
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

		// Fetch tag
		tag := fieldType.Tag.Get(tagKey)
		if tag == "" || !field.CanSet() {
			continue
		}

		// Get value from data
		values, exists := data[tag]
		if !exists || len(values) == 0 {
			continue
		}
		value := values[0]

		// Set field value
		if err := setFieldValue(field, value, tag); err != nil {
			return err
		}
	}

	// Struct validation
	if err := validate.Struct(obj); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return fmt.Errorf("validation error: %s", validationErrors[0].Error())
		}
		return err
	}

	return nil
}

type FormData struct {
	Values map[string][]string
	Files  map[string][]*multipart.FileHeader
}

// bindAndValidateMultipart binds and validates both form values and files
func bindAndValidateMultipart(obj any, data FormData, tagKey string) error {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()
	fileHeaderType := reflect.TypeOf((*multipart.FileHeader)(nil))

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get(tagKey)
		if tag == "" {
			continue
		}

		currentType := field.Type()

		// Handle file fields
		switch {
		case currentType == fileHeaderType:
			if files, exists := data.Files[tag]; exists && len(files) > 0 {
				field.Set(reflect.ValueOf(files[0]))
			}
		case currentType == reflect.SliceOf(fileHeaderType):
			if files, exists := data.Files[tag]; exists {
				field.Set(reflect.ValueOf(files))
			}
		default:
			// Handle regular form values
			if values, exists := data.Values[tag]; exists && len(values) > 0 {
				if err := setFieldValue(field, values[0], tag); err != nil {
					return err
				}
			}
		}
	}

	// Validation using go-playground/validator
	if err := validate.Struct(obj); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return fmt.Errorf("validation error: %s", ve[0].Error())
		}
		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}

// setFieldValue sets a value to a struct field based on its type.
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
	case reflect.Pointer:
		return parseAndSetBoolPointer(field, value, tag) // WARNING! pay attention to this as this can cause problem, but it was necessary to handle bool pointers for false values
	default:
		log.Errorf("unhandled field type: %v", field.Kind())
	}
	return nil
}

// parseAndSetBoolPointer parses and sets a pointer to a boolean field.
func parseAndSetBoolPointer(field reflect.Value, value, tag string) error {
	if field.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer field for '%s'", tag)
	}

	// If the input value is empty, set the field to nil
	if value == "" {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// Parse the boolean value
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value for '%s'", tag)
	}

	// Ensure the pointer is initialized
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem())) // Allocate new bool pointer
	}

	// Set the parsed boolean value
	field.Elem().SetBool(boolValue)
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
