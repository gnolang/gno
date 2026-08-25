package reentrancy

import (
	"strings"
	"testing"
)

// TestRenderDoesNotOverstateSafety guards the page's central claim. The realm
// has no state and no coin path of its own; what it ships is the explanation,
// so the explanation is what a test can protect. Coin transfers running no
// recipient code removes Solidity's automatic-callback vector, but a realm that
// calls out through an interface or callback can still be re-entered, and
// earlier revisions of this page said otherwise in several places.
func TestRenderDoesNotOverstateSafety(t *testing.T) {
	out := Render("")

	for _, want := range []string{
		"does **not** make reentrancy impossible",
		"Checks-Effects-Interactions",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render output must keep the qualified claim %q", want)
		}
	}

	for _, bad := range []string{
		"no reentrancy possible",
		"Reentrancy attacks are not possible",
		"Gno prevents reentrancy attacks",
		"reentrancy isn't a concern",
	} {
		if strings.Contains(out, bad) {
			t.Errorf("Render output claims reentrancy is impossible: %q", bad)
		}
	}
}
