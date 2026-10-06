// PKGPATH: gno.land/r/zzoperatorauth
//
// Regression pin for the per-operator auth-list capability leak
// (reported by @thehowl; pre-existing, fixed here because it is the same
// class as the governance capability leak, and the same audit that found
// that one should have found it).
//
// THE HOLE. `Valoper.Auth() *authorizable.Authorizable` was exported and
// `GetByAddr` is exported and non-crossing. Valoper is returned BY
// VALUE, but `auth` is a pointer field, so the copy shared the realm's
// live Authorizable. Authorizable's gates read
// `rlm.Previous().Address()`, so inside a hostile realm's frame
// Previous() is whoever called it. A valoper OPERATOR merely calling any
// function of a hostile realm (faucet, airdrop, mint) let that realm run
//
//	valopers.GetByAddr(operator).Auth().AddToAuthList(0, cur, attacker)
//
// with Previous() == the operator, i.e. the Authorizable's own owner.
// The write was accepted and PERSISTED. From the next transaction on,
// the attacker acted alone: UpdateKeepRunning to drain the validator,
// UpdateSigningKey to rotate its consensus signing key.
//
// THE FIX is the one layer 2 applies to the realm's governance
// authority: never export the live capability. `Auth()` is removed;
// `AuthOwner()` copies out an address.
//
// WHAT THIS FILE PINS, honestly scoped. The raw-handle attack is a
// compile-time absence now, so — like the foreign-realm filetest —
// reverting the fix fails this on a build error rather than an
// assertion. What it
// drives at RUNTIME is the property that makes the removal safe: the
// exported wrapper is NOT an equivalent route. There `cur.Previous()` is
// the hostile realm rather than the operator, so the superuser check
// refuses, and it refuses even in the exact scenario that used to work —
// the operator calling the hostile realm themselves.
//
// NOTE for whoever edits valopers next: re-adding any exported accessor
// that returns *authorizable.Authorizable (or the *ownable.Ownable
// inside it) reopens this, and this file would still compile.
package zzoperatorauth

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnops/valopers"
)

var (
	operator = testutils.TestAddress("operator")
	attacker = chain.PackageAddress("gno.land/r/zzoperatorauth")
	pubKey   = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

// Pwn is an ordinary exported function of a hostile realm — the airdrop
// the operator was told to claim. It needs no privilege of its own; it
// used to just read the public handle and write through it.
func Pwn(cur realm, victim address) {
	// The attack the fix removes. Uncommenting must not compile:
	//
	//	valopers.GetByAddr(victim).Auth().AddToAuthList(0, cur, attacker)
	//	                           ^^^^^^ undefined
	//
	// What remains is the exported wrapper. It is not equivalent: it
	// derives the principal from THIS frame, where Previous() is
	// gno.land/r/zzoperatorauth and not the operator. Aborts.
	valopers.AddToAuthList(cross(cur), victim, attacker)
	println("UNREACHABLE: hostile realm joined the operator's auth list")
}

func main(cur realm) {
	// The operator registers their profile normally.
	testing.SetOriginCaller(operator)
	testing.SetRealm(testing.NewUserRealm(operator))
	valopers.Register(cross(cur), "Op",
		"a valid description for the operator profile", "cloud", operator, pubKey)

	// The auth list is owned by the operator, and AuthOwner hands back
	// an address — a value, not a capability.
	println("owner is operator:",
		valopers.GetByAddr(operator).AuthOwner() == operator)
	println("attacker is not owner:",
		valopers.GetByAddr(operator).AuthOwner() != attacker)

	// The operator calls the hostile realm once. This single step used
	// to be the whole attack.
	testing.SetOriginCaller(operator)
	testing.SetRealm(testing.NewUserRealm(operator))
	Pwn(cross(cur), operator)
}

// Output:
// owner is operator: true
// attacker is not owner: true

// Error:
// authorizable: caller is not superuser
