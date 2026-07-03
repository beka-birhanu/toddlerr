package status

import "fmt"

// StatusCode defines custom application-specific status codes that extend
// HTTP status semantics with 4-digit granularity.
type StatusCode int

// BadRequest-related errors (4000 - 4009)
const (
	BadRequest                StatusCode = 4000 + iota // Generic bad request
	BadRequestMissingField                             // Required field missing
	BadRequestTypeMismatch                             // Type mismatch
	BadRequestFieldConstraint                          // Field constraint failed
	BadRequestInvalidFormat                            // Invalid format
	BadRequestOutOfRange                               // Value out of range
	BadRequestInvalidValue                             // Invalid value
	BadRequestEnumViolation                            // Enum value not allowed
)

// Unauthorized-related errors (4010 - 4019)
const (
	Unauthorized                  StatusCode = 4010 + iota // Generic unauthorized
	UnauthorizedInvalidCredential                          // Invalid credentials
	UnauthorizedTokenRequired                              // Token required
	UnauthorizedInvalidToken                               // Invalid token
)

// Forbidden-related errors (4030 - 4039)
const (
	Forbidden                   StatusCode = 4030 + iota // Generic forbidden
	ForbiddenNotEnoughPrivilege                          // Insufficient privileges
	ForbiddenOnlyOwners                                  // Allowed for resource owners only
)

// NotFound-related errors (4040 - 4049)
const (
	NotFound         StatusCode = 4040 + iota // Generic not found
	NotFoundResource                          // Resource not found
)

// Conflict-related errors (4090 - 4099)
const (
	Conflict              StatusCode = 4090 + iota // Generic conflict
	ConflictDuplicateData                          // Duplicate data
)

// Server-related errors (5000 - 5009)
const (
	ServerError                     StatusCode = 5000 + iota // Generic server error
	ServerErrorDatabase                                      // Database error
	ServerErrorServiceCommunication                          // Service communication failed
)

var statusCodeMap = map[StatusCode]string{
	BadRequest:                      "BadRequest",
	BadRequestMissingField:          "BadRequest_MissingField",
	BadRequestTypeMismatch:          "BadRequest_TypeMismatch",
	BadRequestFieldConstraint:       "BadRequest_FieldConstraint",
	BadRequestInvalidFormat:         "BadRequest_InvalidFormat",
	BadRequestOutOfRange:            "BadRequest_OutOfRange",
	BadRequestInvalidValue:          "BadRequest_InvalidValue",
	BadRequestEnumViolation:         "BadRequest_EnumViolation",
	Unauthorized:                    "Unauthorized",
	UnauthorizedInvalidCredential:   "Unauthorized_InvalidCredential",
	UnauthorizedTokenRequired:       "Unauthorized_TokenRequired",
	UnauthorizedInvalidToken:        "Unauthorized_InvalidToken",
	Forbidden:                       "Forbidden",
	ForbiddenNotEnoughPrivilege:     "Forbidden_NotEnoughPrivilege",
	ForbiddenOnlyOwners:             "Forbidden_OnlyOwners",
	NotFound:                        "NotFound",
	NotFoundResource:                "NotFound_Resource",
	Conflict:                        "Conflict",
	ConflictDuplicateData:           "Conflict_DuplicateData",
	ServerError:                     "ServerError",
	ServerErrorDatabase:             "ServerError_Database",
	ServerErrorServiceCommunication: "ServerError_ServiceCommunication",
}

// GetErrorName returns the code's symbolic name, or "UnknownStatusCode-N"
// if it isn't one of the defined constants.
func GetErrorName(code StatusCode) string {
	if name, exists := statusCodeMap[code]; exists {
		return name
	}
	return fmt.Sprintf("UnknownStatusCode-%d", code)
}

// suppressMap maps over-detailed status codes to public-safe equivalents.
// BadRequestMissingField, BadRequestTypeMismatch, and BadRequestFieldConstraint
// are intentionally absent — those details are safe to expose to clients.
var suppressMap = map[StatusCode]StatusCode{
	BadRequestOutOfRange:            BadRequest,
	BadRequestInvalidValue:          BadRequest,
	BadRequestEnumViolation:         BadRequest,
	ForbiddenOnlyOwners:             Forbidden,
	ServerErrorDatabase:             ServerError,
	ServerErrorServiceCommunication: ServerError,
}

// SuppressOverDetail returns a neutralized version of the given StatusCode.
func SuppressOverDetail(code StatusCode) StatusCode {
	if suppressed, ok := suppressMap[code]; ok {
		return suppressed
	}
	return code
}
