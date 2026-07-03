package err

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/beka-birhanu/toddlerr/v2/status"
	"github.com/go-playground/validator/v10"
)

var requiredTags = map[string]status.StatusCode{
	"required":             status.BadRequestMissingField,
	"required_if":          status.BadRequestMissingField,
	"required_unless":      status.BadRequestMissingField,
	"required_with":        status.BadRequestMissingField,
	"required_with_all":    status.BadRequestMissingField,
	"required_without":     status.BadRequestMissingField,
	"required_without_all": status.BadRequestMissingField,
}

var formatTags = map[string]status.StatusCode{
	"email":         status.BadRequestInvalidFormat,
	"uuid":          status.BadRequestInvalidFormat,
	"uuid3":         status.BadRequestInvalidFormat,
	"uuid4":         status.BadRequestInvalidFormat,
	"uuid5":         status.BadRequestInvalidFormat,
	"uuid3_rfc4122": status.BadRequestInvalidFormat,
	"uuid4_rfc4122": status.BadRequestInvalidFormat,
	"uuid5_rfc4122": status.BadRequestInvalidFormat,
	"uuid_rfc4122":  status.BadRequestInvalidFormat,
	"base64":        status.BadRequestInvalidFormat,
	"base64url":     status.BadRequestInvalidFormat,
	"base64rawurl":  status.BadRequestInvalidFormat,
	"json":          status.BadRequestInvalidFormat,
	"image":         status.BadRequestInvalidFormat,
}

var enumTags = map[string]status.StatusCode{
	"oneof": status.BadRequestEnumViolation,
}

var valueConstraintTags = map[string]status.StatusCode{
	"eq":             status.BadRequestInvalidValue,
	"ne":             status.BadRequestInvalidValue,
	"eq_ignore_case": status.BadRequestInvalidValue,
	"ne_ignore_case": status.BadRequestInvalidValue,
	"unique":         status.BadRequestInvalidValue,
}

var rangeTags = map[string]status.StatusCode{
	"min": status.BadRequestOutOfRange,
	"max": status.BadRequestOutOfRange,
	"len": status.BadRequestOutOfRange,
	"gt":  status.BadRequestOutOfRange,
	"lt":  status.BadRequestOutOfRange,
	"gte": status.BadRequestOutOfRange,
	"lte": status.BadRequestOutOfRange,
}

var fallbackStatusCode = status.BadRequest

// FromValidationErrors converts a validator.ValidationErrors into an Error.
// PublicStatusCode/PublicMessage reflect only the first field error, even
// with multiple failures. See MapValidationErrors for structType/masking.
func FromValidationErrors(err error, structType ...reflect.Type) *Error {
	if err == nil {
		return nil
	}

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return &Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid input provided",
			ServiceMessage:    fmt.Sprintf("Unknown validation error: %v", err),
			PublicMetaData: map[string]string{
				"error_type": "Validation",
			},
			ServiceMetaData: map[string]string{
				"error_type": "ValidatorErrorUnknown",
				"raw_error":  err.Error(),
			},
		}
	}

	fieldErrors := MapValidationErrors(ve, structType...)
	if len(fieldErrors) == 0 {
		return &Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid input provided",
			ServiceMessage:    fmt.Sprintf("Unknown validation error: %v", err),
			PublicMetaData: map[string]string{
				"error_type": "Validation",
			},
			ServiceMetaData: map[string]string{
				"error_type": "ValidatorErrorUnknown",
				"raw_error":  err.Error(),
			},
		}
	}

	fields := make([]string, 0, len(fieldErrors))
	failures := make([]string, 0, len(fieldErrors))
	serviceMessages := make([]string, 0, len(fieldErrors))
	publicMeta := make(map[string]string)
	serviceMeta := make(map[string]string)

	for _, fe := range fieldErrors {
		failures = append(failures, fmt.Sprintf("%s: %s", fe.Field, fe.Reason))
		serviceMessages = append(serviceMessages, fmt.Sprintf("Field '%s' with value '%v' failed on '%s'", fe.Field, fe.Value, fe.ValidationTag))
		fields = append(fields, fe.Field)

		publicMeta[fe.Field] = fe.Reason
		serviceMeta[fe.Field+"_reason"] = fe.ValidationTag
		serviceMeta[fe.Field+"_status_code"] = fmt.Sprintf("%d", fe.StatusCode)
	}

	return &Error{
		PublicStatusCode:  fieldErrors[0].StatusCode,
		ServiceStatusCode: fieldErrors[0].StatusCode,
		PublicMessage:     fieldErrors[0].Reason,
		ServiceMessage:    strings.Join(serviceMessages, "; "),
		PublicMetaData: map[string]string{
			"error_type": "Validation",
			"fields":     strings.Join(fields, ", "),
			"failures":   strings.Join(failures, "; "),
		},
		ServiceMetaData: map[string]string{
			"error_type": "ValidatorFieldErrors",
			"fields":     strings.Join(fields, ", "),
			"details":    fmt.Sprintf("%v", serviceMeta),
		},
	}
}

