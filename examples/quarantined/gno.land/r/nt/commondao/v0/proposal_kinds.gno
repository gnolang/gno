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
// registered on it. The full catalog is registered at DAO creation and
// kinds are toggled afterwards through CreateSetProposalKindProposal.
const (
	kindText                  = "text"
	kindCouncilUpdate         = "council-update"
	kindAncestorCouncilUpdate = "ancestor-council-update"
	kindSubDAO                = "subdao"
	kindDissolve              = "dissolve"
	kindTreasurySpend         = "treasury-spend"
	kindTreasuryClawback      = "treasury-clawback"
	kindTreasuryFreeze        = "treasury-freeze"
	kindProposalKinds         = "proposal-kinds"
)

// errInvalidProposalArgs reports a proposal kind invoked with the wrong
// args type. Wrappers and kinds live in the same package, so this only
// fires on a realm bug, never on user input.
var errInvalidProposalArgs = errors.New("invalid proposal arguments")

// proposalKindCatalog lists every proposal kind this realm can host, in
// display order. Kinds are stateless singletons that wrap the definition
// constructors; the executor set stays closed because only catalog kinds
// are ever registered on a DAO.
var proposalKindCatalog = []commondao.ProposalKind{
	textKind{},
	councilUpdateKind{},
	ancestorCouncilUpdateKind{},
	subDAOKind{},
	dissolveKind{},
	treasurySpendKind{},
	treasuryClawbackKind{},
	treasuryFreezeKind{},
	proposalKindsKind{},
}

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

