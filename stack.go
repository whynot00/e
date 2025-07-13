package e

import (
	"runtime"
	"strings"
)

// frame represents a single captured stack frame in the trace.
type frame struct {
	funcName string
	file     string
	line     int
	message  string
}

// simplifyFuncName trims package and receiver prefixes from a function name.
func simplifyFuncName(fn string) string {
	if i := strings.LastIndex(fn, "/"); i != -1 {
		fn = fn[i+1:]
	}
	if i := strings.LastIndex(fn, "."); i != -1 {
		return fn[i+1:]
	}
	return fn
}

// captureStackTrace collects and filters the current call stack,
// excluding frames from the Go runtime and known internal packages.
func captureStackTrace() []frame {
	const skipFrames = 3 // skip: Callers → captureStackTrace → WrapRecovered → Recover
	const maxDepth = 32

	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skipFrames, pcs)
	rawFrames := runtime.CallersFrames(pcs[:n])

	var trace []frame

	for {
		fr, more := rawFrames.Next()
		if !more {
			break
		}

		if isInternalFrame(fr.Function) {
			continue
		}

		trace = append(trace, frame{
			funcName: simplifyFuncName(fr.Function),
			file:     fr.File,
			line:     fr.Line,
		})
	}

	return trace
}

// isInternalFrame filters out frames from standard library and internal infrastructure.
//
// This avoids polluting stack traces with frames like `runtime.*`, `log/slog`, `encoding/json`,
// or your own wrapper utilities (Recover, SlogGroup, etc.).
func isInternalFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") ||
		strings.Contains(function, "/log/slog.") ||
		strings.Contains(function, "log/slog.") ||
		strings.Contains(function, "encoding/json.") ||
		strings.Contains(function, ".Recover") ||
		strings.Contains(function, ".SlogGroup")
}
