package governor

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"std"
)

//type Governor struct {
//	members *avl.Tree
//}

var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin
// governor keeper
type Governor struct {
	name        string
	proposals   *avl.Tree // ProposalId => Votes
	votingPower VotingPower
	tally       Tally
	gs          *GovernorSettings
}

func NewGovernor(name string, vp VotingPower, tally Tally, gs *GovernorSettings) *Governor {
	return &Governor{
		name:        name,
		proposals:   avl.NewTree(),
		tally:       tally,
		gs:          gs,
		votingPower: vp,
	}
}

func (gvr *Governor) getName() string {
	return gvr.name
}

func (gvr *Governor) getVersion() string {
	return "1"
}

//func (gvr *Governor) Tally(proposalId string) (ProposalState, *TallyResult, error) {
//	println("tally------")
//	state := gvr.state(proposalId)
//	tallyRes, err := gvr.tally.Tally(proposalId)
//	if err != nil {
//		return Pending, nil, err
//	}
//	return state, tallyRes, nil
//}

func (gvr *Governor) SetVotingDelay(vd int64) {
	gvr.gs.setVotingDelay(vd)
}

// TODO: other settings

// Address of contract through which the governor executes takes action.
// Will be overloaded by module that execute actions through another contract such as a timelock.
func (gvr *Governor) executor() std.Address {
	// TODO: return executor
	return ""
}

func (gvr *Governor) CastVote(proposalId string, vt VoteType) error {
	voter := std.GetOrigCaller()
	return gvr.castVote(proposalId, voter, vt)
}

func (gvr *Governor) castVote(proposalId string, account std.Address, vt VoteType) error {
	_, ok := gvr.proposals.Get(proposalId)
	if !ok {
		panic("Governor: proposal not exist")
	}

	status := gvr.state(proposalId)
	if status != Active {
		panic("Governor: vote not currently active")
	}

	gvr.tally.CastVote(proposalId, account, vt)

	return nil
}

func (gvr *Governor) RenderHome(proposalId string) string {
	state := gvr.state(proposalId)
	snapShot := gvr.proposalSnapshot(proposalId)
	deadLine := gvr.proposalDeadline(proposalId)
	//totalDeposit := std.Coin{Amount: 0, Denom: "ugnot"}
	//totalI, ok := gvr.deposits.Get(proposalId)
	//if ok {
	//	totalDeposit = totalI.(std.Coin)
	//}

	tallyRes, err := gvr.tally.Tally(proposalId)
	if err != nil {
		panic(err)
	}

	str := "\n"
	str += ufmt.Sprintf("* **Proposal state is***: %s\n", state.String())
	str += ufmt.Sprintf("* **Proposal snapShot is***: %d\n", snapShot)
	str += ufmt.Sprintf("* **Proposal deadLine is***: %d\n", deadLine)
	str += ufmt.Sprintf("* **Current height is***: %d\n", std.GetHeight())
	//str += ufmt.Sprintf("* **Proposal deposit is***: %s\n", totalDeposit.String())
	//str += ufmt.Sprintf("* **isBurnDeposit :***: %t\n", isBurnDeposit)

	tallyRes.tr.Iterate("", "", func(key string, value interface{}) bool {
		str += ufmt.Sprintf("* **Proposal vote %s is***: %d\n", key, value.(uint64))
		return false
	})

	return str
}
