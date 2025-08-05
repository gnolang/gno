// PKGPATH: gno.land/r/std
package std

import (
	"std"
)

func main(cur realm) {
	caller := std.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
