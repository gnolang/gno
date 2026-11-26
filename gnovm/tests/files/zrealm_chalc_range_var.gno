// PKGPATH: gno.land/r/zrealm_chalc_range_var
package zrealm_chalc_range_var

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
// gno.land/r/zrealm_chalc_range_var's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_range_var) package{})).count
