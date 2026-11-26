// PKGPATH: gno.land/r/zrealm_lend_inplace_in_cb
package zrealm_lend_inplace_in_cb

var count int

func run(g func()) { g() }

// A literal called where written, inside a callback that runs by value.
func attack() {
	run(func() {
		func() { count++ }()
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
