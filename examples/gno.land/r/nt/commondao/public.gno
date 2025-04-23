package commondao

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/commondao"
)

// TODO: Allow different ownerships of DAOs within the same tree

var reSlug = regexp.MustCompile("^[a-zA-Z]+[a-zA-Z0-9-]*$")

// Invite invites a user to the realm.
// A user invitation is required to start creating new DAOs.
func Invite(invitee std.Address) {
	crossing()

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
	return checkIsInvited(addr)
}

// New creates a new CommonDAO.
func New(name string) *commondao.CommonDAO {
	assertCallerIsRealm() // TODO: Allow users to create DAOs using run TX?

	name = strings.TrimSpace(name)
	assertNameIsNotEmpty(name)

	caller := std.CurrentRealm().Address()
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

	ownership.Set(caller.String(), ids)
	daos.Set(id.String(), dao)
	return dao
}

// NewSubDAO creates a new SubDAO.
func NewSubDAO(name, slug string, parent *commondao.CommonDAO) *commondao.CommonDAO {
	assertCallerIsRealm() // TODO: Allow users to create SubDAOs using run TX?

	if parent == nil {
		panic("parent is required")
	}

	name = strings.TrimSpace(name)
	assertNameIsNotEmpty(name)
	assertSlugIsValid(slug)

	dao := commondao.New(
		commondao.WithParent(parent),
		commondao.WithName(name),
		commondao.WithSlug(slug),
	)

	parentID := parent.TopParent().ID()
	tree := getTree(parentID)
	if tree == nil {
		tree = avl.NewTree()
		trees.Set(makeIDKey(parentID), tree)
	}

	path := dao.Path()
	if tree.Has(path) {
		panic("subDAO slug is already taken")
	}

	tree.Set(path, dao)
	parent.Children().Append(dao)
	return dao
}

// IsOwner checks if an address has ownership of a DAO.
func IsOwner(addr std.Address, daoID uint64) bool {
	return checkIsOwner(addr, daoID)
}

// TransferOwnership transfers DAO access ownership to a different address.
func TransferOwnership(daoID uint64, newOwner std.Address) {
	crossing()

	assertIsOwner(std.PreviousRealm().Address(), daoID)

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
// It panics if caller doesn't have access ownership to it.
// Only toplevel DAOs are returned, to get SubDAOs use `GetSubDAO()`.
func Get(daoID uint64) *commondao.CommonDAO {
	assertCallerIsRealm() // TODO: Allow users to get DAOs using run TX?
	assertIsOwner(std.CurrentRealm().Address(), daoID)

	return mustGetDAO(daoID)
}

// GetSubDAO returns a SubDAO searched by path.
// It panics if caller doesn't have access ownership to the topmost DAO.
func GetSubDAO(rootID uint64, path string) *commondao.CommonDAO {
	assertCallerIsRealm() // TODO: Allow users to get SubDAOs using run TX?
	assertIsOwner(std.CurrentRealm().Address(), rootID)

	return mustGetSubDAO(rootID, path)
}

// GetSize returns the number of SubDAOs under a specified root DAO.
func GetSize(rootID uint64) int {
	assertIsOwner(std.CurrentRealm().Address(), rootID)

	tree := getTree(rootID)
	if tree == nil {
		return 0
	}
	return tree.Size()
}

// Vote submits a vote for a DAO proposal.
func Vote(daoID, proposalID uint64, vote commondao.VoteChoice, reason string) {
	crossing()

	dao := mustGetDAO(daoID)
	caller := std.PreviousRealm().Address()
	assertCallerIsMember(caller, dao)

	err := dao.Vote(caller, proposalID, vote, reason)
	if err != nil {
		panic(err)
	}
}

// Execute executes a DAO proposal.
func Execute(daoID, proposalID uint64) {
	crossing()

	dao := mustGetDAO(daoID)
	caller := std.PreviousRealm().Address()
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

func assertCallerIsRealm() { // Assert for non crossing functions
	if std.CurrentRealm().PkgPath() == "" {
		panic("caller is not a realm")
	}
}

func assertIsOwner(addr std.Address, daoID uint64) {
	if !checkIsOwner(addr, daoID) {
		panic("unauthorized")
	}
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

func assertSlugIsValid(slug string) {
	if !reSlug.MatchString(slug) {
		panic("invalid DAO slug")
	}
}
