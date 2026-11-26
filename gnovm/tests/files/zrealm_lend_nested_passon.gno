// PKGPATH: gno.land/r/tests/vm/lendnestedpasson
package lendnestedpasson

var count int
var saved func()

func esc(x func()) { saved = x }

// one: a wrapper of f is stored, so f is refused.
func one(f func()) {
	esc(func() { f() })
	f()
}

// passon: a lent wrapper of f is passed on to a storing parameter, so the
// wrapper and f both lose their loans.
func passon(f func()) {
	pass(func() { f() })
	f()
}

func pass(g func()) { esc(g) }

func main(cur realm) {
	r := try(func() { one(func() { count++ }) })
	println("one level:", count, r != nil)
	r = try(func() { passon(func() { count++ }) })
	println("lent wrapper passed on:", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// one level: 0 true
// lent wrapper passed on: 0 true
