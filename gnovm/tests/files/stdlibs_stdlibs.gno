// PKGPATH: gno.land/r/test
package test

import "time"

func main() {
	println(time.UTC == nil)
	time.UTC = nil
	println(time.UTC == nil)
	println("done")
}

// Error:
// cannot modify external-realm or non-realm object
