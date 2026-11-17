// PKGPATH: gno.land/r/launderrh_mv_higher_order_chain
package launderrh_mv_higher_order_chain

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Higher-order indirection chain: victim calls fn that calls fn that
// calls attacker's STORED PlainHook, which dispatches /p/-Apply
// against a captured /r/-victim-stamped /p/-pointer.
//
// 1. main → m.Realm = /r/-attacker
// 2. RegisterPlainHook(closureA) — borrow rule #1 → victim. closureA stored.
// 3. main calls runViaTrampoline() — local helper that wraps the
//    victim call inside another function value indirection.
// 4. trampoline() invokes launderrvictim.RunPlainHooks(cross(cur)). borrow rule #1 →
//    m.Realm = /r/-victim.
// 5. RunPlainHooks calls closureA() — /r/-attacker body → borrow rule #1 →
//    m.Realm = /r/-attacker.
// 6. Inside closureA: chainHelper() — local /r/-attacker function
//    that further indirects to captured.Apply(EvilWrite).
// 7. Apply is /p/-method, recv is /r/-victim-stamped → borrow rule #2 →
//    m.Realm = /r/-victim during Apply.
// 8. EvilWrite is /p/-top-level → no shift → write commits under
//    victim authority → SUCCEEDS.
//
// This proves indirection chaining doesn't shield the (B) primitive.
// As long as the captured pointer is /r/-victim-stamped, the eventual
// .Apply call cuts through the wrapper layers.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

var captured *launderpkg.Immutable

func chainHelper() {
	captured.Apply(launderattack.EvilWrite)
}

func trampoline(fn func()) {
	fn()
}

func main(cur realm) {
	launderrvictim.ClearHooks(cross(cur))
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured = &we.Immutable

	launderrvictim.RegisterPlainHook(cross(cur), func() {
		chainHelper()
	})
	trampoline(func() { launderrvictim.RunPlainHooks(cross(cur)) })

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
