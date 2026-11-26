// PKGPATH: gno.land/r/zrealm_lend_range_var
package zrealm_lend_range_var

var count int

func attack() { run(func() { count++ }) }

func run(gs ...func()) {
	for _, f := range gs {
		f()
	}
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_lend_range_var's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_range_var) package{})).count
