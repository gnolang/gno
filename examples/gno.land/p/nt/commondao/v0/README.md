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
├── active proposals:   active + early passed, each with an electorate
│                       snapshot and voting record
├── finished proposals: dismissed / executed / failed / withdrawn
└── children:           sub-DAOs (each a CommonDAO with a parent pointer)
```

## Quick start

```go
import "gno.land/p/nt/commondao/v0"

var dao = commondao.New(
    commondao.WithName("My DAO"),
    commondao.WithCouncilMember(founder),
)

// Council members propose; the council snapshot taken here is the
// proposal's electorate.
p, _ := dao.Propose(founder, myDefinition)

// Electorate members vote; default rule proposals can be decided the
// moment the outcome is settled.
dao.Vote(founder, p.ID(), commondao.ChoiceYes, "")

// Execute runs passed proposals (early passed ones immediately, active
// ones once their voting deadline passes).
dao.Execute(p.ID(), cur)
```

## Voting rules (the constitutional defaults)

Proposal definitions that implement `DefaultTallier` are decided by
`TallyDefault` with integer math over the proposal's **electorate
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

Custom definitions implement `ProposalDefinition.Tally(VotingContext)`
instead and are tallied once, at the deadline. Definitions may also
customize vote choices (`CustomizableVoteChoices`) — but not combined
with `DefaultTallier`.

## Council changes

`UpdateCouncil(add, remove)` applies idempotent set operations: the
final set is `(council ∪ add) \ remove`, duplicate adds and absent
removes are no-ops (so concurrently passed updates merge in execution
order), full replacement in one call is legal, and an update that
would empty a non-empty council returns `ErrEmptyCouncil` — executors
propagate the error to fail the proposal cleanly.

## Proposal lifecycle

`Propose` (capped by `WithMaxActiveProposals`; `CapExempt` definitions
such as council updates bypass the cap, bounded to one active proposal
per creator) → `Vote` (electorate-gated, deadline-gated, rejects
non-active proposals) → `Execute` (early-passed: immediately, still
validating; active: after the deadline, dismissing undecided
proposals) or `Withdraw` (active, zero votes). `Dissolve` dismisses
every in-flight proposal and soft deletes the DAO; deleted DAOs reject
proposals, votes, and executions.

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
DAOs with ownership, invitations, options, and rendering.
