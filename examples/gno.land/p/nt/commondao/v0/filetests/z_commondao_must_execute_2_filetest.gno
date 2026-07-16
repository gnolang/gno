// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/commondao/v0"
)

func main(cur realm) {
	commondao.MustExecute(nil, cur)
}

// Error:
// executable proposal definition is nil
