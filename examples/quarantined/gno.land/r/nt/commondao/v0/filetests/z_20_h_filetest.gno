// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_h_filetest

package z_20_h_filetest

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
	execID uint64
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

	// Pass an execution proposal that drains the sub's own treasury, but do
	// not execute it yet. It is created while the sub is still unfrozen.
	execID = commondao.CreateExecutionProposal(cross(cur), subID, "Drain", "Body", func(_ int, sub realm) error {
		b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
		b.SendCoins(sub.Address(), user, chain.NewCoins(chain.NewCoin("ugnot", 500)))
		return nil
	})
	commondao.Vote(cross(cur), subID, execID, pdao.ChoiceYes, "")

	// The ancestor freezes the sub-DAO treasury before the execution runs.
	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The standing passed execution proposal fails cleanly once the treasury
	// is frozen: the freeze check reruns inside Execute, so the closure never
	// runs and no funds leave (StatusFailed, not a stuck-Passed tx panic).
	commondao.Execute(cross(cur), subID, execID)

	view := commondao.GetView(subID)
	p, _ := view.GetProposal(execID)
	println("status:", string(p.Status()))
	println("reason:", p.StatusReason())
	println("frozen:", view.IsTreasuryFrozen())
	println("treasury:", banker.NewReadonlyBanker().GetCoins(view.Address()).String())
}

// Output:
// status: failed
// reason: commondao: treasury is frozen
// frozen: true
// treasury: 500ugnot
