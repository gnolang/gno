package testing

import (
	"regexp"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	return time.Now().UnixNano()
}

func X_matchString(pat, str string) bool {
	matchRe, err := regexp.Compile(pat)
	if err != nil {
		panic(err)
	}
	return matchRe.MatchString(str)
}

func X_verifyRegex(pat string) (ok bool) {
	_, err := regexp.Compile(pat)
	return err == nil
}

func X_recoverWithStacktrace(m *gnolang.Machine) (gnolang.TypedValue, string) {
	exception := m.Recover()
	if exception == nil {
		return gnolang.TypedValue{}, ""
	}
	return exception.Value, exception.Stacktrace.String()
}
