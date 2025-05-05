// PKGPATH: gno.land/r/test
package test

import "time"

func main() {
	crossing()

	println(time.UTC == nil)
	time.UTC = nil
	println(time.UTC == nil)
	println("done")
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(time) package{})).UTC
