// PKGPATH: gno.land/r/t75b
package t75b

import lib "gno.land/p/demo/tests/t75ifl"

var count int

// A literal lent in place to a method called on an interface value of a
// directly imported /p/ type writes.
func main(cur realm) {
	var e lib.Eacher = lib.Impl{}
	e.Each([]int{1, 2, 3}, func(x int) { count += x })
	println(count)
}

// Output:
// 6
