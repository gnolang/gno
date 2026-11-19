package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadyWatchStartsTheBudgetOnTheMarker: the budget must start when the
// child reports ready, wherever on stdout that report lands. A child that
// writes something else first is still compiling once it has said so, and
// leaving it under the prepare budget would time the compile against the
// wrong bound.
func TestReadyWatchStartsTheBudgetOnTheMarker(t *testing.T) {
	cases := map[string]struct {
		writes    []string
		wantReady int
	}{
		"marker alone":                   {[]string{childReadyMarker + "\n"}, 1},
		"marker split across writes":     {[]string{childReadyMarker[:5], childReadyMarker[5:] + "\n"}, 1},
		"a stray line before the marker": {[]string{"warning: something\n", childReadyMarker + "\n"}, 1},
		"an overlong line before it":     {[]string{strings.Repeat("x", 2*maxReadyLine) + "\n", childReadyMarker + "\n"}, 1},
		"marker then more output":        {[]string{childReadyMarker + "\n", childReadyMarker + "\n", "more\n"}, 1},
		"never ready":                    {[]string{"nothing\n", "here\n"}, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ready := 0
			w := &readyWatch{onReady: func() { ready++ }}
			for _, s := range tc.writes {
				n, err := w.Write([]byte(s))
				require.NoError(t, err)
				require.Equal(t, len(s), n, "a short write would make exec report the child as failed")
			}
			assert.Equal(t, tc.wantReady, ready, "the budget must start exactly once, on the marker")
		})
	}
}
