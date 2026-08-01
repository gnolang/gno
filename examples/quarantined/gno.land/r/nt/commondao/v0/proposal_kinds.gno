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
	kindAmendBylaws           = "amend-bylaws"

	// Opt-in kind. It is part of the catalog (so catalogKind resolves it and
	// manage-kinds can register it by name) but is NOT default-seeded: a DAO
	// gains it only after a supermajority register.
	//
	// It is a realm-side kind (executionKind), not the bare /p/
	// commondao.ExecutionKind: the realm wraps the arbitrary-exec closure with
	// its own Validable freeze policy so a frozen DAO cannot drain its own
	// treasury through an execution proposal (the /p/ kind carries no such
	// policy). The name still matches the /p/ kind's name.
	kindExecution = "execution"
)

// errInvalidProposalArgs reports a proposal kind invoked with the wrong
// args type. Wrappers and kinds live in the same package, so this only
// fires on a realm bug, never on user input.
var errInvalidProposalArgs = errors.New("invalid proposal arguments")

// defaultProposalKinds lists the proposal kinds seeded on every new DAO
// (genesis, user-created or sub-DAO). Kinds are stateless singletons that
// wrap the definition constructors. This is the DAO's starting governance
// surface; the kind set is managed afterwards through the manage-kinds kind.
// Ordering here is immaterial — render.gno owns the presentation order.
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
	amendBylawsKind{},
}

// optInProposalKinds lists the opt-in proposal kinds: they are part of the
// catalog but NOT default-seeded, so a DAO gains them only after a
// supermajority manage-kinds register. The execution kind is realm-side
// (executionKind), not the bare /p/ ExecutionKind, so it can carry the
// realm's freeze policy (see kindExecution).
var optInProposalKinds = []commondao.ProposalKind{
	executionKind{},
}

// proposalKindCatalog lists every proposal kind this realm can host: the
// default-seeded kinds plus the opt-in kind (execution). catalogKind
// resolves against this list, so manage-kinds can register an opt-in kind by
// name that is not seeded. The executor set stays closed because only catalog
// kinds are ever registered on a DAO by name. Ordering here is immaterial —
// render.gno owns the presentation order.
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

