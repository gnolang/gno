// PKGPATH: gno.land/r/zrealm_chalc_deferred_arg
package zrealm_chalc_deferred_arg

var count int

func attack() { run(func() { count++ }) }

// A deferred call through a reused parameter.
func run(g func()) {
	defer g()
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
// gno.land/r/zrealm_chalc_deferred_arg's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_deferred_arg) package{})).count
