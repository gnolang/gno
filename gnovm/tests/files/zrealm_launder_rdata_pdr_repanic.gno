// PKGPATH: gno.land/r/launderrh_pdr_repanic
package launderrh_pdr_repanic

// Vector 9: panic → recover → re-panic chain in attacker land,
// each step probing m.Realm state and a write attempt.
//
// We're staying within attacker realm for the panic/recover cycle
// (no cross-realm panic), so recover() works. Tests that even
// after multiple panic/recover cycles, attacker can never write
// to victim's gImm.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func try(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			println(label, "blocked: true")
		} else {
			println(label, "blocked: false")
		}
	}()
	fn()
}

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	// 1. Panic in attacker fn, recover, then attempt write.
	try("1 panic+recover+write", func() {
		func() {
			defer func() { _ = recover() }()
			panic("attacker self-panic")
		}()
		// After recover, m.Realm should still be attacker.
		imm := launderrvictim.GetImm()
		imm.Field = "pwnd-after-self-recover"
	})

	// 2. Re-panic from defer: panic in inner, recover only checks
	//    val and re-panics. Final outer recover catches.
	try("2 panic→recover→repanic→outer-recover", func() {
		defer func() {
			r := recover()
			if r == nil {
				println("UNEXPECTED inner-2 no recover")
			}
		}()
		// repanic-rethrown sequence:
		defer func() {
			if r := recover(); r != nil {
				panic(r) // re-panic
			}
		}()
		panic("inner-2 panic")
	})

	// 3. Stacked defers, each tries a victim write.
	try("3 stacked-defer victim writes", func() {
		imm := launderrvictim.GetImm()
		defer func() { imm.Field = "pwnd-defer-1" }()
		defer func() { imm.Field = "pwnd-defer-2" }()
		// Force the defers to run via panic.
		panic("force-defers")
	})

	println("after: ", launderrvictim.ReadImm())
}

// Output:
// before: rdata-original
// 1 panic+recover+write blocked: true
// 2 panic→recover→repanic→outer-recover blocked: false
// 3 stacked-defer victim writes blocked: true
// after:  rdata-original
