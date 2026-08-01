package commondao

import (
	"chain"
	"strings"
	"time"

	"gno.land/p/nt/bylaws/v0"
	"gno.land/p/nt/commondao/v0"
)

// CreateTextProposal creates a new general text proposal.
//
// Parameters:
// - daoID: ID of the DAO (required)
// - title: Title of the proposal (required)
// - body: Body of the proposal (required)
// - votingDays: The number of days where proposal accepts votes.
//
// The default voting period is 7 days.
func CreateTextProposal(cur realm, daoID uint64, title, body string, votingDays uint8) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
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

	return mustPropose(dao, caller, kindText, textArgs{title, body, votingPeriod})
}

// CreateCouncilUpdateProposal creates a new proposal to add and/or remove
// council members.
//
// Parameters:
// - daoID: ID of the DAO (required)
// - newMembers: Newline separated list of addresses to add to the council
// - removeMembers: Newline separated list of council addresses to remove
func CreateCouncilUpdateProposal(cur realm, daoID uint64, newMembers, removeMembers string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	args := councilUpdateArgs{dao, parseAddresses(newMembers), parseAddresses(removeMembers)}
	return mustPropose(dao, caller, kindCouncilUpdate, args)
}

// CreateAncestorCouncilUpdateProposal creates a proposal for an ancestor
// DAO to add and/or remove members of a descendant's council
// (docs/CONSTITUTION.md :1531-1532) — the rescue path for a stuck or
// empty descendant council. It is hosted and voted in the ancestor
// (daoID) and decided by supermajority; the proposing DAO must be a
// proper ancestor of the target, verified at proposal validation.
//
// Parameters:
// - daoID: ID of the proposing ancestor DAO (required)
// - targetID: ID of the descendant DAO whose council changes (required)
// - newMembers: Newline separated list of addresses to add to the council
// - removeMembers: Newline separated list of council addresses to remove
func CreateAncestorCouncilUpdateProposal(cur realm, daoID, targetID uint64, newMembers, removeMembers string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	target := mustGetDAO(targetID)
	args := ancestorCouncilUpdateArgs{dao, target, parseAddresses(newMembers), parseAddresses(removeMembers)}
	return mustPropose(dao, caller, kindAncestorCouncilUpdate, args)
}

// CreateSubDAOProposal creates a new proposal to create a new SubDAO.
//
// Parameters:
// - daoID: ID of the parent DAO (required)
// - name: A name for the SubDAO (required)
// - purpose: A purpose for the SubDAO (required)
// - description: A description for the SubDAO
// - members: Newline separated list of initial SubDAO council addresses (required)
func CreateSubDAOProposal(cur realm, daoID uint64, name, purpose, description, members string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	args := subDAOArgs{dao, name, purpose, description, parseAddresses(members)}
	return mustPropose(dao, caller, kindSubDAO, args)
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
	assertCurrent(cur)

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

	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	// The host (nearest live ancestor) gates and votes the proposal, while
	// the definition operates on the dissolved descendant carried in args.
	return mustPropose(dao, caller, kindDissolve, dissolveArgs{dissolveDAO, destination})
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
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	args := treasurySpendArgs{dao, to, chain.NewCoin(denom, amount)}
	return mustPropose(dao, caller, kindTreasurySpend, args)
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
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	target := mustGetDAO(targetID)
	return mustPropose(dao, caller, kindTreasuryClawback, treasuryClawbackArgs{dao, target})
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
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	target := mustGetDAO(targetID)
	return mustPropose(dao, caller, kindTreasuryFreeze, treasuryFreezeArgs{dao, target, frozen})
}

// CreateExecutionProposal creates a proposal that runs an arbitrary
// ExecFunc as the DAO's own sub on approval, through the realm's execution
// kind.
//
// The execution kind is opt-in: it is not seeded on new DAOs and must be
// registered first through a supermajority CreateRegisterKindProposal.
//
// Freeze policy: an execution proposal moves value under the DAO's own
// authority, so it is subject to the treasury freeze exactly like a spend.
// This wrapper fails fast when the treasury is already frozen, and the
// definition re-checks at Execute (so a freeze landing after the proposal
// passed fails it cleanly, StatusFailed, no funds leaving). An ancestor's
// clawback/dissolution is a separate power and is not blocked by freeze.
//
// Sharp edges (known limitations):
//   - The fn closure cannot be encoded in a CLI transaction, so this wrapper
//     is reachable only from a PERSISTENT realm that imports this one and is
//     a council member of the DAO (a realm-in-council). The closure must be
//     authored in that realm so it survives Propose→Execute; a `maketx run`
//     script's closure does not persist and cannot execute later.
//   - A closure that panics or runs out of gas aborts the whole Execute tx, so
//     the proposal is stuck Active (every retry re-aborts) and can never
//     finalize. The only recovery is dissolving the DAO (Dissolve dismisses
//     in-flight proposals). Author closures that return an error instead of
//     panicking so a bad execution fails cleanly (StatusFailed) and releases.
//
// Parameters:
// - daoID: ID of the DAO (required)
// - title: proposal title (raw text, escaped when rendered)
// - body: proposal body (raw text, escaped when rendered)
// - fn: the closure executed on approval (required, non-nil)
func CreateExecutionProposal(cur realm, daoID uint64, title, body string, fn commondao.ExecFunc) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	// Opt-in gate: refuse unless the DAO registered the execution kind.
	// Without this, arbitrary code execution would ride on a name a DAO
	// never opted into. Propose also rejects an unregistered kind, but this
	// fails fast with a clear message before any definition is built.
	assertKindRegistered(dao, kindExecution)

	// Freeze gate (intentional defense in depth): a frozen DAO cannot initiate
	// any treasury movement, so refuse up front here even though the
	// definition's Validate re-checks the same flag at create and at execute.
	// Redundant on purpose — two independent layers on the "no funds leave a
	// frozen DAO" invariant. Without it an execution proposal could drain a
	// frozen DAO's own treasury, defeating an ancestor's freeze.
	if dao.IsTreasuryFrozen() {
		panic(errTreasuryFrozen)
	}

	return mustPropose(dao, caller, kindExecution, executionArgs{title: title, body: body, fn: fn})
}

