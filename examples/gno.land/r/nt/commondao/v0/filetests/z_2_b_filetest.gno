// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_2_b_filetest

package z_2_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")) // @test1
	commondao.New(cross(cur), "", "Purpose", "", "")
}

// Error:
// DAO name is empty
