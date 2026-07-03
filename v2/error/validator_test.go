package err_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	err "github.com/beka-birhanu/toddlerr/v2/error"
	"github.com/beka-birhanu/toddlerr/v2/status"
	"github.com/go-playground/validator/v10"
)

// --- FromValidationErrors ---

func TestFromValidationErrors_Nil(t *testing.T) {
	if err.FromValidationErrors(nil) != nil {
		t.Error("nil input should return nil")
	}
}

func TestFromValidationErrors_NonValidatorError(t *testing.T) {
	e := err.FromValidationErrors(errors.New("something"))
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	if e.PublicStatusCode != status.BadRequest {
		t.Errorf("expected BadRequest, got %d", e.PublicStatusCode)
	}
}

func TestFromValidationErrors_SingleField(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	ve := runValidate(t, S{})

	e := err.FromValidationErrors(ve)
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	if e.PublicStatusCode != status.BadRequestMissingField {
		t.Errorf("single field: expected BadRequestMissingField, got %d", e.PublicStatusCode)
	}
	if !strings.Contains(e.PublicMessage, "Name") {
		t.Errorf("single field: PublicMessage should name the field, got: %s", e.PublicMessage)
	}
}

func TestFromValidationErrors_MultipleFields(t *testing.T) {
	type S struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}
	ve := runValidate(t, S{})

	e := err.FromValidationErrors(ve)
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	// PublicMessage = first field's reason only
	if e.PublicMessage != "Name is required" {
		t.Errorf("multi-field: PublicMessage should be first error only, got: %s", e.PublicMessage)
	}
	// PublicStatusCode = first field's status code
	if e.PublicStatusCode != status.BadRequestMissingField {
		t.Errorf("multi-field: PublicStatusCode should be first field's code, got %d", e.PublicStatusCode)
	}
	// metadata lists all fields
	if !strings.Contains(e.PublicMetaData["fields"], "Name") || !strings.Contains(e.PublicMetaData["fields"], "Email") {
		t.Errorf("multi-field: metadata missing fields: %v", e.PublicMetaData)
	}
	// failures metadata has field-prefixed entries for all errors
	if !strings.Contains(e.PublicMetaData["failures"], "Name:") || !strings.Contains(e.PublicMetaData["failures"], "Email:") {
		t.Errorf("multi-field: failures metadata missing entries: %s", e.PublicMetaData["failures"])
	}
}

func TestFromValidationErrors_ServiceMetaKeySeparator(t *testing.T) {
	type S struct {
		Age int `validate:"min=1"`
	}
	ve := runValidate(t, S{Age: 0})

	e := err.FromValidationErrors(ve)
	if e == nil {
		t.Fatal("expected non-nil error")
	}
	details := e.ServiceMetaData["details"]
	if strings.Contains(details, "Agereason") || strings.Contains(details, "Agestatus_code") {
		t.Error("serviceMeta keys missing underscore separator")
	}
	if !strings.Contains(details, "Age_reason") || !strings.Contains(details, "Age_status_code") {
		t.Errorf("serviceMeta keys not found with underscore separator: %s", details)
	}
}

// --- MapValidationErrors ---

