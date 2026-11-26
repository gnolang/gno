// PKGPATH: gno.land/r/tests/vm/lendinit
package lendinit

import lib "gno.land/p/demo/tests/lendholdlib"

var count int
var fromInit = lib.Ret(func() int { count++; return count })
var kept []func()

func init() {
	lib.Keep(func() { count++ }, &kept)
	h := func() { count++ }
	h()
}

func show(s string, err any) { println(s, count, err) }

// Writes during init are allowed; a literal kept during init is a stored
// value afterwards, refused when called later.
func main(cur realm) {
	println(fromInit)
	show("i01 after init", nil)
	show("i02 init-kept literal", try(func() { kept[0]() }))
	show("i03 init-kept via lib", try(func() { lib.Call(kept[0]) }))
	show("i04 init-kept re-lent wrapper", try(func() { lib.Call(func() { kept[0]() }) }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 1
// i01 after init 2 undefined
// i02 init-kept literal 2 gno.land/r/tests/vm/lendinit's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendinit) package{})).count
// i03 init-kept via lib 2 gno.land/r/tests/vm/lendinit's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendinit) package{})).count
// i04 init-kept re-lent wrapper 2 gno.land/r/tests/vm/lendinit's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendinit) package{})).count
