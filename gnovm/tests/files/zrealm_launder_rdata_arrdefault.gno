// PKGPATH: gno.land/r/launderrh_arrdefault
package launderrh_arrdefault

import "gno.land/r/tests/vm/launderrarrvictim"

type Local struct {
	A launderrarrvictim.A
	B launderrarrvictim.B
}

var (
	g      = &Local{}
	gSlice = make([]launderrarrvictim.A, 1)
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
	println("registered:", launderrarrvictim.Register(&g.A), launderrarrvictim.RegisterBytes(&g.B), launderrarrvictim.RegisterSlice(&gSlice[0]))

	try("array default", func() { g.A[0] = 999 })
	try("byte array default", func() { g.B[0] = 99 })
	try("slice element default", func() { gSlice[0][0] = 999 })
	println("after attacker:", launderrarrvictim.Trusted(), launderrarrvictim.TrustedBytes(), launderrarrvictim.TrustedSlice())

	launderrarrvictim.VictimWrite(cross(cur), 5, 6, 7)
	println("after victim:", launderrarrvictim.Trusted(), launderrarrvictim.TrustedBytes(), launderrarrvictim.TrustedSlice())
}

// Output:
// registered: true true true
// array default blocked: true
// byte array default blocked: true
// slice element default blocked: true
// after attacker: 0 0 0
// after victim: 5 6 7
