package impl

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

var ErrMemberNotFound = errors.New("member not found")

type GovDAO struct {
	pss      ProposalsStatuses
	pssPager *pager.Pager
	router   *mux.Router
}

func NewGovDAO() *GovDAO {
	pss := NewProposalsStatuses()
	d := &GovDAO{
		pss:      pss,
		pssPager: pager.NewPager(pss.Tree, 3, true),
	}

	r := mux.NewRouter()

	r.HandleFunc("/", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(d.renderActiveProposals(req.RawPath))
	})

	r.HandleFunc("", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(d.renderActiveProposals(req.RawPath))
	})

	r.HandleFunc("{pid}", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(d.renderProposal(req.GetVar("pid")))
	})

	r.HandleFunc("{pid}/votes", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(d.renderVotesForProposal(req.GetVar("pid")))
	})

	d.router = r
	_govdao = d

	return d
}

// Setting this to a global variable forces attaching the GovDAO struct to this realm.
var _govdao *GovDAO

func (g *GovDAO) PreCreateProposal(r dao.ProposalRequest) (dao.Author, error) {
	if !g.isValidCall() {
		return nil, errors.New("proposal creation must be done directly by a user")
	}

	return NewUserAuthor(), nil
}

func (g *GovDAO) PostCreateProposal(r dao.ProposalRequest, pid dao.ProposalID) {
	// Tiers Allowed to Vote
	tatv := []string{memberstore.T1, memberstore.T2, memberstore.T3}
	switch v := r.Metadata().(type) {
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

func (g *GovDAO) VoteOnProposal(r dao.VoteRequest) error {
	if !g.isValidCall() {
		return errors.New("proposal voting must be done directly by a user")
	}

	caller := std.OriginCaller()
	mem, tie := memberstore.Get().GetMember(caller)
	if mem == nil {
		return ErrMemberNotFound
	}

	status := g.pss.GetStatus(r.ProposalID)

	if status.Denied || status.Accepted {
		return errors.New(ufmt.Sprintf("proposal closed. Accepted: %v", status.Accepted))
	}

	if !status.IsAllowed(tie) {
		return errors.New("member on specified tier is not allowed to vote on this proposal")
	}

	mVoted, _ := status.AllVotes.GetMember(caller)
	if mVoted != nil {
		return errors.New("already voted on proposal")
	}

	switch r.Option {
	case dao.YesVote:
		status.AllVotes.SetMember(tie, caller, mem)
		status.YesVotes.SetMember(tie, caller, mem)
	case dao.NoVote:
		status.AllVotes.SetMember(tie, caller, mem)
		status.NoVotes.SetMember(tie, caller, mem)
	default:
		return errors.New("voting can only be YES or NO")
	}

	return nil
}

func (g *GovDAO) PreGetProposal(pid dao.ProposalID) error {
	return nil
}

func (g *GovDAO) PostGetProposal(pid dao.ProposalID, p *dao.Proposal) error {
	return nil
}

func (g *GovDAO) PreExecuteProposal(pid dao.ProposalID) (bool, error) {
	if !g.isValidCall() {
		return false, errors.New("proposal execution must be done directly by a user")
	}
	status := g.pss.GetStatus(pid)
	if status.Denied || status.Accepted {
		return false, errors.New(ufmt.Sprintf("proposal already executed. Accepted: %v", status.Accepted))
	}

	if status.YesPercent() >= law.Supermajority {
		status.Accepted = true
		return true, nil
	}

	if status.NoPercent() >= law.Supermajority {
		status.Denied = true
		return false, nil
	}

	return false, errors.New(ufmt.Sprintf("proposal didn't reach supermajority yet: %v", law.Supermajority))
}

func (g *GovDAO) renderActiveProposals(url string) string {
	out := "# Active Proposals:\n"

	page, err := g.pssPager.GetPageByPath(url)
	if err != nil {
		out += ufmt.Sprintf("Error getting selected page: %v", err.Error())
		return out
	}

	for _, item := range page.Items {
		out += g.renderProposal(item.Key)
	}

	out += page.Picker("")

	return out
}

func (g *GovDAO) renderProposal(sPid string) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}
	ps := g.pss.GetStatus(dao.ProposalID(pid))

	p := dao.MustGetProposal(dao.ProposalID(pid))

	out := ""
	out += ufmt.Sprintf("## Proposal with id: %v", sPid)
	out += StringifyProposal(p)
	out += "\n\n"
	out += ps.String()
	out += ufmt.Sprintf("- [Go to votes list](:%v/votes).", sPid)
	out += "\n\n"

	return out
}

func (g *GovDAO) renderVotesForProposal(sPid string) string {
	pid, err := strconv.Atoi(sPid)
	if err != nil {
		panic(err.Error())
	}
	ps := g.pss.GetStatus(dao.ProposalID(pid))

	out := ""
	out += ufmt.Sprintf("## Voters for Proposal with id: %v\n\n", sPid)
	out += StringifyVotes(ps)

	return out
}

func (g *GovDAO) Render(path string) string {
	return g.router.Render(path)
}

func (g *GovDAO) isValidCall() bool {
	// only calls from the dao proxy are allowed
	// return std.OriginPkgAddress() == std.DerivePkgAddr("gno.land/r/gov/dao") // FIX: https://github.com/gnolang/gno/issues/3976
	return true
}

type UserAuthor struct {
	addr string
}

func NewUserAuthor() *UserAuthor {
	return &UserAuthor{
		addr: std.OriginCaller().String(),
	}
}

func (u *UserAuthor) ID() string {
	return "addr:" + u.addr
}

func (u *UserAuthor) String() string {
	return u.addr
}
