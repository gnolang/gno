// PKGPATH: gno.land/r/ctw_recover
// A read-only frame is no realm boundary: a panic raised in another realm's
// read-only code is recovered by the caller, which keeps its own authority.
package ctw_recover

import "gno.land/r/tests/vm/ctw"

var Count int

func main(cur realm) {
	func() {
		defer func() { println("recovered:", recover()) }()
		ctw.Boom()
	}()
	Count++
	println("count:", Count)
	println("value:", ctw.RecoverValue(cross(cur)))
}

// Output:
// recovered: boom
// count: 1
// value: boom
