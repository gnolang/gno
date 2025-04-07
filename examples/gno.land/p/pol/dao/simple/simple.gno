// This is an example of a DAO implementation with a
// poll implementation, with title and description.
//
// Note that you can use some of this Simple in your own
// implementation without using all of it.
// Like you can use an SimpleExecutor without having the
// same Proposal.
package simple

import (
	"errors"
	"std"
	"time"
	"gno.land/p/pol/dao"
)

// We define the status
const (
	SimpleAccepted dao.ProposalChoice = iota
	SimpleRejected
	SimpleWhiteVote
	SimpleChoiceCount
)

// Then we implement, first, the Executor
type SimpleExecutor struct {
	f func()error
}

func (e *SimpleExecutor) Execute() error {
	return e.f()
}

func NewSimpleExecutor(f func()error) *SimpleExecutor {
	return &SimpleExecutor{f}
}

// Then, the Proposal
type SimpleProposal struct {
	Title string
	Description string
	author dao.Member					// The original author.
	votes map[dao.ProposalChoice]uint	// The actual votes.
	voters []dao.Member 				// The ones who alread votes.
	expires time.Time					// When does the poll expires.
	executor *SimpleExecutor			// Execute the poll
	alive bool							// If the poll is alive
}

func (p *SimpleProposal) GetAuthor() dao.Member {
	return p.author
}

func (p *SimpleProposal) HowManyVoters() uint {
	return uint(len(p.voters))
}

// In this vote system, if >50% of the votes are White, its refused
// Else, its just the most between Accepted & Rejected.
func (p *SimpleProposal) IsAccepted() bool {
	total := len(p.voters)
	if (total == 0) {
		return false
	}
	if (float64(p.votes[SimpleWhiteVote]) / float64(total)) > 0.5 {
		return false
	}
	return p.votes[SimpleAccepted] > p.votes[SimpleRejected]
}

func (p *SimpleProposal) ExecuteIfAccepted() error {
	p.alive = false
	if p.IsAccepted() {
		return p.executor.f()
	}
	return nil
}

func (p *SimpleProposal) ExpiresIn() time.Time {
	return p.expires
}

func (p *SimpleProposal) Vote(verdict dao.ProposalChoice) error {
	who := std.PreviousRealm().Address()
	if who.String() == "" || verdict >= SimpleChoiceCount {
		return errors.New("invalid previous realm")
	}
	for _, v := range p.voters {
		if v.Address() == who {
			return errors.New("user already voted")
		}
	}
	p.voters = append(p.voters, dao.Member(who))
	p.votes[verdict] += 1
	return nil
}

func (p *SimpleProposal) IsEnded() bool {
	return !p.alive
}

func NewSimpleProposal(who dao.Member, title, desc string, f func()error) *SimpleProposal {
	exp := time.Now().Add(time.Minute) // One minute for now
	x := &SimpleProposal{
		Title: title,
		Description: desc,
		author: who,
		votes: map[dao.ProposalChoice]uint{},
		voters: []dao.Member{},
		expires: exp,
		executor: NewSimpleExecutor(f),
		alive: true,
	}
	for n := SimpleAccepted; n < SimpleChoiceCount; n++ {
		x.votes[n] = 0
	}
	return x
}

// After all of that we create the DAO and implement the store
type SimpleDAO struct {
	dao.DAO
	props []dao.Proposal
}

func (d *SimpleDAO) GetProposals() []dao.Proposal {
	return d.props
}

func (d *SimpleDAO) Propose(prop dao.Proposal) {
	d.props = append(d.props, prop)
}

func (d *SimpleDAO) GetChoicesNames() map[dao.ProposalChoice]string {
	return map[dao.ProposalChoice]string{
		SimpleAccepted: "Accept",
		SimpleRejected: "Refuse",
		SimpleWhiteVote: "Abstaint",
	}
}

func NewSimpleDAO() *SimpleDAO {
	return &SimpleDAO{
		dao.NewDAO(),
		[]dao.Proposal{},
	}
}
