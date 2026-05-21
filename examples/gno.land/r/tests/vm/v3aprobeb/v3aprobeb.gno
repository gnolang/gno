// Package v3aprobeb is an intermediate-hop realm used to test the v3a
// walk on deeper call chains: /r/A → /r/v3aprobeb → /r/v3aprobe. It
// exports a single function that forwards into /r/v3aprobe.Probe and
// returns the result, so the test can observe what Caller / Self /
// CallerN resolve to at each hop.
package v3aprobeb

import (
	"chain/runtime"

	"gno.land/r/tests/vm/v3aprobe"
)

// ChainProbe forwards into /r/v3aprobe.DeepProbe. From DeepProbe's
// vantage (one frame deeper than ChainProbe), Caller is /r/v3aprobeb,
// CallerN(2) is /r/v3aprobeb's caller (the outermost /r/ realm).
func ChainProbe() (caller runtime.Realm, self runtime.Realm, prev runtime.Realm) {
	return v3aprobe.DeepProbe()
}

// silence unused import when ChainProbe doesn't use runtime directly.
var _ = runtime.Caller
