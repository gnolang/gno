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
│                       proposed (name → New(readonly dao, args))
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
func (textKind) New(dao commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
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
// factory with a readonly view of the host DAO and args to build the
// frozen definition. The council snapshot taken at Propose is the
// proposal's electorate. (The package does not gate who proposes —
// hosting realms do.)
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
`ProposalKind` couples a registry name with a
`New(dao ReadonlyCommonDAO, args)` factory, and
`Propose(creator, kind, args)` accepts exactly the kinds registered
(`WithProposalKind` at construction; `RegisterKind`/`DeregisterKind`
afterwards, typically from a governance proposal executor). The registry
is read only at `Propose`: deregistering a kind blocks new proposals but
never touches in-flight ones, whose definitions were frozen at creation.
`HasKind`/`KindNames` expose the registry, also on the readonly view.

`New` receives only a **`ReadonlyCommonDAO`**, so a kind — including an
externally-authored or user-registered one — cannot mutate the host DAO
(or its tree) at `Propose` time, before the vote. A kind that must mutate
state on execution takes the target `*CommonDAO` through `args`, which
only a trusted caller can populate (an external proposer cannot obtain a
`*CommonDAO`), captures it in the definition, and mutates in its
`Executor` — which runs only after the vote passes.

### The `ExecutionKind` concrete kind

The package ships exactly one concrete kind, `/p/`-typed so any realm can
seed it with `WithProposalKind(ExecutionKind{})` or register it later
with `RegisterKind`:

- **`ExecutionKind`** (`"execution"`) runs an arbitrary `ExecFunc`
  supplied by the proposer (`ExecutionArgs{Title, Body, Fn}`) on
  approval. The `Fn` closure is frozen at `Propose` (vote-integrity), so
  it **must be authored in a persistent realm** — a closure created by a
  `maketx run` script does not persist to `Execute` and cannot run.

A registered foreign-realm kind runs under its **defining** realm's
authority — registering one is a governance trust grant, not a sandbox.

The package ships **no** governance meta-kinds. `RegisterKind` /
`DeregisterKind` are plain registry primitives with no reserved names:
any registered kind can be removed. Managing a DAO's kind set through
governance — and keeping a managing kind un-removable so a DAO can always
recover — is the consuming realm's policy, built on these primitives (see
the reference realm's `manage-kinds` kind).

## Extending commondao in your own realm

The package is pure mechanism; every governance policy lives in the
consuming realm. To add your own proposal type:

1. **Author a `ProposalKind`** — `Name()` plus
   `New(dao ReadonlyCommonDAO, args any) (ProposalDefinition, error)`. Make
   the definition `Executable` if it mutates on approval. If its executor
   moves funds from a DAO other than the host, have the **host** realm
   consume a `Funded`-style contract (`FundingDAOID() uint64`): minting a
   DAO sub needs the host's `cur`, so it is host-consumed, not
   package-dispatched — define it in your realm.
2. **Seed it** — the owning realm holds the handle, so no proposal is
   needed: `commondao.New(WithProposalKind(YourKind{}), …)` at
   construction, or `dao.RegisterKind(YourKind{})` directly.
3. **Author a typed, CLI-friendly wrapper**
   `CreateYourProposal(cur realm, daoID uint64, …params…)` that
   council-gates the caller, builds the args, and calls `Propose`.
4. **Optionally add a governance toggle** — a `manage-kinds`-style kind
   whose executor calls `RegisterKind`/`DeregisterKind`, kept itself
   un-deregisterable, if the council should manage kinds at runtime.

**Trust boundary:** `New` gets only a `ReadonlyCommonDAO`; the mutable
`*CommonDAO` reaches a definition only via args your trusted wrapper
populates; the executor gets the DAO's terminal, RealmSend-only sub. See
the reference realm for a worked example.

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
that one DAO address. Which DAO's sub the host mints is the host's
decision — minting a sub needs the host realm's `cur`, so the package
cannot make it — typically the proposal's own DAO, but a fund-moving
definition may direct the host to a different DAO (e.g. clawback sweeps
the target, not the host). See the reference realm's treasury proposals
and its host-side `Funded` contract for the constitutional pattern.

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
