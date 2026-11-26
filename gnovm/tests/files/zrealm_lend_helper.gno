// PKGPATH: gno.land/r/tests/vm/lendhelper
package lendhelper

var count int
var saved func()
var savedX func(cur realm)

func init() {
	saved = func() { count++ }
	savedX = func(cur realm) { count++ }
}

func target() {}

func show(s string, err any) { println(s, count, err) }

func main(cur realm) {
	show("h01 helper (control)", try(func() { f := func() { count++ }; f(); f() }))
	show("h02 var-then-assign", try(func() {
		var f func()
		f = func() { count++ }
		f()
	}))
	show("h03 if-header shadow", try(func() {
		f := func() {}
		if f := saved; true {
			f()
		}
		f()
	}))
	show("h04 switch-header shadow", try(func() {
		f := func() {}
		switch f := saved; {
		default:
			f()
		}
		f()
	}))
	show("h05 swap", try(func() {
		f, g := func() {}, saved
		f, g = g, f
		f()
	}))
	show("h06 for-post assign", try(func() {
		i := 0
		for f := func() {}; i < 2; f = saved {
			f()
			i++
		}
	}))
	show("h07 goto loop reuse", try(func() {
		i := 0
	again:
		f := func() {}
		if i == 0 {
			i++
			goto again
		}
		f()
		g := saved
		g()
	}))
	show("h08 range assign", try(func() {
		f := func() {}
		for _, f = range []func(){saved} {
		}
		f()
	}))
	show("h09 crossing value no cross", try(func() { g := savedX; g(cur) }))
	show("h10 crossing literal helper", try(func() { f := func(cur realm) { count++ }; f(cur) }))
	show("h11 deferred helper", try(func() { f := func() { count++ }; defer f() }))
	show("h12 helper via defer arg", try(func() {
		f := func() {}
		defer func(g func()) { g() }(saved)
		f()
	}))
	show("h13 multi-def from pair", try(func() { f, g := pair(); f(); g() }))
	show("h14 recursive var helper", try(func() {
		var f func(n int)
		f = func(n int) {
			if n > 0 {
				count++
				f(n - 1)
			}
		}
		f(2)
	}))
	show("h15 closure capture reassign then call", try(func() {
		f := func() {}
		set := func() { f = saved }
		set()
		f()
	}))
	show("h16 pointer to helper", try(func() {
		f := func() {}
		p := &f
		*p = saved
		f()
	}))
	show("h17 declared func reassigned", try(func() { assignTarget(); target() }))
}

func assignTarget() { defer func() { recover() }(); setTarget() }
func setTarget()    {}

func pair() (func(), func()) { return saved, saved }

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// h01 helper (control) 2 undefined
// h02 var-then-assign 3 undefined
// h03 if-header shadow 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h04 switch-header shadow 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h05 swap 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h06 for-post assign 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h07 goto loop reuse 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h08 range assign 3 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h09 crossing value no cross 4 undefined
// h10 crossing literal helper 5 undefined
// h11 deferred helper 6 undefined
// h12 helper via defer arg 6 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h13 multi-def from pair 6 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h14 recursive var helper 8 undefined
// h15 closure capture reassign then call 8 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h16 pointer to helper 8 gno.land/r/tests/vm/lendhelper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhelper) package{})).count
// h17 declared func reassigned 8 undefined
