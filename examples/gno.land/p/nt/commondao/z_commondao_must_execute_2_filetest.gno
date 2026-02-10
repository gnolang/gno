// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/commondao"
)

func main() {
	commondao.MustExecute(nil)
}

// Error:
// executable proposal definition is nil
