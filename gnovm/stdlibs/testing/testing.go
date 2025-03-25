package testing

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (bool, string) {
	panic("only implemented in testing stdlibs")
}

func X_recoverWithStacktrace() (gno.TypedValue, string) {
	panic("only available in testing stdlibs")
}

func X_expectEmit(m *gno.Machine, expectedType string, expectedAttrs []string) bool {
	ctx := std.GetContext(m)
	events := ctx.EventLogger.Events()

	ll := len(events)
	if ll == 0 {
		return false
	}

	lastEvent, ok := events[ll-1].(std.GnoEvent)
	if !ok {
		return false
	}

	if lastEvent.Type != expectedType {
		return false
	}

	if len(expectedAttrs)%2 != 0 {
		return false
	}

	expectedAttrCount := len(expectedAttrs) / 2
	if len(lastEvent.Attributes) != expectedAttrCount {
		return false
	}

	for i := 0; i < len(expectedAttrs); i += 2 {
		expectedKey := expectedAttrs[i]
		expectedValue := expectedAttrs[i+1]
		found := false

		for _, attr := range lastEvent.Attributes {
			if attr.Key == expectedKey && attr.Value == expectedValue {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
