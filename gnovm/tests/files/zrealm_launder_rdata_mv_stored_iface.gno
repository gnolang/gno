// PKGPATH: gno.land/r/launderrh_mv_stored_iface
package launderrh_mv_stored_iface

// Method-value stored through `any` boxing.
//
// 1. Bind we.Apply, box into `any`, store in attacker-local map.
// 2. Later retrieve and type-assert back to the function type.
// 3. Invoke. Does the recv stamp survive the round-trip through
//    interface boxing?
//
// Trace expectation: bound method value carries a closure-like
// structure that retains the recv. After unbox + invoke, Rule 2
// should fire and write succeeds. Method values laundered through
// interface boxing remain dangerous.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/p/demo/tests/launderpkg"
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
// after:  pwnd-via-apply
