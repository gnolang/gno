// PKGPATH: gno.land/r/tests/vm/lendhelpercapassign
package lendhelpercapassign

var count int
var saved func()

func init() { saved = func() { count++ } }

// A local helper reassigned inside a literal it is captured by.
func main(cur realm) {
	println(count)
	println(try(func() {
		f := func() {}
		func() { f = saved }()
		f()
	}))
	println(count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 0
// gno.land/r/tests/vm/lendhelpercapassign's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelpercapassign) package{})).count
// 0
