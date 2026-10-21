package impl

import (
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	implv0 "gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
)

// defaultSupermajority is the law impl/v1 starts with, and the weakest law an
// install proposal may carry.
const defaultSupermajority = 66.66 // Two thirds

func init() {
	law = &Law{
		Supermajority: defaultSupermajority,
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
// this realm's GovDAO and AllowedDAOs = [impl/v0, impl/v1].
//
// The request is built by impl/v0.NewUpgradeDaoImplRequest, so its checks
// (proposer is a member, realm path is well-formed) and its disclosure text are
// v0's. The *GovDAO is handed over inside this proposal rather than exported:
// its PostCreateProposal and ExecuteProposal trust the r/gov/dao proxy and must
// not be reachable from anywhere else. impl/v0 stays allowed, so its own
// proposal builders (a v0 treasury payment that moves the v0 funds, for
// instance) keep executing.
//
// newLaw is the law impl/v1 decides with from its first decision on. impl/v0
// keeps its law in an unexported variable with no accessor, so it cannot be
// read and carried over; the proposer states it instead, the proposal text
// shows it, and the DAO votes on it under impl/v0's own law. It is in force
// from impl/v1's first decision (see installedDAO), so impl/v1 never decides
// anything with a law the DAO did not approve. A law weaker than impl/v1's
// default is refused: a DAO that wants one lowers it by a law-change vote
// after the install, in the open.
//
// reason is caller-supplied and lands in the same rendered description as
// the law line, so it is escaped (sanitize.InlineText): markdown in it cannot
// hide what follows, e.g. by opening an HTML comment.
func NewInstallRequest(cur realm, reason string, newLaw Law) dao.ProposalRequest {
	if !(newLaw.Supermajority >= defaultSupermajority && newLaw.Supermajority <= 100) {
		panic(ufmt.Sprintf("the install law's Supermajority must be in [%v, 100], got %v", defaultSupermajority, newLaw.Supermajority))
	}
	d := &installedDAO{GovDAO: govDAO, law: newLaw}
	text := sanitize.InlineText(reason) + ufmt.Sprintf(
		"\n\n**Law after the install: Supermajority %v%%.** impl/v1 decides with this law, "+
			"carried by this proposal, from its first decision on. impl/v0's current law is "+
			"not read: check that this is the law the DAO means to keep.", newLaw.Supermajority)
	return implv0.NewUpgradeDaoImplRequest(cross(cur), d, "gno.land/r/gov/dao/impl/v1", text)
}

// installedDAO is the DAO value an install proposal hands to dao.UpdateImpl:
// this realm's GovDAO plus the law the proposal carried.
//
// The carried law is applied on the first proxy-validated PreExecuteProposal
// after each installation, the only place a law is read, so no decision is
// ever made with another law. Later law changes (NewChangeLawRequest) stand
// until the next installation.
//
// "Each installation" is read from r/gov/dao's InstallGeneration, which
// UpdateImpl increments on every non-nil DAO assignment, including assigning
// this same value again. The proxy keeps a request's executor, so an install
// request can be submitted twice; executing it again bumps the generation and
// re-applies its law. Nothing a proposal or its executor reports is consulted,
// and an allowlist-only update (nil DAO) does not count as an installation.
type installedDAO struct {
	*GovDAO
	law Law
}

// appliedGeneration is the proxy installation generation at which an
// installedDAO last applied its carried law.
var appliedGeneration uint64

func (d *installedDAO) PreExecuteProposal(_ int, rlm realm, pid dao.ProposalID) (bool, error) {
	// Only a call through the r/gov/dao proxy applies the law; the proxy only
	// calls the installed DAO. Anything else falls through to the same check
	// inside GovDAO.PreExecuteProposal and is refused there.
	// The proxy calls PreExecuteProposal on the current implementation only,
	// so after the check this is the installedDAO the generation belongs to.
	if gen := dao.InstallGeneration(); d.GovDAO.isValidCall(0, rlm) && appliedGeneration != gen {
		l := d.law
		law = &l
		appliedGeneration = gen
	}
	return d.GovDAO.PreExecuteProposal(0, rlm, pid)
}
