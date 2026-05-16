// PKGPATH: gno.land/r/test/sealrlmshadow
package sealrlmshadow

// Realistic Teller/Banker attack shape: the prior embedding filetests
// shadow Hello() string, which proves embedding-bypass in principle.
// This filetest shadows a method whose signature takes `(_ int, rlm
// realm, ...)` — the actual shape of grc20.Teller.Transfer,
// treasury.Banker.Send, and similar rlm-leaking methods — to confirm
// the attacker really can capture cur via the embedding bypass.
//
// Search terms: rlm shadow, capture cur via embedding, Teller attack
// proof, sealed interface embedding bypass for rlm-typed methods.

import (
	"gno.land/p/test/seal"
)

// Capturable is a Sealed-shaped interface with an rlm-typed method,
// matching the Teller/Banker pattern (sealed marker + a method that
// takes rlm).
type Capturable interface {
	seal.Sealed
	DoWithRealm(_ int, rlm realm) string
}

// evilImpl embeds *seal.Impl — gains the promoted isSealed() marker
// satisfying the seal, plus the promoted Hello() — then adds its
// own DoWithRealm that captures rlm. This is the realistic Teller
// attack shape: attacker holds a legit Sealed value, embeds it,
// adds a method that takes rlm and stashes it for later use.
type evilImpl struct {
	*seal.Impl
}

var capturedRealmPkgPath string

func (e *evilImpl) DoWithRealm(_ int, rlm realm) string {
	// rlm captured here — attacker can later use this stash to act
	// AS the realm whose cur was captured (A/B-class leak).
	capturedRealmPkgPath = rlm.PkgPath()
	return "evil captured: " + capturedRealmPkgPath
}

// attack runs in a crossing frame so cur is available.
func attack(cur realm) {
	evil := &evilImpl{Impl: seal.NewImpl("inner")}
	// Assignment SUCCEEDS — *evilImpl satisfies Capturable via the
	// embedded *seal.Impl's promoted isSealed() and Hello(), plus
	// the attacker-defined DoWithRealm.
	var c Capturable = evil
	// Caller hands their cur to what they believe is a sealed,
	// in-package Capturable. Attacker's DoWithRealm captures rlm.
	println(c.DoWithRealm(0, cur))
	println("stash holds:", capturedRealmPkgPath)
}

func main(cur realm) {
	attack(cross2(cur))
}

// Output:
// evil captured: gno.land/r/test/sealrlmshadow
// stash holds: gno.land/r/test/sealrlmshadow
