// Package e provides a lightweight error wrapper with stack trace
// and structured logging support via slog.Group and JSON serialization.
package e

import (
	"errors"
	"log/slog"
	"runtime"
)

// Wrap returns an ErrorWrapper with the current call site.
// If the error is already wrapped, the new frame is prepended.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, "", nil)
}

// WrapWithMessage is like Wrap but also attaches a custom message to the frame.
func WrapWithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return wrapWithSkip(err, 2, msg, nil)
}

func WrapWithFields(err error, fields ...Fields) error {
	if err == nil {
		return nil
	}

	merged := Fields{}
	for _, f := range fields {
		merged.list = append(merged.list, f.list...)
	}

	return wrapWithSkip(err, 2, "", &merged)
}

// wrapWithSkip captures a stack frame at the given depth.
func wrapWithSkip(err error, skip int, msg string, flds *Fields) *ErrorWrapper {
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

		if flds != nil && len(flds.list) > 0 {

			if ew.fields == nil {
				ew.fields = &Fields{}
			}

			ew.fields.list = append(ew.fields.list, flds.list...)
		}

		return ew
	}

	return &ErrorWrapper{
		err:    err,
		frames: []frame{fr},
		fields: flds,
	}
}

// SlogGroup returns a slog.Group containing structured fields with error and stack trace.
func SlogGroup(err error) slog.Attr {

	return slogGroup(err, "error")
}

// SlogGroup returns a slog.Group containing structured fields with custom name, error and stack trace.
func SlogGroupNamed(err error, name string) slog.Attr {

	return slogGroup(err, name)
}

func slogGroup(err error, name string) slog.Attr {

	if err == nil {
		return slog.Group(name,
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

	attrs := []slog.Attr{
		slog.String("error_text", baseErr.Error()),
	}

	if ew != nil && len(ew.frames) > 0 {
		attrs = append(attrs, slog.Any("stack_trace", frames))
	}

	if ew != nil && ew.fields != nil && len(ew.fields.list) > 0 {
		for _, kv := range ew.fields.list {
			attrs = append(attrs, slog.Any(kv.Key, kv.Value))
		}
	}

	anyAttrs := make([]any, len(attrs))
	for i, a := range attrs {
		anyAttrs[i] = a
	}
	return slog.Group("error", anyAttrs...)
}
