package dao_proposal_single

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	dao_core "gno.land/p/demo/daodao/core_v3"
	dao_interfaces "gno.land/p/demo/daodao/interfaces_v3"
	"gno.land/p/demo/jsonutil_v3"
)

type DAOProposalSingleOpts struct {
	/// The threshold a proposal must reach to complete.
	Threshold dao_interfaces.Threshold
	/// The default maximum amount of time a proposal may be voted on
	/// before expiring.
	MaxVotingPeriod time.Duration
	/// The minimum amount of time a proposal must be open before
	/// passing. A proposal may fail before this amount of time has
	/// elapsed, but it will not pass. This can be useful for
	/// preventing governance attacks wherein an attacker aquires a
	/// large number of tokens and forces a proposal through.
	MinVotingPeriod time.Duration // 0 means no minimum
	/// If set to true only members may execute passed
	/// proposals. Otherwise, any address may execute a passed
	/// proposal.
	OnlyMembersExecute bool
	/// Allows changing votes before the proposal expires. If this is
	/// enabled proposals will not be able to complete early as final
	/// vote information is not known until the time of proposal
	/// expiration.
	AllowRevoting bool
	/// Information about what addresses may create proposals.
	// preProposeInfo PreProposeInfo
	/// If set to true proposals will be closed if their execution
	/// fails. Otherwise, proposals will remain open after execution
	/// failure. For example, with this enabled a proposal to send 5
	/// tokens out of a DAO's treasury with 4 tokens would be closed when
	/// it is executed. With this disabled, that same proposal would
	/// remain open until the DAO's treasury was large enough for it to be
	/// executed.
	CloseProposalOnExecutionFailure bool

	Registry *dao_interfaces.MessagesRegistry
}

type daoProposalSingle struct {
	dao_interfaces.IProposalModule

	core      dao_core.IDAOCore
	opts      *DAOProposalSingleOpts
	proposals []dao_interfaces.Proposal
}

func NewDAOProposalSingle(core dao_core.IDAOCore, opts *DAOProposalSingleOpts) *daoProposalSingle {
	if core == nil {
		panic("core cannot be nil")
	}

	if opts == nil {
		panic("opts cannot be nil")
	}

	if opts.Registry == nil {
		panic("opts.Registry cannot be nil")
	}

	if opts.AllowRevoting {
		panic("allow revoting not implemented")
	}

	if opts.OnlyMembersExecute {
		panic("only members execute not implemented")
	}

	if opts.CloseProposalOnExecutionFailure {
		panic("close proposal on execution failure not implemented")
	}

	// TODO: support other threshold types
	threshold := opts.Threshold.ThresholdQuorum
	if threshold == nil {
		panic("opts.Threshold must be of type ThresholdQuorum")
	}

	thresholdPercent := threshold.Threshold.Percent
	if thresholdPercent == nil {
		panic("opts.Threshold.Threshold must be of type Percent")
	}
	if *thresholdPercent > 10000 {
		panic("opts.Threshold.Threshold must be <= 100%")
	}

	quorumPercent := threshold.Quorum.Percent
	if quorumPercent == nil {
		panic("opts.Threshold.Quorum must be of type Percent")
	}
	if *quorumPercent > 10000 {
		panic("opts.Threshold.Quorum must be <= 100%")
	}

	return &daoProposalSingle{core: core, opts: opts}
}

func (d *daoProposalSingle) Render(path string) string {
	minVotingPeriodStr := "No minimum voting period"
	if d.opts.MinVotingPeriod != 0 {
		minVotingPeriodStr = "Min voting period: " + d.opts.MinVotingPeriod.String()
	}

	executeStr := "Any address may execute passed proposals"
	if d.opts.OnlyMembersExecute {
		executeStr = "Only members may execute passed proposals"
	}

	revotingStr := "Revoting is not allowed"
	if d.opts.AllowRevoting {
		revotingStr = "Revoting is allowed"
	}

	closeOnExecFailureStr := "Proposals will remain open after execution failure"
	if d.opts.CloseProposalOnExecutionFailure {
		closeOnExecFailureStr = "Proposals will be closed if their execution fails"
	}

	thresholdStr := ""
	if threshold := d.opts.Threshold.ThresholdQuorum; threshold != nil {
		thresholdStr = "Threshold: " + threshold.Threshold.Percent.String() + "\n\n" +
			"Quorum: " + threshold.Quorum.Percent.String()
	}

	proposalsStr := "## Proposals\n"
	for _, p := range d.proposals {
		messagesStr := ""
		for _, m := range p.Messages {
			messagesStr += "- " + m.(dao_interfaces.ExecutableMessage).String() + "\n"
		}

		proposalsStr += "### #" + strconv.Itoa(p.ID) + " " + p.Title + "\n" +
			"Status: " + p.Status.String() + "\n\n" +
			"Proposed by " + p.Proposer.String() + "\n\n" +
			p.Description + "\n\n" +
			"Votes summary:" + "\n\n" +
			"- Yes: " + strconv.FormatUint(p.Votes.Yes, 10) + "\n" +
			"- No: " + strconv.FormatUint(p.Votes.No, 10) + "\n" +
			"- Abstain: " + strconv.FormatUint(p.Votes.Abstain, 10) + "\n\n" +
			"Total: " + strconv.FormatUint(p.Votes.Total(), 10) + "\n" +
			"#### Messages\n" +
			messagesStr +
			"#### Votes\n"

		// /* dev
		p.Ballots.Iterate("", "", func(k string, v interface{}) bool {
			ballot := v.(dao_interfaces.Ballot)
			proposalsStr += "- " + k + " voted " + ballot.Vote.String() + "\n"
			return false
		})
		// */

		/* test3
		ballotsCount := p.Ballots.Size()
		for i := 0; i < ballotsCount; i++ {
			k, v := p.Ballots.GetByIndex(i)
			ballot := v.(dao_interfaces.Ballot)
			proposalsStr += "- " + k + " voted " + ballot.Vote.String() + "\n"
		}
		*/

		proposalsStr += "\n"
	}

	return "# Single choice proposals module" + "\n" +
		"## Summary" + "\n" +
		"Max voting period: " + d.opts.MaxVotingPeriod.String() + "\n\n" +
		minVotingPeriodStr + "\n\n" +
		executeStr + "\n\n" +
		revotingStr + "\n\n" +
		closeOnExecFailureStr + "\n\n" +
		thresholdStr + "\n\n" +
		proposalsStr
}

