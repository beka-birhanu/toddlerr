# Toddlerr
Toddlerr is a Go package for structured error handling with custom 4-digit status codes, extending standard HTTP status semantics. It helps categorize and manage errors more granularly, with clear separation of public and internal values.

## Why

After rewriting custom error handling in nearly every Go project I've touched, I finally decided to create a reusable library. There's no magic here — just structured error handling with custom status codes, and clear separation of public and private values.

## Usage

### 1. Creating an Error

The `Error` struct represents a custom error with a status code, public message, service message, and metadata. Here’s an example of creating and using the `Error` struct:

```go
package main

import (
	"fmt"
	toddlerr "github.com/beka-birhanu/toddlerr/error"
	"github.com/beka-birhanu/toddlerr/status"
)

func main() {
	err := &toddlerr.Error{
		StatusCode:   status.UnauthorizedInvalidCredential,
		PublicMessage: "Incorrect username or password. Please try again.",
		ServiceMessage: "Incorrect password attempt for user ID: 6eb3d746-1be1-445e-9d76-5aa996754dbd",
		PublicMetaData: map[string]string{
			"error_type": "authentication", 
			"hint":       "Please check your credentials.",
		},
		ServiceMetaData: map[string]string{
			"timestamp": "2025-05-06T10:00:00Z",
			"user_id":   "6eb3d746-1be1-445e-9d76-5aa996754dbd",
		},
	}
    
	// Print error with detailed information
	fmt.Println(err.Error())
    
    /* Output:
    {
      status: Unauthorized_InvalidCredential (4011),
      publicMessage: 'Incorrect username or password. Please try again.',
      serviceMessage: 'Incorrect password attempt for user ID: 6eb3d746-1be1-445e-9d76-5aa996754dbd',
      publicMetaData: {error_type: 'authentication', hint: 'Please check your credentials.'}, 
      serviceMetaData: {timestamp: '2025-05-06T10:00:00Z', user_id: '6eb3d746-1be1-445e-9d76-5aa996754dbd'}
    }
    */
}

```

### 2. Neutralizing Overly Detailed Status Codes

When an error occurs, sensitive internal information (e.g., database details) should not be exposed in public-facing messages. **Neutralizing** maps detailed error codes to more general ones, preventing the leak of internal specifics.

#### Example

```go
err := &toddlerr.Error{
	StatusCode:   status.ServerErrorDatabase,
	PublicMessage: "Internal server error.",
	ServiceMessage: "Database connection failed.",
}

// Print error before neutralization
fmt.Println("Before Neutralization: ", err.Error())

// Neutralize error code
err.NeutralizeOverDetailedStatus()

// Print error after neutralization
fmt.Println("After Neutralization: ", err.Error())
```

#### Output:

Before Neutralization:

```
{
  status: ServerError_Database (5001),
  publicMessage: 'Internal server error.',
  serviceMessage: 'Database connection failed.',
}
```

After Neutralization:

```
{
  status: ServerError (5000),
  publicMessage: 'Internal server error.',
  serviceMessage: 'Database connection failed.',
}
```

### 3. Status Code Mapping

The status codes in this package are organized into groups based on HTTP semantics:

* **4000-4009 (Bad Request)**: Errors related to invalid user input or malformed requests.
* **4010-4019 (Unauthorized)**: Authentication or authorization failures.
* **4030-4039 (Forbidden)**: Access control issues.
* **4040-4049 (Not Found)**: Resources not found.
* **5000-5009 (Server Error)**: Internal server errors or service failures.

Each error code has a corresponding human-readable name, which can be retrieved using `status.GetErrorName(code)`.

### Example:

```go
fmt.Println(status.GetErrorName(status.BadRequestMissingField))
// Output: "BadRequest_MissingField"
```

## Statuses

All status codes extend standard HTTP semantics plus one more digit (**4-digit codes**) to improve clarity in error handling. 

