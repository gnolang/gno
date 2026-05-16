// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_5_b_filetest

package z_5_b_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const user = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Vote(cross2(cur), 404, 1, pcommondao.ChoiceYes, "")
}

// Error:
// DAO not found
