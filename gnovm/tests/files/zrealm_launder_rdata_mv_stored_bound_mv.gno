// PKGPATH: gno.land/r/launderrh_mv_stored_bound_mv
package launderrh_mv_stored_bound_mv

// Stored bound-method-value laundering probe.
//
// 1. Attacker reads victim's exposed /p/-typed pointer field
//    (`&we.Immutable` aliasing victim memory).
// 2. Attacker BINDS the /p/-method-value `captured.BumpToPwn` —
//    type is `func()`, fits PlainHook.
// 3. Attacker registers the bound method value as a PlainHook in
//    /r/-victim state via RegisterPlainHook.
// 4. Attacker calls RunPlainHooks(). Inside RunPlainHooks (Rule 1
//    → m.Realm = /r/-victim), the loop body invokes h(). The bound
//    method value has no Rule-1 body source (it's a /p/-method), but
//    PushFrameCall sees the captured recv pointer is /r/-victim-
//    stamped → Rule 2 → m.Realm = /r/-victim during BumpToPwn → the
//    write commits under victim authority.
//
// This is a STORED variant where the dangerous function is a bound
// /p/-method value, no callback parameter is passed at the call site.
// It exercises the same Rule-2 receiver-borrow primitive as the
// known /p/-method-on-/r/-pointer attack, but through hook storage.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	launderrvictim.ClearHooks()
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured := &we.Immutable
	launderrvictim.RegisterPlainHook(captured.BumpToPwn) // bound mv
	launderrvictim.RunPlainHooks()

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig
// after:  pwnd-via-bound-mv
