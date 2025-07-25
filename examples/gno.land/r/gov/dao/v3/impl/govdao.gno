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
	func(cur realm) {
		// TODO: replace with future attach()
		_govdao = d
	}(cross)

	return d
}

// Setting this to a global variable forces attaching the GovDAO struct to this
// realm. TODO replace with future `attach()`.
var _govdao *GovDAO

func (g *GovDAO) PreCreateProposal(r dao.ProposalRequest) (std.Address, error) {
	if !g.isValidCall() {
		return "", errors.New(ufmt.Sprintf("proposal creation must be done directly by a user or through the r/gov/dao proxy. current realm: %v; previous realm: %v",
			std.CurrentRealm(), std.PreviousRealm()))
	}

	// Verify that the one creating the proposal is a member.
	caller := std.OriginCaller()
	mem, _ := getMembers(cross).GetMember(caller)
	if mem == nil {
		return caller, errors.New("only members can create new proposals")
	}

	return caller, nil
}

func (g *GovDAO) PostCreateProposal(r dao.ProposalRequest, pid dao.ProposalID) {
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
	g.pss.Set(pid.String(), newProposalStatus(tatv))
}

func (g *GovDAO) VoteOnProposal(r dao.VoteRequest) error {
	if !g.isValidCall() {
		return errors.New("proposal voting must be done directly by a user")
	}

	caller := std.OriginCaller()
	mem, tie := getMembers(cross).GetMember(caller)
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

func (g *GovDAO) Render(pkgPath string, path string) string {
	return g.render.Render(pkgPath, path)
}

func (g *GovDAO) isValidCall() bool {
	// We need to verify two cases:
	// 1: r/gov/dao (proxy) functions called directly by an user
	// 2: r/gov/dao/v3/impl methods called directly by an user

	// case 1
	if std.CurrentRealm().PkgPath() == "gno.land/r/gov/dao" {
		// called directly by an user through MsgCall
		if std.PreviousRealm().IsUser() {
			return true
		}
		isMsgRun := std.DerivePkgAddr(std.PreviousRealm().PkgPath()) == std.OriginCaller()
		// called directly by an user through MsgRun
		if isMsgRun {
			return true
		}
	}

	// case 2
	if std.CurrentRealm().IsUser() {
		return true
	}

	return false
}
