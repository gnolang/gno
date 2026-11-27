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

// Error:
// cannot allocate gno.land/r/tests/vm/launderrarrvictim.A in realm gno.land/r/launderrh_arrdefault

// TypeCheckError:
// gno.land/r/launderrh_arrdefault/zrealm_launder_rdata_arrdefault.gno:6:6: cannot hold a value of gno.land/r/tests/vm/launderrarrvictim.A outside its realm (use a pointer); gno.land/r/launderrh_arrdefault/zrealm_launder_rdata_arrdefault.gno:7:2: cannot hold a value of gno.land/r/tests/vm/launderrarrvictim.A outside its realm (use a pointer); gno.land/r/launderrh_arrdefault/zrealm_launder_rdata_arrdefault.gno:8:2: cannot hold a value of gno.land/r/tests/vm/launderrarrvictim.B outside its realm (use a pointer); gno.land/r/launderrh_arrdefault/zrealm_launder_rdata_arrdefault.gno:13:2: cannot hold a value of gno.land/r/tests/vm/launderrarrvictim.A outside its realm (use a pointer); gno.land/r/launderrh_arrdefault/zrealm_launder_rdata_arrdefault.gno:13:11: cannot hold a value of gno.land/r/tests/vm/launderrarrvictim.A outside its realm (use a pointer)
