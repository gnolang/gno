// PKGPATH: gno.land/r/test/commondao_must_execute
package commondao_must_execute

import (
	"gno.land/p/nt/commondao/v0"
)

func main(cur realm) {
	commondao.MustExecute(nil, cur)
}

// Error:
// executable proposal definition is nil