func TestMapValidationErrors_RequiredTag(t *testing.T) {
	type S struct {
		F string `validate:"required"`
	}
	ve := runValidate(t, S{})
	result := err.MapValidationErrors(ve)
	if len(result) == 0 {
		t.Fatal("expected field errors")
	}
	if result[0].StatusCode != status.BadRequestMissingField {
		t.Errorf("expected BadRequestMissingField, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_EmailTag(t *testing.T) {
	type S struct {
		F string `validate:"email"`
	}
	ve := runValidate(t, S{F: "notanemail"})
	result := err.MapValidationErrors(ve)
	if result[0].StatusCode != status.BadRequestInvalidFormat {
		t.Errorf("expected BadRequestInvalidFormat, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_OneofTag(t *testing.T) {
	type S struct {
		F string `validate:"oneof=a b c"`
	}
	ve := runValidate(t, S{F: "z"})
	result := err.MapValidationErrors(ve)
	if result[0].StatusCode != status.BadRequestEnumViolation {
		t.Errorf("expected BadRequestEnumViolation, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_MinTag(t *testing.T) {
	type S struct {
		F int `validate:"min=5"`
	}
	ve := runValidate(t, S{F: 1})
	result := err.MapValidationErrors(ve)
	if result[0].StatusCode != status.BadRequestOutOfRange {
		t.Errorf("expected BadRequestOutOfRange, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_LenTag(t *testing.T) {
	type S struct {
		F string `validate:"len=6"`
	}
	ve := runValidate(t, S{F: "ab"})
	result := err.MapValidationErrors(ve)
	if result[0].StatusCode != status.BadRequestOutOfRange {
		t.Errorf("expected BadRequestOutOfRange, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_UnknownTag(t *testing.T) {
	type S struct {
		F string `validate:"alphanum"`
	}
	ve := runValidate(t, S{F: "!!!"})
	result := err.MapValidationErrors(ve)
	if result[0].StatusCode != status.BadRequest {
		t.Errorf("expected fallback BadRequest, got %d", result[0].StatusCode)
	}
}

func TestMapValidationErrors_MaskedField(t *testing.T) {
	type S struct {
		Password string `validate:"min=8" mask:""`
	}
	ve := runValidate(t, S{Password: "short"})
	result := err.MapValidationErrors(ve, reflect.TypeOf(S{}))
	if result[0].Value != "***" {
		t.Errorf("masked field should report '***' as value, got: %v", result[0].Value)
	}
}

func TestMapValidationErrors_UnmaskedField(t *testing.T) {
	type S struct {
		Name string `validate:"min=8"`
	}
	ve := runValidate(t, S{Name: "short"})
	result := err.MapValidationErrors(ve, reflect.TypeOf(S{}))
	if result[0].Value != "short" {
		t.Errorf("unmasked field should report real value, got: %v", result[0].Value)
	}
}

func TestMapValidationErrors_NonStructType_NoPanic(t *testing.T) {
	type S struct {
		F string `validate:"required"`
	}
	ve := runValidate(t, S{})
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MapValidationErrors panicked with non-struct structType: %v", r)
		}
	}()
	err.MapValidationErrors(ve, reflect.TypeOf([]int{}))
}

// --- generateReason (via MapValidationErrors) ---

func TestGenerateReason_Required(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	ve := runValidate(t, S{})
	result := err.MapValidationErrors(ve)
	if result[0].Reason != "Name is required" {
		t.Errorf("unexpected reason: %s", result[0].Reason)
	}
}

func TestGenerateReason_Format(t *testing.T) {
	type S struct {
		Email string `validate:"email"`
	}
	ve := runValidate(t, S{Email: "bad"})
	result := err.MapValidationErrors(ve)
	if !strings.Contains(result[0].Reason, "valid email") {
		t.Errorf("unexpected reason: %s", result[0].Reason)
	}
}

func TestGenerateReason_Len(t *testing.T) {
	type S struct {
		Code string `validate:"len=6"`
	}
	ve := runValidate(t, S{Code: "ab"})
	result := err.MapValidationErrors(ve)
	if !strings.Contains(result[0].Reason, "exactly") {
		t.Errorf("len reason should say 'exactly': %s", result[0].Reason)
	}
}

func TestGenerateReason_RangeHumanText(t *testing.T) {
	// test via real structs with static tags
	type Smin struct {
		F string `validate:"min=3"`
	}
	type Smax struct {
		F string `validate:"max=2"`
	}
	type Sgt struct {
		F int `validate:"gt=10"`
	}
	type Slt struct {
		F int `validate:"lt=1"`
	}
	type Sgte struct {
		F int `validate:"gte=5"`
	}
	type Slte struct {
		F int `validate:"lte=2"`
	}

	check := func(name, needle string, result []*err.FieldValidationError) {
		t.Helper()
		if len(result) == 0 {
			t.Fatalf("%s: no errors", name)
		}
		if !strings.Contains(result[0].Reason, needle) {
			t.Errorf("%s: reason %q missing %q", name, result[0].Reason, needle)
		}
	}

	check("min", "at least", err.MapValidationErrors(runValidate(t, Smin{F: "x"})))
	check("max", "at most", err.MapValidationErrors(runValidate(t, Smax{F: "xxx"})))
	check("gt", "greater than", err.MapValidationErrors(runValidate(t, Sgt{F: 5})))
	check("lt", "less than", err.MapValidationErrors(runValidate(t, Slt{F: 5})))
	check("gte", "at least", err.MapValidationErrors(runValidate(t, Sgte{F: 3})))
	check("lte", "at most", err.MapValidationErrors(runValidate(t, Slte{F: 5})))
}

func TestGenerateReason_Enum(t *testing.T) {
	type S struct {
		Role string `validate:"oneof=admin user"`
	}
	ve := runValidate(t, S{Role: "superuser"})
	result := err.MapValidationErrors(ve)
	if !strings.Contains(result[0].Reason, "admin") || !strings.Contains(result[0].Reason, "user") {
		t.Errorf("enum reason should list options: %s", result[0].Reason)
	}
}

func TestGenerateReason_Unknown(t *testing.T) {
	type S struct {
		F string `validate:"alphanum"`
	}
	ve := runValidate(t, S{F: "!@#"})
	result := err.MapValidationErrors(ve)
	if !strings.Contains(result[0].Reason, "alphanum") {
		t.Errorf("unknown tag reason should include tag name: %s", result[0].Reason)
	}
}

// --- helpers ---

func runValidate(t *testing.T, s any) validator.ValidationErrors {
	t.Helper()
	v := validator.New()
	verr := v.Struct(s)
	if verr == nil {
		t.Fatal("expected validation error but got none")
	}
	ve, ok := verr.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T", verr)
	}
	return ve
}
