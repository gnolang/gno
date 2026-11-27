// PKGPATH: gno.land/r/zrealm_lend_any_switchout
package zrealm_lend_any_switchout

var count int
var savedFn func()

// The type-switch variable outlives the call: stored, or captured by a
// literal that is kept.
func switchStore(f any) {
	switch g := f.(type) {
	case func():
		savedFn = g
		g()
	}
}

func switchCapture(f any) {
	switch g := f.(type) {
	case func():
		h := func() { g() }
		savedFn = h
		h()
	}
}

// Converted to an interface and passed on: escapes too.
func switchBox(f any) {
	switch g := f.(type) {
	case func():
		sink(any(g))
		g()
	}
}

func sink(x any) {}

func main(cur realm) {
	println("store:", try(func() { switchStore(func() { count++ }) }))
	println("capture:", try(func() { switchCapture(func() { count++ }) }))
	println("box:", try(func() { switchBox(func() { count++ }) }))
	println("later:", try(func() { savedFn() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// store: gno.land/r/zrealm_lend_any_switchout's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_switchout) package{})).count
// capture: gno.land/r/zrealm_lend_any_switchout's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_switchout) package{})).count
// box: gno.land/r/zrealm_lend_any_switchout's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_switchout) package{})).count
// later: gno.land/r/zrealm_lend_any_switchout's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_switchout) package{})).count
// count: 0
