package grc20reg

import (
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// Two tokens created by one realm with the same symbol get distinct identifiers,
// so their event streams can be told apart. This is the case that motivated the
// whole exercise: with grc20.NewToken and a reused seqid they would collide.
func TestEmitterNeverReissuesAnIdentifier(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/emitdup"))

	first, _ := grc20.NewTokenWithEmitter("Same", "DUP", 6, Emitter(), cur)
	second, _ := grc20.NewTokenWithEmitter("Same", "DUP", 6, Emitter(), cur)

	t.Log("first :", first.ID())
	t.Log("second:", second.ID())
	uassert.NotEqual(t, first.ID(), second.ID())
}

// Identifiers stay inside the creating realm's namespace, so one realm cannot
// obtain an identifier that appears to belong to another.
func TestIdentifierIsNamespacedToTheCreatingRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/emitns"))
	tok, _ := grc20.NewTokenWithEmitter("Ns", "NS", 6, Emitter(), cur)

	uassert.True(t, strings.HasPrefix(tok.ID(), "gno.land/r/demo/emitns.NS."))

	to, ok := IssuedTo(tok.ID())
	urequire.True(t, ok, "registry should have a record of the id it issued")
	uassert.Equal(t, "gno.land/r/demo/emitns", to)
}

// The attack that defeats a handed-out counter: copy the value the registry
// returns and replay it. Here the counter lives in the registry's package state,
// so a copy of the returned handle carries none of it.
func TestCopyingTheEmitterCannotReplayAnIdentifier(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/emitcopy"))

	snapshot := *canonicalEmitter // the move that breaks a struct-based generator

	viaRegistry, _ := grc20.NewTokenWithEmitter("A", "CP", 6, Emitter(), cur)
	viaCopy, _ := grc20.NewTokenWithEmitter("B", "CP", 6, &snapshot, cur)

	t.Log("via registry:", viaRegistry.ID())
	t.Log("via copy    :", viaCopy.ID())
	uassert.NotEqual(t, viaRegistry.ID(), viaCopy.ID())
}

// Register reports whether this registry issued the identifier, read from its own
// record rather than from anything the token claims about itself.
func TestRegisterReportsWhetherTheRegistryIssuedTheIdentifier(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/emitowned"))
	owned, _ := grc20.NewTokenWithEmitter("Owned", "OWN", 6, Emitter(), cur)
	Register(cross(cur), owned, "")
	to, ok := IssuedTo(owned.ID())
	uassert.True(t, ok)
	uassert.Equal(t, "gno.land/r/demo/emitowned", to)

	// A legacy token: created without an emitter, still registers, no record.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/emitlegacy"))
	legacy, _ := grc20.NewToken("Legacy", "LEG", 6, 0, cur)
	Register(cross(cur), legacy, "")
	_, ok = IssuedTo(legacy.ID())
	uassert.False(t, ok, "a token grc20 identified itself must not appear as registry-issued")
}
