// PKGPATH: gno.land/r/std
package std

import (
	"chain/runtime/unsafe"
)

func main(cur realm) {
	caller := unsafe.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
