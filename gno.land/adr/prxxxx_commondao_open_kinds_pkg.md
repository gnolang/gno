# ADR: commondao open kind-registration + /p/ default kinds (package layer, M1)

## Context

`pr6012_commondao_proposal_kinds.md` made proposal types a per-DAO registry
(`ProposalKind{Name(); New(dao *CommonDAO, args)}`, `RegisterKind`/
`DeregisterKind`), but the `/p/` package shipped **no** concrete kind: the
9-kind catalog lived realm-side and was closed. `pr6012_commondao_reentrancy.md`
added a realm-**global** `Execute` latch and explicitly flagged the follow-up:
the package `Execute` is exported and unguarded, `Parent()` returns a live
`*CommonDAO`, and a future arbitrary-execution kind's closure must not be able
to reach the package `Execute` directly and bypass the latch.

The maintainer's firm requirements for this milestone:

- **R1** a DAO's kind set is mutable/open, not a closed catalog;
- **R2** `/p/` ships a default arbitrary-execution kind (runs a supplied
  `ExecFunc`);
- **R3** `/p/` ships a default `register-kind` kind (governance-driven
  registration);
- **R4** external realms can construct DAOs with open registration without a
  proposal (constructor/mutators; only cross-DAO/governance registration needs a
  proposal).

Opening registration brings **untrusted** kind code into scope. This ADR covers
only the **package layer** (`/p/nt/commondao/v0`); the reference realm rewire is
M2.

## The crux (trust boundary)

With `New(dao *CommonDAO, args)`, a single council member's `Propose` of a
registered kind runs that kind's `New` **pre-vote**, holding the full mutable
handle. A hostile `New` could `UpdateCouncil`/`Dissolve`/`RegisterKind` — and via
`Parent()`, mutate the whole tree — with zero votes. Convention ("New must be
pure") does not bind untrusted code. This must be closed structurally.

## Decisions

### 1. Narrow `New` to a readonly view

```go
ProposalKind interface {
    Name() string
    New(dao ReadonlyCommonDAO, args any) (ProposalDefinition, error)
}
```

`ReadonlyCommonDAO` exposes no mutators and is not downcastable to `*CommonDAO`,
so a kind that only receives it provably cannot mutate the host or its tree at
Propose time. `New` becomes a pure factory.

### 2. Mutation via trusted args-capture; NO capability interface

Kinds that must mutate receive the target `*CommonDAO` through **`args`**, which
only trusted callers can populate (the owning realm, or realm wrappers that
already hold the handle). An external proposer cannot obtain a `*CommonDAO` to
inject. The definition captures it and mutates in its `Executor()` (post-vote).
An open structural `StateChanging{ChangeState(*CommonDAO, realm)}` interface was
**rejected**: an external type could satisfy it and receive the handle post-vote.
Execution + external kinds carry no handle in args → structurally bounded to the
DAO-scoped `sub`.

`Funded.FundingDAO() *CommonDAO` → `Funded.FundingDAOID() uint64` for the same
reason: the interface must not hand out a mutable handle. The host resolves the
ID to mint the funding sub (M2).

### 3. Two per-DAO latches in the package

- `executing bool` guarded in `Execute` (after the `deleted` check; raise, then
  `defer` lower). Blocks an executor from re-entering `Execute` on the same DAO —
  including a *different* proposal — closing the package-`Execute` bypass the
  reentrancy ADR flagged. This gives external `/p/` consumers same-DAO re-entry
  protection that the realm-global latch never gave them.
- `proposing bool` guarded around the `New` call in `Propose`. Closes re-entrant
  `New` (a captured handle re-entering `Propose` mid-factory).

Both **panic** on violation (unrecoverable across `cross` → whole-tx abort). A
mutable `/p/` global is forbidden by the borrow rule, but a **field on the owning
realm's object is writable** — that is why these are fields, not globals. The
realm keeps its global latch for the cross-DAO/ancestor-dissolution straddle,
which a per-DAO field misses (M2).

### 4. Default kinds (opt-in) + self-brick guard

`/p/` ships `ExecutionKind` (`"execution"`) and `RegisterKindKind`
(`"register-kind"`) as stateless singletons, plus `WithDefaultKinds()`
(batteries-included seeding). `DeregisterKind` **refuses the reserved
`register-kind` name** (returns `ErrCannotDeregisterRegisterKind`, matching the
error-return style — not a panic): deregistering it would brick the only
governance path to add kinds back. The reference realm adds them to its catalog
but does not default-seed them (M2): default-on arbitrary execution is a silent,
irreversible capability expansion; opt-in is reversible.

- `ExecutionArgs{Title, Body string; Fn ExecFunc}` → `executionDef` (7-day
  voting period, supermajority; `Executable`, not `Funded`). The `Fn` closure is
  frozen at Propose (vote-integrity) and must be authored in a **persistent
  realm** — a `maketx run` script's closure does not persist to Execute.
- `RegisterKindArgs{DAO *CommonDAO; Kind ProposalKind}` → `registerKindDef`
  (supermajority; `Executor` → `dao.RegisterKind(kind)`). `New` pre-checks
  `DAO.HasKind` → `ErrProposalKindExists`. A registered foreign-realm kind runs
  under its **defining** realm's authority (borrow rule #1) — a governance trust
  grant, not a sandbox.

## Alternatives considered

- **Keep `New(*CommonDAO)`, registration-is-trust.** Rejected: open registration
  + register-kind admit untrusted kinds; a single member's propose then mutates
  governance pre-vote. Not closable by convention.
- **`StateChanging` capability interface.** Rejected (see decision 2): satisfiable
  by an external type.
- **Package-global latch.** Impossible (mutable `/p/` global forbidden). Per-DAO
  fields are the writable equivalent.
- **Deregister-kind default kind.** Deferred: not required by R1–R4 and expands
  the self-brick surface.

## Consequences

- Breaking `/p/` API: every `ProposalKind.New` implementation must take
  `ReadonlyCommonDAO`; `Funded` implementers must switch to `FundingDAOID`. All
  package tests/filetests updated. **M2 (the reference realm) does not compile
  against the new package until its governance kinds capture `*CommonDAO` via
  args and its `Funded` definitions switch to `FundingDAOID`.**
- New file `default_kinds.gno` (kinds + args + defs + reserved-name consts);
  `WithDefaultKinds()` in `commondao_options.gno`; `executing`/`proposing`
  fields + guards; `ErrCannotDeregisterRegisterKind`.
- Guards mutation-verified: removing `executing` fails the re-entrant-Execute
  test; removing the self-brick guard fails the self-brick test. The
  readonly-New guarantee is enforced by the type system (a `New` reaching a
  mutator does not compile) — documented, not a runtime test.
- The reentrancy ADR's flagged package-`Execute` bypass is now closed at the
  package layer by the per-DAO `executing` field.
