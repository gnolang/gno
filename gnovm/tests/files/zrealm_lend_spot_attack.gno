// PKGPATH: gno.land/r/tests/vm/lendspotattack
package lendspotattack

import plib "gno.land/p/demo/tests/lendcalllib"

var count int
var kept func()
var saved func()
var pg = func() func() { return func() { count++ } }()

func init() { saved = func() { count++ } }

func show(s string, err any) { println(s, count, err) }

// Only the literal called where it is written runs as the realm's code; what
// it returns or stores, an alias of a literal, and a library's own on-the-spot
// literal calling a stored value stay refused.
func main(cur realm) {
	show("a01 inner closure returned, called later", try(func() {
		g := func() func() { return func() { count++ } }()
		g()
	}))
	func() { kept = func() { count++ } }()
	show("a02 inner closure stored, called later", try(func() { kept() }))
	show("a03 literal assigned first, called by another name", try(func() {
		h := (func() { count++ })
		k := h
		k()
	}))
	show("a04 library on-the-spot literal calls a stored value", try(func() { plib.Spot(saved) }))
	show("a06 package-level on-the-spot literal's returned closure", try(func() { pg() }))
	show("a05 control: on-the-spot literal writes", try(func() { func() { count++ }() }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// a01 inner closure returned, called later 0 gno.land/r/tests/vm/lendspotattack's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendspotattack) package{})).count
// a02 inner closure stored, called later 0 gno.land/r/tests/vm/lendspotattack's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendspotattack) package{})).count
// a03 literal assigned first, called by another name 0 gno.land/r/tests/vm/lendspotattack's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendspotattack) package{})).count
// a04 library on-the-spot literal calls a stored value 0 gno.land/r/tests/vm/lendspotattack's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendspotattack) package{})).count
// a06 package-level on-the-spot literal's returned closure 0 gno.land/r/tests/vm/lendspotattack's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendspotattack) package{})).count
// a05 control: on-the-spot literal writes 1 undefined
