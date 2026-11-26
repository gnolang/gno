// PKGPATH: gno.land/r/zrealm_lend_litparam_redecl
package zrealm_lend_litparam_redecl

var count int

// A literal's own parameter reused by :=, inside a named function.
func run(h func()) {
	f := func(g func()) {
		g()
		g, ok := func() {}, true
		_ = ok
	}
	f(h)
}

func attack() { run(func() { count++ }) }

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_lend_litparam_redecl's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_litparam_redecl) package{})).count
