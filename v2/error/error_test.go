package err_test

import (
	"strings"
	"testing"

	err "github.com/beka-birhanu/toddlerr/v2/error"
	"github.com/beka-birhanu/toddlerr/v2/status"
)

// --- Error() ---

func TestError_Error(t *testing.T) {
	e := &err.Error{
		PublicStatusCode:  status.BadRequestMissingField,
		ServiceStatusCode: status.BadRequestMissingField,
		PublicMessage:     "Missing required field",
		ServiceMessage:    "Field 'username' is missing",
		PublicMetaData:    map[string]string{"field": "username"},
		ServiceMetaData:   map[string]string{"requestId": "abc123"},
	}

	got := e.Error()
	want := "{publicStatus: BadRequest_MissingField (4001), serviceStatus: BadRequest_MissingField (4001), publicMessage: 'Missing required field', serviceMessage: 'Field 'username' is missing', publicMetaData: {field: 'username'}, serviceMetaData: {requestId: 'abc123'}}"
	if got != want {
		t.Errorf("Error():\ngot  %s\nwant %s", got, want)
	}
}

func TestError_EmptyMetaData(t *testing.T) {
	e := &err.Error{
		PublicStatusCode: status.BadRequest,
		PublicMessage:    "bad",
	}
	got := e.Error()
	if !strings.Contains(got, "{}") {
		t.Errorf("empty metadata should render as {}: %s", got)
	}
}

func TestError_MetaDataDeterministic(t *testing.T) {
	e := &err.Error{
		PublicStatusCode: status.BadRequest,
		PublicMetaData: map[string]string{
			"z_key": "1",
			"a_key": "2",
			"m_key": "3",
		},
	}
	first := e.Error()
	for i := range 20 {
		if e.Error() != first {
			t.Fatalf("formatMetaData output is non-deterministic on iteration %d", i)
		}
	}
}

// --- NeutralizeOverDetailedStatus ---

func TestNeutralizeOverDetailedStatus(t *testing.T) {
	e := &err.Error{PublicStatusCode: status.ServerErrorDatabase}
	e.NeutralizeOverDetailedStatus()
	if e.PublicStatusCode != status.ServerError {
		t.Errorf("expected ServerError after neutralize, got %d", e.PublicStatusCode)
	}
}

func TestNeutralizeOverDetailedStatus_MissingField_Unchanged(t *testing.T) {
	e := &err.Error{PublicStatusCode: status.BadRequestMissingField}
	e.NeutralizeOverDetailedStatus()
	if e.PublicStatusCode != status.BadRequestMissingField {
		t.Errorf("BadRequestMissingField should not be suppressed, got %d", e.PublicStatusCode)
	}
}
