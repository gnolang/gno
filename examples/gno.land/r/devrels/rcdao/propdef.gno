package rcdao

import (
	"std"
	"time"

	"gno.land/p/nt/commondao"
	"gno.land/r/sys/users"
)

type EventProposal struct {
	title           string
	body            string
	location        string // ie Berlin, Germany
	requestedBudget std.Coins
	beneficiary     std.Address // after the prop passes, coins are sent here
}

func (ep EventProposal) Title() string            { return ep.title }
func (ep EventProposal) Body() string             { return ep.body }
func (ep EventProposal) Location() string         { return ep.location }
func (EventProposal) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (ep EventProposal) Render() string {
	out := ep.body + "\n\n"
	out += ep.location + "\n\n"
	out += "Asking amount: " + ep.requestedBudget.String() + "\n\n"

	name := "Author: " + ep.beneficiary.String()
	user := users.ResolveAddress(ep.beneficiary)
	if user != nil {
		name = "Author: " + user.RenderLink("")
	}

	out += name
	return out
}

func (EventProposal) Tally(r commondao.ReadonlyVotingRecord, set commondao.MemberSet) (bool, error) {
	_, passed := commondao.SelectChoiceBySuperMajority(r, set.Size())
	return passed, nil
}

func (ep EventProposal) Execute() error {
	crossing()

	bank := std.NewBanker(std.BankerTypeRealmSend)
	treasuryCoins := bank.GetCoins(std.CurrentRealm().Address()) // DAO treasury coins

	for i := 0; i < len(ep.requestedBudget); i++ {
		coin := ep.requestedBudget[i]
		if treasuryCoins.AmountOf(coin.Denom) >= coin.Amount {
			panic("cannot pay out requested budget, too little in the treasury")
		}
	}

	return nil
}
