// PKGPATH: gno.land/r/zrealm_lend_recover_param
package zrealm_lend_recover_param

var count int

func attack() { run(func() { count++ }) }

// The parameter is reused by := ; a panic skips it and the deferred literal calls the caller's value.
func run(g func()) {
	defer func() {
		recover()
		g()
	}()
	if count == 0 {
		panic("leave before the :=")
	}
	g, ok := func() {}, true
	_ = ok
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_lend_recover_param's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_recover_param) package{})).count
