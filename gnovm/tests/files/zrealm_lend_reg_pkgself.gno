// PKGPATH: gno.land/r/tests/vm/lendregpkgself
package lendregpkgself

var count int
var self, selfUse = func() { count++ }, func() int { self(); return count }()

func main(cur realm) {
	println("x", count, selfUse)
}

// Error:
// gno.land/r/tests/vm/lendregpkgself: an object written by read-only code would be stored (do the write in a crossing function of gno.land/r/tests/vm/lendregpkgself, called with cross)
