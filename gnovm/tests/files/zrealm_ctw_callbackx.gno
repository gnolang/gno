// PKGPATH: gno.land/r/ctw_callbackx
// A crossing callback called with cross writes its own realm's state.
package ctw_callbackx

import "gno.land/r/tests/vm/ctw"

var sum int

func main(cur realm) {
	ctw.ForEachX(cross(cur), func(cur realm, n int) { sum += n })
	println(sum)
}

// Output:
// 1
