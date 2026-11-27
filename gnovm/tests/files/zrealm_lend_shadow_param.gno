// PKGPATH: gno.land/r/zrealm_lend_shadow_param
package zrealm_lend_shadow_param

var count int

func attack() { run(func() { count++ }) }

// An inner helper shadows the parameter's name; the outer call is the parameter.
func run(g func()) {
	{
		g := func() {}
		g()
	}
	g()
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0
// after: 1
