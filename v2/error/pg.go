package err

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beka-birhanu/toddlerr/v2/status"
	"github.com/lib/pq"
)

const (
	errUniqueViolation  = "23505"
	errForeignKey       = "23503"
	errNotNullViolation = "23502"
	errCheckViolation   = "23514"
)

func sanitizeEntityName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, name)
}

type pgConstraintErrorSpec struct {
	publicStatus  status.StatusCode
	publicMessage string
	errorType     string
	serviceVerb   string
	metaKey       string
}

func (s pgConstraintErrorSpec) build(entityName string, pqErr *pq.Error) *Error {
	publicMeta := map[string]string{
		"error_type":   s.errorType,
		"resourceName": entityName,
	}
	serviceMeta := map[string]string{
		"pgcode":         string(pqErr.Code),
		"error_type":     s.errorType,
		"resourceName":   entityName,
		"error_message":  pqErr.Message,
		"error_severity": pqErr.Severity,
		"raw_error":      pqErr.Error(),
	}
	switch s.metaKey {
	case "constraint":
		serviceMeta["constraint"] = pqErr.Constraint
	case "column":
		serviceMeta["column"] = pqErr.Column
	}

	return &Error{
		PublicStatusCode:  s.publicStatus,
		ServiceStatusCode: s.publicStatus,
		PublicMessage:     fmt.Sprintf(s.publicMessage, entityName),
		PublicMetaData:    publicMeta,
		ServiceMessage:    fmt.Sprintf("%s on %s: %s", s.serviceVerb, entityName, pqErr.Message),
		ServiceMetaData:   serviceMeta,
	}
}

var pgConstraintErrorSpecs = map[string]pgConstraintErrorSpec{
	errUniqueViolation: {
		publicStatus:  status.ConflictDuplicateData,
		publicMessage: "A %s with the same value already exists",
		errorType:     "Data duplication",
		serviceVerb:   "Unique constraint violation",
		metaKey:       "constraint",
	},
	errForeignKey: {
		publicStatus:  status.BadRequest,
		publicMessage: "%s has invalid reference to related data",
		errorType:     "Foreign key violation",
		serviceVerb:   "Foreign key constraint failed",
		metaKey:       "constraint",
	},
	errNotNullViolation: {
		publicStatus:  status.BadRequest,
		publicMessage: "%s is missing required fields",
		errorType:     "Missing field",
		serviceVerb:   "NOT NULL constraint failed",
		metaKey:       "column",
	},
	errCheckViolation: {
		publicStatus:  status.BadRequest,
		publicMessage: "%s failed validation rules",
		errorType:     "Constraint check failed",
		serviceVerb:   "CHECK constraint violation",
		metaKey:       "constraint",
	},
}

// FromDBError maps Postgres errors into application errors.
func FromDBError(err error, entityName string) *Error {
	if err == nil {
		return nil
	}
	entityName = sanitizeEntityName(entityName)

	if errors.Is(err, sql.ErrNoRows) {
		return &Error{
			PublicStatusCode:  status.NotFoundResource,
			ServiceStatusCode: status.NotFoundResource,
			PublicMessage:     fmt.Sprintf("Either %s does not exist or you don't have access", entityName),
			PublicMetaData: map[string]string{
				"error_type":   "Data not found",
				"resourceName": entityName,
			},
			ServiceMessage: fmt.Sprintf("No record found for %s: %s", entityName, err),
			ServiceMetaData: map[string]string{
				"error_type":   "Data not found",
				"resourceName": entityName,
				"raw_error":    err.Error(),
			},
		}
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if spec, ok := pgConstraintErrorSpecs[string(pqErr.Code)]; ok {
			return spec.build(entityName, pqErr)
		}

		return &Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerErrorDatabase,
			PublicMessage:     "A server error occurred. Please try again later.",
			PublicMetaData: map[string]string{
				"error_type":   "Internal database error",
				"resourceName": entityName,
			},
			ServiceMessage: fmt.Sprintf("Unhandled PostgreSQL error for %s: %s", entityName, pqErr.Message),
			ServiceMetaData: map[string]string{
				"pgcode":         string(pqErr.Code),
				"resourceName":   entityName,
				"error_message":  pqErr.Message,
				"error_severity": pqErr.Severity,
				"raw_error":      pqErr.Error(),
			},
		}
	}

	return &Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerErrorDatabase,
		PublicMessage:     "A server error occurred. Please try again later.",
		PublicMetaData: map[string]string{
			"error_type":   "Unknown server error",
			"resourceName": entityName,
		},
		ServiceMessage: fmt.Sprintf("Unexpected DB error for %s: %s", entityName, err),
		ServiceMetaData: map[string]string{
			"error_type":   "Unknown database error",
			"resourceName": entityName,
			"raw_error":    err.Error(),
		},
	}
}
