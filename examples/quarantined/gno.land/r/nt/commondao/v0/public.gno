package commondao

import (
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
// An invitation is required to be able to create new DAOs.
// The caller becomes the DAO owner and a council member, optionally
// together with a newline separated list of additional member addresses.
// Sub-DAOs are created through proposals (see CreateSubDAOProposal).
func New(cur realm, name, description, members string) uint64 {
	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	description = strings.TrimSpace(description)
	assertDAODescriptionIsValid(description)

	caller := cur.Previous().Address()
	if !hasOwnership(caller) {
		orig := unsafe.OriginCaller()
		assertIsInvited(orig)
		invites.Remove(orig.String())
	}

	dao := createDAO(name, description, caller, parseInitialMembers(caller, members)...)
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

// GetOptions returns a copy of the configuration of a DAO.
// It panics if caller doesn't have ownership of the DAO instance.
func GetOptions(cur realm, daoID uint64) Options {
	if !isOwner(cur.Previous().Address(), daoID) {
		panic("unauthorized")
	}

	return *getOptions(daoID)
}

// UpdateOptions changes the hosting configuration of a DAO.
// It panics if caller doesn't have ownership of the DAO instance.
// A maxActiveProposals value below one keeps the current proposal cap.
func UpdateOptions(cur realm, daoID uint64, allowListing, allowRender, allowChildren, allowExecution bool, maxActiveProposals int) {
	if !isOwner(cur.Previous().Address(), daoID) {
		panic("unauthorized")
	}

	*getOptions(daoID) = Options{
		AllowListing:   allowListing,
		AllowRender:    allowRender,
		AllowChildren:  allowChildren,
		AllowExecution: allowExecution,
	}

	// Values below one are ignored by the package cap setter
	mustGetDAO(daoID).SetMaxActiveProposals(maxActiveProposals)
}

// IsOwner checks if an address has ownership of a DAO.
func IsOwner(addr address, daoID uint64) bool {
	return isOwner(addr, daoID)
}

// TransferOwnership transfers DAO ownership to a different address.
// Ownership is indexed by root DAO, so SubDAO IDs are rejected.
func TransferOwnership(cur realm, daoID uint64, newOwner address) {
	caller := cur.Previous().Address()
	assertIsOwner(caller, daoID)

	if mustGetDAO(daoID).Parent() != nil {
		panic("ownership belongs to the root DAO, transfer using the root DAO ID")
	}

	if !newOwner.IsValid() {
		panic("invalid owner address")
	}
	ids := getOwnership(caller)
	for i, id := range ids {
		if id == daoID {
			ownership.Set(caller.String(), append(ids[:i], ids[i+1:]...))
			ownership.Set(newOwner.String(), append(getOwnership(newOwner), id))

			break
		}
	}
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
// rules before its voting deadline) requires execution to be enabled for
// the DAO and the caller to be a council member. Once the voting deadline
// has passed execution is permissionless: the tally is deterministic, so
// anyone can finalize the proposal.
func Execute(cur realm, daoID, proposalID uint64) {
	dao := mustGetDAO(daoID)
	p := dao.GetProposal(proposalID)
	if p == nil {
		panic(commondao.ErrProposalNotFound)
	}

	if !p.HasVotingDeadlinePassed() {
		o := getOptions(dao.ID())
		if !o.AllowExecution {
			panic("proposal execution is not enabled")
		}

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

func hasOwnership(addr address) bool {
	return ownership.Has(addr.String())
}

func isInvited(addr address) bool {
	return invites.Has(addr.String())
}

func isOwner(addr address, daoID uint64) bool {
	// Make sure to check the root DAO ID, in case daoID belongs to a SubDAO.
	// This is required because ownership is assigned to the root DAO.
	daoID = mustGetDAO(daoID).TopParent().ID()

	for _, id := range getOwnership(addr) {
		if id == daoID {
			return true
		}
	}
	return false
}

func assertIsInvited(addr address) {
	if !isInvited(addr) {
		panic("unauthorized")
	}
}

func assertIsOwner(addr address, daoID uint64) {
	if !isOwner(addr, daoID) {
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
