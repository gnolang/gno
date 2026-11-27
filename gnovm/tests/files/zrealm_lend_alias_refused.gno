// PKGPATH: gno.land/r/tests/vm/lendaliasrefused
package lendaliasrefused

import (
	lib "gno.land/p/demo/tests/t76lib"
	other "gno.land/r/tests/vm/t76other"
)

// A local that is not exactly "declared once from a declared function and
// only called" stays a call by value; "name" lines show the by-name twin.

var (
	n       int
	xs      = []int{1, 2, 3}
	saved   func([]int, func(int))
	savedFn func()
	pkgEach = lib.Each
)

func show(s string, err any) {
	println(s, n, err)
	n = 0
}

type Fn func()

func run(f Fn) { f() }

func pass(f func([]int, func(int))) {}

func get() func([]int, func(int)) {
	each := lib.Each
	each(xs, func(x int) { n += x })
	return each
}

func main(cur realm) {
	show("01 reassigned later", try(func() {
		each := lib.Each
		each(xs, func(x int) { n += x })
		each = lib.Keep
	}))
	show("02 method value", try(func() {
		mv := lib.Inner{}.Do
		mv(xs, func(x int) { n += x })
	}))
	show("03 package variable", try(func() {
		ev := lib.EachVar
		ev(xs, func(x int) { n += x })
	}))
	show("04a passed on", try(func() {
		each := lib.Each
		each(xs, func(x int) { n += x })
		pass(each)
	}))
	show("04b stored", try(func() {
		each := lib.Each
		each(xs, func(x int) { n += x })
		saved = each
	}))
	show("05 returned", try(func() { get() }))
	show("06a captured by a stored literal", try(func() {
		each := lib.Each
		savedFn = func() { each(xs, func(x int) {}) }
		each(xs, func(x int) { n += x })
	}))
	show("06b captured by a literal handed out", try(func() {
		each := lib.Each
		lib.Call(func() { each(xs, func(x int) {}) })
		each(xs, func(x int) { n += x })
	}))
	show("07 other realm, name ", try(func() { other.Each(xs, func(x int) { n += x }) }))
	show("07 other realm, alias", try(func() { oe := other.Each; oe(xs, func(x int) { n += x }) }))
	show("08 multi-assign", try(func() {
		a, b := lib.Each, lib.Keep
		a(xs, func(x int) { n += x })
		b(xs, func(x int) { n += x })
	}))
	show("09 package-level var", try(func() { pkgEach(xs, func(x int) { n += x }) }))
	show("10 literal not in place, name ", try(func() {
		f := func(x int) { n += x }
		lib.Each(xs, f)
	}))
	show("10 literal not in place, alias", try(func() {
		f := func(x int) { n += x }
		each := lib.Each
		each(xs, f)
	}))
	show("11 deferred, name ", try(func() { defer lib.Each(xs, func(x int) { n += x }) }))
	show("11 deferred, alias", try(func() {
		each := lib.Each
		defer each(xs, func(x int) { n += x })
	}))
	show("12 address taken", try(func() {
		each := lib.Each
		p := &each
		each(xs, func(x int) { n += x })
		_ = p
	}))
	show("13 method expression, name ", try(func() { lib.Inner.Do(lib.Inner{}, xs, func(x int) { n += x }) }))
	show("13 method expression, alias", try(func() {
		me := lib.Inner.Do
		me(lib.Inner{}, xs, func(x int) { n += x })
	}))
	// Accepted fail-closed gap: refused although the by-name twin writes.
	show("14 in a converted literal, nested call, name ", try(func() {
		run(func() {
			func() { lib.Each(xs, func(x int) { n += x }) }()
		})
	}))
	show("14 in a converted literal, nested call, alias", try(func() {
		run(func() {
			each := lib.Each
			func() { each(xs, func(x int) { n += x }) }()
		})
	}))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 01 reassigned later 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 02 method value 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 03 package variable 0 gno.land/r/tests/vm/lendaliasrefused's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 04a passed on 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 04b stored 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 05 returned 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 06a captured by a stored literal 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 06b captured by a literal handed out 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 07 other realm, name  0 gno.land/r/tests/vm/lendaliasrefused's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 07 other realm, alias 0 gno.land/r/tests/vm/lendaliasrefused's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 08 multi-assign 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 09 package-level var 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 10 literal not in place, name  0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 10 literal not in place, alias 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 11 deferred, name  0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 11 deferred, alias 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 12 address taken 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 13 method expression, name  6 undefined
// 13 method expression, alias 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
// 14 in a converted literal, nested call, name  6 undefined
// 14 in a converted literal, nested call, alias 0 gno.land/r/tests/vm/lendaliasrefused's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendaliasrefused) package{})).n
