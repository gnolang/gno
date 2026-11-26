// PKGPATH: gno.land/r/tests/vm/lendctlstored
package lendctlstored

var count int
var saved func()

func init() { saved = func() { count++ } }

func main(cur realm) {
	println(count)
	g := saved
	println(try(func() { g() }))
	println(count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 0
// gno.land/r/tests/vm/lendctlstored's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendctlstored) package{})).count
// 0
