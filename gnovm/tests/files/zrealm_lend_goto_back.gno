// PKGPATH: gno.land/r/zrealm_lend_goto_back
package zrealm_lend_goto_back

var count int

func attack() { run(func() { count++ }) }

// A bare var assigned once from a literal, with a goto back over the call.
func run(g func()) {
	n := 0
	var f func()
L:
	if n == 1 {
		f()
		return
	}
	f = func() { g() }
	n++
	goto L
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_lend_goto_back's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_goto_back) package{})).count
