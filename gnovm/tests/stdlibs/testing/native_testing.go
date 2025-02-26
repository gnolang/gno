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
	if len(m.Exceptions) == 0 {
		return gnolang.TypedValue{}, ""
	}

	// If the exception is out of scope, this recover can't help; return nil.
	if m.PanicScope <= m.DeferPanicScope {
		return gnolang.TypedValue{}, ""
	}

	exception := &m.Exceptions[len(m.Exceptions)-1]

	// If the frame the exception occurred in is not popped, it's possible that
	// the exception is still in scope and can be recovered.
	if !exception.Frame.Popped {
		// If the frame is not the current frame, the exception is not in scope; return nil.
		// This retrieves the second most recent call frame because the first most recent
		// is the call to recover itself.
		if frame := m.LastCallFrame(2); frame == nil || (frame != nil && frame != exception.Frame) {
			return gnolang.TypedValue{}, ""
		}
	}

	m.Exceptions = nil

	return exception.Value, exception.Stacktrace.String()
}
