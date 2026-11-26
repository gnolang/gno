// PKGPATH: gno.land/r/zrealm_chalc_escape_capture
package zrealm_chalc_escape_capture

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
// gno.land/r/zrealm_chalc_escape_capture's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_escape_capture) package{})).count
