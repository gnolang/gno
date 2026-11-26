// PKGPATH: gno.land/r/zrealm_lend_param_redecl
package zrealm_lend_param_redecl

var count int

func attack() { run(func() { count++ }) }

// g is the parameter; the := below reuses it, so the birth rule counts it once.
func run(g func()) {
	g()
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
// gno.land/r/zrealm_lend_param_redecl's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_param_redecl) package{})).count
