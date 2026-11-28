// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_b_filetest

package z_6_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const owner = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.Execute(cross(cur), 404, 1)
}

// Error:
// DAO not found
