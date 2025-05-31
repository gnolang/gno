package uassert

import "strings"

func fail(t TestingT, customMsgs []string, failureMessage string, args ...any) bool {
	customMsg := ""
	if len(customMsgs) > 0 {
		customMsg = strings.Join(customMsgs, " ")
	}
	if customMsg != "" {
		failureMessage += " - " + customMsg
	}
	t.Errorf(failureMessage, args...)
	return false
}

func checkDidPanic(f any) (didPanic bool, message string) {
	didPanic = true
	defer func() {
		r := recover()

		if r == nil {
			message = "nil"
			return
		}

		err, ok := r.(error)
		if ok {
			message = err.Error()
			return
		}

		errStr, ok := r.(string)
		if ok {
			message = errStr
			return
		}

		message = "recover: unsupported type"
	}()
	switch f := f.(type) {
	case func():
		f()
	case func(realm):
		f(cross)
	default:
		panic("f must be of type func() or func(realm)")
	}
	didPanic = false
	return
}
