// PKGPATH: gno.land/r/zrealm_chalb_helper_nested
package zrealm_chalb_helper_nested

var count int

func attack() {
	var f func()
	f = func() { count++ }
	func() { f = func() { count += 100 } }()
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
// gno.land/r/zrealm_chalb_helper_nested's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalb_helper_nested) package{})).count
