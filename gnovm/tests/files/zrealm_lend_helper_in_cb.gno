// PKGPATH: gno.land/r/zrealm_lend_helper_in_cb
package zrealm_lend_helper_in_cb

var count int

func run(g func()) { g() }

// A local helper declared inside a callback that runs by value.
func attack() {
	run(func() {
		h := func() { count++ }
		h()
	})
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0
// after: 1
