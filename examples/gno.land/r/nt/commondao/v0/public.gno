package commondao

import (
	// unsafe is used only for the CurrentRealm call in assertRunningPath when
	// checking the deployment path. Caller authentication uses cur.Previous().
	"chain/runtime/unsafe"
	"strings"

	"gno.land/p/nt/commondao/v0"
)

// assertCurrent guards every public crossing entry. Authentication reads
// cur.Previous() and Execute mints cur.Sub(...) for the executor; both
// require cur to be the live top-of-frame realm (AGENTS.md / interrealm-v2).
// Every entry is only reachable via a cross today, so this is
// defense-in-depth against a future non-crossing caller.
func assertCurrent(cur realm) {
	if !cur.IsCurrent() {
		panic("commondao: cur realm is not current")
	}
}

// assertRunningPath fails closed if this realm is deployed at a package
// path other than pkgPath. Treasury addresses derive from the pkgPath
// const (daoAddress), while Execute mints the DAO sub under the running
// realm's own path (cur.Sub). Those agree only when the running path is
// pkgPath; a copy deployed elsewhere would read balances at one address
// and send from another, turning a clean StatusFailed into a mismatch. We
// reject that at genesis rather than silently diverge.
func assertRunningPath() {
	if got := unsafe.CurrentRealm().PkgPath(); got != pkgPath {
		panic("commondao: realm deployed at " + got + ", expected " + pkgPath)
	}
}

// assertCallerIsUser rejects a code realm creating a DAO on someone's
// behalf. New keys the invite and the creators set on the caller, and
// invites are granted to people, so the caller must be the user itself:
// either a direct MsgCall (IsUserCall) or that user's own MsgRun frame
// (IsUserRun, whose address is the signer's own). Both are IsUser; a
// published realm is neither.
//
// Relaying is what this excludes. If a realm could redeem an invite, one
// invited realm would become a DAO factory for un-invited callers — the
// vector pr6012_commondao_ownership_rescope.md closed by keying on the
// transaction origin. Keying on the caller closes it too, but only while
// the caller cannot be a realm.
//
// It bounds who may *create* a DAO, not who may sit on one: the members
// argument may still name realm addresses, and sub-DAO councils are set
// by proposal (see CreateSubDAOProposal). Handing a DAO to a realm is a
// council decision, not a creation-time one.
func assertCallerIsUser(cur realm) {
	if !cur.Previous().IsUser() {
		panic("only a user can create a DAO")
	}
}

// Invite invites a user to the realm.
// A user invitation is required to start creating new DAOs.
func Invite(cur realm, invitee address) {
	assertCurrent(cur)

	if !invitee.IsValid() {
		panic("invalid address")
	}

	dao := mustGetDAO(CommonDAOID)
	caller := cur.Previous().Address()
	if !dao.Council().Has(caller) {
		panic("unauthorized")
	}

	invites.Set(invitee.String(), caller.String())
}

// IsInvited checks if an address has an invitation to the realm.
func IsInvited(addr address) bool {
	return isInvited(addr)
}

// New creates a new CommonDAO and returns its ID.
// The caller must be a user, not a realm (see assertCallerIsUser), and must
// hold an invite. The invite is consumed when that caller creates a DAO for
// the first time, after which the caller may create further DAOs freely.
// The caller becomes a council member along with any additional member
// addresses listed on separate lines; those may be realm addresses, so a
// DAO can be handed to a realm by seating it here and resigning. DAOs with
// a parent are created through proposals (see CreateSubDAOProposal).
func New(cur realm, name, purpose, description, members string) uint64 {
	assertCurrent(cur)
	assertCallerIsUser(cur)
	caller := cur.Previous().Address()

	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	purpose = strings.TrimSpace(purpose)
	assertDAOPurposeIsValid(purpose)

	description = strings.TrimSpace(description)
	assertDAODescriptionIsValid(description)

	// The invite is consumed once. The direct user caller is then recorded and
	// can create further DAOs without another invite.
	if !isCreator(caller) {
		assertIsInvited(caller)
		invites.Remove(caller.String())
		creators.Set(caller.String(), struct{}{})
	}

	dao := createDAO(name, purpose, description, parseInitialMembers(caller, members)...)
	return dao.ID()
}

// parseInitialMembers returns the caller plus the parsed additional
// members, deduplicated.
func parseInitialMembers(caller address, members string) []address {
	addrs := parseAddresses(members)
	if containsAddress(addrs, caller) {
		return addrs
	}
	return append(addrs, caller)
}

// SetListed adds or removes a DAO from the realm's public home index.
// Listing is cosmetic — it affects only how this realm presents the DAO
// in its own UI — so any single council member of the DAO may toggle it,
// like Resign. It defaults to off.
func SetListed(cur realm, daoID uint64, listed bool) {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	if dao.IsDeleted() {
		panic(commondao.ErrDAOIsDeleted)
	}

	assertCallerIsCouncilMember(cur.Previous().Address(), dao)
	setListed(daoID, listed)
}

// IsListed reports whether a DAO appears in the realm's public home index.
func IsListed(daoID uint64) bool {
	return isListed(daoID)
}

// GetView returns a read only view of a common DAO searched by ID.
func GetView(daoID uint64) commondao.ReadonlyCommonDAO {
	return mustGetDAO(daoID).Readonly()
}

