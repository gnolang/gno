package commondao

import (
	// unsafe is used only for OriginCaller in New: invitations target
	// EOAs, so the invite is keyed to and consumed by the transaction
	// origin. Caller authentication everywhere uses cur.Previous().
	"chain/runtime/unsafe"
	"strings"

	"gno.land/p/nt/commondao/v0"
)

// Invite invites a user to the realm.
// A user invitation is required to start creating new DAOs.
func Invite(cur realm, invitee address) {
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
// An invitation is required to start creating DAOs: the transaction
// origin (an EOA) must hold an invite, which is consumed on its first
// creation; afterwards that origin may create further DAOs freely. The
// caller becomes a council member, optionally together with a newline
// separated list of additional member addresses. Sub-DAOs are created
// through proposals (see CreateSubDAOProposal).
func New(cur realm, name, purpose, description, members string) uint64 {
	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	purpose = strings.TrimSpace(purpose)
	assertDAOPurposeIsValid(purpose)

	description = strings.TrimSpace(description)
	assertDAODescriptionIsValid(description)

	// The right to create DAOs is granted to invited EOAs, so gate on the
	// transaction origin. The invite is consumed once; the origin is then
	// recorded so it can create further DAOs without another invite.
	orig := unsafe.OriginCaller()
	if !isCreator(orig) {
		assertIsInvited(orig)
		invites.Remove(orig.String())
		creators.Set(orig.String(), struct{}{})
	}

	caller := cur.Previous().Address()
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

// Vote submits a vote for a DAO proposal.
// Voting is allowed to the members of the proposal's electorate: the
// council snapshot taken when the proposal was created.
func Vote(cur realm, daoID, proposalID uint64, vote commondao.VoteChoice, reason string) {
	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	err := dao.Vote(caller, proposalID, vote, reason)
	if err != nil {
		panic(err)
	}
}

// Execute executes a DAO proposal.
//
// Executing a proposal that passed early (decided by the default Council
// rules before its voting deadline) requires the caller to be a council
// member. Once the voting deadline has passed execution is permissionless:
// the tally is deterministic, so anyone can finalize the proposal.
func Execute(cur realm, daoID, proposalID uint64) {
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

	err := dao.Execute(proposalID, cur)
	if err != nil {
		panic(err)
	}
}

// Withdraw withdraws an active DAO proposal that has no votes.
// Only the proposal creator can withdraw it.
func Withdraw(cur realm, daoID, proposalID uint64) {
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
