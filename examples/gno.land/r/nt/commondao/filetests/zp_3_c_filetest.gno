// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_c_filetest

package zp_3_c_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.CreateDissolutionProposal(cross(cur), 404, "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx")
}

// Error:
// DAO not found
