package commondao

import (
	"std"
	"time"

	"gno.land/p/nt/commondao"
)

// Must implement a proposal struct and define the five methods below
type Proposal struct {
	TitleProp       string
	DescriptionProp string
}

func (c *Proposal) Title() string {
	return c.TitleProp
}

func (c *Proposal) Body() string {
	return c.DescriptionProp
}

// The voting period allowed for each proposal created
func (c *Proposal) VotingPeriod() time.Duration {
	return 24 * time.Hour // Voting period is 24 hours here
}

// The tally decides if the proposal is passed according to your own rule
func (c *Proposal) Tally(r commondao.ReadonlyVotingRecord, m commondao.MemberSet) (bool, error) {
	yesCount := r.VoteCount(commondao.ChoiceYes)
	noCount := r.VoteCount(commondao.ChoiceNo)
	return yesCount > noCount, nil
}

// Execute will execute what you decide once the proposal is considered passed
func (p *Proposal) Execute() error {
	return nil // here execute does not do anything
}

// Create the proposal and store it inside the ActiveProposal storage
func CreateProposal(title, desc string) {
	caller := std.PreviousRealm().Address()

	prop := &Proposal{
		TitleProp:       title,
		DescriptionProp: desc,
	}

	DAO.MustPropose(caller, prop) // Panics if the creation of the proposal fails, use Propose() if you don't want the possible panic
}

// Allows member of the DAO to vote YES or NO on a proposal and give their reasoning
func Vote(vote bool, proposalID uint64, reason string) {
	var choice commondao.VoteChoice

	if vote {
		choice = commondao.ChoiceYes // Other types of vote choice can be found in 'p/nt/commondao/proposal.gno'
	} else {
		choice = commondao.ChoiceNo
	}

	caller := std.PreviousRealm().Address()

	DAO.Vote(caller, proposalID, choice, reason)
}

func ExecutionPassedProposals(proposalID uint64) {
	// Calls top level method which does the following
	//	- Verify if the proposal is active
	//	- Verify that the deadline is reached
	//	- Tally the votes
	//	- Run our previously defined Execute()
	//	- Set the proposal status and move it out of Active Proposal if necessary
	DAO.Execute(proposalID)
}
