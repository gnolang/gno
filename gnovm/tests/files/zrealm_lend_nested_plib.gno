// PKGPATH: gno.land/r/tests/vm/lendnestedplib
package lendnestedplib

import lib "gno.land/p/demo/tests/lendholdlib"

var count int

// one: a /p/ function keeps a wrapper of f, so f is refused.
func one(f func()) {
	var box []func()
	lib.Keep(func() { f() }, &box)
	f()
}

// two: a /p/ function keeps a wrapper of a wrapper of f, so f is refused too.
func two(f func()) {
	mid(func() { f() })
	f()
}

func mid(g func()) {
	var box []func()
	lib.Keep(func() { g() }, &box)
}

func main(cur realm) {
	r := try(func() { one(func() { count++ }) })
	println("one level:", count, r != nil)
	r = try(func() { two(func() { count++ }) })
	println("two levels /p/:", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// one level: 0 true
// two levels /p/: 0 true
