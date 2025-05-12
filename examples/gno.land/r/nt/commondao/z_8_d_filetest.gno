// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"gno.land/r/nt/commondao"
)

func main() {
	crossing()

	commondao.GetSubDAO(1, "foo/bar")
}

// Error:
// unauthorized
