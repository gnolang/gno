// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_e_filetest

package z_20_e_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// No-op register: the text kind is already seeded, so a council vote to
	// register it again is rejected at creation (a vote must be about a real
	// change).
	commondao.CreateRegisterKindProposal(cross(cur), daoID, "text")
}

// Error:
// proposal kind is already enabled
