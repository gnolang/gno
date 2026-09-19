// PKGPATH: gno.land/r/launderrh_mv_stored_iface
package launderrh_mv_stored_iface

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Method-value stored through `any` boxing.
//
// 1. Bind we.Apply, box into `any`, store in attacker-local map.
// 2. Later retrieve and type-assert back to the function type.
// 3. Invoke. Does the recv stamp survive the round-trip through
//    interface boxing?
//
// Trace expectation: bound method value carries a closure-like
// structure that retains the recv. After unbox + invoke, borrow rule #2
// should fire and write succeeds. Method values laundered through
// interface boxing remain dangerous.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	store := map[string]any{}
	store["go"] = we.Apply // box

	fn := store["go"].(func(func(*launderpkg.Immutable))) // unbox
	fn(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,0)>.Field
