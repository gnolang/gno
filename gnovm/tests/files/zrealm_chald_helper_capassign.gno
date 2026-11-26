// PKGPATH: gno.land/r/tests/vm/chaldcapasg
package chaldcapasg

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
// gno.land/r/tests/vm/chaldcapasg's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldcapasg) package{})).count
// 0
