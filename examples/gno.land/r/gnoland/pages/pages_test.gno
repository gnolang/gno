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

<div class='columns-3'><div>

## [Events](/r/gnoland/pages:p/events)
**[Learn More](/r/gnoland/pages:p/events)**

</div><div>

## [Game of Realms](/r/gnoland/pages:p/gor)
**[Learn More](/r/gnoland/pages:p/gor)**

</div><div>

## [Getting Started](/r/gnoland/pages:p/start)
**[Learn More](/r/gnoland/pages:p/start)**

</div><div>

## [Tokenomics](/r/gnoland/pages:p/tokenomics)
**[Learn More](/r/gnoland/pages:p/tokenomics)**

</div></div>
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
