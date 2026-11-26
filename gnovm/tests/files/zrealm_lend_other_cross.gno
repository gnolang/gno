// PKGPATH: gno.land/r/tests/vm/lendothercross
package lendothercross

import (
	plib "gno.land/p/demo/tests/lendcalllib"
	"gno.land/r/tests/vm/lendcaller"
)

var count int

// Another realm's crossing callee runs the literal in its own crossing call:
// refused, and the panic crosses back unrecovered.
func main(cur realm) {
	plib.Call(func() { count++ })
	println("o03 control: library callee writes", count)
	lendcaller.CallX(cross(cur), func() { count++ })
	println("o04 not reached", count)
}

// Output:
// o03 control: library callee writes 1

// Error:
// gno.land/r/tests/vm/lendothercross's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendothercross) package{})).count
