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

func compare(t TestingT, e1, e2 interface{}) int {
	t.Helper()
	switch v1 := e1.(type) {
	case int:
		v2 := e2.(int)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case int8:
		v2 := e2.(int8)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case int16:
		v2 := e2.(int16)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case int32:
		v2 := e2.(int32)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case int64:
		v2 := e2.(int64)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case uint:
		v2 := e2.(uint)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case uint8:
		v2 := e2.(uint8)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case uint16:
		v2 := e2.(uint16)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case uint32:
		v2 := e2.(uint32)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case uint64:
		v2 := e2.(uint64)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case float32:
		v2 := e2.(float32)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case float64:
		v2 := e2.(float64)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	case string:
		v2 := e2.(string)
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	default:
		fail(t, []string{}, "compare: unsupported type")
		return 0
	}
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	switch v := object.(type) {
	case *int, *int8, *int16, *int32, *int64:
		return v == nil
	case *uint, *uint8, *uint16, *uint32, *uint64:
		return v == nil
	case *float32, *float64:
		return v == nil
	case *string:
		return v == nil
	case *bool:
		return v == nil
	case *[]int, *[]string, *[]float64, *[]bool:
		return v == nil
	case *map[string]int, *map[string]string, *map[string]bool:
		return v == nil
	case *func():
		return v == nil
	case []int:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []bool:
		return len(v) == 0
	case map[string]int:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	case map[string]bool:
		return len(v) == 0
	case func():
		return v == nil
	default:
		return false
	}
}
