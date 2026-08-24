// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_1_b_filetest

package zp_1_b_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateTextProposal(cross(cur), 404, "", "", 0)
}

// Error:
// DAO not found
