package grc20reg

import (
	"chain"
	"strconv"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/avl/v0"
)

// This file makes grc20reg a grc20.Emitter: it issues token identifiers and
// records the events of the tokens that opted into it.
//
// A token realm opts in at construction:
//
//	Token, ledger := grc20.NewTokenWithEmitter("Foo", "FOO", 4, grc20reg.Emitter(), cur)
//	key := grc20reg.Register(cross(cur), Token, "")
//
// What that buys, and why it holds:
//
//   - Distinct identifiers. sequences below lives in this realm's package state
//     and is never returned to a caller, so it cannot be copied and replayed.
//     Two tokens from one realm therefore cannot share an identifier, which is
//     what made their event streams inseparable before.
//
//   - An unforgeable label on every event. tokenEmitter.Emit calls chain.Emit
//     from this package, so the VM stamps pkg_path = gno.land/r/demo/defi/grc20reg.
//     A counterfeit grc20.Emitter cannot borrow that: its events carry its own
//     path. Consumers filter on pkg_path and need trust nothing in-band.
//
//   - Provenance checkable from this realm's own records. issued below maps every
//     identifier this realm handed out to the realm it was handed to, so Register
//     verifies provenance without asking the token to describe its emitter — and
//     so without ever type-asserting or handing back a value that carries the
//     Emit capability.
//
// Tokens built with grc20.NewToken still work and still register. They emit
// under grc20's own pkg_path and keep the ambiguity, which is the point: opting
// out is visible rather than silent.

type emitter struct{}

type tokenEmitter struct{ id string }

var (
	// realm path -> int64, the per-realm issuance counter. Package state, never
	// handed out: this is the difference between a counter that cannot be
	// replayed and one that can.
	sequences = avl.NewTree()

	// token identifier -> issuing realm path. Register reads this instead of
	// trusting anything the token reports about itself.
	issued = avl.NewTree()

	canonicalEmitter = &emitter{}
)

var (
	_ grc20.Emitter      = (*emitter)(nil)
	_ grc20.TokenEmitter = (*tokenEmitter)(nil)
)

// Emitter returns this realm as a grc20.Emitter, for grc20.NewTokenWithEmitter.
//
// Unlike handing out a counter, handing out this value grants nothing: it holds
// no state, and every method that mutates state runs here.
func Emitter() grc20.Emitter {
	return canonicalEmitter
}

// IssueToken assigns a fresh identifier to a token being created by the calling
// realm. It is a crossing method, so the creating realm is read from the live
// frame rather than taken as an argument — an argument could name any realm.
func (e *emitter) IssueToken(cur realm, symbol string) grc20.TokenEmitter {
	rlmPath := cur.Previous().PkgPath()
	if rlmPath == "" {
		panic("grc20reg: token issuer must be a realm")
	}

	n := int64(0)
	if v := sequences.Get(rlmPath); v != nil {
		n = v.(int64)
	}
	n++
	sequences.Set(rlmPath, n)

	id := rlmPath + "." + symbol + "." + strconv.FormatInt(n, 10)
	issued.Set(id, rlmPath)

	return &tokenEmitter{id: id}
}

// TokenID returns the identifier this handle is permanently bound to.
func (t *tokenEmitter) TokenID() string { return t.id }

// Emit records one token event. Called from this package, so the VM stamps this
// realm's path onto the event.
//
// The handle is bound to a single identifier for life, so a copy of it conveys
// no authority its holder did not already have.
func (t *tokenEmitter) Emit(kind string, attrs ...string) {
	chain.Emit(kind, attrs...)
}

// IssuedTo returns the realm this registry issued id to, and whether it issued
// id at all.
func IssuedTo(id string) (string, bool) {
	v := issued.Get(id)
	if v == nil {
		return "", false
	}
	return v.(string), true
}
