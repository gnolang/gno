// PKGPATH: gno.land/r/zrealm_chalb_helper_ref
package zrealm_chalb_helper_ref

var count int

func attack() {
	var f func()
	f = func() { count++ }
	p := &f
	*p = func() { count += 100 }
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
// gno.land/r/zrealm_chalb_helper_ref's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalb_helper_ref) package{})).count
