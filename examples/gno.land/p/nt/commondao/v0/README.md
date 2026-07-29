> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# commondao

Governance primitives following the Common DAO Spec
(`docs/CONSTITUTION.md`, Appendix): a `CommonDAO` is a **Council** (a
set of addresses with equal voting power), a proposal lifecycle, and an
optional sub-DAO tree.

```
CommonDAO
├── council:            *addrset.Set — who may vote
├── kinds:              registered ProposalKind factories — what may be
│                       proposed (name → New(dao, args))
├── active proposals:   active + early passed, each with an electorate
│                       snapshot and voting record
├── finished proposals: dismissed / executed / failed / withdrawn
├── treasury:           a derived address + frozen flag (funds moved by
│                       the hosting realm, never by this package)
└── children:           sub-DAOs (each a CommonDAO with a parent pointer)
```

## Quick start

```go
import "gno.land/p/nt/commondao/v0"

// A proposal kind names one proposal type and builds its definitions.
type textKind struct{}

func (textKind) Name() string { return "text" }
func (textKind) New(dao *commondao.CommonDAO, args any) (commondao.ProposalDefinition, error) {
    text, ok := args.(string) // validate args, build the definition
    if !ok || text == "" {
        return nil, errors.New("a proposal text is required")
    }
    return textDefinition{text}, nil
}

var dao = commondao.New(
    commondao.WithName("My DAO"),
    commondao.WithCouncilMember(founder),
    commondao.WithProposalKind(textKind{}), // a type is proposable iff registered
)

// Propose looks the kind up in the DAO's registry and calls its New
// factory with the host DAO and args to build the frozen definition.
// The council snapshot taken at Propose is the proposal's electorate.
// (The package does not gate who proposes — hosting realms do.)
p, _ := dao.Propose(founder, "text", "hello world")

// Electorate members vote; default rule proposals can be decided the
// moment the outcome is settled.
dao.Vote(founder, p.ID(), commondao.ChoiceYes, "")

// Execute runs passed proposals (early passed ones immediately, active
// ones once their voting deadline passes). The host mints a DAO-scoped
// sub-identity and passes it as the executor's value-movement authority.
dao.Execute(p.ID(), sub)
```

## Voting rules (the constitutional defaults)

Every proposal definition returns a `Threshold()`; proposals are decided
by `TallyDefault` with integer math over the proposal's **electorate
snapshot** E (the council at `Propose` time):

```
D = |E| - abstains                 // the tally denominator
supermajority:   pass ⇔ D > 0 && 3*yes >= 2*D   ("two thirds or more")
simple majority: pass ⇔ D > 0 && 2*yes > D       ("more than half")
dismiss (both):       ⇔ 2*no > D
undecided at deadline ⇒ dismissed
```

Abstaining shrinks the denominator (deference); not voting counts
against passage (silence is opposition). Votes are re-evaluated after
every ballot — including changed votes — so a proposal **passes or is
dismissed the moment the outcome is mathematically settled** and an
early-passed proposal may be executed before its deadline.

Vote choices are fixed at YES/NO/ABSTAIN.

## Council changes

`UpdateCouncil(add, remove)` applies idempotent set operations: the
final set is `(council ∪ add) \ remove`, duplicate adds and absent
removes are no-ops (so concurrently passed updates merge in execution
order), full replacement in one call is legal, and an update that
would empty a non-empty council returns `ErrEmptyCouncil` — executors
propagate the error to fail the proposal cleanly.

## Proposal kinds

Proposal types are registered on the DAO, not passed per proposal: a
`ProposalKind` couples a registry name with a `New(dao, args)` factory,
and `Propose(creator, kind, args)` accepts exactly the kinds registered
(`WithProposalKind` at construction; `RegisterKind`/`DeregisterKind`
afterwards, typically from a governance proposal executor). The registry
is read only at `Propose`: deregistering a kind blocks new proposals but
never touches in-flight ones, whose definitions were frozen at creation.
`HasKind`/`KindNames` expose the registry, also on the readonly view.

## Proposal lifecycle

`Propose` (kind-gated as above; capped via `SetMaxActiveProposals`;
`CapExempt` definitions such as council updates bypass the cap, bounded
to one active proposal per creator) → `Vote` (electorate-gated,
deadline-gated, rejects non-active proposals) → `Execute`
(early-passed: immediately, still validating; active: after the
deadline, dismissing undecided proposals) or `Withdraw` (active, zero
votes). `Dissolve` dismisses every in-flight proposal and soft deletes
the DAO; deleted DAOs reject proposals, votes, and executions.

## Treasury

The package stores a treasury `address` (`WithAddress`, `Address()`) and
a frozen flag (`SetTreasuryFrozen`, `IsTreasuryFrozen`) but never moves
funds — hosting realms derive the address (typically a realm
sub-identity via `chain.DerivePkgSubAddr`) and enforce the frozen flag.
`Execute` runs the executor with the DAO-scoped sub-identity the host
passes as its value-movement authority: a fund-moving definition builds
its banker from that `sub`, so value moves are structurally bounded to
that one DAO address. A definition implementing `Funded` names the DAO
whose sub funds it (e.g. clawback sweeps the target, not the host). See
the reference realm's treasury proposals for the constitutional pattern.

## Realm boundaries

A `*CommonDAO` is a mutable handle for the realm that owns it:

1. Do not ACCEPT a `*CommonDAO` from an untrusted caller.
2. Do not RETURN a `*CommonDAO` — return `dao.Readonly()`, a
   `ReadonlyCommonDAO` view whose whole reachable graph is read-only
   (`ReadonlyProposal` flattens `Title()`/`Body()` and never exposes
   the `ProposalDefinition`, whose executor would otherwise be
   callable under your realm's authority).
3. Do not TRUST a readonly view received from an untrusted caller —
   it is a live handle over the sender's data.

See `gno.land/r/nt/commondao/v0` for the reference realm hosting many
DAOs with invitations, council governance, treasuries, and rendering.