// CreateRegisterKindProposal creates a proposal to register one of the
// realm's catalog proposal kinds on a DAO by name, through the permanent
// manage-kinds kind (e.g. register "execution").
//
// The proposal is hosted and voted in the DAO itself and decided by
// supermajority; on approval the named catalog kind is registered, so new
// proposals of that kind can be created. The manage-kinds kind is seeded
// on every DAO, so this path is always available.
//
// Parameters:
// - daoID: ID of the DAO (required)
// - kindName: name of the catalog proposal kind to register (required)
func CreateRegisterKindProposal(cur realm, daoID uint64, kindName string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	return mustPropose(dao, caller, kindManageKinds, manageKindsProposal{dao: dao, name: kindName})
}

// CreateDeregisterKindProposal creates a proposal to deregister a proposal
// kind from a DAO by name, through the permanent manage-kinds kind.
//
// The proposal is hosted and voted in the DAO itself and decided by
// supermajority; on approval the kind is deregistered, which blocks new
// proposals of that kind while in-flight ones still vote and execute. The
// manage-kinds kind itself cannot be deregistered, so a DAO always keeps
// the ability to manage its kind set (and to re-register a catalog kind by
// name).
//
// Parameters:
// - daoID: ID of the DAO (required)
// - kindName: name of the proposal kind to deregister (required)
func CreateDeregisterKindProposal(cur realm, daoID uint64, kindName string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	return mustPropose(dao, caller, kindManageKinds, manageKindsProposal{dao: dao, remove: true, name: kindName})
}

// CreateAmendBylawsProposal creates a proposal to add, amend or remove one
// of the DAO's bylaws/mandates documents with a verifiable diff patch (see
// gno.land/p/nt/bylaws/v0). The payload is an encoded patch — build it with
// AmendBylawsPayload (e.g. through a vm/qeval query) or with
// bylaws.Diff(...).Encode() from a realm. The patch pins the sha256 of the
// document text it was diffed against, so an amendment racing a concurrent
// change to the same document fails cleanly instead of clobbering it; a
// patch already stale at creation is rejected here. Amendments are decided
// by simple majority (the Constitution's bar for charter, bylaws and
// mandate changes).
//
// Parameters:
// - daoID: ID of the DAO (required)
// - payload: encoded bylaws patch (required)
func CreateAmendBylawsProposal(cur realm, daoID uint64, payload string) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	patch, err := bylaws.DecodePatch(payload)
	if err != nil {
		panic(err)
	}

	return mustPropose(dao, caller, kindAmendBylaws, amendBylawsProposal{
		daoID: daoID,
		set:   bylawsOf(daoID),
		patch: patch,
	})
}

// AmendBylawsPayload builds the CreateAmendBylawsProposal payload that
// changes a DAO's document at path to the proposed text: a new path adds a
// document, empty proposed text removes one. It diffs against the
// document's current text and pins its hash, so build the payload fresh
// (e.g. through a vm/qeval query) and propose promptly — a payload built
// against superseded text is rejected. Read-only.
func AmendBylawsPayload(daoID uint64, path, proposed string) string {
	mustGetDAO(daoID)

	var (
		cur    string
		exists bool
	)
	if set := bylawsView(daoID); set != nil {
		cur, exists = set.Get(path)
	}

	p, err := bylaws.DiffTexts(path, cur, proposed, exists)
	if err != nil {
		panic(err)
	}
	return p.Encode()
}

// assertKindRegistered panics unless a proposal kind is registered on the
// DAO. Used to gate the opt-in propose paths so they work only after the DAO
// registered the kind through governance.
func assertKindRegistered(dao *commondao.CommonDAO, name string) {
	if !dao.HasKind(name) {
		panic("proposal kind is not registered: " + name)
	}
}

// mustPropose submits a proposal through one of the DAO's registered
// proposal kinds and validates it for the current state, panicking on any
// error. Validation also reruns inside Execute, so this only fails fast at
// creation.
func mustPropose(dao *commondao.CommonDAO, caller address, kind string, args any) uint64 {
	p, err := dao.Propose(caller, kind, args)
	if err != nil {
		panic(err)
	}

	if err = p.Validate(); err != nil {
		panic(err)
	}

	return p.ID()
}

// parseAddresses parses a newline separated list of addresses,
// deduplicated, panicking on invalid entries.
func parseAddresses(s string) []address {
	var addrs []address
	for _, raw := range strings.Split(s, "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		addr := address(raw)
		if !addr.IsValid() {
			panic("invalid address: " + addr.String())
		}

		if !containsAddress(addrs, addr) {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

// containsAddress checks if an address is present in a list.
func containsAddress(addrs []address, addr address) bool {
	for _, a := range addrs {
		if a == addr {
			return true
		}
	}
	return false
}
