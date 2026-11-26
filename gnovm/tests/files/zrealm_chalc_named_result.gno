// PKGPATH: gno.land/r/zrealm_chalc_named_result
package zrealm_chalc_named_result

var count int

func attack() { run(func() { count++ }) }

// return h fills the named result g; the deferred in-place literal calls it.
func run(h func()) (g func()) {
	g, ok := func() {}, true
	_ = ok
	defer func() { g() }()
	return h
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_chalc_named_result's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_named_result) package{})).count
