// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_h_filetest

package zp_2_h_filetest

import (
	"strings"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	name  = strings.Repeat("A", 61)
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(user))

}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross(cur), daoID, name, "Purpose", "", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// DAO name is too long, max length is 60 characters
