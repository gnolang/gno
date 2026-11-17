// PKGPATH: gno.land/r/launderrh_pdr_victim_recover
package launderrh_pdr_victim_recover

// Vector 5: /r/-victim's RecoverAndRetry calls the attacker's callback
// h(gImm), which panics with the readonly violation. The callback is the
// attacker's code entered read-only, no realm boundary, so the victim's own
// recover catches the panic and nothing was written. The attacker's direct
// retry is refused too, caught by its own recover.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	recovered := launderrvictim.RecoverAndRetry(cross(cur), func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-victim-recover"
	})

	// Inspect recovered exception. We expect a readonly panic, not nil.
	if recovered == nil {
		println("UNEXPECTED: nil recover")
	} else {
		println("recovered some exception (write was blocked)")
	}

	// Attacker tries again, this time direct.
	defer func() {
		if r := recover(); r != nil {
			println("retry blocked:", r != nil)
		}
	}()

	imm := launderrvictim.GetImm()
	imm.Field = "pwnd-direct-after"

	println("after: ", launderrvictim.ReadImm())
}

// Output:
// before: rdata-original
// recovered some exception (write was blocked)
// retry blocked: true
