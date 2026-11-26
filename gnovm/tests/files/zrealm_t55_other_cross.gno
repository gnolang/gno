// PKGPATH: gno.land/r/tests/vm/t55othercross
package t55othercross

import (
	plib "gno.land/p/demo/tests/t55plib"
	"gno.land/r/tests/vm/t55other"
)

var count int

// Another realm's crossing callee runs the literal in its own crossing call:
// refused, and the panic crosses back unrecovered.
func main(cur realm) {
	plib.Call(func() { count++ })
	println("o03 control: library callee writes", count)
	t55other.CallX(cross(cur), func() { count++ })
	println("o04 not reached", count)
}

// Output:
// o03 control: library callee writes 1

// Error:
// gno.land/r/tests/vm/t55othercross's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/t55othercross) package{})).count
