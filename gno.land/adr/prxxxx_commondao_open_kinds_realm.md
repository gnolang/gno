# ADR: commondao open kind-registration — reference realm rewire (M2)

## Context

`prxxxx_commondao_open_kinds_pkg.md` (M1) narrowed the `/p/` package:
`ProposalKind.New` now takes a `ReadonlyCommonDAO` (was `*CommonDAO`), and
`Funded.FundingDAO() *CommonDAO` became `FundingDAOID() uint64`. M1 shipped the
`/p/` default kinds (`ExecutionKind` `"execution"`, `RegisterKindKind`
`"register-kind"`) but did **not** touch the reference realm
(`r/nt/commondao/v0`), which therefore no longer compiled against the package:
its 9 kinds implemented the old `New(*CommonDAO, ...)` signature and its `Funded`
definitions returned `*CommonDAO`. M1 explicitly predicted this rewire.

This ADR covers only the reference realm.

## Decisions

### 1. Host handle via trusted args-capture (all self-mutating kinds)

Each realm kind's `New` now takes `commondao.ReadonlyCommonDAO`. A kind that
mutates DAO state on execution can no longer read the mutable host from `New`'s
param, so the trusted `Create*` wrapper — which already holds the handle via
`mustGetDAO` — passes it through the kind's args struct, and the definition
captures it and mutates in its `Executor()` (unchanged post-vote behavior).

- Kinds that already carried their target via args barely changed
  (`ancestor-council-update`, `dissolve`, `treasury-clawback`, `treasury-freeze`
  — their args gained the proposing-ancestor `host` where the executor/validator
  needs it, `dissolve` already self-contained).
- The **self-host** kinds now carry the host explicitly: `council-update` and
  `proposal-kinds` gained `dao *CommonDAO`; `subdao` gained `parent`;
  `treasury-spend` gained `dao`. External proposers cannot obtain a `*CommonDAO`
  to inject, so args-capture is the trust boundary (per M1 decision 2).

`proposalKindsKind.New` also does its no-op `HasKind` pre-checks against the
args handle (same host the readonly view would wrap).

### 2. `Funded` → `FundingDAOID`; host resolves the sub

`treasury-spend` returns the host id, `treasury-clawback` the target id,
`dissolve` the dissolved-descendant id. `public.Execute` resolves it:

```go
op := daoID
if f, ok := p.Definition().(commondao.Funded); ok { op = f.FundingDAOID() }
sub := cur.Sub(subpathOf(op))
```

### 3. Opt-in catalog: execution + register-kind (Q5)

`defaultProposalKinds` (the 9 kinds seeded on every DAO) is split from
`proposalKindCatalog` (those 9 **plus** the /p/ `ExecutionKind{}` +
`RegisterKindKind{}`). `catalogKind(name)` resolves against the full catalog, so
the existing supermajority `CreateSetProposalKindProposal` can enable an opt-in
kind; `newDAOOptions` seeds only `defaultProposalKinds`. Default-on arbitrary
execution is a silent, irreversible capability expansion; opt-in is reversible.

### 4. Propose paths for the two /p/ default kinds (M1 residual)

`CreateExecutionProposal(cur, daoID, title, body, fn)` and
`CreateRegisterKindProposal(cur, daoID, kind)` are council-gated realm wrappers.
Their `ExecFunc`/`ProposalKind` params are not CLI-encodable, so they are
reachable only from a persistent realm that imports this one (the execution
closure must be authored there to survive Propose→Execute). Both are gated by
`assertKindEnabled(dao, kind)` so they work only after the DAO enabled the kind
through governance — failing fast with a clear message before any definition is
built (`Propose` would otherwise reject the unregistered kind with a vaguer
`ErrProposalKindNotFound`). Because the params are non-CLI-encodable, neither
gets a create-link in the DAO render page.

### 5. Render: escape proposal bodies by default (inversion)

The `/p/` `executionDef`/`registerKindDef` return raw, user-influenced
Title/Body but are package-typed, so they cannot implement the realm's private
raw-text marker. Rather than special-case them (impossible to type-assert an
unexported foreign type, and fail-open for any future raw-body kind), the render
layer was **inverted**: proposal bodies are `md.EscapeText`-escaped by default,
and only definitions that assemble their own markdown opt out via a new
`trustedMarkdownBody` marker (the 8 markdown-building realm defs). Titles were
already escaped universally. The /p/ default kinds fall into the escaped-by-
default bucket automatically. This is fail-safe: a new raw-body kind that
forgets to opt out is escaped, not leaked. No existing filetest output changed
(the markdown-building defs render verbatim exactly as before; the text kind was
already escaped).

## Alternatives considered

- **Keep raw-text opt-in escaping, detect execution at render.** Rejected:
  `executionDef` is unexported in `/p/`, not type-assertable from the realm, and
  the proposal carries no kind name — no render-time signal exists. Opt-in
  escaping is also fail-open.
- **Wrap the /p/ execution def in a realm def that implements the marker.**
  Rejected: the propose path uses the /p/ `ExecutionKind` (per the milestone),
  whose `New` builds the /p/ def internally; the realm cannot substitute a wrapper.
- **Add create-links for the opt-in kinds.** Rejected: `ExecFunc`/`ProposalKind`
  are not fillable from a web `$help` link.
- **Rely solely on `Propose`'s unregistered-kind rejection instead of
  `assertKindEnabled`.** Rejected: a vaguer error and no explicit opt-in
  invariant at the realm surface.

## Consequences

- The realm compiles against the M1 package again; the `/p/` package is
  unchanged by this milestone and stays green.
- The 9 governance kinds stay default-seeded; execution + register-kind are
  opt-in per DAO via the existing supermajority toggle.
- The realm keeps its global `executing` latch (`reentrancy.gno`) for the
  cross-DAO/ancestor-dissolution straddle the per-DAO fields miss.
- New filetests `z_20_a..d`: execution end-to-end, register-kind end-to-end,
  render escaping of a hostile execution title/body, and the opt-in enable-gate.
  The gate is mutation-verified: removing `assertKindEnabled` from
  `CreateExecutionProposal` flips `z_20_b`'s error from
  `proposal kind is not enabled: execution` to `proposal kind not found`,
  failing the golden.
- The readonly-`New` guarantee remains type-enforced: a realm kind's `New`
  cannot reach a mutator through its `ReadonlyCommonDAO` param.
