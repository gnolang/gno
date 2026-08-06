// PKGPATH: gno.land/r/demo/grc20fakeemitter

// A counterfeit Emitter is free to return whatever identifier it likes, so
// NewTokenWithEmitter checks the issued id against the live crossing frame
// before adopting it. That keeps Token.ID()'s namespace guarantee identical to
// NewToken's no matter which Emitter is supplied: a realm cannot obtain a token
// that appears to belong to somebody else.
package grc20fakeemitter

import (
	"gno.land/p/demo/tokens/grc20"
)

type fakeEmitter struct{}

func (f *fakeEmitter) IssueToken(cur realm, symbol string) grc20.TokenEmitter {
	// Claim a victim realm's namespace.
	return &fakeTokenEmitter{id: "gno.land/r/demo/grc20victim.VIC.1"}
}

type fakeTokenEmitter struct{ id string }

func (f *fakeTokenEmitter) TokenID() string                   { return f.id }
func (f *fakeTokenEmitter) Emit(kind string, attrs ...string) {}

func main(cur realm) {
	grc20.NewTokenWithEmitter("Fake", "VIC", 6, &fakeEmitter{}, cur)
}

// Error:
// emitter issued an id outside the calling realm's namespace
