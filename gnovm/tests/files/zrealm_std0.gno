// PKGPATH: gno.land/r/std
package std

import "chain/runtime"

func main(cur realm) {
	caller := runtime.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
