package governor

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/gor/timers"
	"gno.land/p/demo/ufmt"
	"std"
	"time"
)

type ProposalState int

const (
	Pending ProposalState = iota
	Active
	Canceled
	Defeated
	Vetoed
	Succeeded
	Queued
	Expired
	Executed
	QuorumNotReach
)

func (ps ProposalState) String() string {
	switch ps {
	case Pending:
		return "Pending"
	case Active:
		return "Active"
	case Canceled:
		return "Canceled"
	case Defeated:
		return "Defeated"
	case Succeeded:
		return "Succeeded"
	case Queued:
		return "Queued"
	case Expired:
		return "Expired"
	case Executed:
		return "Executed"
	default:
		return ufmt.Sprintf("%d", int(ps))
	}
}

// state of proposal
type proposalStatus struct {
	voteStart *timers.BlockNumber
	voteEnd   *timers.BlockNumber
	//executed  bool
	//canceled bool
	state ProposalState
}

func NewProposalCore() *proposalStatus {
	return &proposalStatus{
		voteStart: &timers.BlockNumber{},
		voteEnd:   &timers.BlockNumber{},
	}
}

type proposalMeta struct {
	title   string
	author  std.Address
	detail  string
	summary string
	url     string
}

type ProposalKind uint8

const (
	TextProposal = iota
	ParamChangeProposal
	SendProposal
	ExecutableProposal
)

func (pk ProposalKind) String() string {
	switch pk {
	case TextProposal:
		return "text_proposal"
	case ParamChangeProposal:
		return "param_change_proposal"
	case SendProposal:
		return "send_proposal"
	case ExecutableProposal:
		return "executable_proposal"
	default:
		panic("no such kind of proposal type")
	}
}

func str2ProposalKind(str string) ProposalKind {
	switch str {
	case "text_proposal":
		return TextProposal
	case "param_change_proposal":
		return ParamChangeProposal
	case "send_proposal":
		return SendProposal
	case "executable_proposal":
		return ExecutableProposal
	default:
		panic("invalid proposal type")
	}
}

type proposal struct {
	proposalID  string
	propKind    ProposalKind
	payee       std.Address
	amount      std.Coins
	meta        *proposalMeta
	core        *proposalStatus
	deposits    *avl.Tree // proposalId <=> Amount, total amount
	depositList *avl.Tree // proposalId <=> (caller <=> amount)
	proposer    std.Address
	createAt    time.Time
	// params []string
	// callable /r/demo/xxx
}

func (gvr *Governor) getProposal(propID string) (*proposal, error) {
	propI, ok := gvr.proposals.Get(propID)
	if !ok {
		return nil, ErrProposalNotExist
	}
	proposal := propI.(*proposal)
	return proposal, nil
}

func (gvr *Governor) proposalSnapshot(propID string) int64 {
	proposalI, ok := gvr.proposals.Get(propID)
	if !ok {
		panic("Governor: proposal not exist")
	}
	proposal := proposalI.(*proposal)
	return proposal.core.voteStart.GetDeadline()
}

func (gvr *Governor) proposalDeadline(propID string) int64 {
	proposalI, found := gvr.proposals.Get(propID)
	if !found {
		panic("Governor: proposal not exist")
	}
	proposal := proposalI.(*proposal)
	return proposal.core.voteEnd.GetDeadline()
}

func (gvr *Governor) Propose(propID string, title, detail, summary, url string, payee std.Address, amount std.Coins, propKind string) string {
	prop := &proposal{
		proposalID: propID,
		propKind:   str2ProposalKind(propKind),
		proposer:   std.GetOrigCaller(), // EOA account
		payee:      payee,
		amount:     amount,
		core:       NewProposalCore(),
	}
	meta := &proposalMeta{
		title:   title,
		detail:  detail,
		summary: summary,
		url:     url,
	}
	prop.meta = meta

	if gvr.proposals.Has(propID) {
		panic("proposal already exist")
	}

	snapshot := std.GetHeight() + gvr.gs.getVotingDelay() //start
	deadline := snapshot + gvr.gs.getVotingPeriod()       // end

	prop.core.voteStart.SetDeadline(snapshot)
	prop.core.voteEnd.SetDeadline(deadline)

	gvr.proposals.Set(propID, prop)
	// TODO: emit ProposalCreatedEvent
	return propID
}

func (gvr *Governor) Execute(propID string) {
	if !gvr.proposals.Has(propID) {
		panic(("proposal not exist"))
	}
	proposal, _ := gvr.getProposal(propID)

	// check state
	status := gvr.state(propID)
	if status != Succeeded && status != Queued {
		panic("Governor: proposal not successful")
	}

	// TODO: call callable

	if proposal.propKind == SendProposal {
		// get banker
		// send coin to payee
		banker := std.GetBanker(std.BankerTypeOrigSend)
		pkgaddr := std.GetOrigPkgAddr()
		// TODO: use std.Coins constructors, this isn't generally safe.
		banker.SendCoins(pkgaddr, proposal.payee, proposal.amount)
	}

	// TODO: do the transfer to payees

	proposal.core.state = Executed
	gvr.proposals.Set(propID, proposal) // update state
	// TODO: emit ProposalExecuted(proposalId);
}

// Internal cancel mechanism: locks up the proposal timer, preventing it from being re-submitted. Marks it as
// canceled to allow distinguishing it from executed proposals.
func (gvr *Governor) cancel(propID string) string {
	proposal, err := gvr.getProposal(propID)
	if err != nil {
		panic(err.Error())
	}

	status := gvr.state(propID)

	if status == Canceled || status == Expired || status == Executed {
		panic("Governor: proposal not active")
	}

	proposal.core.state = Canceled
	gvr.proposals.Set(propID, proposal) // update state

	// TODO: emit ProposalCanceled(proposalId);
	return propID
}

func (gvr *Governor) state(propID string) ProposalState {
	if !gvr.proposals.Has(propID) {
		panic(("proposal not exist"))
	}
	proposal, _ := gvr.getProposal(propID)

	if proposal.core.state == Executed {
		return Executed
	}

	if proposal.core.state == Canceled {
		return Canceled
	}

	start := gvr.proposalSnapshot(propID)
	deadline := gvr.proposalDeadline(propID)

	height := std.GetHeight()

	if height <= start {
		return Pending
	}

	if height <= deadline {
		return Active
	}

	// other than executed, canceled, pending, active
	// check voting status
	if gvr.quorumReached(gvr.votingPower.Total()) {
		tallyRes, err := gvr.tally.Tally(propID)
		if err != nil {
			panic(err)
		}
		if gvr.vetoReached(tallyRes) {
			return Vetoed
		} else if !gvr.voteSucceeded(tallyRes) {
			return Defeated
		}
	} else {
		return QuorumNotReach
	}
	return Succeeded
}

//func (gvr *Governor) RenderProposal(propID string) string {
//
//}
