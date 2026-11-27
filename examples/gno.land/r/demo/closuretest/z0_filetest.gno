// PKGPATH: gno.land/r/demo/main

package main

import "gno.land/r/demo/closuretest"

func main(cur realm) {
	println(closuretest.Step(cross(cur)))
	println(closuretest.Step(cross(cur)))
	println(closuretest.Render(""))
}

// Output:
// count=3
// count=6
// closuretest: count=6 history=0
