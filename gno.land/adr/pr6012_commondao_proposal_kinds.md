# ADR: commondao proposal kinds registered on the DAO

## Context

Proposal creation used to take a definition per proposal:
`CommonDAO.Propose(creator, def ProposalDefinition)`. Which proposal
types a DAO supports was therefore not a property of the DAO at all —
any code holding the handle could submit any definition, the realm's
`Create*` wrappers were the only (realm-local) discipline, and there was
no way for a DAO's governance to renounce or restore a capability. The
goal: proposal types are **registered on the DAO** — seeded at
construction, changed only by governance — and a definition is no longer
passed with each proposal.

## Decisions

### Kind registry in the package contract

```go
ProposalKind interface {
    Name() string                                           // registry key
    New(dao *CommonDAO, args any) (ProposalDefinition, error) // factory
}

func (dao *CommonDAO) Propose(creator address, kind string, args any) (*Proposal, error)
```

The DAO carries a name→kind registry: seeded via `WithProposalKind`
options at construction, mutated only through `RegisterKind` /
`DeregisterKind` (narrow mutators on the realm-held handle, the
`SetTreasuryFrozen` precedent), read via `HasKind` / `KindNames` (also
on `ReadonlyCommonDAO`). `Propose` looks the kind up (absent →
`ErrProposalKindNotFound`), calls its factory, and then follows the
previous path unchanged — cap + `CapExempt`, electorate snapshot,
definition frozen. The definition-accepting `Propose` is **removed**, so
the gate is part of the package contract: every consumer of `/p/`
gets per-DAO kind gating, not just this realm.

`New` receives the host DAO from `Propose` rather than via args: the
package knows the host, so a kind author cannot mis-wire it (the same
error class as a mis-wired `FundingDAO`, prevented structurally).
Targets and parameters travel in `args`, a per-kind struct type-asserted
inside the factory.

### Realm catalog, full registration at creation

The realm defines a closed catalog of nine realm-authored kinds —
`text`, `council-update`, `ancestor-council-update`, `subdao`,
`dissolve`, `treasury-spend`, `treasury-clawback`, `treasury-freeze`,
`proposal-kinds` — each a stateless singleton whose factory wraps the
existing definition constructor. Every DAO-creation site (genesis, user
`New`, sub-DAO creation) registers the full catalog, so the default is
all-enabled and behavior-compatible. The executor set stays closed: this
realm only ever registers catalog kinds, so no foreign executor can
reach a DAO sub through it (see the exec-scope ADR for why that
matters).

One kind deliberately diverges from "definition takes the host":
dissolution is hosted in the nearest live ancestor while its definition
operates on the dissolved descendant, so the dissolve kind's args carry
the descendant and its factory ignores the host parameter.

### Governance: register/deregister as a proposal

`CreateSetProposalKindProposal(daoID, kindName, enabled)` — council
gated, self-hosted, **supermajority** (a capability change is as
sensitive as council self-mutation). Its factory validates at creation:
the name must be in the catalog, no-ops are rejected (enable an enabled
kind / disable an absent one), and **`proposal-kinds` itself can never
be disabled** — the self-brick guard replacing the old idea of skipping
a gate, since the gate now lives in the package chokepoint. The executor
calls `RegisterKind`/`DeregisterKind` and returns their error, so a race
between two toggles fails cleanly as `StatusFailed`. It moves no funds
and does not implement `Funded`.

### Vote-integrity

The registry is read **only at Propose**. Deregistering a kind never
touches in-flight proposals — they keep their frozen definitions and
still vote and execute. Disabling withdraws only the ability to create
new proposals of that kind.

## Alternatives considered

- **Per-DAO disabled-type bitmask over a fixed enum** (an earlier
  iteration of this PR) — rejected. The gate was realm-local (any other
  realm embedding `/p/` got no gating), definitions were still passed
  per proposal, and it kept two parallel sources of truth (enum +
  mask) beside the actual definition types.
- **Permissionless kind registration by users** — rejected. Kinds carry
  executors (code), and an executor receives the DAO's sub; untrusted
  executors are the treasury-drain scenario the exec-scope ADR
  documents. Registration is realm-authored (catalog) here; downstream
  realms author their own kinds.
- **`New(args)` without the DAO parameter** — rejected; kind authors
  would thread the host manually, recreating the mis-wire class the
  `Funded` review flagged.

## Consequences

- Breaking `/p/` API (quarantined realm + package, no live state):
  `Propose(creator, kind, args)` replaces `Propose(creator, def)`; new
  `ProposalKind`, `WithProposalKind`, `RegisterKind`, `DeregisterKind`,
  `HasKind`, `KindNames`, and `ErrProposalKind{NotFound,Exists,Required}`.
  Package tests/filetests propose through registered probe kinds.
- Realm surface: `proposal_kinds.gno` (catalog, governance kind,
  `CreateSetProposalKindProposal`, `IsProposalKindEnabled`); `Create*`
  wrapper signatures unchanged; the DAO page's Create Proposal section
  lists only registered kinds (goldens gained one additive entry).
- Tests: z_19_a (gate: deregistered kind fails at create with "proposal
  kind not found", others work), z_19_b (vote-integrity), z_19_c
  (disable/re-enable round-trip), z_19_d (supermajority pin, 5-member
  council), z_19_e (non-member rejected), z_19_f (self-brick guard).
  Mutation-verified: threshold weakened → exactly z_19_d fails; guard
  removed → exactly z_19_f fails; partial catalog at creation → 61
  filetests fail. Both gno suites and all five commondao txtars green.
