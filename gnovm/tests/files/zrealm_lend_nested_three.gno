// PKGPATH: gno.land/r/tests/vm/lendnestedthree
package lendnestedthree

var count int
var saved func()

func esc(x func()) { saved = x }

// one: a wrapper of f is stored, so f is refused.
func one(f func()) {
	esc(func() { f() })
	f()
}

// three: a wrapper three levels around f is stored, so f is refused too.
func three(f func()) {
	mid1(func() { f() })
	f()
}

func mid1(g func()) { mid2(func() { g() }) }
func mid2(h func()) { esc(func() { h() }) }

func main(cur realm) {
	r := try(func() { one(func() { count++ }) })
	println("one level:", count, r != nil)
	r = try(func() { three(func() { count++ }) })
	println("three levels:", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// one level: 0 true
// three levels: 0 true
