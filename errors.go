// Package e provides a lightweight error wrapper with stack trace
// and structured logging support via slog.Group.
package e

import (
	"errors"
	"log/slog"
	"runtime"
	"strings"
)

// frame represents a single stack frame in the error trace.
type frame struct {
	funcName string
	file     string
	line     int
	message  string
}

// ErrorWrapper wraps an error with a stack trace.
// It supports error unwrapping and slog-compatible structured logging.
type ErrorWrapper struct {
	err    error
	frames []frame
}

// Error returns the original error message.
func (e *ErrorWrapper) Error() string {
	return e.err.Error()
}

// Unwrap allows ErrorWrapper to be compatible with errors.Is and errors.As.
func (e *ErrorWrapper) Unwrap() error {
	return e.err
}

// Wrap returns a new ErrorWrapper with a captured stack frame.
// Returns nil if the input error is nil.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, "")
}

// WrapWithMessage returns a new ErrorWrapper with a stack frame and a custom message.
// Returns nil if the input error is nil.
func WrapWithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, msg)
}

// wrapWithSkip captures a stack frame at the given skip level.
// If the error is already an ErrorWrapper, it prepends the new frame.
func wrapWithSkip(err error, skip int, msg string) *ErrorWrapper {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file, line = "unknown", 0
	}
	funcName := runtime.FuncForPC(pc).Name()

	newFrame := frame{
		funcName: simplifyFuncName(funcName),
		file:     file,
		line:     line,
		message:  msg,
	}

	var ew *ErrorWrapper
	if errors.As(err, &ew) {
		ew.frames = append([]frame{newFrame}, ew.frames...)
		return ew
	}

	return &ErrorWrapper{
		err:    err,
		frames: []frame{newFrame},
	}
}

// SlogGroup returns a structured slog.Attr with error message and full stack trace.
// It inserts an additional frame to reflect the log call site.
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
	var baseErr error = err
	var frames []map[string]any

	if errors.As(err, &ew) {
		baseErr = ew.err
		for _, f := range ew.frames {
			frameMap := map[string]any{
				"function": f.funcName,
				"file":     f.file,
				"line":     f.line,
			}
			if f.message != "" {
				frameMap["message"] = f.message
			}
			frames = append(frames, frameMap)
		}
	} else {
		frames = append(frames, map[string]any{
			"message": baseErr.Error(),
		})
	}

	if pc, file, line, ok := runtime.Caller(1); ok {
		funcName := simplifyFuncName(runtime.FuncForPC(pc).Name())
		logFrame := map[string]any{
			"function": funcName,
			"file":     file,
			"line":     line,
		}
		frames = append([]map[string]any{logFrame}, frames...)
	}

	return slog.Group("error",
		slog.String("error_text", baseErr.Error()),
		slog.Any("stack_trace", frames),
	)
}

// simplifyFuncName trims package prefixes from a function name for readability.
func simplifyFuncName(fn string) string {
	if i := strings.LastIndex(fn, "/"); i != -1 {
		fn = fn[i+1:]
	}
	if i := strings.Index(fn, "."); i != -1 {
		return fn[i+1:]
	}
	return fn
}
