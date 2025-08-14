package sloghelpers

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

type Error struct {
	err error

	mu    sync.RWMutex
	attrs map[string]any
}

func (o *Error) Error() string {
	if o.err != nil {
		return o.err.Error()
	}

	return fmt.Sprintf("%v", o.attrs)
}

func (o *Error) Unwrap() error {
	return o.err
}

func (o *Error) addStackTrace() {
	sbuff := make([]uintptr, 50)
	length := runtime.Callers(3, sbuff[:])
	stack := sbuff[:length]

	frames := runtime.CallersFrames(stack)
	var stackTrace []string
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			continue
		}

		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d - %s", frame.File, frame.Line, frame.Function))

		if !more {
			break
		}
	}

	o.attrs["stack"] = stackTrace
}

func (o *Error) mapToSlice() []any {
	var res []any

	o.mu.RLock()
	defer o.mu.RUnlock()

	for k, v := range o.attrs {
		res = append(res, k, v)
	}

	return res
}

func NewError(err error, keyVals ...any) error {
	attrs := make(map[string]any)

	var e *Error
	addStack := true
	if errors.As(err, &e) {
		addStack = false
		e.mu.Lock()
		for k, v := range e.attrs {
			attrs[k] = v
		}

		e.mu.Unlock()
	}

	for i, l := 0, len(attrs); i < l; i++ {
		k, ok := keyVals[i].(string)
		if !ok {
			continue
		}
		i++
		attrs[k] = keyVals[i]
	}

	res := &Error{
		err:   err,
		attrs: attrs,
	}

	if addStack {
		res.addStackTrace()
	}

	return res
}

func GetArgs(err error, attrs ...any) []any {
	var e *Error

	if errors.As(err, &e) {
		return append(e.mapToSlice(), attrs)
	}

	return attrs
}
