// PKGPATH: gno.land/r/zrealm_chalc_redefine
package zrealm_chalc_redefine

var count int

func attack() { run(func() { count++ }) }

// A helper redeclared with a caller's value by a second :=.
func run(g func()) {
	f := func() {}
	f, x := g, 1
	_ = x
	f()
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_chalc_redefine's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_redefine) package{})).count
