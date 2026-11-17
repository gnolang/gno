package impl

import (
	"chain"
	"chain/runtime/unsafe"
	"errors"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/memberstore/v0"
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

	// GovDAO is this realm's type, so d belongs to this realm wherever it is
	// stored; only this realm's crossing code writes it.
	return d
}

func (g *GovDAO) PreCreateProposal(cur realm, r dao.ProposalRequest) (address, error) {
	if !g.isValidCall(cur) {
		return "", errors.New(ufmt.Sprintf("proposal creation must be done directly by a user or through the r/gov/dao proxy. caller realm: %v; caller's previous: %v",
			cur.Previous(), cur.Previous().Previous()))
	}

	// Verify that the one creating the proposal is a member.
	caller := unsafe.OriginCaller()
	mem, _ := getMembers(cross(cur)).GetMember(caller)
	if mem == nil {
		return caller, errors.New("only members can create new proposals")
	}

	return caller, nil
}

func (g *GovDAO) PostCreateProposal(cur realm, r dao.ProposalRequest, pid dao.ProposalID) {
	mustBeProxy(cur)
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

func (g *GovDAO) VoteOnProposal(cur realm, r dao.VoteRequest) error {
	if !g.isValidCall(cur) {
		return errors.New("proposal voting must be done directly by a user")
	}

	caller := unsafe.OriginCaller()
	mem, tie := getMembers(cross(cur)).GetMember(caller)
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

func (g *GovDAO) PreExecuteProposal(cur realm, pid dao.ProposalID) (bool, error) {
	if !g.isValidCall(cur) {
		return false, errors.New("proposal execution must be done directly by a user")
	}
	status := g.pss.GetStatus(pid)
	if status == nil {
		// Unknown to this implementation: either an unknown id, or a
		// proposal created before a DAO upgrade replaced this GovDAO
		// (statuses live on the instance, so a fresh instance has none).
		// Mirrors VoteOnProposal above, which returns this same error for
		// the same lookup. Without it the nil deref panics with an opaque
		// "runtime error: nil pointer dereference" that reads like a VM
		// fault. It reports the condition only — see the ADR for why the
		// proposal still cannot be resolved.
		return false, errors.New("proposal not found")
	}
	if status.Denied || status.Accepted {
		return false, errors.New(ufmt.Sprintf("proposal already executed. Accepted: %v", status.Accepted))
	}

	if status.YesPercent(0, cur) >= law.Supermajority {
		status.Accepted = true
		return true, nil
	}

	if status.NoPercent(0, cur) >= law.Supermajority {
		status.Denied = true
		return false, nil
	}

	return false, errors.New(ufmt.Sprintf("proposal didn't reach supermajority yet: %v", law.Supermajority))
}

func (g *GovDAO) ExecuteProposal(cur realm, pid dao.ProposalID, e dao.Executor) error {
	// Executing crosses into e as this realm, a gov/dao subpackage that
	// executors trust, so only the proxy may drive it.
	mustBeProxy(cur)
	if e == nil {
		panic("an executor is required to execute the proposal")
	}

	status := g.pss.GetStatus(pid)
	if status == nil {
		panic("proposal not found")
	}

	err := e.Execute(cross(cur))
	if err != nil {
		status.Accepted = false
		status.Denied = true
		// Clamped on the way in, not just on the way out. The error comes from
		// the proposal's executor — third-party code — and this assignment is a
		// write to realm storage, paid for by whoever executes the proposal
		// rather than by whoever wrote the executor. Bounding it here means the
		// realm never stores a reason larger than it can display. The clamp at
		// the render site stays, because reasons stored before this change are
		// still unbounded.
		status.DeniedReason = "execution failed: " + clampField(err.Error(), maxRenderedReason)
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
// The impl methods are crossing: cur.Previous() is the realm that crossed in,
// which must be the proxy, and the proxy's own Previous() is its caller.
//
// This is also what makes the unsafe.OriginCaller()-based membership
// checks safe. Those key on the transaction origin (an EOA), which is
// correct only if the origin is the immediate caller. isValidCall
// guarantees exactly that: it admits prev only when prev.IsUser() (a
// direct EOA call or the origin's own ephemeral run realm) or when
// prev's package address equals the origin — never a third-party realm
// in the middle. Relaxing this to allow realm intermediaries would let
// one member's vote be cast under another origin: keep the two in sync.
// mustBeProxy panics unless the r/gov/dao proxy crossed in. These methods are
// crossing, so any realm holding a GovDAO could otherwise call them.
func mustBeProxy(cur realm) {
	if cur.Previous().PkgPath() != "gno.land/r/gov/dao" {
		panic("only the gno.land/r/gov/dao proxy may call this")
	}
}

func (g *GovDAO) isValidCall(cur realm) bool {
	proxy := cur.Previous()
	if proxy.PkgPath() != "gno.land/r/gov/dao" {
		return false
	}
	prev := proxy.Previous()
	// MsgCall: proxy was called directly by an EOA (UserRealm).
	if prev.IsUser() {
		return true
	}
	// MsgRun: proxy was called from the ephemeral run realm; that
	// realm's package address equals the EOA OriginCaller.
	return chain.PackageAddress(prev.PkgPath()) == unsafe.OriginCaller()
}
