// PKGPATH: gno.land/r/fvforgero
package fvforgero

import (
	"gno.land/r/tests/vm/fvforge"
	"gno.land/r/tests/vm/fvx"
)

func main(cur realm) {
	p := fvforge.Forge()
	println(fvx.Read(cross(cur), p))
}

// Error:
// illegal conversion to external realm type
