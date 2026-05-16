// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_b_filetest

package z_6_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.Execute(cross2(cur), 404, 1)
}

// Error:
// DAO not found
