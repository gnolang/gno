package impl

import (
	"chain"
	"chain/runtime"
	"errors"

	"gno.land/p/nt/ufmt/v0"
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

	// Attach to package var (impl owns _govdao). Plain assignment is
	// fine — we're in impl's package, no realm transition needed.
	// TODO: replace with future attach().
	_govdao = d

	return d
}

// Setting this to a global variable forces attaching the GovDAO struct to this
// realm. TODO replace with future `attach()`.
var _govdao *GovDAO

func (g *GovDAO) PreCreateProposal(_ int, rlm realm, r dao.ProposalRequest) (address, error) {
	if !g.isValidCall(0, rlm) {
		return "", errors.New(ufmt.Sprintf("proposal creation must be done directly by a user or through the r/gov/dao proxy. caller realm: %v; caller's previous: %v",
			rlm, rlm.Previous()))
	}

	// Verify that the one creating the proposal is a member.
	caller := runtime.OriginCaller()
	mem, _ := getMembers(cross2(rlm)).GetMember(caller)
	if mem == nil {
		return caller, errors.New("only members can create new proposals")
	}

	return caller, nil
}

func (g *GovDAO) PostCreateProposal(_ int, rlm realm, r dao.ProposalRequest, pid dao.ProposalID) {
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

func (g *GovDAO) VoteOnProposal(_ int, rlm realm, r dao.VoteRequest) error {
	if !g.isValidCall(0, rlm) {
		return errors.New("proposal voting must be done directly by a user")
	}

	caller := runtime.OriginCaller()
	mem, tie := getMembers(cross2(rlm)).GetMember(caller)
	if mem == nil {
		return ErrMemberNotFound
	}

	status := g.pss.GetStatus(r.ProposalID)
	if status == nil {
		return errors.New("proposal not found")
	}

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
	case dao.AbstainVote:
		status.AllVotes.SetMember(tie, caller, mem)
		status.AbstainVotes.SetMember(tie, caller, mem)
	default:
		return errors.New("voting can only be YES, NO, or ABSTAIN")
	}

	return nil
}

func (g *GovDAO) PreExecuteProposal(_ int, rlm realm, pid dao.ProposalID) (bool, error) {
	if !g.isValidCall(0, rlm) {
		return false, errors.New("proposal execution must be done directly by a user")
	}
	status := g.pss.GetStatus(pid)
	if status.Denied || status.Accepted {
		return false, errors.New(ufmt.Sprintf("proposal already executed. Accepted: %v", status.Accepted))
	}

	if status.YesPercent(0, rlm) >= law.Supermajority {
		status.Accepted = true
		return true, nil
	}

	if status.NoPercent(0, rlm) >= law.Supermajority {
		status.Denied = true
		return false, nil
	}

	return false, errors.New(ufmt.Sprintf("proposal didn't reach supermajority yet: %v", law.Supermajority))
}

func (g *GovDAO) ExecuteProposal(_ int, rlm realm, pid dao.ProposalID, e dao.Executor) error {
	if e == nil {
		panic("an executor is required to execute the proposal")
	}

	status := g.pss.GetStatus(pid)
	if status == nil {
		panic("proposal not found")
	}

	err := e.Execute(cross2(rlm))
	if err != nil {
		status.Accepted = false
		status.Denied = true
		status.DeniedReason = "execution failed: " + err.Error()
	}
	return err
}

func (g *GovDAO) Render(cur realm, pkgPath string, path string) string {
	// Same-realm dispatch: pass cur through as data (non-crossing).
	return g.render.Render(0, cur, pkgPath, path)
}

// isValidCall verifies that the impl method is being invoked from the
// r/gov/dao proxy via a legitimate user transaction (MsgCall or MsgRun).
//
// The proxy passes its own crossing-frame Cur as rlm when calling the
// impl methods. rlm.IsCurrent() rejects stale or stashed realm values —
// a malicious realm cannot replay a captured proxy cur to impersonate
// the proxy. After the IsCurrent() check:
//   - rlm.PkgPath() == "gno.land/r/gov/dao" identifies the proxy
//     unforgeably (pkg path is set at mint time by installCrossingCur).
//   - rlm.Previous() is the caller of the proxy.
//
// The "called directly by user bypassing the proxy" path (formerly
// case 2: runtime.CurrentRealm().IsUser()) is gone — the impl methods
// are non-crossing-with-realm-arg now, so a direct user MsgCall to
// them would not provide a valid rlm; the IsCurrent() check rejects
// any forged realm. The proxy is the only legitimate entrypoint.
func (g *GovDAO) isValidCall(_ int, rlm realm) bool {
	if !rlm.IsCurrent() {
		return false
	}
	if rlm.PkgPath() != "gno.land/r/gov/dao" {
		return false
	}
	prev := rlm.Previous()
	// MsgCall: proxy was called directly by an EOA (UserRealm).
	if prev.IsUser() {
		return true
	}
	// MsgRun: proxy was called from the ephemeral run realm; that
	// realm's package address equals the EOA OriginCaller.
	return chain.PackageAddress(prev.PkgPath()) == runtime.OriginCaller()
}
