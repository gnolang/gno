// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_c_filetest

package zp_3_c_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateDissolutionProposal(cross2(cur), 404)
}

// Error:
// DAO not found
