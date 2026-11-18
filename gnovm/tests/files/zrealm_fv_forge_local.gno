// PKGPATH: gno.land/r/fvforgelocal
package fvforgelocal

import (
	"gno.land/r/tests/vm/fvforge"
	"gno.land/r/tests/vm/fvx"
)

func main(cur realm) {
	p := fvforge.ForgeLocal()
	println(fvx.Read(cross(cur), p))
}

// Error:
// illegal conversion to external realm type
