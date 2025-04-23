package commondao

import (
	"std"
	"strings"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
)

// CreateMembersUpdateProposal creates a new proposal to add and/or remove DAO members.
//
// Proposals can be created for standalone DAOs or SubDAOs when DAO is hierarchical.
// The `subPath` parameter is required to create proposals for a SubDAOs.
//
// Parameters:
// - daoID: ID of the root DAO (required)
// - subPath: Optional path to a SubDAO
// - newMembers: Comma separated list of addresses to add as members
// - removeMembers: Comma separated list of member addresses to remove
func CreateMembersUpdateProposal(daoID uint64, subPath, newMembers, removeMembers string) uint64 {
	crossing()

	var dao *commondao.CommonDAO
	if subPath != "" {
		dao = mustGetSubDAO(daoID, subPath)
	} else {
		dao = mustGetDAO(daoID)
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
