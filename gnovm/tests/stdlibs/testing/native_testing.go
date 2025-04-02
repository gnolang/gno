package testing

import (
	"regexp"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	return time.Now().UnixNano()
}

func X_matchString(pat, str string) (bool, string) {
	matchRe, err := regexp.Compile(pat)
	if err != nil {
		return false, err.Error()
	}
	return matchRe.MatchString(str), ""
}

func X_recoverWithStacktrace(m *gnolang.Machine) (gnolang.TypedValue, string) {
	//fmt.Println("---native, recover with stacktrace...")
	exception := m.Recover()
	//fmt.Println("recovered from panic:", exception)
	if exception == nil {
		return gnolang.TypedValue{}, ""
	}
	exception.Recovered = true
	return exception.Value, exception.Stacktrace.String()
}
