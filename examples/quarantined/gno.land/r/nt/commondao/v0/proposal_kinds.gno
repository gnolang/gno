package commondao

import (
	"chain"
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// Proposal kind names: the per-DAO registry keys for every proposal type
// this realm can host. A DAO accepts proposals of exactly the kinds
// registered on it. The default kinds are registered at DAO creation and
// the kind set is managed afterwards through the manage-kinds kind
// (CreateRegisterKindProposal / CreateDeregisterKindProposal).
const (
	kindText                  = "text"
	kindCouncilUpdate         = "council-update"
	kindAncestorCouncilUpdate = "ancestor-council-update"
	kindSubDAO                = "subdao"
	kindDissolve              = "dissolve"
	kindTreasurySpend         = "treasury-spend"
	kindTreasuryClawback      = "treasury-clawback"
	kindTreasuryFreeze        = "treasury-freeze"
	kindManageKinds           = "manage-kinds"

	// Opt-in kind from the /p/ package. It is part of the catalog (so
	// catalogKind resolves it and manage-kinds can register it by name) but
	// is NOT default-seeded: a DAO gains it only after a supermajority
	// register. The name must match the package's kind name.
	kindExecution = "execution" // commondao.ExecutionKind
)

// errInvalidProposalArgs reports a proposal kind invoked with the wrong
// args type. Wrappers and kinds live in the same package, so this only
// fires on a realm bug, never on user input.
var errInvalidProposalArgs = errors.New("invalid proposal arguments")

// defaultProposalKinds lists the proposal kinds seeded on every new DAO
// (genesis, user-created or sub-DAO), in display order. Kinds are stateless
// singletons that wrap the definition constructors. This is the DAO's
// starting governance surface; the kind set is managed afterwards through
// the manage-kinds kind.
var defaultProposalKinds = []commondao.ProposalKind{
	textKind{},
	councilUpdateKind{},
	ancestorCouncilUpdateKind{},
	subDAOKind{},
	dissolveKind{},
	treasurySpendKind{},
	treasuryClawbackKind{},
	treasuryFreezeKind{},
	manageKindsKind{},
}

// optInProposalKinds lists the opt-in /p/ proposal kinds: they are part of
// the catalog but NOT default-seeded, so a DAO gains them only after a
// supermajority manage-kinds register.
var optInProposalKinds = []commondao.ProposalKind{
	commondao.ExecutionKind{},
}

// proposalKindCatalog lists every proposal kind this realm can host, in
// display order: the default-seeded kinds plus the opt-in /p/ kind
// (execution). catalogKind resolves against this list, so manage-kinds can
// register an opt-in kind by name that is not seeded. The executor set stays
// closed because only catalog kinds are ever registered on a DAO by name.
var proposalKindCatalog = append(
	append([]commondao.ProposalKind{}, defaultProposalKinds...),
	optInProposalKinds...,
)

// catalogKind returns a catalog kind by name, or nil when the name is not
// part of the catalog.
func catalogKind(name string) commondao.ProposalKind {
	for _, k := range proposalKindCatalog {
		if k.Name() == name {
			return k
		}
	}
	return nil
}

// IsProposalKindEnabled checks if a proposal kind is enabled (registered)
// on a DAO.
func IsProposalKindEnabled(daoID uint64, name string) bool {
	return mustGetDAO(daoID).HasKind(name)
}

// textArgs carries CreateTextProposal parameters to the text kind.
type textArgs struct {
	title        string
	body         string
	votingPeriod time.Duration
}

// textKind creates general text proposals.
type textKind struct{}

func (textKind) Name() string { return kindText }

func (textKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(textArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTextPropDefinition(a.title, a.body, a.votingPeriod), nil
}

// councilUpdateArgs carries CreateCouncilUpdateProposal parameters to the
// council-update kind. dao is the host DAO whose own council the executor
// mutates: New receives only a readonly view, so the trusted wrapper passes
// the mutable host handle through args (captured from its own mustGetDAO).
type councilUpdateArgs struct {
	dao    *commondao.CommonDAO
	add    []address
	remove []address
}

// councilUpdateKind creates proposals that add and/or remove members of
// the host DAO's own council.
type councilUpdateKind struct{}

func (councilUpdateKind) Name() string { return kindCouncilUpdate }

func (councilUpdateKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(councilUpdateArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newCouncilUpdatePropDefinition(a.dao, a.add, a.remove), nil
}

// ancestorCouncilUpdateArgs carries CreateAncestorCouncilUpdateProposal
// parameters to the ancestor-council-update kind. host is the proposing
// ancestor DAO (read for the ancestry check); target is the descendant
// whose council the executor mutates. Both handles come from the trusted
// wrapper via args, since New receives only a readonly view.
type ancestorCouncilUpdateArgs struct {
	host   *commondao.CommonDAO
	target *commondao.CommonDAO
	add    []address
	remove []address
}

// ancestorCouncilUpdateKind creates proposals for the host DAO, as an
// ancestor, to add and/or remove members of a descendant's council.
type ancestorCouncilUpdateKind struct{}

func (ancestorCouncilUpdateKind) Name() string { return kindAncestorCouncilUpdate }

func (ancestorCouncilUpdateKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(ancestorCouncilUpdateArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newAncestorCouncilUpdatePropDefinition(a.host, a.target, a.add, a.remove), nil
}

// subDAOArgs carries CreateSubDAOProposal parameters to the subdao kind.
// parent is the host DAO the new SubDAO is created under; the executor
// mutates it (wiring the child in), so the trusted wrapper passes the
// mutable host handle through args, since New receives only a readonly view.
type subDAOArgs struct {
	parent      *commondao.CommonDAO
	name        string
	purpose     string
	description string
	members     []address
}

// subDAOKind creates proposals that add a SubDAO under the host DAO.
type subDAOKind struct{}

func (subDAOKind) Name() string { return kindSubDAO }

func (subDAOKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(subDAOArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newSubDAOPropDefinition(a.parent, a.name, a.purpose, a.description, a.members), nil
}

// dissolveArgs carries CreateDissolutionProposal parameters to the
// dissolve kind, including the DAO being dissolved: the proposal is hosted
// in the nearest live ancestor (the host DAO that Propose passes to New),
// so the definition must operate on the dissolved descendant carried here,
// never on the host.
type dissolveArgs struct {
	dissolveDAO *commondao.CommonDAO
	destination address // sweep destination, root DAOs only
}

// dissolveKind creates proposals that dissolve a DAO or SubDAO.
type dissolveKind struct{}

func (dissolveKind) Name() string { return kindDissolve }

func (dissolveKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(dissolveArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newDissolvePropDefinition(a.dissolveDAO, a.destination), nil
}

// treasurySpendArgs carries CreateTreasurySpendProposal parameters to the
// treasury-spend kind. dao is the host DAO whose own treasury funds the
// spend (its sub is the funding source, see FundingDAOID); the trusted
// wrapper passes the mutable host handle through args, since New receives
// only a readonly view.
type treasurySpendArgs struct {
	dao  *commondao.CommonDAO
	to   address
	coin chain.Coin
}

// treasurySpendKind creates proposals that send coins from the host DAO's
// own treasury.
type treasurySpendKind struct{}

func (treasurySpendKind) Name() string { return kindTreasurySpend }

func (treasurySpendKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasurySpendArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasurySpendPropDefinition(a.dao, a.to, a.coin), nil
}

// treasuryClawbackArgs carries CreateTreasuryClawbackProposal parameters
// to the treasury-clawback kind. host is the proposing ancestor DAO (read
// for the ancestry check); target is the descendant whose treasury the
// executor sweeps (its sub is the funding source, see FundingDAOID). Both
// handles come from the trusted wrapper via args, since New receives only
// a readonly view.
type treasuryClawbackArgs struct {
	host   *commondao.CommonDAO
	target *commondao.CommonDAO
}

// treasuryClawbackKind creates proposals for the host DAO, as an ancestor,
// to sweep a descendant DAO's treasury to the descendant's parent.
type treasuryClawbackKind struct{}

func (treasuryClawbackKind) Name() string { return kindTreasuryClawback }

func (treasuryClawbackKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasuryClawbackArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasuryClawbackPropDefinition(a.host, a.target), nil
}

// treasuryFreezeArgs carries CreateTreasuryFreezeProposal parameters to
// the treasury-freeze kind. host is the proposing ancestor DAO (read for
// the ancestry / orphan-rescue check); target is the descendant whose
// treasury the executor freezes or unfreezes. Both handles come from the
// trusted wrapper via args, since New receives only a readonly view.
type treasuryFreezeArgs struct {
	host   *commondao.CommonDAO
	target *commondao.CommonDAO
	frozen bool
}

// treasuryFreezeKind creates proposals for the host DAO, as an ancestor,
// to freeze or unfreeze a descendant DAO's treasury.
type treasuryFreezeKind struct{}

func (treasuryFreezeKind) Name() string { return kindTreasuryFreeze }

func (treasuryFreezeKind) New(_ commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasuryFreezeArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasuryFreezePropDefinition(a.host, a.target, a.frozen), nil
}

// manageKindsArgs carries the manage-kinds wrapper parameters to the
// manage-kinds kind. dao is the host DAO whose registry the executor
// mutates: New receives only a readonly view, so the trusted wrapper passes
// the mutable host handle through args (captured from its own mustGetDAO).
//
// A proposal is one of three shapes, all populated only by a trusted
// wrapper:
//   - register by name: remove=false, name set, kind=nil (catalog lookup);
//   - register by value: remove=false, kind set (foreign, name ignored);
//   - deregister by name: remove=true, name set.
type manageKindsArgs struct {
	dao    *commondao.CommonDAO
	remove bool
	name   string
	kind   commondao.ProposalKind
}

// manageKindsKind creates governance proposals that register or deregister
// a proposal kind on the host DAO — the DAO's one permanent capability to
// manage which kinds it accepts. Registering adds a kind (by catalog name,
// or by value for a foreign kind supplied by a realm caller); deregistering
// removes one, which blocks new proposals of that kind while in-flight ones
// still vote and execute. It is decided by supermajority and cannot itself
// be deregistered (the self-brick guard below), so a DAO always keeps the
// ability to manage its kind set.
type manageKindsKind struct{}

func (manageKindsKind) Name() string { return kindManageKinds }

func (manageKindsKind) New(dao commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(manageKindsArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}

	if a.remove {
		// Deregister by name. Reject no-ops so a council vote is always
		// about an actual change, and reject the self-brick: manage-kinds
		// is the only un-deregisterable kind, so a DAO can never lose the
		// ability to manage its kind set.
		if !dao.HasKind(a.name) {
			return nil, errors.New("proposal kind is not enabled")
		}
		if a.name == kindManageKinds {
			return nil, errors.New("the manage-kinds kind cannot be deregistered")
		}
		return manageKindsPropDefinition(a), nil
	}

	// Register. A foreign kind supplied by a realm caller (by value) is not
	// catalog-resolved and carries its own name; otherwise the name must
	// resolve against the catalog. Either way reject a no-op (already
	// registered) so a council vote is always about an actual change.
	name := a.name
	if a.kind != nil {
		name = a.kind.Name()
		if name == "" {
			return nil, errors.New("proposal kind requires a name")
		}
	} else if catalogKind(a.name) == nil {
		return nil, errors.New("unknown proposal kind")
	}
	if dao.HasKind(name) {
		return nil, errors.New("proposal kind is already enabled")
	}

	return manageKindsPropDefinition(a), nil
}

// manageKindsPropDefinition defines a proposal type for registering or
// deregistering one of the realm's proposal kinds on a DAO. Deregistering
// only blocks new proposals: the kind registry is read at Propose time
// only, so in-flight proposals of a deregistered kind still vote and
// execute.
type manageKindsPropDefinition struct {
	dao    *commondao.CommonDAO
	remove bool
	name   string
	kind   commondao.ProposalKind
}

// displayName returns the kind name this proposal acts on: the supplied
// kind's own name for a register-by-value, else the catalog name.
func (p manageKindsPropDefinition) displayName() string {
	if p.kind != nil {
		return p.kind.Name()
	}
	return p.name
}

// Title returns the proposal title as raw text: the renderer escapes
// every definition title, so escaping the kind name here (unlike in Body,
// which the renderer trusts as markdown) would double-escape it.
func (p manageKindsPropDefinition) Title() string {
	if p.remove {
		return "Deregister Proposal Kind: " + p.displayName()
	}
	return "Register Proposal Kind: " + p.displayName()
}

func (manageKindsPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// isTrustedMarkdownBody marks Body as self-assembled markdown; the embedded
// kind name is escaped in Body itself (defense in depth).
func (manageKindsPropDefinition) isTrustedMarkdownBody() {}

// Threshold returns the tally threshold: changing which proposal kinds a
// DAO accepts alters its governance surface, so the supermajority default
// applies.
func (manageKindsPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

// Validate re-asserts the self-brick at Execute time (Validate reruns
// inside Execute): the manage-kinds kind can never be deregistered, so a
// DAO always keeps the ability to manage its kind set. This is a
// defense-in-depth second layer behind the New check, now that the /p/
// package no longer locks any name.
func (p manageKindsPropDefinition) Validate() error {
	if p.remove && p.name == kindManageKinds {
		return errors.New("the manage-kinds kind cannot be deregistered")
	}
	return nil
}

func (p manageKindsPropDefinition) Body() string {
	var b strings.Builder

	// The kind name is validated (against the catalog for a register by
	// name, or the supplied kind's own name); it is escaped anyway as
	// defense in depth.
	b.WriteString(md.Paragraph(md.Bold("Proposal Kind:") + "\n" + md.EscapeText(p.displayName())))

	action := "registered: new proposals of this kind can be created."
	if p.remove {
		action = "deregistered: no new proposal of this kind can be created. " +
			"In-flight proposals of the kind still vote and execute."
	}
	b.WriteString(md.Paragraph(md.Bold("Effect:") + "\nThe proposal kind is " + action))

	return b.String()
}

func (p manageKindsPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

// execute registers or deregisters the kind, returning any registry error
// unchanged so a race between two concurrently passed manage-kinds
// proposals fails the later one cleanly (StatusFailed) instead of panicking
// the transaction. catalogKind is a process-global immutable lookup, so a
// name valid at New still resolves here; the executor does not re-resolve
// or re-validate beyond the Validate self-brick. It moves no funds, so the
// definition is not Funded and ignores sub.
func (p manageKindsPropDefinition) execute(_ int, sub realm) error {
	if p.remove {
		return p.dao.DeregisterKind(p.name)
	}
	if p.kind != nil {
		return p.dao.RegisterKind(p.kind)
	}
	return p.dao.RegisterKind(catalogKind(p.name))
}
