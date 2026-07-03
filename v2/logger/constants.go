package logger

import (
	"reflect"
	"time"
)

// File is currently unused internally; exported for callers that want a
// consistent field key for file-related log fields.
const (
	File = "file"
)

// LogTypeSYS is the "logType" value stamped on every log line emitted by
// this package.
const (
	LogTypeSYS = "SYS"
)

const separator = "|"

const (
	maskTag       = "mask"
	sliceByteMask = "X@BQ1"
	stringMask    = "***"
)

// TypeSliceOfBytes and TypeTime are reflect.Type sentinels masking uses to
// special-case []byte (never masked) and time.Time (copied whole, since its
// fields are unexported).
var (
	TypeSliceOfBytes = reflect.TypeOf([]byte(nil))
	TypeTime         = reflect.TypeOf(time.Time{})
)
