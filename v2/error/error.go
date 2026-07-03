package err

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beka-birhanu/toddlerr/v2/status"
)

// Error is an application error with separate public-facing and
// internal service fields to avoid leaking sensitive details.
type Error struct {
	PublicStatusCode  status.StatusCode
	ServiceStatusCode status.StatusCode
	PublicMessage     string
	ServiceMessage    string
	PublicMetaData    map[string]string
	ServiceMetaData   map[string]string
}

// Error implements the error interface. Contains both public and service
// internals — do NOT use in API responses.
func (e *Error) Error() string {
	return fmt.Sprintf(
		"{publicStatus: %s (%d), serviceStatus: %s (%d), publicMessage: '%s', serviceMessage: '%s', publicMetaData: %s, serviceMetaData: %s}",
		status.GetErrorName(e.PublicStatusCode),
		e.PublicStatusCode,
		status.GetErrorName(e.ServiceStatusCode),
		e.ServiceStatusCode,
		e.PublicMessage,
		e.ServiceMessage,
		formatMetaData(e.PublicMetaData),
		formatMetaData(e.ServiceMetaData),
	)
}

func formatMetaData(metaData map[string]string) string {
	if len(metaData) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(metaData))
	for k := range metaData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k + ": '" + metaData[k] + "'")
	}
	b.WriteString("}")
	return b.String()
}

// NeutralizeOverDetailedStatus replaces the public status code with a safe
// generalized equivalent to avoid leaking internal error specifics.
func (e *Error) NeutralizeOverDetailedStatus() {
	e.PublicStatusCode = status.SuppressOverDetail(e.PublicStatusCode)
}
