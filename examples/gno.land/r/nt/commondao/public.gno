package commondao

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/nt/commondao"
)

// TODO: Limit the number of DAOs per address, maybe discuss fees ranges to avoid spam

// Invite invites a user to the realm.
// A user invitation is required to start creating new DAOs.
func Invite(cur realm, invitee std.Address) {
	if !invitee.IsValid() {
		panic("invalid address")
	}

	dao := mustGetDAO(CommonDAOID)
	caller := std.PreviousRealm().Address()
	if !dao.Members().Has(caller) {
		panic("unauthorized")
	}

	invites.Set(invitee.String(), caller.String())
}

// IsInvited check if an address has an invitation to the realm.
func IsInvited(addr std.Address) bool {
	return isInvited(addr)
}

// New creates a new CommonDAO.
// An invitation is required to be able to create new DAOs.
func New(cur realm, name string) *commondao.CommonDAO {
	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	caller := std.CurrentRealm().Address()

	// If caller doesn't own a DAO already
	// Add them to the dao owners list
	// Then use up their invite
	if !hasOwnership(caller) {
		// why make the distinction here?
		orig := std.OriginCaller()
		assertIsInvited(orig)
		invites.Remove(orig.String())
	}

	return createDAO(name, caller)
}

// NewSubDAO creates a new SubDAO.
func NewSubDAO(cur realm, name string, parentID uint64) *commondao.CommonDAO {
	caller := std.CurrentRealm().Address()
	assertIsOwner(caller, parentID)

	parentOptions := getOptions(parentID)
	if !parentOptions.AllowChildren {
		panic("SubDAO support is not enabled")
	}

	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	parent := getDAO(parentID)
	subDAO := createSubDAO(parent, name)
	parent.Children().Append(subDAO)
	return subDAO
}

// GetOptions returns the configuration of a DAO.
// It panics if caller doesn't have ownership of the DAO instance.
func GetOptions(daoID uint64) *Options {
	if !isOwner(std.CurrentRealm().Address(), daoID) {
		panic("unauthorized")
	}

	return getOptions(daoID)
}

// IsOwner checks if addr is the owner of the specified DAO.
func IsOwner(addr std.Address, daoID uint64) bool {
	return isOwner(addr, daoID)
}

// TransferOwnership transfers ownership of a DAO to a different address.
func TransferOwnership(cur realm, daoID uint64, newOwner std.Address) {
	assertIsOwner(std.PreviousRealm().Address(), daoID)

	if !newOwner.IsValid() {
		panic("invalid owner address")
	}

	caller := std.PreviousRealm().Address()
	ids := getOwnership(caller)

	for i, id := range ids {
		if id == daoID {
			daoOwnerships.Set(caller.String(), append(ids[:i], ids[i+1:]...))
			daoOwnerships.Set(newOwner.String(), append(getOwnership(newOwner), id))

			break
		}
	}
}

// Get returns a common DAO searched by ID.
// It panics if caller doesn't have daoOwnerships of the DAO instance.
// Only toplevel DAOs are returned, to get SubDAOs use `GetSubDAO()`.
func Get(daoID uint64) *commondao.CommonDAO {
	assertIsOwner(std.CurrentRealm().Address(), daoID)
	return mustGetDAO(daoID)
}

// GetSize returns the number of SubDAOs under a specified root DAO,
// including the root DAO.
func GetSize(rootID uint64) int {
	assertIsOwner(std.CurrentRealm().Address(), rootID)

	tree := getTree(rootID)
	if tree == nil {
		return 0
	}

	return tree.Size()
}

// Vote submits a vote for a DAO proposal.
func Vote(cur realm, daoID, proposalID uint64, vote commondao.VoteChoice, reason string) {
	dao := mustGetDAO(daoID)
	o := getOptions(dao.ID())
	if !o.AllowVoting {
		panic(ufmt.Sprintf("voting in DAO `%s` (ID:%d) is not enabled", dao.Name(), dao.ID()))
	}

	caller := std.PreviousRealm().Address()
	assertCallerIsMember(caller, dao)

	err := dao.Vote(caller, proposalID, vote, reason)
	if err != nil {
		panic(err)
	}
}

// Execute executes a DAO proposal.
// This function is callable by any DAO member
func Execute(cur realm, daoID, proposalID uint64) {
	dao := mustGetDAO(daoID)
	o := getOptions(dao.ID())
	if !o.AllowExecution {
		panic(ufmt.Sprintf("proposal execution in DAO `%s` (ID:%d) is not enabled", dao.Name(), dao.ID()))
	}

	caller := std.PreviousRealm().Address()
	assertCallerIsMember(caller, dao)

	err := dao.Execute(proposalID)
	if err != nil {
		panic(err)
	}
}

// hasOwnership checks if addr is an owner of a DAO
func hasOwnership(addr std.Address) bool {
	_, found := daoOwnerships.Get(addr.String())
	return found
}

// isInvited checks if addr has been invited to use the CommonDAO realm
func isInvited(addr std.Address) bool {
	_, found := invites.Get(addr.String())
	return found
}

// isOwner checks if addr owns dao with id daoID
func isOwner(addr std.Address, daoID uint64) bool {
	// Make sure to check the root DAO ID, in case daoID belongs to a SubDAO.
	// This is required because ownership is assigned to the root DAO.

	// so the whole dao tree has one owner? why?
	daoID = mustGetDAO(daoID).TopParent().ID()

	for _, id := range getOwnership(addr) {
		if id == daoID {
			return true
		}
	}
	return false
}

func assertIsInvited(addr std.Address) {
	if !isInvited(addr) {
		panic("unauthorized")
	}
}

func assertIsOwner(addr std.Address, daoID uint64) {
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

// assertCallerIsMember asserts that the caller is part of the specified dao
func assertCallerIsMember(caller std.Address, dao *commondao.CommonDAO) {
	if !dao.Members().Has(caller) {
		panic("caller is not a DAO member")
	}
}
