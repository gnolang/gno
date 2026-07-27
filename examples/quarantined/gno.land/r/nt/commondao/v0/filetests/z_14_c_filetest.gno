// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_14_c_filetest

package z_14_c_filetest

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
	rootID = commondao.New(cross(cur), "Root", "", commondao.AllowChildren(true))
	subID = commondao.NewSubDAO(cross(cur), "Sub", rootID, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
}

// Error:
// treasury proposals are not enabled
