package e_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/whynot00/e"
)

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

func TestMarshalJSON_SimpleError(t *testing.T) {
	err := e.Wrap(errors.New("basic error"))

	data, errMarshal := json.Marshal(err)
	if errMarshal != nil {
		t.Fatalf("unexpected marshal error: %v", errMarshal)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if jsonMap["error"] != "basic error" {
		t.Errorf("expected error message to be 'basic error', got: %v", jsonMap["error"])
	}

	stack, ok := jsonMap["stack_trace"].([]interface{})
	if !ok || len(stack) == 0 {
		t.Errorf("expected non-empty stack_trace, got: %v", jsonMap["stack_trace"])
	}
}

func TestMarshalJSON_WithMessage(t *testing.T) {
	baseErr := errors.New("db connection failed")
	wrapped := e.WrapWithMessage(baseErr, "initialization failed")

	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result struct {
		Error      string `json:"error"`
		StackTrace []struct {
			File     string `json:"file"`
			Function string `json:"function"`
			Line     int    `json:"line"`
			Message  string `json:"message"`
		} `json:"stack_trace"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result.Error != "db connection failed" {
		t.Errorf("unexpected error text: %v", result.Error)
	}

	foundMsg := false
	for _, f := range result.StackTrace {
		if strings.Contains(f.Message, "initialization failed") {
			foundMsg = true
			break
		}
	}

	if !foundMsg {
		t.Errorf("expected to find custom message 'initialization failed' in stack trace")
	}
}
