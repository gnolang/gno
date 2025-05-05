package commondao

import (
	"std"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
)

// CreateTextProposal creates a new general text proposal.
//
// Parameters:
// - daoID: ID of the root DAO (required)
// - title: Title of the proposal (required)
// - body: Body of the proposal (required)
// - votingDays: The number of days where proposal accepts votes.
//
// The default voting period is 7 days.
func CreateTextProposal(daoID uint64, title, body string, votingDays uint8) uint64 {
	crossing()

	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowTextProposals {
		panic("forbidden")
	}

	if votingDays > 30 {
		panic("maximum proposal voting period is 30 days")
	}

	caller := std.PreviousRealm().Address()
	assertCallerIsMember(caller, dao)

	var votingPeriod time.Duration
	if votingDays == 0 {
		votingPeriod = time.Hour * 24 * 7
	} else {
		votingPeriod = time.Hour * 24 * time.Duration(votingDays)
	}

	def := NewTextPropDefinition(title, body, commondao.QuorumOneThird, votingPeriod)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateMembersUpdateProposal creates a new proposal to add and/or remove DAO members.
//
// Parameters:
// - daoID: ID of the root DAO (required)
// - newMembers: Comma separated list of addresses to add as members
// - removeMembers: Comma separated list of member addresses to remove
func CreateMembersUpdateProposal(daoID uint64, newMembers, removeMembers string) uint64 {
	crossing()

	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowMembersUpdate {
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
