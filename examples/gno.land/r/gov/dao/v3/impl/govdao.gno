package impl

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

var ErrMemberNotFound = errors.New("member not found")

type GovDAO struct {
	pss    ProposalsStatuses
	render *render
}

func NewGovDAO() *GovDAO {
	pss := NewProposalsStatuses()
	d := &GovDAO{
		pss: pss,
	}

	d.render = NewRender(d)

	// There was no realm, from main(), so it succeeded, And
	// when returning, there was no finalization.  We don't
	// finalize anyways because there wasn't a realm boundary.
	// XXX make filetest main package a realm.
	//
	// filetest.init() ->
	//   v3/init.Init() ->
	//     NewGovDAO() ->
	//       returns an unsaved DAO NOTE NO REALM!
	//     dao.UpdateImpl =>
	//       saves dao under
	//
	// r/gov/dao.CrossPropposal() ->
	//   proposals.SetProposal(),
	//     that proposal lives in r/gov/dao.
	// r/gov/dao.ExecuteProposal() ->
	//   g.PreExecuteProposal() ->
	//     XXX g.test = 1 fails, owned by gov/dao.
	//
	//
	_govdao = d

	return d
}

// Setting this to a global variable forces attaching the GovDAO struct to this
// realm. TODO replace with future `attach()`.
var _govdao *GovDAO

func (g *GovDAO) PreCreateProposal(ctx *dao.Context, r dao.ProposalRequest) (std.Address, error) {
	crossing()

	if !g.isValidCall(ctx) {
		return "", errors.New("proposal creation must be done directly by a user")
	}

	return std.OriginCaller(), nil
}

func (g *GovDAO) PostCreateProposal(ctx *dao.Context, r dao.ProposalRequest, pid dao.ProposalID) {
	crossing()

	// Tiers Allowed to Vote
	tatv := []string{memberstore.T1, memberstore.T2, memberstore.T3}
	switch v := r.Filter().(type) {
	case FilterByTier:
		// only members from T1 are allowed to vote when adding new members to T1
		if v.Tier == memberstore.T1 {
			tatv = []string{memberstore.T1}
		}
		// only members from T1 and T2 are allowed to vote when adding new members to T2
		if v.Tier == memberstore.T2 {
			tatv = []string{memberstore.T1, memberstore.T2}
		}
	}

	pids := ufmt.Sprintf("%v", int(pid))
	g.pss.Set(pids, newProposalStatus(tatv))
}

func (g *GovDAO) VoteOnProposal(ctx *dao.Context, r dao.VoteRequest) error {
	crossing()

	if !g.isValidCall(ctx) {
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

func (g *GovDAO) PreGetProposal(ctx *dao.Context, pid dao.ProposalID) error {
	crossing()

	return nil
}

func (g *GovDAO) PostGetProposal(ctx *dao.Context, pid dao.ProposalID, p *dao.Proposal) error {
	crossing()

	return nil
}

func (g *GovDAO) PreExecuteProposal(ctx *dao.Context, pid dao.ProposalID) (bool, error) {
	crossing()

	if !g.isValidCall(ctx) {
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

func (g *GovDAO) Render(ctx *dao.Context, path string) string {
	return g.render.Render(path, ctx.CurrentRealm.PkgPath())
}

func (g *GovDAO) isValidCall(ctx *dao.Context) bool {
	// We need to verify two cases:
	// 1: r/gov/dao (proxy) was called directly by an user
	// 2: r/gov/dao/v3/impl was called directly by an user

	// case 1
	// if ctx.CurrentRealm.PkgPath() == "gno.land/r/gov/dao" && ctx.PrevRealm.IsUser() {
	// 	return true
	// }

	// case 2
	// if ctx.CurrentRealm.PkgPath() == "gno.land/r/gov/dao/v3/impl" && ctx.PrevRealm.IsUser() {
	// 	return true
	// }

	// return false

	return true // TODO: FIX: https://github.com/gnolang/gno/issues/4066
}
