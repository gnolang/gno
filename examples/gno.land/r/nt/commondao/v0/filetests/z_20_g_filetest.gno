// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_g_filetest

package z_20_g_filetest

import (
	"chain"
	"chain/banker"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
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
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

	// Fund the sub-DAO and enable the execution kind on it.
	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 500)))
	{
		pID := commondao.CreateRegisterKindProposal(cross(cur), subID, "execution")
		commondao.Vote(cross(cur), subID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), subID, pID)
	}

	// The ancestor freezes the sub-DAO treasury.
	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The sub-DAO is frozen by a proper ancestor. Freeze means no funds can
	// leave, so an execution proposal that would drain the sub's own treasury
	// is refused at CREATE time (fail-fast), defeating any attempt to bypass
	// the ancestor's freeze through arbitrary execution.
	commondao.CreateExecutionProposal(cross(cur), subID, "Drain", "Body", func(_ int, sub realm) error {
		b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
		b.SendCoins(sub.Address(), user, chain.NewCoins(chain.NewCoin("ugnot", 500)))
		return nil
	})
}

// Error:
// commondao: treasury is frozen
