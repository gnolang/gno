package govdao

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

var DAOPath = "gno.land/r/demo/govdao"
var ErrInvalidRealm = errors.New("this can be only executed by govDAO realm")

var router *mux.Router

var proposals Proposals // TODO move inactive proposals to an archive
var proposalsStatuses ProposalsStatuses
var proposalIDSeq int

func init() {
	proposals = Proposals{avl.NewTree()}
	proposalsStatuses = ProposalsStatuses{avl.NewTree()}

	router = mux.NewRouter()
	router.HandleFunc("law", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(lawStringifier(theLaw, members))
	})

	router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderActiveProposals())
	})
}

func renderActiveProposals() string {
	out := "## Active Propoposals:\n"
	proposals.Iterate("", "", func(key string, valueProposal interface{}) bool {
		p, _ := valueProposal.(*Proposal)
		ps := proposalsStatuses.GetStatus(key)
		if ps.Inactive {
			return false
		}

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

func Render(p string) string {
	return router.Render(p)
}

func Vote(pid int, vote string) {
	pids := ufmt.Sprintf("%v", pid)
	ps := proposalsStatuses.GetStatus(pids)
	if ps == nil {
		panic(ufmt.Sprintf("proposal with pid %v not found", pids))
	}

	if ps.Accepted || ps.Denied {
		panic("Proposal out of the voting period")
	}

	caller := std.GetOrigCaller()

	m, tn := members.GetMember(caller)
	if m == nil {
		panic("Caller is not a member so cannot vote")
	}

	mVoted, _ := ps.AllVotes.GetMember(caller)
	if mVoted != nil {
		panic("already voted on proposal")
	}

	switch strings.ToLower(vote) {
	case "yes", "y":
		ps.YesVotes.SetMember(tn, caller, m)
		ps.AllVotes.SetMember(tn, caller, m)
	case "no", "n":
		ps.NoVotes.SetMember(tn, caller, m)
		ps.AllVotes.SetMember(tn, caller, m)
	default:
		panic(ufmt.Errorf("wrong vote string used: %v", vote))
	}

	ts := theLaw.Tiers
	totalPower := members.GetTotalPower(ts)
	yesPower := ps.YesVotes.GetTotalPower(ts)
	noPower := ps.NoVotes.GetTotalPower(ts)

	if (yesPower/totalPower)*100 >= float64(theLaw.VotingPercentageNeeded) {
		ps.Accepted = true
		return
	}

	if (noPower/totalPower)*100 >= float64(theLaw.VotingPercentageNeeded) {
		ps.Denied = true
		return
	}
}

func WithdrawPrpoposal(pid int) {
	caller := std.GetOrigCaller()
	err := theLaw.CheckProposalWithdrawn(members, caller)
	if err != nil {
		panic(err)
	}

	// TODO: withdraw proposal by ID
}

func MemberProposalAdd(addr std.Address, tier string) int {
	caller := std.GetOrigCaller()
	err := theLaw.CheckMemberAddProposal(members, caller, tier, addr)
	if err != nil {
		panic(err)
	}
	exec := NewExecutor(addNewMemberCallback, &addMemberMetadata{
		Tier:   tier,
		Addr:   addr,
		Member: &Member{},
	})

	prop := &Proposal{
		Author:      caller,
		Title:       "New Member porposal",
		Description: "This prpoposal is looking to add a new member to the board.",
		Executor:    exec,
	}

	return newProposal(prop)
}

func MemberProposalPromote(addr std.Address, tier string) int {
	caller := std.GetOrigCaller()

	err := theLaw.CheckMemberPromoteProposal(members, caller, tier, addr)
	if err != nil {
		panic(err)
	}

	m, tn := members.GetMember(addr)
	if m == nil {
		panic("proposed address is not a member")
	}

	exec := NewExecutor(promoteNewMemberCallback, &promoteMemberMetadata{
		FromTier: tn,
		ToTier:   tier,
		Addr:     addr,
	})

	prop := &Proposal{
		Author:      caller,
		Title:       "Promote Member porposal",
		Description: "This prpoposal is looking to promote a member to a new tier.",
		Executor:    exec,
	}

	return newProposal(prop)
}

func MemberProposalWithdrawn(addr std.Address) int {
	caller := std.GetOrigCaller()

	err := theLaw.CheckMemberWithdrawnProposal(members, caller, addr)
	if err != nil {
		panic(err)
	}

	// TODO: implement proposal

	return 0
}

func LawProposalChange(law *Law) int {
	caller := std.GetOrigCaller()

	exec := NewExecutor(changeTheLawCallback, &lawMetadata{
		NewLaw: law,
	})

	prop := &Proposal{
		Author:      caller,
		Title:       "New Law proposal",
		Description: "This prpoposal will change the implementation of the law. Read carefully the source code and the metadata string representation to verify that everything is as expected.",
		Executor:    exec,
	}

	return newProposal(prop)
}

func NewProposal(p *Proposal) int {
	p.Author = std.GetOrigCaller()
	return newProposal(p)
}

func newProposal(p *Proposal) int {
	pid := proposalIDSeq
	pids := ufmt.Sprintf("%v", pid)
	proposals.Set(pids, p)
	proposalsStatuses.Set(pids, newProposalStatus())
	proposalIDSeq = proposalIDSeq + 1

	return pid
}