// CreateSetProposalKindProposal creates a proposal to enable or disable
// one of the realm's proposal kinds on a DAO.
//
// The proposal is hosted and voted in the DAO itself and decided by
// supermajority. Enabling registers the catalog kind on the DAO and
// disabling deregisters it, which blocks new proposals of that kind while
// in-flight ones still vote and execute. The proposal-kinds kind itself
// can never be disabled, so a DAO always keeps the ability to re-enable
// kinds.
//
// Parameters:
// - daoID: ID of the DAO (required)
// - kindName: name of the proposal kind to enable or disable (required)
// - enabled: true to enable the kind, false to disable it
func CreateSetProposalKindProposal(cur realm, daoID uint64, kindName string, enabled bool) uint64 {
	assertCurrent(cur)

	dao := mustGetDAO(daoID)
	caller := cur.Previous().Address()
	assertCallerIsCouncilMember(caller, dao)

	return mustPropose(dao, caller, kindProposalKinds, setProposalKindArgs{kindName, enabled})
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

func (textKind) New(_ *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(textArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTextPropDefinition(a.title, a.body, a.votingPeriod), nil
}

// councilUpdateArgs carries CreateCouncilUpdateProposal parameters to the
// council-update kind.
type councilUpdateArgs struct {
	add    []address
	remove []address
}

// councilUpdateKind creates proposals that add and/or remove members of
// the host DAO's own council.
type councilUpdateKind struct{}

func (councilUpdateKind) Name() string { return kindCouncilUpdate }

func (councilUpdateKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(councilUpdateArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newCouncilUpdatePropDefinition(dao, a.add, a.remove), nil
}

// ancestorCouncilUpdateArgs carries CreateAncestorCouncilUpdateProposal
// parameters to the ancestor-council-update kind.
type ancestorCouncilUpdateArgs struct {
	target *commondao.CommonDAO
	add    []address
	remove []address
}

// ancestorCouncilUpdateKind creates proposals for the host DAO, as an
// ancestor, to add and/or remove members of a descendant's council.
type ancestorCouncilUpdateKind struct{}

func (ancestorCouncilUpdateKind) Name() string { return kindAncestorCouncilUpdate }

func (ancestorCouncilUpdateKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(ancestorCouncilUpdateArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newAncestorCouncilUpdatePropDefinition(dao, a.target, a.add, a.remove), nil
}

// subDAOArgs carries CreateSubDAOProposal parameters to the subdao kind.
type subDAOArgs struct {
	name        string
	purpose     string
	description string
	members     []address
}

// subDAOKind creates proposals that add a SubDAO under the host DAO.
type subDAOKind struct{}

func (subDAOKind) Name() string { return kindSubDAO }

func (subDAOKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(subDAOArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newSubDAOPropDefinition(dao, a.name, a.purpose, a.description, a.members), nil
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

func (dissolveKind) New(_ *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(dissolveArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newDissolvePropDefinition(a.dissolveDAO, a.destination), nil
}

// treasurySpendArgs carries CreateTreasurySpendProposal parameters to the
// treasury-spend kind.
type treasurySpendArgs struct {
	to   address
	coin chain.Coin
}

// treasurySpendKind creates proposals that send coins from the host DAO's
// own treasury.
type treasurySpendKind struct{}

func (treasurySpendKind) Name() string { return kindTreasurySpend }

func (treasurySpendKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasurySpendArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasurySpendPropDefinition(dao, a.to, a.coin), nil
}

// treasuryClawbackArgs carries CreateTreasuryClawbackProposal parameters
// to the treasury-clawback kind.
type treasuryClawbackArgs struct {
	target *commondao.CommonDAO
}

// treasuryClawbackKind creates proposals for the host DAO, as an ancestor,
// to sweep a descendant DAO's treasury to the descendant's parent.
type treasuryClawbackKind struct{}

func (treasuryClawbackKind) Name() string { return kindTreasuryClawback }

func (treasuryClawbackKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasuryClawbackArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasuryClawbackPropDefinition(dao, a.target), nil
}

// treasuryFreezeArgs carries CreateTreasuryFreezeProposal parameters to
// the treasury-freeze kind.
type treasuryFreezeArgs struct {
	target *commondao.CommonDAO
	frozen bool
}

// treasuryFreezeKind creates proposals for the host DAO, as an ancestor,
// to freeze or unfreeze a descendant DAO's treasury.
type treasuryFreezeKind struct{}

func (treasuryFreezeKind) Name() string { return kindTreasuryFreeze }

func (treasuryFreezeKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(treasuryFreezeArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	return newTreasuryFreezePropDefinition(dao, a.target, a.frozen), nil
}

// setProposalKindArgs carries CreateSetProposalKindProposal parameters to
// the proposal-kinds kind.
type setProposalKindArgs struct {
	kindName string
	enabled  bool
}

// proposalKindsKind creates governance proposals that enable or disable
// one of the realm's proposal kinds on the host DAO.
type proposalKindsKind struct{}

func (proposalKindsKind) Name() string { return kindProposalKinds }

func (proposalKindsKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(setProposalKindArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}

	if catalogKind(a.kindName) == nil {
		return nil, errors.New("unknown proposal kind")
	}

	// Self-brick guard: a DAO must always keep the ability to re-enable
	// kinds, so the governance kind itself can never be disabled.
	if !a.enabled && a.kindName == kindProposalKinds {
		return nil, errors.New("the proposal-kinds kind cannot be disabled")
	}

	// Reject no-ops so a council vote is always about an actual change.
	if a.enabled && dao.HasKind(a.kindName) {
		return nil, errors.New("proposal kind is already enabled")
	}
	if !a.enabled && !dao.HasKind(a.kindName) {
		return nil, errors.New("proposal kind is not enabled")
	}

	return setProposalKindPropDefinition{
		dao:      dao,
		kindName: a.kindName,
		enabled:  a.enabled,
	}, nil
}

// setProposalKindPropDefinition defines a proposal type for enabling or
// disabling one of the realm's proposal kinds on a DAO. Disabling only
// blocks new proposals: the kind registry is read at Propose time only, so
// in-flight proposals of a disabled kind still vote and execute.
type setProposalKindPropDefinition struct {
	dao      *commondao.CommonDAO
	kindName string
	enabled  bool
}

func (p setProposalKindPropDefinition) Title() string {
	if p.enabled {
		return "Enable Proposal Kind: " + p.kindName
	}
	return "Disable Proposal Kind: " + p.kindName
}

func (setProposalKindPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// Threshold returns the tally threshold: changing which proposal kinds a
// DAO accepts alters its governance surface, so the supermajority default
// applies.
func (setProposalKindPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (p setProposalKindPropDefinition) Body() string {
	var b strings.Builder

	// The kind name is validated against the catalog at proposal creation;
	// it is escaped anyway as defense in depth.
	b.WriteString(md.Paragraph(md.Bold("Proposal Kind:") + "\n" + md.EscapeText(p.kindName)))

	action := "enabled: new proposals of this kind can be created."
	if !p.enabled {
		action = "disabled: no new proposal of this kind can be created. " +
			"In-flight proposals of the kind still vote and execute."
	}
	b.WriteString(md.Paragraph(md.Bold("Effect:") + "\nThe proposal kind is " + action))

	return b.String()
}

func (p setProposalKindPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

// execute toggles the kind, returning any registry error so a race between
// two concurrently passed toggles fails the later proposal cleanly
// (StatusFailed) instead of panicking the transaction. It moves no funds,
// so the definition is not Funded and ignores sub.
func (p setProposalKindPropDefinition) execute(_ int, sub realm) error {
	if p.enabled {
		return p.dao.RegisterKind(catalogKind(p.kindName))
	}
	return p.dao.DeregisterKind(p.kindName)
}
