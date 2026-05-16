// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_b_filetest

package zp_0_b_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateMembersUpdateProposal(cross2(cur), 404, "", "")
}

// Error:
// DAO not found