func (d *daoProposalSingle) Propose(title string, description string, messages []dao_interfaces.ExecutableMessage) {
	// TODO: auth
	d.proposals = append(d.proposals, dao_interfaces.Proposal{
		ID:          len(d.proposals),
		Title:       title,
		Description: description,
		Messages:    messages,
		Proposer:    std.GetOrigCaller(),
		Ballots:     avl.NewTree(), // dev
		// Ballots: avl.NewMutTree(), // test3
		Status: dao_interfaces.ProposalStatusOpen,
	})
}

func (d *daoProposalSingle) GetBallot(proposalID int, memberAddress std.Address) dao_interfaces.Ballot {
	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	proposal := d.proposals[proposalID]
	ballot, has := proposal.Ballots.Get(memberAddress.String())
	if !has {
		panic("ballot does not exist")
	}
	return ballot.(dao_interfaces.Ballot)
}

func (d *daoProposalSingle) Vote(proposalID int, vote dao_interfaces.Vote, rationale string) {
	voter := std.PrevRealm().Addr()

	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	proposal := d.proposals[proposalID]
	// TODO: check proposal expiration

	votePower := d.core.VotingModule().VotingPower(voter)
	if votePower == 0 {
		panic("you're not a member")
	}

	// TODO: handle revoting
	if ok := proposal.Ballots.Has(voter.String()); ok {
		panic("you already voted")
	}
	proposal.Ballots.Set(voter.String(), dao_interfaces.Ballot{
		Vote:      vote,
		Power:     votePower,
		Rationale: rationale,
	})

	proposal.Votes.Add(vote, votePower)

	d.updateStatus(proposalID)
}

func (d *daoProposalSingle) Execute(proposalID int) {
	executer := std.GetOrigCaller()

	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	prop := d.proposals[proposalID]

	d.updateStatus(proposalID)
	if prop.Status != dao_interfaces.ProposalStatusPassed {
		panic("proposal is not passed")
	}

	d.opts.Registry.ExecuteMessages(prop.Messages)

	d.proposals[proposalID].Status = dao_interfaces.ProposalStatusExecuted
}

// FIXME: should probably return a copy for safety
func (d *daoProposalSingle) Proposals() []dao_interfaces.Proposal {
	return d.proposals
}

func (d *daoProposalSingle) ProposalsJSON() string {
	return jsonutil.FormatSlice(d.proposals)
}

func (d *daoProposalSingle) Threshold() dao_interfaces.Threshold {
	return d.opts.Threshold
}

func (d *daoProposalSingle) updateStatus(proposalID int) {
	proposal := d.proposals[proposalID]
	if proposal.Status == dao_interfaces.ProposalStatusOpen && d.isPassed(proposalID) {
		d.proposals[proposalID].Status = dao_interfaces.ProposalStatusPassed
		return
	}
}

func (d *daoProposalSingle) isPassed(proposalID int) bool {
	proposal := d.proposals[proposalID]

	// TODO: support other threshold types
	threshold := d.opts.Threshold.ThresholdQuorum.Threshold
	quorum := d.opts.Threshold.ThresholdQuorum.Quorum

	totalPower := d.core.VotingModule().TotalPower()

	if !doesVoteCountPass(proposal.Votes.Total(), totalPower, quorum) {
		return false
	}

	// TODO: handle expiration
	options := totalPower - proposal.Votes.Abstain
	return doesVoteCountPass(proposal.Votes.Yes, options, threshold)
}

func doesVoteCountPass(yesVotes uint64, options uint64, percent dao_interfaces.PercentageThreshold) bool {
	if options == 0 {
		return false
	}
	percentValue := uint64(*percent.Percent)
	votes := yesVotes * 10000
	threshold := options * percentValue
	return votes >= threshold
}
