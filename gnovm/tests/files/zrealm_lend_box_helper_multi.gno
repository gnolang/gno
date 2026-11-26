// PKGPATH: gno.land/r/zrealm_lend_box_helper_multi
package zrealm_lend_box_helper_multi

var count int

func attack() {
	var f func()
	var x int
	f = func() { count++ }
	f, x = func() { count += 100 }, 1
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
// gno.land/r/zrealm_lend_box_helper_multi's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_box_helper_multi) package{})).count
