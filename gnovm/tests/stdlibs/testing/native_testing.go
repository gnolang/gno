package testing

import (
	"regexp"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	return time.Now().UnixNano()
}

func X_matchString(pat, str string) (result bool, err error) {
	var matchRe *regexp.Regexp
	if matchRe, err = regexp.Compile(pat); err != nil {
		return
	}
	return matchRe.MatchString(str), nil
}

func X_recoverWithStacktrace(m *gnolang.Machine) (gnolang.TypedValue, string) {
	exception := m.Recover()
	if exception == nil {
		return gnolang.TypedValue{}, ""
	}
	return exception.Value, exception.Stacktrace.String()
}
