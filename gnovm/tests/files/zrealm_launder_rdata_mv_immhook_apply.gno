// PKGPATH: gno.land/r/launderrh_mv_immhook_apply
package launderrh_mv_immhook_apply

// Stored ImmHook (func(*launderrvictim.Immutable)) that uses its arg
// — given by victim — to dispatch /p/-Apply on an EMBEDDED /p/-pointer
// reached from the victim's /r/-Immutable.
//
// But our victim's Immutable is JUST { Field string }, no inner
// /p/-pointer. So this variant probes whether the attacker can still
// reach an exploit by capturing the /p/-Apply-bearing pointer from
// elsewhere and using the supplied arg as a "trigger".
//
// Effectively: ImmHook called from victim with i=gImm gives us
// nothing extra over RegisterPlainHook. But: the hook can CAPTURE the
// /p/-aliased pointer separately. The closure body is /r/-attacker
// → Rule 1 → m.Realm = /r/-attacker. From there, captured.Apply
// dispatches via Rule 2 to /r/-victim, then EvilWrite as
// top-level /p/.
//
// Outcome: SUCCEEDS — write commits via Apply, identical primitive
// to the PlainHook variant. The ImmHook signature does NOT provide
// extra safety; the captured pointer carries the exploit.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/launderrvictim"
)

var captured *launderpkg.Immutable

func main(cur realm) {
	launderrvictim.ClearHooks()
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured = &we.Immutable

	launderrvictim.RegisterHook(func(_ *launderrvictim.Immutable) {
		captured.Apply(launderattack.EvilWrite)
	})
	launderrvictim.RunHooks()

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig
// after:  pwnd-via-apply
