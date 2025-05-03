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
