package e

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type RecoverOpts struct {
	Message      string
	WithoutStack bool
	Context      context.Context
	RecoverOnly  bool
	Fatal        bool
}

func Recover(opts *RecoverOpts, callback func(error)) {

	if r := recover(); r != nil {

		msg := getPanicString(r)

		var stack []frame

		if !opts.WithoutStack {
			stack = getPanicTrace()
		}

		if !opts.RecoverOnly {

			callback(&ErrorWrapper{err: errors.New(msg), frames: stack})
		}
	}

}

func getPanicString(r any) string {

	var msg string
	switch v := r.(type) {
	case error:
		msg = v.Error()
	default:
		msg = fmt.Sprintf("%v", v)
	}

	return msg
}

func getPanicTrace() []frame {

	pcs := make([]uintptr, 10)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	wrapperFrames := make([]frame, 0)
	for {
		fr, more := frames.Next()
		if !more {
			break
		}

		if strings.HasPrefix(fr.Function, "runtime.") {
			continue
		}

		wrapperFrames = append(wrapperFrames, frame{funcName: simplifyFuncName(fr.Function), file: fr.File, line: fr.Line})
	}

	return wrapperFrames
}
