// PKGPATH: gno.land/r/tests/vm/lendvalues
package lendvalues

import lib "gno.land/p/demo/tests/lendholdlib"

var count int
var saved func()
var t = &T{}

type T struct{ n int }

func (t *T) Inc() { count++ }

type Runner interface{ Run() }

type F func()

func (f F) Call() { f() }

func init() { saved = func() { count++ } }

func show(s string, err any) { println(s, count, err) }

func main(cur realm) {
	show("v01 defer stored", try(func() { defer saved() }))
	show("v02 defer alias", try(func() { g := saved; defer g() }))
	show("v03 method value", try(func() { mv := t.Inc; mv() }))
	show("v04 defer method value", try(func() { mv := t.Inc; defer mv() }))
	show("v05 method expr (named)", try(func() { (*T).Inc(t) }))
	show("v06 map element", try(func() { m := map[string]func(){"a": func() { count++ }}; m["a"]() }))
	show("v07 struct field literal", try(func() { s := struct{ f func() }{func() { count++ }}; s.f() }))
	show("v08 converted literal call", try(func() { F(func() { count++ })() }))
	show("v09 returned then called", try(func() { mk()() }))
	show("v10 iface Runner holds lib.Fn", try(func() { var r Runner = lib.Fn(func() { count++ }); r.Run() }))
	show("v11 realm named type method", try(func() { F(func() { count++ }).Call() }))
	show("v12 any assert", try(func() { var x any = func() { count++ }; x.(func())() }))
	show("v13 lent lit panics, defer inside", try(func() {
		lib.Call(func() {
			defer func() { recover(); count++ }()
			panic("x")
		})
	}))
	show("v14 variadic lend", try(func() { lib2(func() { count++ }) }))
	show("v15 lend to own non-escaping", try(func() { own(func() { count++ }) }))
	show("v16 lend to own escaping", try(func() { ownKeep(func() { count++ }) }))
	show("v17 own kept called", try(func() { kept() }))
	show("v18 lent lit lends again", try(func() { own(func() { own(func() { count++ }) }) }))
	show("v19 closure of lent param", try(func() { ownWrap(func() { count++ }) }))
	show("v20 shadowed param name", try(func() { ownShadow(func() { count++ }) }))
}

var kept func()

func own(f func())     { f() }
func ownKeep(f func()) { kept = f }
func ownWrap(f func()) { g := func() { f() }; g() }

func ownShadow(f func()) {
	{
		f := saved
		_ = f
	}
	f()
}

func lib2(fs ...func()) { fs[0]() }

func mk() func() { return func() { count++ } }

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// v01 defer stored 0 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v02 defer alias 0 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v03 method value 0 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(3,0)>
// v04 defer method value 0 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(3,0)>
// v05 method expr (named) 1 undefined
// v06 map element 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v07 struct field literal 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v08 converted literal call 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v09 returned then called 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v10 iface Runner holds lib.Fn 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v11 realm named type method 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v12 any assert 1 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v13 lent lit panics, defer inside 2 undefined
// v14 variadic lend 2 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v15 lend to own non-escaping 3 undefined
// v16 lend to own escaping 3 undefined
// v17 own kept called 3 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v18 lent lit lends again 4 undefined
// v19 closure of lent param 4 gno.land/r/tests/vm/lendvalues's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendvalues) package{})).count
// v20 shadowed param name 5 undefined
