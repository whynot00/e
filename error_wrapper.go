package e

import (
	"encoding/json"
)

// Fields is an ordered collection of key–value pairs that can be attached
// to an ErrorWrapper for structured logging and serialization.
type Fields struct {
	list []fieldKV
}

// fieldKV is an internal helper that stores one key/value pair.
type fieldKV struct {
	Key   string
	Value any
}

// List returns an independent copy of all key-value pairs stored
// in Fields.  Modifications to the returned slice do not affect
// the original Fields instance.
func (f Fields) List() []fieldKV {
	cp := make([]fieldKV, len(f.list))
	copy(cp, f.list)
	return cp
}

// Get retrieves the value associated with the given key k.
// If the key does not exist, it returns nil.
func (f Fields) Get(k string) any {
	for _, kv := range f.list {
		if kv.Key == k {
			return kv.Value
		}
	}
	return nil
}

// Field returns a new Fields containing a single key/value pair.
// It is intended for use with WrapWithFields or as a starting point
// for chaining additional fields.
func Field(key string, value any) Fields {
	return Fields{list: []fieldKV{{Key: key, Value: value}}}
}

// ErrorWrapper wraps an underlying error with stack-trace frames
// and optional custom fields.
type ErrorWrapper struct {
	err    error
	frames []frame
	fields *Fields
}

// Error returns the underlying error message.
func (e *ErrorWrapper) Error() string { return e.err.Error() }

// Unwrap implements errors.Unwrap, allowing errors.Is / errors.As to work.
func (e *ErrorWrapper) Unwrap() error { return e.err }

// StackTrace returns a shallow copy of the captured stack frames.
func (e *ErrorWrapper) StackTrace() []frame {
	cp := make([]frame, len(e.frames))
	copy(cp, e.frames)
	return cp
}

// Fields returns a deep copy of the custom fields attached to the error.
// If no fields were attached, a zero value is returned.
func (e *ErrorWrapper) Fields() Fields {
	if e.fields == nil {
		return Fields{}
	}
	cp := Fields{list: make([]fieldKV, len(e.fields.list))}
	copy(cp.list, e.fields.list)
	return cp
}

// MarshalJSON implements json.Marshaler and outputs a single JSON object
// containing the original error text, stack trace and any custom fields.
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

	out := map[string]any{
		"error":       e.err.Error(),
		"stack_trace": stack,
	}

	if e.fields != nil && len(e.fields.list) > 0 {
		for _, kv := range e.fields.list {
			out[kv.Key] = kv.Value
		}
	}

	return json.Marshal(out)
}

// frameJSON is the public representation of a single frame in the stack trace.
type frameJSON struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
	Message  string `json:"message,omitempty"`
}
