package impl

import (
	"gno.land/r/gov/dao"
	implv0 "gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
)

func init() {
	law = &Law{
		Supermajority: 66.66, // Two thirds
	}
}

func Render(cur realm, in string) string {
	// Same-realm: pass cur to govDAO.Render (also crossing, but same realm
	// so use the literal cur form rather than cross(cur)).
	return govDAO.Render(cur, cur.PkgPath(), in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(cur realm, addr address) {
	// AGENTS.md: check IsCurrent() before cur.Previous() in a crossing function.
	if !cur.IsCurrent() {
		panic("AddMember: realm value is not the caller's live cur")
	}
	// address args are not VM-validated (the raw MsgCall string is stored), so
	// reject a non-bech32 addr here — matching InitWithUsers. This keeps the
	// member store free of unauthenticable keys and of markdown/pipe/HTML
	// metachars that would otherwise inject into the members render table.
	if !addr.IsValid() {
		panic("invalid member address: " + addr.String())
	}
	caller := cur.Previous()
	if !caller.IsUser() {
		panic("this function must be called by an EOA through msg call or msg run")
	}
	m, t := memberstore.Get(0, cur).GetMember(caller.Address())
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 && t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	m.RemoveInvitationPoint()

	if err := memberstore.Get(0, cur).SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

// NewInstallRequest builds the proposal that installs this implementation
// (impl/v1) in place of impl/v0. It is how v1 goes live: a v0-era member
// creates it, the DAO votes on it, and executing it runs dao.UpdateImpl with
// this realm's singleton and AllowedDAOs = [impl/v0, impl/v1].
//
// The request is built by impl/v0.NewUpgradeDaoImplRequest, so its checks
// (proposer is a member, realm path is well-formed) and its disclosure text are
// v0's. Passing the singleton from here, rather than exporting it, keeps the
// *GovDAO out of other realms' hands: its PostCreateProposal and
// ExecuteProposal trust the r/gov/dao proxy and must not be reachable from
// anywhere else. impl/v0 stays allowed, so its own proposal builders (a v0
// treasury payment that moves the v0 funds, for instance) keep executing.
func NewInstallRequest(cur realm, reason string) dao.ProposalRequest {
	return implv0.NewUpgradeDaoImplRequest(cross(cur), govDAO, "gno.land/r/gov/dao/impl/v1", reason)
}
