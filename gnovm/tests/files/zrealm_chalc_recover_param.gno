// PKGPATH: gno.land/r/zrealm_chalc_recover_param
package zrealm_chalc_recover_param

var count int

func attack() { run(func() { count++ }) }

// The parameter is reused by := ; a panic skips it and the deferred literal calls the caller's value.
func run(g func()) {
	defer func() {
		recover()
		g()
	}()
	if count == 0 {
		panic("leave before the :=")
	}
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
// gno.land/r/zrealm_chalc_recover_param's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalc_recover_param) package{})).count
