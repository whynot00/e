package e_test

import (
	"encoding/json"
	"errors"
	"strings"
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

func TestMarshalJSON_SingleFrameWithMessage(t *testing.T) {
	origErr := errors.New("sql: no rows in result set")
	wrapped := e.WrapWithMessage(origErr, "fetching user data failed")

	jsonBytes, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal error: %v", err)
	}

	jsonStr := string(jsonBytes)

	if !strings.Contains(jsonStr, `"error":"sql: no rows in result set"`) {
		t.Errorf("missing original error in JSON: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"message":"fetching user data failed"`) {
		t.Errorf("missing custom message in JSON: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"file":`) || !strings.Contains(jsonStr, `"function":`) {
		t.Errorf("missing stack frame info in JSON: %s", jsonStr)
	}
}

func TestMarshalJSON_MultipleFrames(t *testing.T) {
	baseErr := errors.New("unexpected EOF")
	// Симулируем несколько обёрток
	wrapped := e.WrapWithMessage(baseErr, "step 3 failed")
	wrapped = e.WrapWithMessage(wrapped, "step 2 failed")
	wrapped = e.Wrap(wrapped)

	jsonBytes, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal error: %v", err)
	}

	jsonStr := string(jsonBytes)

	if strings.Count(jsonStr, `"file":`) < 3 {
		t.Errorf("expected at least 3 stack frames, got: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"error":"unexpected EOF"`) {
		t.Errorf("missing base error in JSON: %s", jsonStr)
	}
}

func TestMarshalJSON_NilError(t *testing.T) {
	var err error = nil

	if wrapped := e.Wrap(err); wrapped != nil {
		t.Errorf("expected nil result from Wrap(nil), got non-nil")
	}
}
