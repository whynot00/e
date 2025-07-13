package e

import "testing"

func TestIsInternalFrame(t *testing.T) {
	tests := []struct {
		fn   string
		want bool
	}{
		{"runtime.main", true},
		{"encoding/json.Marshal", true},
		{"log/slog.newHandler", true},
		{"github.com/user/project/pkg/e.Recover", true},
		{"github.com/user/project/pkg/e.SlogGroup", true},
		{"main.handleRequest", false},
	}

	for _, tt := range tests {
		got := isInternalFrame(tt.fn)
		if got != tt.want {
			t.Errorf("isInternalFrame(%q) = %v; want %v", tt.fn, got, tt.want)
		}
	}
}

func TestSimplifyFuncName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github.com/user/project/pkg/module.Func", "Func"},
		{"project/module.Func", "Func"},
		{"Func", "Func"},
		{"github.com/user/project/pkg/module.(*Type).Method", "Method"},
	}

	for _, tt := range tests {
		got := simplifyFuncName(tt.input)
		if got != tt.want {
			t.Errorf("simplifyFuncName(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}
