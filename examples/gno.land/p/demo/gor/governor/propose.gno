package governor

import "gno.land/p/demo/gor/timers"
import std "std"
import "gno.land/p/demo/ufmt"

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
type ProposalCore struct {
	voteStart *timers.BlockNumber
	voteEnd   *timers.BlockNumber
	//executed  bool
	//canceled bool
	state ProposalState
}

type ProposalMeta struct {
	Title   string
	Author  string
	Detail  string
	Summary string
	Url     string
}

type ProposalKind uint8

const (
	TextProposal = iota
	ParamChangeProposal
	SendProposal
	ExecutableProposal
)

type Proposal struct {
	ProposalKind ProposalKind
	Payee        std.Address
	Meta         *ProposalMeta
	Core         *ProposalCore
	// params []string
	// callable /r/demo/xxx
	deposits    *avl.Tree // proposalId <=> Amount, total amount
	depositList *avl.Tree // proposalId <=> (caller <=> amount)
}

// generate proposalId
func hashProposal(prop *Proposal) string {
	// TODO: get real hash of `prop`
	return "test" + prop.Meta.Title
}


func (gvr *Governor) getProposal(proposalId string) (*Proposal, error){
	propI, ok := gvr.proposals.Get(proposalId)
	if !ok{
		return nil, ErrProposalNotExist
	}
	proposal := propI.(*Proposal)
	return proposal, nil
}


func (gvr *Governor) proposalSnapshot(proposalId string) int64 {
	proposalI, ok := gvr.proposals.Get(proposalId)
	if !ok {
		panic("Governor: proposal not exist")
	}
	proposal := proposalI.(*Proposal)
	return proposal.Core.voteStart.GetDeadline()
}

func (gvr *Governor) proposalDeadline(proposalId string) int64 {
	proposalI, found := gvr.proposals.Get(proposalId)
	if !found {
		panic("Governor: proposal not exist")
	}
	proposal := proposalI.(*Proposal)
	return proposal.Core.voteEnd.GetDeadline()
}

func (gvr *Governor) Propose(prop *Proposal) string {
	std.AssertOriginCall()

	proposalId := hashProposal(prop) // only hash meta

	_, ok := gvr.getProposal.Get(proposalId)
	if ok{
		panic("proposal already exist")
	}

	snapshot := std.GetHeight() + gvr.gs.getVotingDelay() //start
	deadline := snapshot + gvr.gs.getVotingPeriod()       // end

	prop.Core.voteStart.SetDeadline(snapshot)
	prop.Core.voteEnd.SetDeadline(deadline)

	gvr.proposals.Set(proposalId, prop)
	// TODO: emit ProposalCreatedEvent
	return proposalId
}

func (gvr *Governor) Execute(proposalId string) {
	proposal, err := gvr.getProposal(proposalId)
	if err != nil{
		panic(err.Error())
	}

	// check state
	status := gvr.state(proposalId)
	if status != Succeeded && status != Queued {
		panic("Governor: proposal not successful")
	}

	// TODO: call callable
	// TODO: do the transfer to payees

	proposal.Core.state = Executed
	gvr.proposals.Set(proposalId, proposal) // update state
	// TODO: emit ProposalExecuted(proposalId);
}

// Internal cancel mechanism: locks up the proposal timer, preventing it from being re-submitted. Marks it as
// canceled to allow distinguishing it from executed proposals.
func (gvr *Governor) cancel(proposal *Proposal) string {
	// TODO: descriptionHash?
	proposalId := hashProposal(proposal)
	proposal, err := gvr.getProposal(proposalId)
	if err != nil{
		panic(err.Error())
	}

	status := gvr.state(proposalId)

	if status == Canceled || status == Expired || status == Executed) {
		panic("Governor: proposal not active")
	}

	proposal.Core.state = Canceled
	gvr.proposals.Set(proposalId, proposal) // update state

	// TODO: emit ProposalCanceled(proposalId);
	return proposalId
}

func (gvr *Governor) state(proposalId string) ProposalState {
	proposal, err := gvr.getProposal(proposalId)
	if err != nil{
		panic(err.Error())
	}

	if proposal.Core.state == Executed {
		return Executed
	}

	if proposal.Core.state == Canceled {
		return  Canceled
	}

	start := gvr.proposalSnapshot(proposalId)
	deadline := gvr.proposalDeadline(proposalId)


	height := std.GetHeight()

	if height <= start {
		return  Pending
	}


	if height <= deadline {
		return  Active
	}

	// other than executed, canceled, pending, active
	// check voting status
	if gvr.quorumReached(gvr.govToken.TotalSupply()) {
		tallyRes, err := gvr.tally.Tally(proposalId)
		if err != nil {
			panic(err)
		}
		if gvr.vetoReached(tallyRes) {
			return Vetoed
			println("vetoed--------")
		} else if !gvr.voteSucceeded(tallyRes){
			return  Defeated
			println("fail--------")
		}
	} else {
		return QuorumNotReach
	}
	return Succeeded
}