| HTTP Status (First 3 Digits)   | Custom Status Code Range  (Custom Status Code)|
|----------------|------------------------------------------------|
| 400 Bad Request| 4000 - 4009                                     |
|                | - 4000: BadRequest                              |
|                | - 4001: BadRequestMissingField                 |
|                | - 4002: BadRequestTypeMismatch                 |
|                | - 4003: BadRequestFieldConstraint              |
|                | - 4004: BadRequestInvalidFormat                |
|                | - 4005: BadRequestOutOfRange                   |
|                | - 4006: BadRequestInvalidValue                 |
|                | - 4007: BadRequestEnumViolation                |
| 401 Unauthorized| 4010 - 4019                                     |
|                | - 4010: Unauthorized                           |
|                | - 4011: UnauthorizedInvalidCredential          |
|                | - 4012: UnauthorizedTokenRequired              |
|                | - 4013: UnauthorizedInvalidToken               |
| 403 Forbidden   | 4030 - 4039                                     |
|                | - 4030: Forbidden                              |
|                | - 4031: ForbiddenNotEnoughPrivilege            |
|                | - 4032: ForbiddenOnlyOwners                    |
| 404 Not Found   | 4040 - 4049                                     |
|                | - 4040: NotFound                               |
|                | - 4041: NotFoundResource                       |
| 409 Conflict   |  4090 - 4099                                | 
|                | - 4090: Conflict                             |
|                | - 4090: ConflictDuplicateData |
| 500 Server Error| 5000 - 5009                                     |
|                | - 5000: ServerError                            |
|                | - 5001: ServerErrorDatabase                    |
|                | - 5002: ServerErrorServiceCommunication        |

## Error mappers
It includes error mapper for postgresql and validator erros. 

### `FromDBError(err error, entityName string) *error.Error`

**Purpose:**
Maps low-level PostgreSQL errors into structured, application-specific error types that include detailed status codes, public-safe messages, and internal metadata for debugging.

#### Handles:

| Database Error Type                    | Mapped Application Error                |
| -------------------------------------- | --------------------------------------- |
| `sql.ErrNoRows`                        | `status.NotFoundResource`               |
| PostgreSQL Unique Constraint (`23505`) | `status.ConflictDuplicateData`          |
| Foreign Key Violation (`23503`)        | `status.BadRequest` (invalid reference) |
| Not Null Violation (`23502`)           | `status.BadRequest` (missing field)     |
| Check Constraint (`23514`)             | `status.BadRequest` (failed validation) |
| Unhandled PostgreSQL Error             | `status.ServerErrorDatabase`            |
| Unknown Errors                         | `status.ServerErrorDatabase`            |

## `FromValidationErrors` — Structured Validation Error Handler

```go
func FromValidationErrors(err error) *error.Error
```

### 📖 Description

`FromValidationErrors` converts `go-playground/validator` validation errors into a structured application-level error (`*error.Error`). It makes it easy to return clear and helpful feedback to users while preserving rich diagnostic details for logging and debugging.

### Usage Example

```go
type SignupInput struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"required,gte=18"`
}

func ValidateSignupInput(input SignupInput) *error.Error {
    err := validator.New().Struct(input)
    return error.FromValidationErrors(err)
}
```

### 🧭 Tag-to-Status Mapping

| Category          | Tags                                  | Status Code                        |
| ----------------- | ------------------------------------- | ---------------------------------- |
| Required          | `required`, `required_with`, ...      | `status.BadRequestMissingField`    |
| Format / Pattern  | `email`, `uuid`, `json`, ...          | `status.BadRequestInvalidFormat`   |
| Range / Length    | `min`, `max`, `len`, `gt`, `lte`, ... | `status.BadRequestOutOfRange`      |
| Enum / One of     | `oneof`                               | `status.BadRequestEnumViolation`   |
| Value Constraints | `eq`, `ne`, `unique`, ...             | `status.BadRequestInvalidValue`    |
| Unknown           | Anything not explicitly mapped        | `status.BadRequest` |

---

### Output Example

```json
{
  "PublicStatusCode": 4001,
  "PublicMessage": "Invalid input in one or more fields",
  "PublicMetaData": {
    "error_type": "Validation",
    "fields": "Email, Age",
    "failures": "Email: Email must be a valid email; Age: Age must be gte 18"
  },
  "ServiceMetaData": {
    "error_type": "ValidatorFieldErrors",
    "fields": "Email, Age",
    "details": {
      "Emailreason": "email",
      "Emailstatus_code": "4003",
      "Agereason": "gte",
      "Agestatus_code": "4004"
    }
  }
}
```
