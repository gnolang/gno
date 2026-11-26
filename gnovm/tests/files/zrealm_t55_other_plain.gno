// PKGPATH: gno.land/r/tests/vm/t55otherplain
package t55otherplain

import (
	plib "gno.land/p/demo/tests/t55plib"
	"gno.land/r/tests/vm/t55other"
)

var count int

// A literal handed to another realm's function is not lent: that code is not
// this realm's import-tree library.
func main(cur realm) {
	r := try(func() { plib.Call(func() { count++ }) })
	println("o01 control: library callee writes", count, r)
	r = try(func() { t55other.Call(func() { count++ }) })
	println("o02 another realm's non-crossing callee", count, r)
}

func try(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return
}

// Output:
// o01 control: library callee writes 1 undefined
// o02 another realm's non-crossing callee 1 gno.land/r/tests/vm/t55otherplain's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/t55otherplain) package{})).count