// FieldValidationError is a single field's validation failure, mapped to a
// status code and a human-readable reason.
type FieldValidationError struct {
	Field         string            `json:"field"`
	Value         any               `json:"value"`
	Reason        string            `json:"reason"`
	ValidationTag string            `json:"validation_tag"`
	StatusCode    status.StatusCode `json:"status_code"`
}

// MapValidationErrors converts each validator.FieldError into a
// FieldValidationError. structType is optional (only the first is used); if
// its Kind (after unwrapping pointers) isn't Struct, it is silently ignored
// rather than erroring. When given, fields tagged `mask:""` report "***" as
// Value instead of the real value.
func MapValidationErrors(ve validator.ValidationErrors, structType ...reflect.Type) []*FieldValidationError {
	var st reflect.Type
	if len(structType) > 0 {
		st = structType[0]
		for st != nil && st.Kind() == reflect.Ptr {
			st = st.Elem()
		}
		if st != nil && st.Kind() != reflect.Struct {
			st = nil
		}
	}

	var result []*FieldValidationError
	for _, fe := range ve {
		value := fe.Value()
		if st != nil {
			if sf, ok := st.FieldByName(fe.StructField()); ok {
				if _, masked := sf.Tag.Lookup("mask"); masked {
					value = "***"
				}
			}
		}
		result = append(result, &FieldValidationError{
			Field:         fe.Field(),
			Value:         value,
			Reason:        generateReason(fe),
			ValidationTag: fe.Tag(),
			StatusCode:    mapTagToStatusCode(fe),
		})
	}

	return result
}

var rangeTagText = map[string]string{
	"min": "at least",
	"max": "at most",
	"gt":  "greater than",
	"lt":  "less than",
	"gte": "at least",
	"lte": "at most",
}

func generateReason(fe validator.FieldError) string {
	isInMap := func(m map[string]status.StatusCode, key string) bool {
		_, ok := m[key]
		return ok
	}
	tag := fe.Tag()
	field := fe.Field()
	param := fe.Param()

	switch {
	case isInMap(requiredTags, tag):
		return fmt.Sprintf("%s is required", field)
	case isInMap(formatTags, tag):
		return fmt.Sprintf("%s must be a valid %s", field, tag)
	case tag == "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case isInMap(rangeTags, tag):
		text := rangeTagText[tag]
		if text == "" {
			text = tag
		}
		return fmt.Sprintf("%s must be %s %s", field, text, param)
	case isInMap(enumTags, tag):
		return fmt.Sprintf("%s must be one of [%s]", field, param)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

func mapTagToStatusCode(fe validator.FieldError) status.StatusCode {
	tag := fe.Tag()

	if code, ok := requiredTags[tag]; ok {
		return code
	}
	if code, ok := formatTags[tag]; ok {
		return code
	}
	if code, ok := enumTags[tag]; ok {
		return code
	}
	if code, ok := valueConstraintTags[tag]; ok {
		return code
	}
	if code, ok := rangeTags[tag]; ok {
		return code
	}
	return fallbackStatusCode
}
