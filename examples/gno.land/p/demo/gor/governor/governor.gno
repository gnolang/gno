package governor

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"std"
)

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

func (gvr *Governor) GetName() string {
	return gvr.name
}

func (gvr *Governor) GetVersion() string {
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

// TODO: other settings
// update all configs
func (gvr *Governor) SetVotingDelay(vd int64) {
	gvr.gs.setVotingDelay(vd)
}

// Address of contract through which the governor executes takes action.
// Will be overloaded by module that execute actions through another contract such as a timelock.
func (gvr *Governor) executor() std.Address {
	// TODO: return executor
	return ""
}

func (gvr *Governor) CastVote(propID string, vt string) error {
	voter := std.GetOrigCaller()
	voteType := str2VoteType(vt)
	return gvr.castVote(propID, voter, voteType)
}

func (gvr *Governor) castVote(propID string, account std.Address, vt VoteType) error {
	_, ok := gvr.proposals.Get(propID)
	if !ok {
		panic("Governor: proposal not exist")
	}

	status := gvr.state(propID)
	if status != Active {
		panic("Governor: vote not currently active")
	}

	err := gvr.tally.CastVote(propID, account, vt)

	return err
}

func (gvr *Governor) RenderHome(propID string) string {
	state := gvr.state(propID)
	snapShot := gvr.proposalSnapshot(propID)
	deadLine := gvr.proposalDeadline(propID)
	//totalDeposit := std.Coin{Amount: 0, Denom: "ugnot"}
	//totalI, ok := gvr.deposits.Get(proposalId)
	//if ok {
	//	totalDeposit = totalI.(std.Coin)
	//}

	str := "\n"
	str += ufmt.Sprintf("***gvr name***: %s\n", gvr.GetName())
	str += ufmt.Sprintf("***gvr version***: %s\n", gvr.GetVersion())
	str += ufmt.Sprintf("***Proposal state is***: %s\n", state.String())
	str += ufmt.Sprintf("***Proposal snapShot is***: %d\n", snapShot)
	str += ufmt.Sprintf("***Proposal deadLine is***: %d\n", deadLine)
	str += ufmt.Sprintf("***Current height is***: %d\n", std.GetHeight())
	//str += ufmt.Sprintf("* **Proposal deposit is***: %s\n", totalDeposit.String())
	//str += ufmt.Sprintf("* **isBurnDeposit :***: %t\n", isBurnDeposit)

	tallyRes, err := gvr.tally.Tally(propID)
	if err != nil {
		println(err.Error())
	} else {
		tallyRes.tr.Iterate("", "", func(key string, value interface{}) bool {
			str += ufmt.Sprintf("***Proposal vote %s is***: %d\n", key, value.(uint64))
			return false
		})
	}

	return str
}