// HasProposalKind reports whether a proposal kind is registered on a DAO.
func HasProposalKind(daoID uint64, name string) bool {
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

// errTreasuryFrozen reports an execution proposal blocked because the host
// DAO's treasury is frozen. It is a sentinel so the freeze gate is greppable
// and testable (the create-time panic and the Validate-time failure share
// this one message).
var errTreasuryFrozen = errors.New("commondao: treasury is frozen")

// executionArgs carries CreateExecutionProposal parameters to the execution
// kind: a title, a body, and the closure executed on approval. Unlike the
// governance kinds, the execution kind captures no mutable handle — its only
// use of the DAO is a freeze-flag read, so its definition holds the readonly
// host view Propose passes to New rather than a *CommonDAO from args.
type executionArgs struct {
	title string
	body  string
	fn    commondao.ExecFunc
}

// executionKind creates proposals that run an arbitrary ExecFunc as the host
// DAO's own sub on approval. It is the realm-side counterpart of the /p/
// commondao.ExecutionKind: identical arbitrary-exec mechanism, but wrapped
// with the realm's Validable freeze policy (executionPropDefinition.Validate)
// so a frozen DAO cannot drain its own treasury through it — the /p/ kind
// carries no such policy. This is the pattern the /p/ extension docs
// recommend for an arbitrary-exec closure: wrap it with your realm's own
// checks.
type executionKind struct{}

func (executionKind) Name() string { return kindExecution }

func (executionKind) New(dao commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	a, ok := args.(executionArgs)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	if a.fn == nil {
		return nil, commondao.ErrExecutionFuncRequired
	}
	// The definition captures the readonly host view Propose passed here, not a
	// handle from args: its only use of the DAO is the Validate freeze read. No
	// mutable handle means no host-identity pin is needed — the view is the
	// host by construction.
	return executionPropDefinition{dao: dao, title: a.title, body: a.body, fn: a.fn}, nil
}

// executionPropDefinition defines a proposal that runs an arbitrary ExecFunc
// as the host DAO's own sub on approval. Its Validate blocks execution while
// the host treasury is frozen, so an execution proposal can never move funds
// out of a frozen DAO — matching the treasury-spend freeze gate.
type executionPropDefinition struct {
	dao   commondao.ReadonlyCommonDAO
	title string
	body  string
	fn    commondao.ExecFunc
}

// Title and Body return raw, user-supplied text: like the /p/ execution
// definition (executionDef) and unlike the self-assembled bodies, this
// definition does not implement trustedMarkdownBody, so the renderer escapes
// both.
func (p executionPropDefinition) Title() string { return p.title }
func (p executionPropDefinition) Body() string  { return p.body }

func (executionPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// Threshold returns the tally threshold: arbitrary execution runs code under
// the DAO's authority, so the supermajority default applies.
func (executionPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

// Validate runs at proposal creation and again inside Execute: a treasury
// frozen after the proposal passed still blocks it cleanly (StatusFailed, no
// funds leave) instead of letting the closure run against a frozen DAO. Freeze
// = no self-initiated treasury movement, spend and execution alike.
func (p executionPropDefinition) Validate() error {
	if p.dao.IsTreasuryFrozen() {
		return errTreasuryFrozen
	}
	return nil
}

func (p executionPropDefinition) Executor() commondao.ExecFunc {
	return p.fn
}

// manageKindsProposal is both the manage-kinds args struct and the proposal
// definition it produces: manageKindsKind.New validates it and returns it
// unchanged (one type serves both the args and definition roles). dao is the
// host DAO whose registry the executor
// mutates: New receives only a readonly view, so the trusted wrapper passes
// the mutable host handle through args (captured from its own mustGetDAO).
//
// A proposal is one of two shapes, both populated only by a trusted
// wrapper and both by name (the name resolves against the realm catalog):
//   - register:   remove=false, name set;
//   - deregister: remove=true,  name set.
//
// Registering a foreign kind by value is intentionally not offered here: on
// this realm such a kind would be inert (no propose path). The by-value
// capability stays available in /p/ (WithProposalKind / RegisterKind) for a
// downstream realm that authors its own propose wrapper (see /p/ doc.gno,
// "Extending commondao in your own realm").
type manageKindsProposal struct {
	dao    *commondao.CommonDAO
	remove bool
	name   string
}

// manageKindsKind creates governance proposals that register or deregister
// a catalog proposal kind on the host DAO by name — the DAO's one permanent
// capability to manage which kinds it accepts. Registering adds a kind;
// deregistering removes one, which blocks new proposals of that kind while
// in-flight ones still vote and execute. It is decided by supermajority and
// cannot itself be deregistered (the self-brick guard below), so a DAO
// always keeps the ability to manage its kind set.
type manageKindsKind struct{}

func (manageKindsKind) Name() string { return kindManageKinds }

func (manageKindsKind) New(dao commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	p, ok := args.(manageKindsProposal)
	if !ok {
		return nil, errInvalidProposalArgs
	}
	// Defense in depth: this kind captures a mutable host handle from args (it
	// must, to register/deregister on Execute) while the checks below read the
	// readonly host Propose passed here. Pin them to the same DAO so a future
	// wrapper can never validate against one registry and mutate another (the
	// trusted wrapper always passes matching handles today).
	if p.dao.ID() != dao.ID() {
		return nil, errInvalidProposalArgs
	}

	if p.remove {
		// Deregister by name. Reject no-ops so a council vote is always
		// about an actual change, and reject the self-brick: manage-kinds
		// is the only un-deregisterable kind, so a DAO can never lose the
		// ability to manage its kind set.
		if !dao.HasKind(p.name) {
			return nil, errors.New("proposal kind is not registered: " + p.name)
		}
		if p.name == kindManageKinds {
			return nil, errors.New("the manage-kinds kind cannot be deregistered")
		}
		return p, nil
	}

	// Register by name: the name must resolve against the realm catalog, and
	// reject a no-op (already registered) so a council vote is always about
	// an actual change.
	if catalogKind(p.name) == nil {
		return nil, errors.New("unknown proposal kind")
	}
	if dao.HasKind(p.name) {
		return nil, errors.New("proposal kind is already registered")
	}

	return p, nil
}

// Title returns the proposal title as raw text: the renderer escapes
// every definition title, so escaping the kind name here (unlike in Body,
// which the renderer trusts as markdown) would double-escape it.
func (p manageKindsProposal) Title() string {
	if p.remove {
		return "Deregister Proposal Kind: " + p.name
	}
	return "Register Proposal Kind: " + p.name
}

func (manageKindsProposal) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// isTrustedMarkdownBody marks Body as self-assembled markdown; the embedded
// kind name is escaped in Body itself (defense in depth).
func (manageKindsProposal) isTrustedMarkdownBody() {}

// Threshold returns the tally threshold: changing which proposal kinds a
// DAO accepts alters its governance surface, so the supermajority default
// applies.
func (manageKindsProposal) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

// Validate re-asserts the self-brick at Execute time (Validate reruns
// inside Execute): the manage-kinds kind can never be deregistered, so a
// DAO always keeps the ability to manage its kind set. This is a
// defense-in-depth second layer behind the New check.
func (p manageKindsProposal) Validate() error {
	if p.remove && p.name == kindManageKinds {
		return errors.New("the manage-kinds kind cannot be deregistered")
	}
	return nil
}

func (p manageKindsProposal) Body() string {
	var b strings.Builder

	// The kind name is validated against the catalog; it is escaped anyway
	// as defense in depth.
	b.WriteString(md.Paragraph(md.Bold("Proposal Kind:") + "\n" + md.EscapeText(p.name)))

	action := "registered: new proposals of this kind can be created."
	if p.remove {
		action = "deregistered: no new proposal of this kind can be created. " +
			"In-flight proposals of the kind still vote and execute."
	}
	b.WriteString(md.Paragraph(md.Bold("Effect:") + "\nThe proposal kind is " + action))

	return b.String()
}

func (p manageKindsProposal) Executor() commondao.ExecFunc {
	return p.execute
}

// execute registers or deregisters the catalog kind, returning any registry
// error unchanged so a race between two concurrently passed manage-kinds
// proposals fails the later one cleanly (StatusFailed) instead of panicking
// the transaction. catalogKind is a process-global immutable lookup, so a
// name valid at New still resolves here; the executor does not re-resolve
// or re-validate beyond the Validate self-brick. It moves no funds, so the
// definition is not Funded and ignores sub.
func (p manageKindsProposal) execute(_ int, sub realm) error {
	if p.remove {
		return p.dao.DeregisterKind(p.name)
	}
	return p.dao.RegisterKind(catalogKind(p.name))
}
