# Package e

`e` is a lightweight Go package providing enhanced error wrapping with stack trace capture and structured logging support using Go's `log/slog` package.

---

## Features

- Wrap errors with automatic stack trace capture
- Attach custom contextual messages to errors
- Compatible with Go standard library `errors.Is` and `errors.As`
- Produce structured logs with detailed error trace using `slog.Group`
- Simplifies function names in stack traces for better readability

---

## Installation

```bash
go get github.com/whynot00/e
```

---

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

Structured logging with slog
Integrate with log/slog for rich structured logs:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout))

err := e.WrapWithMessage(someError, "additional context")

logger.Error("operation failed", e.SlogGroup(err))
```
This logs the error message along with a detailed stack trace and custom messages.

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
Returns a `slog.Attr` containing the error message and stack trace as a slog.Group, suitable for structured logging.


## Example

```go
package main

import (
    "errors"
    "log/slog"
    "os"

    "github.com/yourusername/e"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout))

    baseErr := errors.New("sql: no rows in result set")
    err := e.WrapWithMessage(baseErr, "fetching user data failed")

    logger.Error("error occurred", e.SlogGroup(err))
}
```

### Output:
```json
{
  "level": "error",
  "msg": "error occurred",
  "error": {
    "error_text": "sql: no rows in result set",
    "stack_trace": [
      {
        "function": "main",
        "file": "/path/to/main.go",
        "line": 42,
        "message": "fetching user data failed"
      }
    ]
  }
}
```
