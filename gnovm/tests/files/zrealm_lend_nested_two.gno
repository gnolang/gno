// PKGPATH: gno.land/r/tests/vm/lendnestedtwo
package lendnestedtwo

var count int
var saved func()

func esc(x func()) { saved = x }

// one: a wrapper of f is stored, so f is refused.
func one(f func()) {
	esc(func() { f() })
	f()
}

// two: a wrapper of a wrapper of f is stored, so f is refused too.
func two(f func()) {
	mid(func() { f() })
	f()
}

func mid(g func()) { esc(func() { g() }) }

func main(cur realm) {
	r := try(func() { one(func() { count++ }) })
	println("one level:", count, r != nil)
	r = try(func() { two(func() { count++ }) })
	println("two levels:", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// one level: 0 true
// two levels: 0 true