// GetBylawsDoc returns the text of a DAO's bylaws/mandates document, or
// an empty string when the document does not exist (a stored document is
// never empty).
func GetBylawsDoc(daoID uint64, path string) string {
	mustGetDAO(daoID)

	if set := bylawsView(daoID); set != nil {
		text, _ := set.Get(path)
		return text
	}
	return ""
}

// ListBylawsDocs returns the sorted paths of a DAO's bylaws/mandates
// documents under a prefix (empty prefix lists all).
func ListBylawsDocs(daoID uint64, prefix string) []string {
	mustGetDAO(daoID)

	if set := bylawsView(daoID); set != nil {
		return set.List(prefix)
	}
	return nil
}

// Vote submits a vote for a DAO proposal.
// Voting is allowed to the members of the proposal's electorate: the
// council snapshot taken when the proposal was created.
func Vote(cur realm, daoID, proposalID uint64, vote commondao.VoteChoice, reason string) {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	err := dao.Vote(caller, proposalID, vote, reason)
	if err != nil {
		panic(err)
	}
}

// Funded is the optional contract a proposal definition implements when its
// executor moves funds from a DAO other than the proposal's host: it names,
// by ID, the DAO whose sub-identity address funds the executor. Execute
// resolves that DAO, mints its terminal RealmSend-only sub and passes it to
// the ExecFunc; a definition that does not implement Funded receives the
// host DAO's own sub by default. The returned ID must identify the DAO the
// definition validates its fund movement against (e.g. the DAO being spent,
// swept or dissolved).
//
// It lives realm-side, not in /p/: minting a DAO sub needs the host realm's
// cur (cur.Sub), so only the host — never the package — can honor it. /p/
// dispatches CapExempt/Executable/Validable itself, but Execute (the host)
// is the sole consumer of Funded, so the package has no reason to know it.
type Funded interface {
	// FundingDAOID returns the ID of the DAO whose sub-address funds the
	// executor.
	FundingDAOID() uint64
}

// Execute executes a DAO proposal.
//
// Executing a proposal that passed early (decided by the default Council
// rules before its voting deadline) requires the caller to be a council
// member. Once the voting deadline has passed execution is permissionless:
// the tally is deterministic, so anyone can finalize the proposal.
func Execute(cur realm, daoID, proposalID uint64) {
	assertCurrent(cur)

	// Re-entrancy latch: a proposal executor must not trigger another
	// Execute. Without this, an executor could finalize a second proposal
	// mid-run — e.g. dissolve its own DAO and leave this frame finalizing on
	// a deleted DAO. The latch is global (one executor per tx); it does not
	// block Vote/Create*, so an executor may still act as its DAO elsewhere.
	enterExecute()
	defer leaveExecute()

	dao := mustGetDAO(daoID)
	p := dao.GetProposal(proposalID)
	if p == nil {
		panic(commondao.ErrProposalNotFound)
	}

	// Before the deadline, only a council member may execute an
	// early-passed proposal. Once the deadline passes the tally is
	// deterministic, so finalization is permissionless.
	if !p.HasVotingDeadlinePassed() {
		assertCallerIsCouncilMember(cur.Previous().Address(), dao)
	}

	// Mint the sub-identity that funds the executor and pass it in. The
	// operative DAO is the host by default; a fund-moving definition
	// (Funded) may name a different DAO by ID (e.g. clawback sweeps the
	// target, sub-DAO dissolution sweeps the dissolved descendant). Non-fund
	// executors ignore the sub. The sub is terminal and RealmSend-only, so
	// the executor can move value only from this one DAO address.
	op := daoID
	if f, ok := p.Definition().(Funded); ok {
		op = f.FundingDAOID()
	}
	sub := cur.Sub(subpathOf(op))

	err := dao.Execute(proposalID, sub)
	if err != nil {
		panic(err)
	}
}

// Withdraw withdraws an active DAO proposal that has no votes.
// Only the proposal creator can withdraw it.
func Withdraw(cur realm, daoID, proposalID uint64) {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	p := dao.GetProposal(proposalID)
	if p == nil {
		panic(commondao.ErrProposalNotFound)
	}

	if p.Creator() != cur.Previous().Address() {
		panic("only the proposal creator can withdraw it")
	}

	if err := dao.Withdraw(proposalID); err != nil {
		panic(err)
	}
}

// Resign removes the caller from a DAO council.
// The last remaining council member cannot resign.
func Resign(cur realm, daoID uint64) {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	if dao.IsDeleted() {
		panic(commondao.ErrDAOIsDeleted)
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	if err := dao.UpdateCouncil(nil, []address{caller}); err != nil {
		panic(err)
	}
}

func isInvited(addr address) bool {
	return invites.Has(addr.String())
}

func assertIsInvited(addr address) {
	if !isInvited(addr) {
		panic("unauthorized")
	}
}

func assertDAONameIsValid(name string) {
	if name == "" {
		panic("DAO name is empty")
	}

	if len(name) > 60 {
		panic("DAO name is too long, max length is 60 characters")
	}
}

func assertDAOPurposeIsValid(purpose string) {
	if purpose == "" {
		panic("DAO purpose is empty")
	}

	if len(purpose) > 250 {
		panic("DAO purpose is too long, max length is 250 characters")
	}
}

func assertDAODescriptionIsValid(description string) {
	if len(description) > 250 {
		panic("DAO description is too long, max length is 250 characters")
	}
}

func assertCallerIsCouncilMember(caller address, dao *commondao.CommonDAO) {
	if !dao.Council().Has(caller) {
		panic("caller is not a council member")
	}
}
