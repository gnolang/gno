// PKGPATH: gno.land/r/tests/vm/recoverfvinitabort
package recoverfvinitabort

import "gno.land/p/demo/tests/recoverfv"

// Twin: at init a library method on the library's stored object can write
// it, so its panic is not recovered by this realm's init.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func init() {
	rec := try(func() { recoverfv.Shared().Set("x") })
	println("not reached", rec)
}

func main(cur realm) {
	println("main")
}

// Error:
// int is not of type string
