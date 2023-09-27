package dao_proposal_single

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/dao_utils"
	"gno.land/p/demo/teritori/ujson"
)

type DAOProposalSingleOpts struct {
	/// The threshold a proposal must reach to complete.
	Threshold Threshold
	/// The default maximum amount of time a proposal may be voted on
	/// before expiring.
	MaxVotingPeriod dao_utils.Duration
	/// The minimum amount of time a proposal must be open before
	/// passing. A proposal may fail before this amount of time has
	/// elapsed, but it will not pass. This can be useful for
	/// preventing governance attacks wherein an attacker aquires a
	/// large number of tokens and forces a proposal through.
	MinVotingPeriod dao_utils.Duration // 0 means no minimum
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
}

func (opts DAOProposalSingleOpts) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "threshold", Value: opts.Threshold},
		{Key: "maxVotingPeriod", Value: opts.MaxVotingPeriod},
		{Key: "minVotingPeriod", Value: opts.MinVotingPeriod},
		{Key: "onlyMembersExecute", Value: opts.OnlyMembersExecute},
		{Key: "allowRevoting", Value: opts.AllowRevoting},
		{Key: "closeProposalOnExecutionFailure", Value: opts.CloseProposalOnExecutionFailure},
	})
}

type DAOProposalSingle struct {
	dao_interfaces.IProposalModule

	core      dao_interfaces.IDAOCore
	opts      *DAOProposalSingleOpts
	proposals []*Proposal
}

func NewDAOProposalSingle(core dao_interfaces.IDAOCore, opts *DAOProposalSingleOpts) *DAOProposalSingle {
	if core == nil {
		panic("core cannot be nil")
	}

	if opts == nil {
		panic("opts cannot be nil")
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

	if opts.MaxVotingPeriod == nil {
		panic("max voting period cannot be nil")
	}

	// TODO: support other threshold types
	switch opts.Threshold.(type) {
	case *ThresholdThresholdQuorum:
		threshold := opts.Threshold.(*ThresholdThresholdQuorum)
		switch threshold.Threshold.(type) {
		case *PercentageThresholdMajority:
			panic("not implemented")
		case *PercentageThresholdPercent:
			if *threshold.Threshold.(*PercentageThresholdPercent) > 10000 {
				panic("opts.Threshold.Threshold must be <= 100%")
			}
		default:
			panic("unknown Threshold type")
		}
		switch threshold.Quorum.(type) {
		case *PercentageThresholdMajority:
			panic("not implemented")
		case *PercentageThresholdPercent:
			if *threshold.Quorum.(*PercentageThresholdPercent) > 10000 {
				panic("opts.Threshold.Quorum must be <= 100%")
			}
		default:
			panic("unknown PercentageThreshold type")
		}
	default:
		panic("unsupported Threshold type")
	}

	return &DAOProposalSingle{core: core, opts: opts}
}

func (d *DAOProposalSingle) Render(path string) string {
	minVotingPeriodStr := "No minimum voting period"
	if d.opts.MinVotingPeriod != nil {
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
	switch d.opts.Threshold.(type) {
	case *ThresholdThresholdQuorum:
		threshold := d.opts.Threshold.(*ThresholdThresholdQuorum)
		thresholdStr = "Threshold: " + threshold.Threshold.String() + "\n\n" +
			"Quorum: " + threshold.Quorum.String()
	default:
		panic("unsupported Threshold type")
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

		p.Ballots.Iterate("", "", func(k string, v interface{}) bool {
			ballot := v.(Ballot)
			proposalsStr += "- " + k + " voted " + ballot.Vote.String() + "\n"
			return false
		})

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

func (d *DAOProposalSingle) Core() dao_interfaces.IDAOCore {
	return d.core
}

func (d *DAOProposalSingle) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "SingleChoiceProposal",
		Version: "0.1.0",
	}
}

func (d *DAOProposalSingle) ConfigJSON() string {
	return ujson.FormatAny(d.opts)
}

func (d *DAOProposalSingle) Propose(title string, description string, messages []dao_interfaces.ExecutableMessage) int {
	// TODO: creation policy

	totalPower := d.core.VotingModule().TotalPowerAtHeight(0)

	expiration := d.opts.MaxVotingPeriod.AfterCurrentBlock()
	minVotingPeriod := dao_utils.Expiration(nil)
	if d.opts.MinVotingPeriod != nil {
		minVotingPeriod = d.opts.MinVotingPeriod.AfterCurrentBlock()
	}

	id := len(d.proposals)

	prop := Proposal{
		ID:              id,
		Title:           title,
		Description:     description,
		Proposer:        std.PrevRealm().Addr(),
		StartHeight:     std.GetHeight(),
		MinVotingPeriod: minVotingPeriod,
		Expiration:      expiration,
		Threshold:       d.opts.Threshold.Clone(),
		TotalPower:      totalPower,
		Messages:        messages,
		Status:          ProposalStatusOpen,
		Ballots:         avl.NewTree(),
		AllowRevoting:   d.opts.AllowRevoting,
	}
	prop.updateStatus()
	d.proposals = append(d.proposals, &prop)
	return id
}

func (d *DAOProposalSingle) GetBallot(proposalID int, memberAddress std.Address) Ballot {
	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	proposal := d.proposals[proposalID]
	ballot, has := proposal.Ballots.Get(memberAddress.String())
	if !has {
		panic("ballot does not exist")
	}
	return ballot.(Ballot)
}

type VoteWithRationale struct {
	Vote      Vote
	Rationale string
}

func (v *VoteWithRationale) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "vote", Value: &v.Vote},
		{Key: "rationale", Value: &v.Rationale},
	})
}

