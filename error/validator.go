package error

import (
	"fmt"
	"strings"

	"github.com/beka-birhanu/toddler/status"
	"github.com/go-playground/validator/v10"
)

// --- Tag to StatusCode Mapping Categories ---

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

func FromValidationErrors(err error) *Error {
	if err == nil {
		return nil
	}

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validator error, treat as internal error
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

	fieldErrors := MapValidationErrors(ve)

	// Combine messages and metadata
	fields := make([]string, 0, len(fieldErrors))
	publicMessages := make([]string, 0, len(fieldErrors))
	serviceMessages := make([]string, 0, len(fieldErrors))
	publicMeta := make(map[string]string)
	serviceMeta := make(map[string]string)

	var finalStatus status.StatusCode
	if len(fieldErrors) != 1 {
		finalStatus = status.BadRequest
	} else {
		finalStatus = fieldErrors[0].StatusCode
	}

	for _, fe := range fieldErrors {
		publicMessages = append(publicMessages, fmt.Sprintf("%s: %s", fe.Field, fe.Reason))
		serviceMessages = append(serviceMessages, fmt.Sprintf("Field '%s' with value '%v' failed on '%s'", fe.Field, fe.Value, fe.ValidationTag))
		fields = append(fields, fe.Field)

		publicMeta[fe.Field] = fe.Reason
		serviceMeta[fe.Field+"reason"] = fe.ValidationTag
		serviceMeta[fe.Field+"status_code"] = fmt.Sprintf("%d", fe.StatusCode)
	}

	return &Error{
		PublicStatusCode:  finalStatus,
		ServiceStatusCode: finalStatus,
		PublicMessage:     "Invalid input in one or more fields",
		ServiceMessage:    strings.Join(serviceMessages, "; "),
		PublicMetaData: map[string]string{
			"error_type": "Validation",
			"fields":     strings.Join(fields, ", "),
			"failures":   strings.Join(publicMessages, "; "),
		},
		ServiceMetaData: map[string]string{
			"error_type": "ValidatorFieldErrors",
			"fields":     strings.Join(fields, ", "),
			"details":    fmt.Sprintf("%v", serviceMeta),
		},
	}
}

type FieldValidationError struct {
	Field         string            `json:"field"`
	Value         any               `json:"value"`
	Reason        string            `json:"reason"`
	ValidationTag string            `json:"validation_tag"`
	StatusCode    status.StatusCode `json:"status_code"`
}

func MapValidationErrors(ve validator.ValidationErrors) []*FieldValidationError {
	var result []*FieldValidationError

	for _, fe := range ve {
		result = append(result, &FieldValidationError{
			Field:         fe.Field(),
			Value:         fe.Value(),
			Reason:        generateReason(fe),
			ValidationTag: fe.Tag(),
			StatusCode:    mapTagToStatusCode(fe),
		})
	}

	return result
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
		return fmt.Sprintf("%s must be %s %s", field, tag, param)
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
