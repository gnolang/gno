// PKGPATH: gno.land/r/zrealm_chalb_helper_if
package zrealm_chalb_helper_if

var count int

func attack() {
	var f func()
	if count == 0 {
		f = func() { count++ }
	}
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
// gno.land/r/zrealm_chalb_helper_if's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_chalb_helper_if) package{})).count
