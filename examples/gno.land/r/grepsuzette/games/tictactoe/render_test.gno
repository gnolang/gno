package tictactoe

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	for _, v := range []struct {
		path          string
		expectedRepr  string
		expectedX     int
		expectedY     int
		expectedDebug bool
	}{
		{"", "---------", -1, -1},
		{"move=a1", "---------", 0, 0},
		{"move=a2", "---------", 0, 1},
		{"move=a3", "---------", 0, 2},
		{"move=b1", "---------", 1, 0},
		{"move=b2", "---------", 1, 1},
		{"move=b3", "---------", 1, 2},
		{"move=c1", "---------", 2, 0},
		{"move=c2", "---------", 2, 1},
		{"move=c3", "---------", 2, 2},
		{"move=c1&debug", "---------", 2, 0, true},
		{"&", "---------", -1, -1},
		{"state=-X--OX-O-&move=b2", "-X--OX-O-", 1, 1},
		{"move=b2?state=-X--OX-O-", "-X--OX-O-", 1, 1},
		{"state=XOXX-O-O-&move=c1", "XOXX-O-O-", 2, 0},
		{"move=c1?state=XOXX-O-O-", "XOXX-O-O-", 2, 0},
		{"debug", "---------", -1, -1, true},
	} {
		g, x, y, debug := parsePath(v.path)
		repr := g.ToRepr()
		if repr != v.expectedRepr || x != v.expectedX || y != v.expectedY {
			t.Errorf(
				"failed to parsePath for '%s', expected x=%d y=%d repr=%s, got x=%d x=%d repr=%s",
				v.path,
				v.expectedX, v.expectedY, v.expectedRepr,
				x, y, repr,
			)
		}
	}
}
