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

func X_expectEmit(m *gno.Machine, expectedType string, expectedAttrs []string, eventIndex int, partialMatch bool) bool {
	ctx := std.GetContext(m)
	events := ctx.EventLogger.Events()

	verifier := std.ExepectEventType(expectedType).(*std.EventVerifierImpl)
	verifier.WithEventIndex(eventIndex)
	if partialMatch {
		verifier.WithPartialMatch()
	}

	// process expected attributes in pairs (key-value)
	for i := 0; i < len(expectedAttrs); i += 2 {
		if i+1 >= len(expectedAttrs) {
			return false
		}
		verifier.WithAttribute(expectedAttrs[i], expectedAttrs[i+1])
	}

	gnoEvents := make([]std.GnoEvent, len(events))
	for i, event := range events {
		if e, ok := event.(std.GnoEvent); ok {
			gnoEvents[i] = e
		} else {
			return false
		}
	}

	// actual verification
	return verifier.Verify(gnoEvents)
}
