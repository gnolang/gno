// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_b_filetest

package zp_2_b_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateSubDAOProposal(cross(cur), 404, "", "")
}

// Error:
// DAO not found
