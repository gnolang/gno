// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_5_a_filetest

package z_5_a_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	pID   uint64
	vote  pdao.VoteChoice = pdao.ChoiceYes
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "", string(user))

	// Create a new proposal
	testing.SetRealm(testing.NewUserRealm(user))
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

func main(cur realm) {
	// User must be the caller to Vote()
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Vote(cross(cur), daoID, pID, vote, "")

	p, _ := commondao.GetView(daoID).GetProposal(pID)
	record := p.VotingRecord()
	if record.Size() != 1 {
		panic("expected a single vote")
	}

	println(record.HasVoted(user))
	record.Iterate(0, record.Size(), false, func(v pdao.Vote) bool {
		println(v.Choice == vote)
		return false
	})
}

// Output:
// true
// true
