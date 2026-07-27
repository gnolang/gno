package commondao

import (
	"chain"
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
		panic("text proposals are not enabled")
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
		panic("council update proposals are not enabled")
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
		panic("SubDAO proposals are not enabled")
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
// SubDAOs can only be dissolved by the parent DAO, which owns and
// controls its sub-DAOs (docs/CONSTITUTION.md :1507). When the parent is
// itself already dissolved, the proposal is hosted in the nearest
// non-dissolved ancestor, so orphans below a dissolved middle DAO remain
// dissolvable.
//
// Dissolution sweeps any remaining treasury balance. A sub-DAO's sweep
// goes to its parent and destination must be empty; a root DAO has no
// parent, so a valid destination address is required.
//
// Parameters:
// - daoID: ID of the DAO to dissolve (required)
// - destination: sweep destination, root DAOs only
func CreateDissolutionProposal(cur realm, daoID uint64, destination address) uint64 {
	// When DAO to dissolve is a SubDAO make sure that proposal is created
	// in the parent DAO, or in the nearest non-dissolved ancestor when
	// parents were dissolved first.
	dao := mustGetDAO(daoID)
	dissolveDAO := dao
	if parent := dao.Parent(); parent != nil {
		for parent.IsDeleted() && parent.Parent() != nil {
			parent = parent.Parent()
		}
		dao = parent
	}

	o := getOptions(dao.ID())
	if !o.AllowDissolutionProposals {
		panic("dissolution proposals are not enabled")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	def := newDissolvePropDefinition(dissolveDAO, destination)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateTreasurySpendProposal creates a new proposal to send coins from
// the DAO's own treasury (docs/CONSTITUTION.md :1542-1543).
//
// Parameters:
// - daoID: ID of the DAO whose treasury is spent (required)
// - to: recipient address (required)
// - denom: coin denomination, e.g. "ugnot" (required)
// - amount: coin amount, must be positive (required)
func CreateTreasurySpendProposal(cur realm, daoID uint64, to address, denom string, amount int64) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowTreasuryProposals {
		panic("treasury proposals are not enabled")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	def := newTreasurySpendPropDefinition(dao, to, chain.NewCoin(denom, amount))
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateTreasuryClawbackProposal creates a new proposal for an ancestor
// DAO to sweep a descendant DAO's full treasury balance to the target's
// parent (docs/CONSTITUTION.md :1507). The proposing DAO must be a
// proper ancestor of the target; a target's own options can never block
// an ancestor's clawback.
//
// Parameters:
// - daoID: ID of the proposing ancestor DAO (required)
// - targetID: ID of the descendant DAO to claw back (required)
func CreateTreasuryClawbackProposal(cur realm, daoID, targetID uint64) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowTreasuryProposals {
		panic("treasury proposals are not enabled")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	target := mustGetDAO(targetID)
	def := newTreasuryClawbackPropDefinition(dao, target)
	p, err := dao.Propose(caller, def)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// CreateTreasuryFreezeProposal creates a new proposal for an ancestor DAO
// to freeze or unfreeze a descendant DAO's treasury. While frozen, no
// treasury spend can execute. Only a proper ancestor can unfreeze — the
// frozen DAO's own council cannot.
//
// Parameters:
// - daoID: ID of the proposing ancestor DAO (required)
// - targetID: ID of the descendant DAO to freeze or unfreeze (required)
// - frozen: true to freeze the target's treasury, false to unfreeze
func CreateTreasuryFreezeProposal(cur realm, daoID, targetID uint64, frozen bool) uint64 {
	dao := mustGetDAO(daoID)
	o := getOptions(daoID)
	if !o.AllowTreasuryProposals {
		panic("treasury proposals are not enabled")
	}

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	target := mustGetDAO(targetID)
	def := newTreasuryFreezePropDefinition(dao, target, frozen)
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
