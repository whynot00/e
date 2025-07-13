package e

import (
	"errors"
	"fmt"
	"os"
)

// RecoverOpts defines behavior for panic recovery.
//
// It controls whether to include a stack trace, invoke a callback,
// or exit the process after recovering from a panic.
type RecoverOpts struct {
	// WithoutStack disables capturing stack trace when recovering from panic.
	// Useful when you want minimal error information or you're handling tracing/logging manually.
	WithoutStack bool

	// RecoverOnly suppresses all side effects (callback or channel send).
	// The panic is still recovered, but no error will be propagated.
	RecoverOnly bool

	// Fatal forces the application to terminate with exit code 1 after recovering the panic.
	// Useful in CLI tools, workers, or when panic is considered unrecoverable.
	Fatal bool
}

// WrapRecovered wraps the recovered panic value `r` into an error with optional stack trace.
//
// It is intended to be used internally by recovery helpers, but can also be reused in custom handlers.
func WrapRecovered(opts *RecoverOpts, r any) error {
	message := formatPanicMessage(r)

	var stack []frame
	if opts == nil || !opts.WithoutStack {
		stack = captureStackTrace()
	}

	return &ErrorWrapper{
		err:    errors.New(message),
		frames: stack,
	}
}

// Recover is a general-purpose recovery helper.
// It must be used with `defer`:
//
//	defer e.Recover(opts, func(err error) { ... })
//
// If a panic occurs, it wraps the value in an error and calls the provided callback.
// If Fatal is true, the process exits after the callback is executed.
func Recover(opts *RecoverOpts, callback func(error)) {
	if r := recover(); r != nil {
		err := WrapRecovered(opts, r)

		if opts == nil || !opts.RecoverOnly {
			callback(err)
		}

		if opts != nil && opts.Fatal {
			os.Exit(1)
		}
	}
}

// RecoverToChannel is a recovery helper for use in goroutines or workers.
// Instead of calling a callback, it sends the recovered error into the provided channel.
//
// Use it with `defer` inside goroutines:
//
//	go func() {
//	    defer e.RecoverToChannel(opts, errChan)
//	    ...
//	}()
func RecoverToChannel(opts *RecoverOpts, errChan chan<- error) {
	if r := recover(); r != nil {
		err := WrapRecovered(opts, r)

		if opts == nil || !opts.RecoverOnly {
			// Use select to avoid panic if channel is full (e.g., buffered without reader)
			select {
			case errChan <- err:
			default:
				// Optionally: log or ignore overflow
			}
		}

		if opts != nil && opts.Fatal {
			os.Exit(1)
		}
	}
}

// formatPanicMessage converts any recovered panic value into a readable error message.
func formatPanicMessage(r any) string {
	switch v := r.(type) {
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}
