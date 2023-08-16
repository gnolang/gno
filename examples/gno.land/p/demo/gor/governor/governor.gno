package governor

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"std"
)

var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin

type Gvr interface {
	Membership
}

// governor keeper
type Governor struct {
	name string
	Membership
	proposals *avl.Tree // ProposalId => Votes
	ballots   *ballots
	gs        *GovernorSettings
}

func NewGovernor(name string, ms Membership, gs *GovernorSettings) *Governor {
	return &Governor{
		name:       name,
		proposals:  avl.NewTree(),
		ballots:    NewBallots(),
		gs:         gs,
		Membership: ms,
	}
}

func (gvr *Governor) GetName() string {
	return gvr.name
}

func (gvr *Governor) GetVersion() string {
	return "1"
}

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

	err := gvr.ballots.CastVote(propID, account, vt)

	return err
}

func (gvr *Governor) RenderHome(propID string) string {
	str := "\n"
	str += ufmt.Sprintf("***gvr name***: %s\n", gvr.GetName())
	str += ufmt.Sprintf("***gvr version***: %s\n", gvr.GetVersion())

	// render Ms
	ms_str := gvr.RenderMembership()
	str += ufmt.Sprintf("***Ms***: %s\n", ms_str)

	state := gvr.state(propID)
	snapShot := gvr.proposalSnapshot(propID)
	deadLine := gvr.proposalDeadline(propID)
	//totalDeposit := std.Coin{Amount: 0, Denom: "ugnot"}
	//totalI, ok := gvr.deposits.Get(proposalId)
	//if ok {
	//	totalDeposit = totalI.(std.Coin)
	//}

	str += ufmt.Sprintf("***Current height is***: %d\n", std.GetHeight())
	str += ufmt.Sprintf("***Proposal snapShot is***: %d\n", snapShot)
	str += ufmt.Sprintf("***Proposal deadLine is***: %d\n", deadLine)
	//str += ufmt.Sprintf("* **Proposal deposit is***: %s\n", totalDeposit.String())
	//str += ufmt.Sprintf("* **isBurnDeposit :***: %t\n", isBurnDeposit)

	tallyRes, err := gvr.ballots.Tally(propID)
	if err != nil {
		println(err.Error())
	} else {
		tallyRes.tr.Iterate("", "", func(key string, value interface{}) bool {
			str += ufmt.Sprintf("***Proposal vote %s is***: %d\n", key, int64(value.(int32)))
			return false
		})
	}

	str += ufmt.Sprintf("***Proposal state is***: %s\n", state.String())

	return str
}
