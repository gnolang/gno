// PKGPATH: gno.land/r/test
package test

import "time"

func main(cur realm) {
	println(time.UTC == nil)
	time.UTC = nil
	println(time.UTC == nil)
	println("done")
}

// Error:
// gno.land/r/test/stdlibs.gno:8:2-10: cannot directly mutate time.UTC from gno.land/r/test
