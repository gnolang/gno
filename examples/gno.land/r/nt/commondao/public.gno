package commondao

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/commondao"
)

// Invite invites a user to the realm.
// A user invitation is required to start creating new DAOs.
func Invite(invitee std.Address) {
	if !invitee.IsValid() {
		panic("invalid address")
	}

	dao := mustGetDAO(CommonDAOID)
	caller := std.OriginCaller()
	if !dao.Members().Has(caller) {
		panic("unauthorized")
	}

	invites.Set(invitee.String(), caller.String())
}

// IsInvited check if an address has an invitation to the realm.
func IsInvited(addr std.Address) bool {
	return checkIsInvited(addr)
}

// New creates a new CommonDAO.
// An invitation is required to be able to create new DAOs.
func New(name string) *commondao.CommonDAO {
	name = strings.TrimSpace(name)
	assertNameIsNotEmpty(name)

	// Require an invitation when caller creates the first DAO
	caller := std.PreviousRealm().Address()
	if !checkOwnershipExists(caller) {
		orig := std.OriginCaller()
		assertIsInvited(orig)
		invites.Remove(orig.String())
	}

	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithName(name),
	)
	ids := append(getOwnership(caller), uint64(id))
	daoOptions := defaultOptions

	ownership.Set(caller.String(), ids)
	daos.Set(id.String(), dao)
	options.Set(id.String(), &daoOptions)

	return dao
}

// NewSubDAO creates a new SubDAO.
func NewSubDAO(name string, parentID uint64) *commondao.CommonDAO {
	caller := std.PreviousRealm().Address()
	assertIsOwner(caller, parentID)

	parentOptions := getOptions(parentID)
	if !parentOptions.AllowChildren {
		panic("SubDAO support is not enabled")
	}

	name = strings.TrimSpace(name)
	assertNameIsNotEmpty(name)

	parent := getDAO(parentID)
	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithParent(parent),
		commondao.WithName(name),
	)

	daoOptions := defaultOptions
	options.Set(id.String(), &daoOptions)

	daos.Set(id.String(), dao)

	rootID := parent.TopParent().ID()
	tree := getTree(rootID)
	if tree == nil {
		tree = avl.NewTree()
		trees.Set(makeIDKey(rootID), tree)
	}

	tree.Set(id.String(), dao)
	parent.Children().Append(dao)

	return dao
}

// GetOptions returns the configuration of a DAO.
// It panics if caller doesn't have ownership of the DAO instance.
func GetOptions(daoID uint64) *Options {
	if !checkIsOwner(std.PreviousRealm().Address(), daoID) {
		panic("unauthorized")
	}

	return getOptions(daoID)
}

// IsOwner checks if an address has ownership of a DAO.
func IsOwner(addr std.Address, daoID uint64) bool {
	return checkIsOwner(addr, daoID)
}

// TransferOwnership transfers DAO ownership to a different address.
func TransferOwnership(daoID uint64, newOwner std.Address) {
	assertCallerIsOwner(daoID)

	if !newOwner.IsValid() {
		panic("invalid owner address")
	}

	caller := std.PreviousRealm().Address()
	ids := getOwnership(caller)
	for i, id := range ids {
		if id == daoID {
			ownership.Set(caller.String(), append(ids[:i], ids[i+1:]...))
			ownership.Set(newOwner.String(), append(getOwnership(newOwner), id))

			break
		}
	}
}

// Get returns a common DAO searched by ID.
// It panics if caller doesn't have ownership of the DAO instance.
// Only toplevel DAOs are returned, to get SubDAOs use `GetSubDAO()`.
func Get(daoID uint64) *commondao.CommonDAO {
	assertCallerIsOwner(daoID)
	return mustGetDAO(daoID)
}

// GetSize returns the number of SubDAOs under a specified root DAO.
func GetSize(rootID uint64) int {
	assertCallerIsOwner(rootID)

	tree := getTree(rootID)
	if tree == nil {
		return 0
	}

	return tree.Size()
}

// Vote submits a vote for a DAO proposal.
func Vote(daoID, proposalID uint64, vote commondao.VoteChoice, reason string) {
	dao := mustGetDAO(daoID)

	o := getOptions(dao.ID())
	if !o.AllowVoting {
		panic("voting is not enabled")
	}

	caller := std.OriginCaller()
	assertCallerIsMember(caller, dao)

	err := dao.Vote(caller, proposalID, vote, reason)
	if err != nil {
		panic(err)
	}
}

// Execute executes a DAO proposal.
func Execute(daoID, proposalID uint64) {
	dao := mustGetDAO(daoID)

	o := getOptions(dao.ID())
	if !o.AllowExecution {
		panic("proposal execution is not enabled")
	}

	caller := std.OriginCaller()
	assertCallerIsMember(caller, dao)

	err := dao.Execute(proposalID)
	if err != nil {
		panic(err)
	}
}

func checkOwnershipExists(addr std.Address) bool {
	_, found := ownership.Get(addr.String())
	return found
}

func checkIsInvited(addr std.Address) bool {
	_, found := invites.Get(addr.String())
	return found
}

func checkIsOwner(addr std.Address, daoID uint64) bool {
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

func assertIsInvited(addr std.Address) {
	if !checkIsInvited(addr) {
		panic("unauthorized")
	}
}

func assertIsOwner(addr std.Address, daoID uint64) {
	if !checkIsOwner(addr, daoID) {
		panic("unauthorized")
	}
}

func assertCallerIsOwner(daoID uint64) {
	assertIsOwner(std.PreviousRealm().Address(), daoID)
}

func assertNameIsNotEmpty(name string) {
	if name == "" {
		panic("DAO name is empty")
	}
}

func assertCallerIsMember(caller std.Address, dao *commondao.CommonDAO) {
	if !dao.Members().Has(caller) {
		panic("caller is not a DAO member")
	}
}
