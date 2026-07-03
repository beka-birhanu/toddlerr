# toddlerr/v2

Intro, motivation, and basic usage are in the main [README](../README.md). This doc covers only what's new from v1.

## What's new in the error package

1. **Better error messages from `FromValidationErrors`**

   In v1, multiple errors collapsed to "invalid input on one or more fields" — not helpful to the frontend devs. Now we return the first error message from `validator.ValidationErrors`.

2. **Masking sensitive fields**

   ```go
   type CreateUserRequest struct {
       ...
       Password string `validate:"required,min=8"`
       ...
   }
   ```

   In v1, logging `err.Error()` exposes the password, e.g. "Password must be at least 8 chars long: value = 123".

   In v2, add `mask:""` to the field to mask it: "Password must be at least 8 chars long: value = \*\*\*"

   Usage:

   ```go
   type CreateUserRequest struct {
       ...
       Password string `validate:"required,min=8" mask:""`
       ...
   }
   appErr := toddlerr.FromValidationErrors(err, reflect.TypeOf(CreateUserRequest{}))
   ```

## Logger

[zap](https://github.com/uber-go/zap) wrapped logger with optional field masking and child logger support via `With()` from a singleton parent.

### Usage

```go
// main.go — build one logger at startup and store it as a package-level var.
// WithStdout(): write JSON logs to stdout. MaskEnabled(): honor `mask:""` tags.
var Log = logger.NewLogger(
    logger.WithStdout(),
    logger.MaskEnabled(),
)
defer Log.Close() // flushes any custom writers on shutdown

// Per request/goroutine, derive a child with With() instead of reusing Log
// directly — it shares Log's core but carries its own fields, and Log itself
// is never mutated, so it stays safe to share across goroutines.
reqLog := Log.With(logger.Str("correlation_id", requestID))
reqLog.Info(ctx, "request started")

// Tag a field `mask:""` to redact it from log output wherever it's logged
// via logger.Any — useful for passwords, tokens, PII, etc.
type LoginRequest struct {
    Username string
    Password string `mask:""`
}
Log.Info(ctx, "login attempt", logger.Any("req", req)) // Password logs as "***"
```
