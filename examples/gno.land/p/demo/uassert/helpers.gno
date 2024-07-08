package uassert

import "strings"

func fail(t TestingT, customMsgs []string, failureMessage string, args ...interface{}) bool {
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

func autofail(t TestingT, success bool, customMsgs []string, failureMessage string, args ...interface{}) bool {
	if success {
		return true
	}
	return fail(t, customMsgs, failureMessage, args...)
}

func checkDidPanic(f func()) (didPanic bool, message string) {
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
	f()
	didPanic = false
	return
}

func compare(t TestingT, e1, e2 interface{}, op string) bool {
	t.Helper()
	switch v1 := e1.(type) {
	case int:
		v2 := e2.(int)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case int8:
		v2 := e2.(int8)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case int16:
		v2 := e2.(int16)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case int32:
		v2 := e2.(int32)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case int64:
		v2 := e2.(int64)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case uint:
		v2 := e2.(uint)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case uint8:
		v2 := e2.(uint8)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case uint16:
		v2 := e2.(uint16)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case uint32:
		v2 := e2.(uint32)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case uint64:
		v2 := e2.(uint64)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case float32:
		v2 := e2.(float32)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	case float64:
		v2 := e2.(float64)
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	default:
		return fail(t, []string{}, "uassert.compare: unsupported types or operation")
	}
	return false
}
