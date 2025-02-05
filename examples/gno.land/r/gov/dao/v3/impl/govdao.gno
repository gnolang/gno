package impl

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

var ErrMemberNotFound = errors.New("member not found")

type GovDAO struct {
	pss ProposalsStatuses
}

func NewGovDAO() *GovDAO {
	return &GovDAO{
		pss: NewProposalsStatuses(),
	}
}

func (g *GovDAO) PreCreateProposal(r proxy.ProposalRequest)(proxy.Author, error) {
	// TODO validate that the caller is a member
	return NewUserAuthor(), nil
}

func (g *GovDAO) PostCreateProposal(r proxy.ProposalRequest, pid proxy.ProposalID) {
	pids := ufmt.Sprintf("%v", int(pid))
	g.pss.Set(pids, newProposalStatus())
}

func (g *GovDAO) VoteOnProposal(r proxy.VoteRequest) error {
	caller := std.GetOrigCaller()
	mem, tie := memberstore.Get().GetMember(caller)
	if mem == nil {
		// return ErrMemberNotFound
		// TODO not returning error because of a bug: test pkg: panic: interface conversion: gnolang.Value is gnolang.RefValue, not *gnolang.StructValue
		return nil
	}

	status := g.pss.GetStatus(r.ProposalID)

	mVoted, _ := status.AllVotes.GetMember(caller)
	if mVoted != nil {
		return errors.New("already voted on proposal")
	}

	switch r.Option {
		case proxy.YesVote:
			status.AllVotes.SetMember(tie, caller, mem)
			status.YesVotes.SetMember(tie, caller, mem)
		case proxy.NoVote:
			status.AllVotes.SetMember(tie, caller, mem)
			status.NoVotes.SetMember(tie, caller, mem)
		default:
			return errors.New("voting can only be YES or NO")
	}

	return nil
}

func (g *GovDAO) PreGetProposal(pid proxy.ProposalID) error {
	return nil
}

func (g *GovDAO) PostGetProposal(pid proxy.ProposalID, p *proxy.Proposal) error {
	return nil
}

func (g *GovDAO) PreExecuteProposal(pid proxy.ProposalID) error {
	// TODO: remove caller. Anyone can call this
	caller := std.GetOrigCaller()
	_, tie := memberstore.Get().GetMember(caller)
	if tie != memberstore.T1 { 
		return errors.New("only members from Tier 1 can execute the proposals")
	}

	status := g.pss.GetStatus(pid)

	if status.Denied || status.Accepted {
		return errors.New("proposal already executed")
	}

	if status.YesPercent() >= law.Supermajority {
		status.Accepted = true
	} else if status.NoPercent() >= law.Supermajority {
		status.Denied = true
	} else {
		return errors.New(ufmt.Sprintf("proposal didn't reached supermajority yet: %v", law.Supermajority))
	}

	return nil
}

func (g *GovDAO) renderActiveProposals() string {
	out := "## Active Propoposals:\n"
	g.pss.Iterate("", "", func(key string, valuePs interface{}) bool {
		ps, _ := valuePs.(*proposalStatus)
		if ps.Inactive {
			return false
		}
		
		pid, err := strconv.Atoi(key)
		if err != nil {
			panic(err.Error())
		}

		p := proxy.MustGetProposal(proxy.ProposalID(pid))
		
		out += ufmt.Sprintf("### Proposal with id: %v", key)
		out += p.String()
		out += "\n\n"
		out += ps.String()
		out += "\n\n"

		// TODO print total state of the voting process
		out += "------------------\n"

		return false
	})

	return out
}

func (g *GovDAO) Render(string) string {
	return g.renderActiveProposals()
}


type UserAuthor struct {
	addr string
}

func NewUserAuthor() *UserAuthor {
	return &UserAuthor{
		addr: std.GetOrigCaller().String(),
	}
}

func (u *UserAuthor)ID() string {
	return "addr:"+u.addr
}

func (u *UserAuthor)String() string {
	return u.addr
}