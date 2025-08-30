package sloghelpers

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type Error struct {
	err error

	args map[string]any
}

func (o Error) Error() string {
	if o.err != nil {
		return o.err.Error()
	}

	return fmt.Sprintf("%v", o.args)
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

	o.args["stack"] = stackTrace
}

func (o *Error) mapToSlice() []any {
	var res []any

	for k, v := range o.args {
		res = append(res, k, v)
	}

	return res
}

func NewError(err error, keyVals ...any) error {
	args := make(map[string]any)

	var e *Error
	addStack := true
	if errors.As(err, &e) {
		addStack = false
		for k, v := range e.args {
			args[k] = v
		}
	}

	for i, l := 0, len(keyVals); i < l; i++ {
		k, ok := keyVals[i].(string)

		if !ok || i+1 >= l {
			continue
		}
		i++
		args[k] = keyVals[i]
	}

	res := &Error{
		err:  err,
		args: args,
	}

	if addStack {
		res.addStackTrace()
	}

	return res
}

func (o *Error) GetArgs(args ...any) []any {
	return append(o.mapToSlice(), args...)
}

func (o *Error) HasArg(name string) bool {
	_, ok := o.args[name]
	return ok
}

func (o *Error) GetArg(name string) any {
	if v, ok := o.args[name]; ok {
		return v
	}

	return nil
}
