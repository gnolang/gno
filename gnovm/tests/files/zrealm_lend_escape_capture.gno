// PKGPATH: gno.land/r/zrealm_lend_escape_capture
package zrealm_lend_escape_capture

var count int

func run(g func()) { g() }

// A helper captured by a literal that leaves as a value.
func attack() {
	h := func() { count++ }
	h()
	count = 0
	run(func() { h() })
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_lend_escape_capture's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_escape_capture) package{})).count
