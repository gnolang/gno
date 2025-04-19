package commondao

import (
	"std"
	"strings"

	"gno.land/p/moul/addrset"
)

// CreateMembersUpdateProposal creates a new proposal to add and/or remove DAO members.
//
// Parameters:
// - daoID: ID of the root DAO (required)
// - newMembers: Comma separated list of addresses to add as members
// - removeMembers: Comma separated list of member addresses to remove
func CreateMembersUpdateProposal(daoID uint64, newMembers, removeMembers string) uint64 {
	crossing()

	dao := mustGetDAO(daoID)
	options := getOptions(daoID)
	if !options.AllowMembersUpdate {
		panic("forbidden")
	}

	caller := std.PreviousRealm().Address()
	assertCallerIsMember(caller, dao)

	toAdd := mustParseStringToAddrset(newMembers)
	toRemove := mustParseStringToAddrset(removeMembers)
	def := NewMembersPropDefinition(dao, toAdd, toRemove)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	return p.ID()
}

func mustParseStringToAddrset(s string) (set addrset.Set) {
	for _, raw := range strings.Split(s, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		addr := std.Address(raw)
		if !addr.IsValid() {
			panic("invalid address: " + addr.String())
		}

		if !set.Has(addr) {
			set.Add(addr)
		}
	}
	return set
}
