// PKGPATH: gno.land/p/demo

package demo

import (
	"internal/bytealg"
)

func main() {
	println(bytealg.PrimeRK)
}

// Error:
// gno.land/p/demo/import12.gno:6:2-20: cannot import stdlib internal/ package outside of standard library
