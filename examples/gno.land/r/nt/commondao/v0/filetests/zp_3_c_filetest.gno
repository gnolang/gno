// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_c_filetest

package zp_3_c_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateDissolutionProposal(cross(cur), 404, "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
}

// Error:
// DAO not found
