package commondao

import (
	"std"
	"strings"

	"gno.land/p/nt/commondao"
)

// TODO: Consider using an interface instead of a ref to CommonDAO type
// TODO: Add support for creating hierarchical DAOs (require slug value)

func Invite(invitee std.Address) {
	// TODO: Use an authority to restrict inviters when `p/moul/authz` is available
	caller := std.PrevRealm().Addr()
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

	caller := std.PrevRealm().Addr()
	if !checkOwnershipExists(caller) {
		orig := std.GetOrigCaller()
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

// IsOwner checks if an address has ownership of a DAO.
func IsOwner(addr std.Address, daoID uint64) bool {
	return checkIsOwner(addr, daoID)
}

// TransferOwnership transfers DAO access ownership to a different address.
func TransferOwnership(daoID uint64, newOwner std.Address) {
	caller := std.PrevRealm().Addr()
	assertIsOwner(caller, daoID)

	ids := getOwnership(caller)
	for i, id := range ids {
		if id == daoID {
			ownership.Set(caller.String(), append(ids[:i], ids[i+1:]...))
			ownership.Set(newOwner.String(), append(getOwnership(newOwner), id))

			break
		}
	}
}

// GetByID returns a DAO searched by its ID.
// It panics if caller doesn't have access ownership to it.
func GetByID(daoID uint64) *commondao.CommonDAO {
	assertCallerIsRealm() // TODO: Allow users to get DAOs using run TX?

	caller := std.PrevRealm().Addr()
	assertIsOwner(caller, daoID)

	return mustGetDAO(daoID)
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

func assertCallerIsRealm() {
	if std.PrevRealm().PkgPath() == "" {
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
