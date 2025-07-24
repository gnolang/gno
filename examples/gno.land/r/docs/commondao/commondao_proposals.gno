package commondao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/nt/commondao"
)

// Proposal implements the commondao.Proposal interface
// It contains anything useful to you in a proposal and defines how the proposal behaves
type proposal struct {
	TitleProp       string
	DescriptionProp string
	CallerProp      std.Address
}

// Title returns the title of the proposal
func (p *proposal) Title() string {
	return p.TitleProp
}

// Body returns the full description of the proposal
func (p *proposal) Body() string {
	return p.DescriptionProp
}

// VotingPeriod defines the length of time the proposal is open for voting
// Another possible implementation would be to have a constructor
// So that each proposal can implement its own duration and tally
func (c *proposal) VotingPeriod() time.Duration {
	return 24 * time.Hour // Voting period is 24 hours here
}

// Tally determines whether the proposal passes based on voting results
// There are multiple ways to tally votes, pick the one most suited to your use case
func (c *proposal) Tally(r commondao.ReadonlyVotingRecord, m commondao.MemberSet) (bool, error) {
	yesCount := r.VoteCount(commondao.ChoiceYes)

	memberCount := DAO.Members().Size()
	fivePercent := (memberCount*5 + 99) / 100 // fivePercent calculate what is 5% of the DAO's member

	return yesCount >= fivePercent, nil // The proposal is considered as "passed" as soon as 5% of members voted yes
}

// Execute will execute what you decide once the proposal is considered passed
func (p *proposal) Execute(cur realm) error {
	// Here the proposal executed is the addition of the creator of the proposal to the DAO
	if DAO.Members().Has(p.CallerProp) {
		return errors.New(p.CallerProp.String() + " is already a DAO member")
	}

	DAO.Members().Add(p.CallerProp)
	return nil
}

// CreateProposal constructs and submits a new proposal to the DAO
// The proposal is stored in the ActiveProposal storage
func CreateProposal(cur realm, title, desc string) {
	caller := std.PreviousRealm().Address()

	// Set the new proposal's title and description
	prop := &proposal{
		TitleProp:       title,
		DescriptionProp: desc,
		CallerProp:      caller,
	}

	DAO.MustPropose(caller, prop) // Panics if the creation of the proposal fails, use Propose() if you don't want the possible panic
}

// Vote allows a DAO member to vote on an existing proposal
// Allows member of the DAO to vote YES or NO on a proposal and give their reasoning
func Vote(cur realm, vote bool, proposalID uint64, reason string) {
	var choice commondao.VoteChoice

	if vote {
		choice = commondao.ChoiceYes // Other types of vote choice can be found in 'p/nt/commondao/proposal.gno'
	} else {
		choice = commondao.ChoiceNo
	}

	caller := std.PreviousRealm().Address()

	err := DAO.Vote(caller, proposalID, choice, reason) // logs in the user's voting choice on a specified proposal
	if err != nil {
		panic(err)
	}
	ExecutionPassedProposals(cross)
}

// ExecutionPassedProposals finalizes a proposal if it passes voting criteria
//
// This method performs the following:
// - Checks if the proposal is active and its deadline has passed
// - Tallies the votes using the Tally() method
// - If passed, executes the associated Execute() logic
// - Updates the proposal's status and  moves it out of Active Proposal if necessary
func ExecutionPassedProposals(cur realm) {
	DAO.ActiveProposals().Iterate(0, DAO.ActiveProposals().Size(), false, func(p *commondao.Proposal) bool {
		if time.Now().After(p.VotingDeadline()) {
			_ = DAO.Execute(p.ID())
		}
		return true
	})
}
