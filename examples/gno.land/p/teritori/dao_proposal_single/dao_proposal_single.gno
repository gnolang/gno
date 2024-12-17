package dao_proposal_single

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/teritori/dao_interfaces"
	"gno.land/p/teritori/dao_utils"
	"gno.land/p/demo/json"
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

func (opts DAOProposalSingleOpts) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"threshold":                       opts.Threshold.ToJSON(),
		"maxVotingPeriod":                 opts.MaxVotingPeriod.ToJSON(),
		"minVotingPeriod":                 opts.MinVotingPeriod.ToJSON(),
		"onlyMembersExecute":              json.BoolNode("", opts.OnlyMembersExecute),
		"allowRevoting":                   json.BoolNode("", opts.AllowRevoting),
		"closeProposalOnExecutionFailure": json.BoolNode("", opts.CloseProposalOnExecutionFailure),
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

	sb := strings.Builder{}
	sb.WriteString("Max voting period: ")
	sb.WriteString(d.opts.MaxVotingPeriod.String())
	sb.WriteString("\n\n")
	sb.WriteString(minVotingPeriodStr)
	sb.WriteString("\n\n")
	sb.WriteString(executeStr)
	sb.WriteString("\n\n")
	sb.WriteString(revotingStr)
	sb.WriteString("\n\n")
	sb.WriteString(closeOnExecFailureStr)
	sb.WriteString("\n\n")
	switch d.opts.Threshold.(type) {
	case *ThresholdThresholdQuorum:
		threshold := d.opts.Threshold.(*ThresholdThresholdQuorum)
		sb.WriteString("Threshold: ")
		sb.WriteString(threshold.Threshold.String())
		sb.WriteString("\n\nQuorum: ")
		sb.WriteString(threshold.Quorum.String())

	case *ThresholdAbsolutePercentage:
		threshold := d.opts.Threshold.(*ThresholdAbsolutePercentage)
		sb.WriteString("Threshold (Absolute Percentage): ")
		sb.WriteString(threshold.Value.String())

	case *ThresholdAbsoluteCount:
		threshold := d.opts.Threshold.(*ThresholdAbsoluteCount)
		sb.WriteString("Threshold (Absolute Count): ")
		sb.WriteString(strconv.FormatUint(uint64(*threshold), 10))

	default:
		panic("unsupported Threshold type")
	}

	sb.WriteString("\n\n")
	sb.WriteString("Proposals:\n\n")
	for _, p := range d.proposals {
		sb.WriteString("\\#")
		sb.WriteString(strconv.Itoa(p.ID))
		sb.WriteString(": ")
		sb.WriteString(p.Title)
		sb.WriteString(":\n\n")
		sb.WriteString("  Status: ")
		sb.WriteString(p.Status.String())
		sb.WriteString("\n\n")
		sb.WriteString("  Proposed by ")
		sb.WriteString(p.Proposer.String())
		sb.WriteString("\n\n")
		sb.WriteString("  ")
		sb.WriteString(p.Description)
		sb.WriteString("\n\n")
		sb.WriteString("  Votes summary:")
		sb.WriteString("\n\n")
		sb.WriteString("  - Yes: ")
		sb.WriteString(strconv.FormatUint(p.Votes.Yes, 10))
		sb.WriteRune('\n')
		sb.WriteString("  - No: ")
		sb.WriteString(strconv.FormatUint(p.Votes.No, 10))
		sb.WriteRune('\n')
		sb.WriteString("  - Abstain: ")
		sb.WriteString(strconv.FormatUint(p.Votes.Abstain, 10))
		sb.WriteString("\n\n")
		sb.WriteString("  Total: ")
		sb.WriteString(strconv.FormatUint(p.Votes.Total(), 10))
		sb.WriteString("\n\n")

		sb.WriteString("  Messages:")
		sb.WriteString("\n\n")
		for _, m := range p.Messages {
			sb.WriteString(m.(dao_interfaces.ExecutableMessage).Type())
			sb.WriteString("\n\n")
			sb.WriteString(m.(dao_interfaces.ExecutableMessage).String())
			sb.WriteString("\n\n")
		}

		sb.WriteString("  Votes:")
		sb.WriteString("\n\n")

		p.Ballots.Iterate("", "", func(k string, v interface{}) bool {
			ballot := v.(Ballot)
			sb.WriteString(k)
			sb.WriteString(" voted ")
			sb.WriteString(ballot.String())
			sb.WriteString("\n\n")
			return false
		})

		sb.WriteRune('\n')
	}

	return sb.String()
}

func (d *DAOProposalSingle) Core() dao_interfaces.IDAOCore {
	return d.core
}

func (d *DAOProposalSingle) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "gno.land/p/teritori/dao_proposal_single",
		Version: "0.1.0",
	}
}

func (d *DAOProposalSingle) ConfigJSON() string {
	b, err := json.Marshal(d.opts.ToJSON())
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (d *DAOProposalSingle) Propose(title string, description string, messages []dao_interfaces.ExecutableMessage) int {
	// TODO: creation policy

	totalPower := d.core.VotingModule().TotalPowerAtHeight(0, []string{})

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

func (v *VoteWithRationale) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	v.Vote.FromJSON(obj["vote"])
	v.Rationale = obj["rationale"].MustString()
}

func (d *DAOProposalSingle) VoteJSON(proposalID int, voteJSON string) {
	var v VoteWithRationale
	v.FromJSON(json.Must(json.Unmarshal([]byte(voteJSON))))

	voter := std.PrevRealm().Addr()

	if len(d.proposals) <= proposalID || proposalID < 0 {
		panic("proposal does not exist")
	}

	proposal := d.proposals[proposalID]

	if proposal.Expiration.IsExpired() {
		panic("proposal is expired")
	}

	votePower := d.core.VotingModule().VotingPowerAtHeight(voter, proposal.StartHeight, []string{})
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
	Messages    []*json.Node
}

func (pr *ProposalRequest) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	pr.Title = obj["title"].MustString()
	pr.Description = obj["description"].MustString()
	pr.Messages = obj["messages"].MustArray()
}

func (d *DAOProposalSingle) ProposeJSON(proposalJSON string) int {
	var req ProposalRequest
	req.FromJSON(json.Must(json.Unmarshal([]byte(proposalJSON))))
	msgs := d.core.Registry().MessagesFromJSON(req.Messages)
	return d.Propose(req.Title, req.Description, msgs)
}

func (d *DAOProposalSingle) Proposals() []*Proposal {
	return d.proposals
}

func (d *DAOProposalSingle) ProposalsJSON(limit int, startAfter string, reverse bool) string {
	iSlice := make([]*json.Node, len(d.proposals))
	for i, p := range d.proposals {
		iSlice[i] = p.ToJSON()
	}

	return json.ArrayNode("", iSlice).String()
}

func (d *DAOProposalSingle) ProposalJSON(proposalID int) string {
	if proposalID < 0 || proposalID >= len(d.proposals) {
		panic("proposal does not exist")
	}

	return d.proposals[proposalID].ToJSON().String()
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
