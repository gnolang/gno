// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_d_filetest

package z_9_d_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", string(owner))

	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

func main(cur realm) {
	// Owner is a council member but not the proposal creator
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.Withdraw(cross(cur), daoID, pID)
}

// Error:
// only the proposal creator can withdraw it
