// PKGPATH: gno.land/r/demo/main

package main

import "gno.land/r/demo/closuretest"

func main(cur realm) {
	println(closuretest.Accumulate(cross(cur), 7))
	println(closuretest.Accumulate(cross(cur), 8))
	println(closuretest.Render(""))
}

// Output:
// history length=1
// history length=2
// closuretest: count=0 history=2
