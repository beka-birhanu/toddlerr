package err_test

import (
	"database/sql"
	"errors"
	"testing"

	toddlerr "github.com/beka-birhanu/toddlerr/v2/error"
	"github.com/beka-birhanu/toddlerr/v2/status"
	"github.com/lib/pq"
)

func TestFromDBError_Nil(t *testing.T) {
	if toddlerr.FromDBError(nil, "user") != nil {
		t.Error("nil input should return nil")
	}
}

func TestFromDBError_NoRows(t *testing.T) {
	e := toddlerr.FromDBError(sql.ErrNoRows, "user")
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	if e.PublicStatusCode != status.NotFoundResource {
		t.Errorf("expected NotFoundResource, got %d", e.PublicStatusCode)
	}
}

func TestFromDBError_UniqueViolation(t *testing.T) {
	err := &pq.Error{Code: "23505", Message: "duplicate key"}
	e := toddlerr.FromDBError(err, "user")
	if e.PublicStatusCode != status.ConflictDuplicateData {
		t.Errorf("expected ConflictDuplicateData, got %d", e.PublicStatusCode)
	}
}

func TestFromDBError_ForeignKey(t *testing.T) {
	err := &pq.Error{Code: "23503", Message: "fk violation"}
	e := toddlerr.FromDBError(err, "order")
	if e.PublicStatusCode != status.BadRequest {
		t.Errorf("expected BadRequest for FK violation, got %d", e.PublicStatusCode)
	}
}

func TestFromDBError_NotNull(t *testing.T) {
	err := &pq.Error{Code: "23502", Message: "null value"}
	e := toddlerr.FromDBError(err, "product")
	if e.PublicStatusCode != status.BadRequest {
		t.Errorf("expected BadRequest for NOT NULL violation, got %d", e.PublicStatusCode)
	}
}

func TestFromDBError_CheckViolation(t *testing.T) {
	err := &pq.Error{Code: "23514", Message: "check failed"}
	e := toddlerr.FromDBError(err, "item")
	if e.PublicStatusCode != status.BadRequest {
		t.Errorf("expected BadRequest for CHECK violation, got %d", e.PublicStatusCode)
	}
}

func TestFromDBError_UnknownPQCode(t *testing.T) {
	err := &pq.Error{Code: "99999", Message: "unknown"}
	e := toddlerr.FromDBError(err, "thing")
	if e.PublicStatusCode != status.ServerError {
		t.Errorf("expected ServerError for unknown pg code, got %d", e.PublicStatusCode)
	}
	if e.ServiceStatusCode != status.ServerErrorDatabase {
		t.Errorf("expected ServerErrorDatabase service code, got %d", e.ServiceStatusCode)
	}
}

func TestFromDBError_NonPQError(t *testing.T) {
	e := toddlerr.FromDBError(errors.New("connection refused"), "db")
	if e.PublicStatusCode != status.ServerError {
		t.Errorf("expected ServerError for non-pq error, got %d", e.PublicStatusCode)
	}
	if e.ServiceStatusCode != status.ServerErrorDatabase {
		t.Errorf("expected ServerErrorDatabase service code, got %d", e.ServiceStatusCode)
	}
}

func TestFromDBError_EntityNameLogInjection(t *testing.T) {
	injected := "user\nX-Injected: header"
	e := toddlerr.FromDBError(sql.ErrNoRows, injected)
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	if containsNewline(e.ServiceMessage) {
		t.Error("ServiceMessage contains newline — log injection not sanitized")
	}
	if containsNewline(e.PublicMessage) {
		t.Error("PublicMessage contains newline — log injection not sanitized")
	}
	for k, v := range e.PublicMetaData {
		if containsNewline(v) {
			t.Errorf("PublicMetaData[%q] contains newline — log injection not sanitized", k)
		}
	}
	for k, v := range e.ServiceMetaData {
		if containsNewline(v) {
			t.Errorf("ServiceMetaData[%q] contains newline — log injection not sanitized", k)
		}
	}
}

func TestFromDBError_ServiceContainsEntityName(t *testing.T) {
	e := toddlerr.FromDBError(sql.ErrNoRows, "invoice")
	if e.PublicMetaData["resourceName"] != "invoice" {
		t.Errorf("resourceName missing from metadata: %v", e.PublicMetaData)
	}
}

func containsNewline(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' {
			return true
		}
	}
	return false
}
