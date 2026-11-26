// PKGPATH: gno.land/r/tests/vm/lendcrossval
package lendcrossval

import (
	"gno.land/p/nt/avl/v0"

	lib "gno.land/p/demo/tests/lendholdlib"
)

var count int
var saved func()
var savedX func(cur realm)
var tree avl.Tree
var setter func(string, any) bool

func init() {
	saved = func() { count++ }
	savedX = func(cur realm) { count++ }
	setter = tree.Set
}

func show(s string, err any) { println(s, count, tree.Size(), err) }

// A stored non-crossing value is refused (control); the same closure typed
// crossing and called without cross is not.
func main(cur realm) {
	show("c01 stored value (control)", try(func() { g := saved; g() }))
	show("c02 stored crossing value, no cross", try(func() { g := savedX; g(cur) }))
	show("c04 stored lib method value", try(func() { g := setter; g("k", 1) }))
	show("c05 lib method value lent to lib", try(func() { lib.Call(func() { setter("k2", 1) }) }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// c01 stored value (control) 0 0 gno.land/r/tests/vm/lendcrossval's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendcrossval) package{})).count
// c02 stored crossing value, no cross 1 0 undefined
// c04 stored lib method value 1 0 cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// c05 lib method value lent to lib 1 0 cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
