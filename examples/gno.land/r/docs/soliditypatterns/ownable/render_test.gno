package ownable

import (
	"strings"
	"testing"
)

// TestRenderKeepsAccurateStackWalkWarning guards the page's caller-identity
// claim. Inside a crossing function unsafe.PreviousRealm() and cur.Previous()
// return the same realm, so the warning must not call that shape an exploit.
// The reason for preferring cur.Previous() is that the helper does not require
// its caller to be crossing, not that the two disagree today.
func TestRenderKeepsAccurateStackWalkWarning(t *testing.T) {
	out := Render("")

	for _, want := range []string{
		"returns the same realm as `cur.Previous()`",
		"does not name the immediate caller",
		"Derive caller identity inside the crossing function",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render output must keep the qualified claim %q", want)
		}
	}

	for _, bad := range []string{
		"This is **insecure**",
		"launder the owner check",
		"outermost",
	} {
		if strings.Contains(out, bad) {
			t.Errorf("Render output overstates the stack-walk hazard: %q", bad)
		}
	}
}
