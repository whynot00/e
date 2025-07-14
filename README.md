# Package e - Error Wrapper for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/whynot00/e.svg)](https://pkg.go.dev/github.com/whynot00/e) [![Go Report Card](https://goreportcard.com/badge/github.com/whynot00/e)](https://goreportcard.com/report/github.com/whynot00/e) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`e` is a lightweight Go package providing enhanced error wrapping with stack trace capture and structured logging support using Go's `log/slog` package.

## Features

- Wrap errors with automatic stack trace capture
- Attach custom contextual messages to errors
- Compatible with Go standard library `errors.Is` and `errors.As`
- Produce structured logs with detailed error trace using `slog.Group`
- Simplifies function names in stack traces for better readability

## Installation

```bash
go get github.com/whynot00/e
```

## Usage

### Wrapping errors
Wrap an existing error to capture the current stack frame:

```go
err := errors.New("database error")

wrappedErr := e.Wrap(err)
```

Wrap with a custom message describing context:
```go
wrappedErr := e.WrapWithMessage(err, "failed to load user profile")
```

### Structured logging with slog
Integrate with `log/slog` for rich structured logs:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout))

err := e.WrapWithMessage(someError, "additional context")

logger.Error("operation failed", e.SlogGroup(err))
```
Example output:
```json
{
  "level": "error",
  "msg": "operation failed",
  "error": {
    "error_text": "some error message",
    "stack_trace": [
      {
        "file": "/path/to/file.go",
        "function": "functionName",
        "line": 42,
        "message": "additional context"
      },
      {
        "file": "/path/to/other.go",
        "function": "otherFunction",
        "line": 10
      }
    ]
  }
}
```
This logs the error message along with a detailed stack trace and custom messages.

### JSON serialization
Wrapped errors implement `json.Marshaler`, producing structured JSON including error message and stack trace:
```go
jsonData, err := json.Marshal(wrappedErr)
if err != nil {
    // handle error
}
fmt.Println(string(jsonData))
```
Example output:
```json
{
  "error": "sql: no rows in result set",
  "stack_trace": [
    {
      "file": "/path/to/main.go",
      "function": "main",
      "line": 15
    },
    {
      "file": "/path/to/main.go",
      "function": "work",
      "line": 22
    },
    {
      "file": "/path/to/main.go",
      "function": "anotherWork",
      "line": 32,
      "message": "fetching user data failed"
    }
  ]
}
```

### Custom Fields Support
You can now attach structured key-value fields to wrapped errors for richer context in logs or serialized output.
Creating an error with fields:
```go
err := someOperation()
if err != nil {
    err = e.WrapWithFields(err,
        e.Field("user_id", 42),
        e.Field("operation", "database insert"),
    )
}
```
Example output (structured via `slog.Group` or JSON):
```json
{
  "error": {
    "error_text": "insert failed",
    "stack_trace": [...],
    "user_id": 42,
    "operation": "database insert"
  }
}
```

### Panic recovery
The package provides helpers for safe panic recovery with optional stack trace capture and structured handling.

```go
defer e.Recover(nil, func(err error) {
    log.Println("panic recovered:", err)
})
```
Or send to a channel in a goroutine:

```go
errCh := make(chan error, 1)
go func() {
    defer e.RecoverToChannel(nil, errCh)
    // potentially panicking code
}()
```
Example output (structured via `slog.Group` or JSON):
```json
{
  "error": "panic: runtime error: index out of range",
  "stack_trace": [
    {
      "file": "/app/service.go",
      "function": "handleRequest",
      "line": 87
    },
    {
      "file": "/app/controller.go",
      "function": "Serve",
      "line": 45
    }
  ]
}
```
Supports `Fatal` termination,`RecoverOnly` suppression, and optional `WithoutStack` mode.

## API
```go
func Wrap(err error) error
```
Wraps an error with a captured stack frame; returns nil if input is nil.

```go
func WrapWithMessage(err error, msg string) error
```
Wraps an error with a stack frame and attaches a custom message.

```go
func SlogGroup(err error) slog.Attr
```
Returns a `slog.Attr` containing the error message and stack trace as a `slog.Group`, suitable for structured logging.

```go
func WrapRecovered(opts *RecoverOpts, r any) error
```
Wraps a value returned from `recover()` into an `error`, by default capturing a filtered stack trace.

```go
func Recover(opts *RecoverOpts, callback func(error))
```
Intercepts and recovers from a `panic`.
Wraps the recovered value as an `error` and invokes the provided `callback`.

```go
func RecoverToChannel(opts *RecoverOpts, errChan chan<- error)
```
Alternative to `Recover` for use in goroutines.
Recovers from a `panic` and sends the error into the provided channel.
Returns a `slog.Attr` containing the error message and stack trace as a slog.Group, suitable for structured logging.

