// PKGPATH: gno.land/r/tests/vm/t55lenddefer
package t55lenddefer

import plib "gno.land/p/demo/tests/t55plib"

var count int

// A callee's defer runs before the callee returns, so a lent literal still
// holds its loan there; a stored value never had one.
func main(cur realm) {
	r := try(func() { plib.DeferCall(func() { count++ }) })
	println("d01 callee defers the lent literal", count, r)
	saved := func() { count++ }
	r = try(func() { plib.DeferCall(saved) })
	println("d02 control: callee defers a variable", count, r)
}

func try(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return
}

// Output:
// d01 callee defers the lent literal 1 undefined
// d02 control: callee defers a variable 1 gno.land/r/tests/vm/t55lenddefer's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/t55lenddefer) package{})).count
