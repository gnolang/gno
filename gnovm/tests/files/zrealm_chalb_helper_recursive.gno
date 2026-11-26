// PKGPATH: gno.land/r/zrealm_chalb_helper_recursive
package zrealm_chalb_helper_recursive

var count int

// The regexp shape: a recursive local helper is a helper, so it writes.
func run() {
	var f func(n int)
	f = func(n int) {
		if n > 0 {
			count++
			f(n - 1)
		}
	}
	f(3)
}

func main(cur realm) {
	println("before:", count)
	run()
	println("after:", count)
}

// Output:
// before: 0
// after: 3
