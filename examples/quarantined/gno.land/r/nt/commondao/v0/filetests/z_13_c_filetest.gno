// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_13_c_filetest

package z_13_c_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	rootID uint64
	subID  uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "",
		commondao.AllowChildren(true),
		commondao.AllowDissolutionProposals(true),
	)
	subID = commondao.NewSubDAO(cross(cur), "Sub", rootID, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateDissolutionProposal(cross(cur), subID, owner)
}

// Error:
// sub-DAO dissolution sweeps to the parent DAO; destination must be empty
