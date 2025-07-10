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
