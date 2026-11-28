// PKGPATH: gno.land/r/launder_rdata_arrdefault_nested
package launder_rdata_arrdefault_nested

// An unnamed array nested in a realm-declared type (a row of a declared
// array, a [3]int or [4]byte field of a declared struct) is stamped with the
// declaring realm, so the realm that built the default cannot keep writing
// to it after handing the pointer over.

import v "gno.land/r/tests/vm/launderrarrnested"

type Local struct {
	N v.N
	S v.S
	P v.P
}

var g = &Local{}

func try(label string, fn func()) {
	defer func() {
		r := recover()
		println(label, "blocked:", r != nil)
	}()
	fn()
}

func main(cur realm) {
	println("registered:", v.RegisterN(&g.N), v.RegisterS(&g.S), v.RegisterP(&g.P))
	try("inner array of declared array", func() { g.N[0][0] = 999 })
	try("array field of declared struct", func() { g.S.Arr[0] = 999 })
	try("byte array field of declared struct", func() { g.P.Buf[0] = 99 })
	println("after attacker:", v.Trusted())
}

// Output:
// registered: true true true
// inner array of declared array blocked: true
// array field of declared struct blocked: true
// byte array field of declared struct blocked: true
// after attacker: 0 0 0
