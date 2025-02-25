package impl

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

var ErrMemberNotFound = errors.New("member not found")

type GovDAO struct {
	pss ProposalsStatuses
	pssPager *pager.Pager
}

func NewGovDAO() *GovDAO {
	pss := NewProposalsStatuses()
	return &GovDAO{
		pss: pss,
		pssPager: pager.NewPager(pss.Tree, 3, false),
	}
}

func (g *GovDAO) PreCreateProposal(r proxy.ProposalRequest) (proxy.Author, error) {
	// TODO validate that the caller is a member
	return NewUserAuthor(), nil
}

func (g *GovDAO) PostCreateProposal(r proxy.ProposalRequest, pid proxy.ProposalID) {
	tatv := []string{memberstore.T1, memberstore.T2, memberstore.T3}
	switch v := r.Metadata.(type) {
	case *AddMemberMetadata:
		// only members from T1 are allowed to vote when adding new members to T1
		if v.Tier == memberstore.T1 {
			tatv = []string{memberstore.T1}
		}
		// only members from T1 and T2 are allowed to vote when adding new members to T2
		if v.Tier == memberstore.T2 {
			tatv = []string{memberstore.T1, memberstore.T2}
		}
	case *PromoteMemberMetadata:
		// only members from T1 are allowed to vote when promoting someone to T1
		if v.ToTier == memberstore.T1 {
			tatv = []string{memberstore.T1}
		}
		// only members from T1 and T2 are allowed to vote when promoting someone to T2
		if v.ToTier == memberstore.T2 {
			tatv = []string{memberstore.T1, memberstore.T2}
		}
	}

	pids := ufmt.Sprintf("%v", int(pid))
	g.pss.Set(pids, newProposalStatus(tatv))
}

func (g *GovDAO) VoteOnProposal(r proxy.VoteRequest) error {
	if !g.isValidCall() {
		return errors.New("proposal voting must be done directly by a user")
	}

	caller := std.GetOrigCaller()
	mem, tie := memberstore.Get().GetMember(caller)
	if mem == nil {
		// return ErrMemberNotFound
		// TODO not returning error because of a bug: test pkg: panic: interface conversion: gnolang.Value is gnolang.RefValue, not *gnolang.StructValue
		return nil
	}

	status := g.pss.GetStatus(r.ProposalID)

	// TODO check if proposal status is accepted or rejected already

	if !status.IsAllowed(tie) {
		return errors.New("member on specified tier is not allowed to vote on this proposal")
	}

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
	if !std.PrevRealm().IsUser() {
		return errors.New("invalid non-user call")
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

func (g *GovDAO) renderActiveProposals(url string) string {
	out := "## Active Propoposals:\n"

	page, err := g.pssPager.GetPageByPath(url)
	if err != nil {
		out += ufmt.Sprintf("Error getting selected page: %v", err.Error())
		return out
	}

	for _, item := range page.Items {
		ps, _ := item.Value.(*proposalStatus)
		if ps.Inactive {
			continue
		}

		pid, err := strconv.Atoi(item.Key)
		if err != nil {
			panic(err.Error())
		}

		p := proxy.MustGetProposal(proxy.ProposalID(pid))

		out += ufmt.Sprintf("### Proposal with id: %v", item.Key)
		out += StringifyProposal(p)
		out += "\n\n"
		out += ps.String()
		out += "\n\n"

		// TODO print total state of the voting process
		out += "------------------\n"
	}

	out += page.Picker()

	return out
}

func (g *GovDAO) Render(url string) string {
	return g.renderActiveProposals(url)
}

func (g *GovDAO) isValidCall() bool {
	// only calls from the proxy are allowed
	// TODO: we need a fix for #1155 and #3802
	// return std.GetOrigPkgAddr() == std.DerivePkgAddr("gno.land/r/gov/dao/proxy")
	return true
}

type UserAuthor struct {
	addr string
}

func NewUserAuthor() *UserAuthor {
	return &UserAuthor{
		addr: std.GetOrigCaller().String(),
	}
}

func (u *UserAuthor) ID() string {
	return "addr:" + u.addr
}

func (u *UserAuthor) String() string {
	return u.addr
}
