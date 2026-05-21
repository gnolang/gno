package testing

import (
	"regexp"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	return time.Now().UnixNano()
}

func X_cycleCount(m *gnolang.Machine) int64 {
	return m.Cycles
}

func X_gasConsumed(m *gnolang.Machine) int64 {
	if m.GasMeter == nil {
		return 0
	}
	return m.GasMeter.GasConsumed()
}

func X_allocBytes(m *gnolang.Machine) int64 {
	if m.Alloc == nil {
		return 0
	}
	return m.Alloc.TotalAllocBytes()
}

func X_allocCount(m *gnolang.Machine) int64 {
	if m.Alloc == nil {
		return 0
	}
	return m.Alloc.NumAllocs()
}

func X_matchString(pat, str string) (bool, string) {
	matchRe, err := regexp.Compile(pat)
	if err != nil {
		return false, err.Error()
	}
	return matchRe.MatchString(str), ""
}

func X_recoverWithStacktrace(m *gnolang.Machine) (gnolang.TypedValue, string) {
	exception := m.Recover()
	if exception == nil {
		return gnolang.TypedValue{}, ""
	}
	return exception.Value, exception.Stacktrace.String()
}
