// PKGPATH: gno.land/r/tests/vm/chaldctl
package chaldctl

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
// gno.land/r/tests/vm/chaldctl's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldctl) package{})).count
// 0
