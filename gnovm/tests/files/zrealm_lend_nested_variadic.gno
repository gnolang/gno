// PKGPATH: gno.land/r/tests/vm/lendnestedvariadic
package lendnestedvariadic

var count int
var saved func()

func escv(xs ...func()) { saved = xs[0] }

// one: a wrapper of f goes to a variadic slot and is stored, so f is refused.
func one(f func()) {
	escv(func() { f() })
	f()
}

// two: a wrapper of a wrapper of f goes to a variadic slot, so f is refused too.
func two(f func()) {
	mid(func() { f() })
	f()
}

func mid(g func()) { escv(func() { g() }) }

func main(cur realm) {
	r := try(func() { one(func() { count++ }) })
	println("one level:", count, r != nil)
	r = try(func() { two(func() { count++ }) })
	println("two levels variadic:", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// one level: 0 true
// two levels variadic: 0 true
