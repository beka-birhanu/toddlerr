package logger

// Str builds a string Field.
func Str(key, val string) Field {
	return Field{Key: key, Val: val}
}

// Int builds an int Field.
func Int(key string, val int) Field {
	return Field{Key: key, Val: val}
}

// Int64 builds an int64 Field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Val: val}
}

// Float64 builds a float64 Field.
func Float64(key string, val float64) Field {
	return Field{Key: key, Val: val}
}

// Bool builds a bool Field.
func Bool(key string, val bool) Field {
	return Field{Key: key, Val: val}
}

// Any builds a Field from any value. Structs/slices/maps/pointers go through
// masking (if enabled) and proto.Message/JSON-string values are logged as
// structured objects, not raw strings.
func Any(key string, val any) Field {
	return Field{Key: key, Val: val}
}
