package gnopages

import (
	"std"
	"strings"
	"testing"
)

func TestPackage(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"))

	author := std.GetOrigCaller()

	// by default, lorem ipsum posts
	{
		got := Render("")
		expected := `
# Gnoland's Pages

## [&#9656; Events](/r/gnoland/pages:p/events)

## [&#9656; Game of Realms](/r/gnoland/pages:p/gor)

## [&#9656; Getting Started](/r/gnoland/pages:p/start)

## [&#9656; Tokenomics](/r/gnoland/pages:p/tokenomics)
`
		assertMDEquals(t, got, expected)
	}
}

func assertMDEquals(t *testing.T, got, expected string) {
	t.Helper()
	expected = strings.TrimSpace(expected)
	got = strings.TrimSpace(got)
	if expected != got {
		t.Errorf("invalid render output.\nexpected %q.\ngot      %q.", expected, got)
	}
}
