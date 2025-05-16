// PKGPATH: gno.land/p/demo

package demo

import (
	"internal/bytealg"
)

func main() {
	println(bytealg.PrimeRK)
}

// Error:
// gno.land/p/demo/files/import12.gno:6:2: cannot import stdlib internal/ package outside of standard library
