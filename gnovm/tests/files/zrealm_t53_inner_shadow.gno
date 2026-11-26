// PKGPATH: gno.land/r/zrealm_t53_inner_shadow
package zrealm_t53_inner_shadow

var count int

var stored = func() { count += 100 }

// The inner g is a new variable of the inner block: a true local helper.
// The outer g is the parameter and still holds the stored value.
func run(g func()) {
	{
		g := func() { count++ }
		g()
	}
	println("inner:", count)
	g()
}

func main(cur realm) {
	println("before:", count)
	run(stored)
	println("after:", count)
}

// Output:
// before: 0
// inner: 1

// Error:
// gno.land/r/zrealm_t53_inner_shadow's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t53_inner_shadow) package{})).count