func (d *DAOProposalSingle) VoteJSON(proposalID int, voteJSON string) {
	var v VoteWithRationale
	ujson.ParseAny(voteJSON, &v)

	voter := std.PrevRealm().Addr()

	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	proposal := d.proposals[proposalID]
	// TODO: check proposal expiration

	votePower := d.core.VotingModule().VotingPowerAtHeight(voter, proposal.StartHeight)
	if votePower == 0 {
		panic("not registered")
	}

	// TODO: handle revoting
	if ok := proposal.Ballots.Has(voter.String()); ok {
		panic("already voted")
	}
	proposal.Ballots.Set(voter.String(), Ballot{
		Vote:      v.Vote,
		Power:     votePower,
		Rationale: v.Rationale,
	})

	proposal.Votes.Add(v.Vote, votePower)

	proposal.updateStatus()
}

func (d *DAOProposalSingle) Execute(proposalID int) {
	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}
	prop := d.proposals[proposalID]

	prop.updateStatus()
	if prop.Status != ProposalStatusPassed {
		panic("proposal is not passed")
	}

	for _, m := range prop.Messages {
		d.core.Registry().Execute(m)
	}

	prop.Status = ProposalStatusExecuted
}

type ProposalRequest struct {
	Title       string
	Description string
	Messages    *ujson.JSONASTNode
}

func (pr *ProposalRequest) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "title", Value: &pr.Title},
		{Key: "description", Value: &pr.Description},
		{Key: "messages", Value: &pr.Messages},
	})
}

func (d *DAOProposalSingle) ProposeJSON(proposalJSON string) int {
	var req ProposalRequest
	ujson.ParseAny(proposalJSON, &req)
	msgs := d.core.Registry().MessagesFromJSON(req.Messages.String()) // TODO: optimize
	return d.Propose(req.Title, req.Description, msgs)
}

func (d *DAOProposalSingle) Proposals() []*Proposal {
	return d.proposals
}

func (d *DAOProposalSingle) ProposalsJSON(limit int, startAfter string, reverse bool) string {
	iSlice := make([]interface{}, len(d.proposals))
	for i, p := range d.proposals {
		iSlice[i] = p
	}
	return ujson.FormatSlice(iSlice)
}

func (d *DAOProposalSingle) ProposalJSON(proposalID int) string {
	if proposalID < 0 || proposalID >= len(d.proposals) {
		panic("proposal does not exist")
	}
	return ujson.FormatAny(d.proposals[proposalID])
}

func (d *DAOProposalSingle) Threshold() Threshold {
	return d.opts.Threshold
}

func (proposal *Proposal) updateStatus() {
	if proposal.Status == ProposalStatusOpen && proposal.isPassed() {
		proposal.Status = ProposalStatusPassed
		return
	}
}

func (proposal *Proposal) isPassed() bool {
	switch proposal.Threshold.(interface{}).(type) {
	case *ThresholdAbsolutePercentage:
		panic("'isPassed' not implemented for 'ThresholdAbsolutePercentage'")
	case *ThresholdThresholdQuorum:
		thresholdObj := proposal.Threshold.(*ThresholdThresholdQuorum)

		threshold := thresholdObj.Threshold
		quorum := thresholdObj.Quorum

		totalPower := proposal.TotalPower

		if !doesVoteCountPass(proposal.Votes.Total(), totalPower, quorum) {
			return false
		}

		// TODO: handle expiration
		options := totalPower - proposal.Votes.Abstain
		return doesVoteCountPass(proposal.Votes.Yes, options, threshold)
	case *ThresholdAbsoluteCount:
		panic("'isPassed' not implemented for 'ThresholdAbsoluteCount'")
	default:
		panic("unknown Threshold type")
	}
}

func doesVoteCountPass(yesVotes uint64, options uint64, percent PercentageThreshold) bool {
	switch percent.(type) {
	case *PercentageThresholdMajority:
		panic("'doesVoteCountPass' not implemented for 'PercentageThresholdMajority'")
	case *PercentageThresholdPercent:
		if options == 0 {
			return false
		}
		percentValue := uint64(*percent.(*PercentageThresholdPercent))
		votes := yesVotes * 10000
		threshold := options * percentValue
		return votes >= threshold
	default:
		panic("unknown PercentageThreshold type")
	}
}
