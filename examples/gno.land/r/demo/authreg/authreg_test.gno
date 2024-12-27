package authreg

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
)

func TestEntityID(t *testing.T) {
	cases := []struct {
		name         string
		provider     string
		subPath      string
		res          string
		panicMessage string
	}{
		{
			name:     "good",
			provider: "alice",
			subPath:  "savings",
			res:      "/alice/savings",
		},
		{
			name:         "mal_backtrack",
			provider:     "eve",
			subPath:      "../alice/savings",
			panicMessage: `invalid sub-path "../alice/savings"`,
		},
		{
			name:     "mal_backtrack_dumb",
			provider: "eve",
			subPath:  "/../alice/savings",
			res:      "/eve/alice/savings",
		},
		{
			name:         "mal_backtrack_hidden",
			provider:     "eve",
			subPath:      "todobien/very/deep/../../../../alice/savings",
			panicMessage: `invalid sub-path "../alice/savings"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func() {
				res := entityID(tc.provider, tc.subPath)
				urequire.Equal(t, tc.res, res)
			}
			if tc.panicMessage != "" {
				urequire.PanicsWithMessage(t, tc.panicMessage, run)
			} else {
				urequire.NotPanics(t, run)
			}
		})
	}
}

type eveAuthToken struct {
	source std.Realm
}

func (t *eveAuthToken) Source() std.Realm {
	return t.source
}
