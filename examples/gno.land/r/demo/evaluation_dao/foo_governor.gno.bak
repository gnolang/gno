package foo_governor

import (
	"gno.land/p/demo/gor/governor"
	"gno.land/p/demo/grc/exts/grc20votes"
	"std"
)

var g20v *grc20votes.Grc20Votes
var gvr *governor.Governor

func init() {
	g20v = grc20votes.NewGrc20Votes("gov", "GOV", 10)
	minDeposit := std.Coin{
		Denom:  "ugnot",
		Amount: 64000000,
	}
	// new votes counter
	// voting delay, voting period, minDeposit, proposalThreshold
	gcs := governor.NewGovernorSettings(0, 5, minDeposit, 0, 40, 67, 34)

	// quorum, veto, succeed
	//tally := governor.NewGovernorCountingBase(40, 34, 67)

	gvr = governor.NewGovernor("test_governor", tally, gcs)
}

func SetVotingDelay(vd int64) {
	gvr.SetVotingDelay(14)
}

func Mint(to std.Address, amount uint64) error {
	err := g20v.Mint(to, amount)
	if err != nil {
		return err
	}
	return nil
}

func TotalSupply() uint64 {
	return g20v.TotalSupply()
}

func Transfer(from std.Address, to std.Address, amount uint64) {
	g20v.Transfer(from, to, amount)
}

// NOTE: explicit delegate, includ self delegate
// is this same with transfer
// func Delegate
// NOTE: a proposal should be concrete that what is proposed is what will be executed.
// In solidity, this is implemented by `target.call`, using same parameters in propose and execution process,
// in gno, maybe we could simply propose a "execution code" that is visible to voters, e.g. `https://test2.gno.land/r/boards/misc.gno`,
// which is clear what will be proposed and what will be executed.

// also, reentrancy problem should also be cosidered.
func Propose(proposal *governor.Proposal) string {
	proposalId := gvr.Propose(proposal)
	println("done propose")
	return proposalId
}

func CastVote(proposalId string, support governor.VoteType) uint64 {
	weight := gvr.CastVote(proposalId, support)
	return weight
}

func Deposit(proposalId string) {
	std.AssertOriginCall()
	gvr.Deposit(proposalId)
}

func UnDeposit(proposalId string) error {
	std.AssertOriginCall()
	err := gvr.Undeposit(proposalId)
	if err != nil {
		return err
	}
	return nil
}

func GetDeposit(proposalId string) (std.Coin, bool) {
	return gvr.GetDeposit(proposalId), false
}

func GetVotes(account std.Address, blockNumber int64) uint64 {
	return gvr.GetVotes(account, blockNumber)
}

func BalanceOf(owner std.Address) (uint64, error) {
	b, err := g20v.BalanceOf(owner)
	if err != nil {
		return 0, err
	} else {
		return b, nil
	}
}

func Tally(proposalId string) (state governor.ProposalState, isBurnDeposit bool, tResult governor.TallyResult) {
	state, isBurnDeposit, tResult = gvr.Tally(proposalId)
	return
}

func RenderHome(proposalId string) string {
	return gvr.RenderHome(proposalId)
}
