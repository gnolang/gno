// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_2_b_filetest

package z_2_b_filetest

import (
	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	commondao.New(cross2(cur), "")
}

// Error:
// DAO name is empty
