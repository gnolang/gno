// PKGPATH: gno.land/r/zrealm_chalb_helper_param
package zrealm_chalb_helper_param

var count int

func attack() { run(func() { count++ }) }

func run(g func()) {
	var f func()
	f = g
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
// gno.land/r/zrealm_chalb_helper_param's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalb_helper_param) package{})).count
