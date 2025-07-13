// Package e provides a lightweight error wrapper with stack trace
// and structured logging support via slog.Group and JSON serialization.
package e

import (
	"encoding/json"
	"errors"
	"log/slog"
	"runtime"
)

// ErrorWrapper wraps an error with its stack trace frames and optional messages.
// It supports error unwrapping, structured logging, and JSON serialization.
type ErrorWrapper struct {
	err    error
	frames []frame
}

// Wrap returns an ErrorWrapper with the current call site.
// If the error is already wrapped, the new frame is prepended.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, "")
}

// WrapWithMessage is like Wrap but also attaches a custom message to the frame.
func WrapWithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, msg)
}

// wrapWithSkip captures a stack frame at the given depth.
func wrapWithSkip(err error, skip int, msg string) *ErrorWrapper {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file, line = "unknown", 0
	}
	funcName := runtime.FuncForPC(pc).Name()

	fr := frame{
		funcName: simplifyFuncName(funcName),
		file:     file,
		line:     line,
		message:  msg,
	}

	var ew *ErrorWrapper
	if errors.As(err, &ew) {
		ew.frames = append([]frame{fr}, ew.frames...)
		return ew
	}

	return &ErrorWrapper{
		err:    err,
		frames: []frame{fr},
	}
}

// Error returns the wrapped error’s message.
func (e *ErrorWrapper) Error() string {
	return e.err.Error()
}

// Unwrap makes ErrorWrapper compatible with errors.Is and errors.As.
func (e *ErrorWrapper) Unwrap() error {
	return e.err
}

// StackTrace returns a copy of the captured stack frames.
func (e *ErrorWrapper) StackTrace() []frame {
	return e.frames
}

// SlogGroup returns a slog.Group containing structured fields with error and stack trace.
func SlogGroup(err error) slog.Attr {
	if err == nil {
		return slog.Group("error",
			slog.String("error_text", "nil"),
			slog.Any("stack_trace", []map[string]any{
				{"message": "nil"},
			}),
		)
	}

	var ew *ErrorWrapper
	var baseErr = err
	var frames []map[string]any

	if errors.As(err, &ew) {
		baseErr = ew.err
		for _, f := range ew.frames {
			entry := map[string]any{
				"function": f.funcName,
				"file":     f.file,
				"line":     f.line,
			}
			if f.message != "" {
				entry["message"] = f.message
			}
			frames = append(frames, entry)
		}
	} else {
		frames = append(frames, map[string]any{
			"message": baseErr.Error(),
		})
	}

	if ew == nil || ew.frames == nil {
		return slog.Group("error",
			slog.String("error_text", baseErr.Error()),
		)
	}

	return slog.Group("error",
		slog.String("error_text", baseErr.Error()),
		slog.Any("stack_trace", frames),
	)
}

// MarshalJSON allows ErrorWrapper to be serialized into structured JSON with full trace.
func (e *ErrorWrapper) MarshalJSON() ([]byte, error) {
	stack := make([]frameJSON, 0, len(e.frames))
	for _, f := range e.frames {
		stack = append(stack, frameJSON{
			File:     f.file,
			Function: f.funcName,
			Line:     f.line,
			Message:  f.message,
		})
	}

	return json.Marshal(errorJSON{
		Error:      e.err.Error(),
		StackTrace: stack,
	})
}

// frameJSON is the exported form of a stack frame used in JSON serialization.
type frameJSON struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
	Message  string `json:"message,omitempty"`
}

// errorJSON is the root structure returned when marshaling ErrorWrapper.
type errorJSON struct {
	Error      string      `json:"error"`
	StackTrace []frameJSON `json:"stack_trace"`
}
