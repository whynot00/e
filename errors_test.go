package e_test

import (
	"errors"
	"testing"

	"github.com/whynot00/e"
)

func TestWrap_Nil(t *testing.T) {
	if e.Wrap(nil) != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

func TestWrapWithMessage_Nil(t *testing.T) {
	if e.WrapWithMessage(nil, "msg") != nil {
		t.Error("expected nil when wrapping nil error with message")
	}
}

func TestWrap_CapturesFrame(t *testing.T) {
	orig := errors.New("original error")
	wrapped := e.Wrap(orig)

	if wrapped == nil {
		t.Fatal("wrapped error is nil")
	}
	if wrapped.Error() != orig.Error() {
		t.Errorf("unexpected error message: got %q, want %q", wrapped.Error(), orig.Error())
	}
}

func TestWrapWithMessage_AttachesMessage(t *testing.T) {
	err := errors.New("db failure")
	wrapped := e.WrapWithMessage(err, "connecting to db")

	attr := e.SlogGroup(wrapped)

	group := attr.Value.Group()
	if group == nil {
		t.Fatal("expected slog.Group value")
	}

	var foundText, foundStack bool
	for _, g := range group {
		switch g.Key {
		case "error_text":
			val, ok := g.Value.Any().(string)
			if !ok {
				t.Fatal("error_text is not string")
			}
			if val == "db failure" {
				foundText = true
			}
		case "stack_trace":
			stack, ok := g.Value.Any().([]map[string]any)
			if !ok || len(stack) == 0 {
				t.Fatal("stack_trace is missing or empty")
			}
			for _, f := range stack {
				if msg, ok := f["message"]; ok && msg == "connecting to db" {
					foundStack = true
				}
			}
		}
	}

	if !foundText {
		t.Error("missing or incorrect error_text")
	}
	if !foundStack {
		t.Error("missing custom message in stack trace")
	}
}

func TestWrap_DoubleWrapAddsFrames(t *testing.T) {
	err := errors.New("root")
	w1 := e.WrapWithMessage(err, "first")
	w2 := e.WrapWithMessage(w1, "second")

	attr := e.SlogGroup(w2)
	group := attr.Value.Group()
	if group == nil {
		t.Fatal("expected slog.Group")
	}

	var frameCount int
	for _, g := range group {
		if g.Key == "stack_trace" {
			stack := g.Value.Any().([]map[string]any)
			frameCount = len(stack)
		}
	}

	if frameCount < 2 {
		t.Errorf("expected at least 2 frames, got %d", frameCount)
	}
}

func TestSlogGroup_PlainError(t *testing.T) {
	err := errors.New("simple error")
	attr := e.SlogGroup(err)

	group := attr.Value.Group()
	if group == nil {
		t.Fatal("expected slog.Group")
	}

	found := false
	for _, g := range group {
		if g.Key == "error_text" {
			val, ok := g.Value.Any().(string)
			if !ok {
				t.Fatal("error_text is not string")
			}
			if val == "simple error" {
				found = true
			}
		}
	}
	if !found {
		t.Error("plain error_text not found in slog group")
	}
}

func TestSlogGroup_Nil(t *testing.T) {
	attr := e.SlogGroup(nil)

	if attr.Key != "error" {
		t.Errorf("unexpected slog group key: got %s, want error", attr.Key)
	}
}

func TestWrapWithFields_Nil(t *testing.T) {
	if e.WrapWithFields(nil, e.Field("k", "v")) != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

func TestWrapWithFields_SingleField(t *testing.T) {
	err := errors.New("root")
	wrapped := e.WrapWithFields(err, e.Field("retry", 3))

	ew, ok := wrapped.(*e.ErrorWrapper)
	if !ok {
		t.Fatalf("want *ErrorWrapper, got %T", wrapped)
	}
	if ew.Error() != err.Error() {
		t.Errorf("message mismatch")
	}
	if v := ew.Fields().Get("retry"); v != 3 {
		t.Errorf("retry=%v, want 3", v)
	}
}

func TestWrapWithFields_MultipleFields(t *testing.T) {
	err := errors.New("root")
	wrapped := e.WrapWithFields(err,
		e.Field("retry", 3),
		e.Field("timeout", "5s"),
		e.Field("user_id", 42),
	)

	fs := wrapped.(*e.ErrorWrapper).Fields()
	if fs.Get("retry") != 3 || fs.Get("timeout") != "5s" || fs.Get("user_id") != 42 {
		t.Errorf("fields mismatch: %v", fs.List())
	}
}

func TestWrapWithFields_NoFields(t *testing.T) {
	err := errors.New("root")
	wrapped := e.WrapWithFields(err)

	if len(wrapped.(*e.ErrorWrapper).Fields().List()) != 0 {
		t.Error("expected zero fields")
	}
}

func TestWrapWithFields_PreservesOrder(t *testing.T) {
	err := errors.New("root")
	wrapped := e.WrapWithFields(err,
		e.Field("first", 1),
		e.Field("second", 2),
		e.Field("third", 3),
	)

	list := wrapped.(*e.ErrorWrapper).Fields().List()
	keys := make([]string, len(list))
	for i, kv := range list {
		keys[i] = kv.Key
	}
	want := []string{"first", "second", "third"}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("wrong order at %d: %q vs %q", i, keys[i], want[i])
		}
	}
}

func TestWrapWithFields_DoubleWrapAppendsFields(t *testing.T) {
	root := errors.New("root")
	w1 := e.WrapWithFields(root, e.Field("a", "outer"))
	w2 := e.WrapWithFields(w1, e.Field("b", "inner"))

	fs := w2.(*e.ErrorWrapper).Fields()
	if fs.Get("a") != "outer" || fs.Get("b") != "inner" {
		t.Errorf("fields after double wrap: %v", fs.List())
	}
}
