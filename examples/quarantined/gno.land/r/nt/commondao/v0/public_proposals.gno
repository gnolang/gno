package commondao

import (
	"strings"
	"time"

	"gno.land/p/moul/addrset"
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
func CreateTextProposal(cur realm, daoID uint64, title, body string, votingDays uint8) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowTextProposals {
		panic("forbidden")
	}

	if votingDays > 30 {
		panic("maximum proposal voting period is 30 days")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	var votingPeriod time.Duration
	if votingDays == 0 {
		votingPeriod = time.Hour * 24 * 7
	} else {
		votingPeriod = time.Hour * 24 * time.Duration(votingDays)
	}

	def := NewTextPropDefinition(title, body, votingPeriod)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateCouncilUpdateProposal creates a new proposal to add and/or remove
// council members.
//
// Parameters:
// - daoID: ID of the root DAO (required)
// - newMembers: Newline separated list of addresses to add to the council
// - removeMembers: Newline separated list of council addresses to remove
func CreateCouncilUpdateProposal(cur realm, daoID uint64, newMembers, removeMembers string) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowCouncilUpdate {
		panic("forbidden")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	toAdd := mustParseStringToAddrset(newMembers)
	toRemove := mustParseStringToAddrset(removeMembers)
	def := newCouncilUpdatePropDefinition(dao, toAdd, toRemove)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateSubDAOProposal creates a new proposal to create a new SubDAO.
//
// Parameters:
// - daoID: ID of the parent DAO (required)
// - name: A name for the SubDAO (required)
// - members: Newline separated list of initial SubDAO council addresses (required)
func CreateSubDAOProposal(cur realm, daoID uint64, name, members string) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !(o.AllowSubDAOProposals && o.AllowChildren) {
		panic("forbidden")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	daoMembers := mustParseStringToAddrset(members)
	def := newSubDAOPropDefinition(dao, name, &daoMembers)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateDissolutionProposal creates a new proposal to dissolve a DAO or SubDAO.
//
// SubDAOs can only be dissolved by the the parent DAO.
//
// Parameters:
// - daoID: ID of the DAO to dissolve (required)
func CreateDissolutionProposal(cur realm, daoID uint64) uint64 {
	// When DAO to dissolve is a SubDAO make sure that proposal is created in the parent DAO
	dao := mustGetDAO(daoID)
	dissolveDAO := dao
	if parent := dao.Parent(); parent != nil {
		dao = parent
	}

	o := getOptions(dao.ID())
	if !o.AllowDissolutionProposals {
		panic("forbidden")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	def := newDissolvePropDefinition(dissolveDAO)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

func mustParseStringToAddrset(s string) (set addrset.Set) {
	if s == "" {
		return
	}

	for _, raw := range strings.Split(s, "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		addr := address(raw)
		if !addr.IsValid() {
			panic("invalid address: " + addr.String())
		}

		if !set.Has(addr) {
			set.Add(addr)
		}
	}
	return set
}
