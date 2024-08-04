package governor

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"std"
)

var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin

type ExecuteHook func() error

// governor keeper
type Governor struct {
	name string
	Membership
	proposals     *avl.Tree // ProposalId => Votes
	lastPropCount uint64
	ballots       *ballots
	gs            *GovernorSettings
	beforeExecute ExecuteHook
	afterExecute  ExecuteHook
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

// TODO: change to registerHook()
// for sendHook, etc
func (gvr *Governor) SetBeforeExecute(eh ExecuteHook) {
	gvr.beforeExecute = eh
}

func (gvr *Governor) SetAfterExecute(eh ExecuteHook) {
	gvr.afterExecute = eh
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

func (gvr *Governor) CastVoteWithMemo(propID string, vt string, memo string) error {
	voter := std.GetOrigCaller()
	voteType := str2VoteType(vt)
	return gvr.castVote(propID, voter, voteType, memo)
}

func (gvr *Governor) castVote(propID string, voter std.Address, vt VoteType, memo string) error {
	if !gvr.proposals.Has(propID) {
		panic("Governor: proposal not exist")
	}

	status := gvr.state(propID)
	if status != Active {
		panic("Governor: vote not currently active")
	}

	// xxx: how to choose the height? start? or arbitrary between the voting period?
	// snapshot mechanism is introduced to fix flash loan issues, that means the voting power of current
	// height is not qualified to vote, since it would mutate(flash) to make double vote.
	// so voting power of an account in `history` is required.
	// we can require voter has a voting power at proposal's starting height, since when the proposal is submitted is not 100%
	// predictable, so a deliberate `preparation` is impossible.
	// e.g. say a proposal voting period is from 100 ~ 110, we can use the snapshot at 100, that is the start of voting period.

	// Usually, we need to get voting power at or near a specific height.
	weight := gvr.VotingPowerFromMemberAt(voter, gvr.proposalSnapshot(propID))

	println("gvr cast vote, voter:", voter.String())
	println("gvr cast vote, weight:", weight)

	err := gvr.ballots.CastVote(propID, voter, vt, weight, memo)

	return err
}

func (gvr *Governor) RenderHome(propID string) string {
	str := "\n"
	str += ufmt.Sprintf("***************************Render Governor*************************** \n")
	str += ufmt.Sprintf("***gvr name***: %s\n", gvr.GetName())
	str += ufmt.Sprintf("***gvr version***: %s\n", gvr.GetVersion())

	// render Ms
	ms_str := gvr.RenderMembership()
	str += ufmt.Sprintf("***Membership***: %s\n", ms_str)

	//totalDeposit := std.Coin{Amount: 0, Denom: "ugnot"}
	//totalI, ok := gvr.deposits.Get(proposalId)
	//if ok {
	//	totalDeposit = totalI.(std.Coin)
	//}

	str += ufmt.Sprintf("*********************Proposal********************* \n")

	str += ufmt.Sprintf("***Current height is***: %d\n", std.GetHeight())

	var state ProposalState
	if propID != "" {
		state = gvr.state(propID)
		snapShot := gvr.proposalSnapshot(propID)
		deadLine := gvr.proposalDeadline(propID)
		str += ufmt.Sprintf("***Proposal ID is***: %s\n", propID)
		str += ufmt.Sprintf("***Proposal snapShot is***: %d\n", snapShot)
		str += ufmt.Sprintf("***Proposal deadLine is***: %d\n", deadLine)
	}

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

	if propID != "" {
		str += ufmt.Sprintf("***Proposal state is***: %s\n", state.String())
	}

	str += ufmt.Sprintf("***************************End*************************** \n")

	return str
}
