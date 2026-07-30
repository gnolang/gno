# ADR: commondao unify proposal-kinds + register-kind into one manage-kinds kind

> **Amended by the no-smell cleanup (same PR series).** The final shape drops
> three residuals this ADR introduced:
>
> - **Reference-realm registration is by-NAME only.** The by-value wrapper
>   `CreateRegisterCustomKindProposal` (decisions 1, 2, 5; and the "kept and
>   documented" alternative) is **DELETED**: on this realm a by-value-registered
>   foreign kind is inert (no propose path), so it was governance dead weight.
>   The `kind commondao.ProposalKind` field is dropped from the manage-kinds
>   args/definition — the shape is now `{dao, remove, name}`, register and
>   deregister both by name. The by-value capability stays available in `/p/`
>   (`WithProposalKind` / `RegisterKind`) as a documented downstream pattern
>   (see `/p/` `doc.gno` / README "Extending commondao in your own realm"), not
>   a governance wrapper. The two surviving wrappers are
>   `CreateRegisterKindProposal` (by name) and `CreateDeregisterKindProposal`.
> - **The args/definition twin is collapsed.** `manageKindsArgs` and
>   `manageKindsPropDefinition` were byte-identical structs joined by a
>   rename-only conversion; they are now ONE type, `manageKindsProposal` (the
>   definition IS the args type — `New` type-asserts it, validates, returns it).
> - **"enabled" → "registered".** `IsProposalKindEnabled` →
>   `HasProposalKind`; `assertKindEnabled` → `assertKindRegistered`; the error
>   strings become "already registered" / "proposal kind is not registered:
>   <name>". `Funded` also moved from `/p/` to the realm (host-consumed, not
>   package-dispatched; see `prxxxx_commondao_open_kinds_pkg.md`). Filetest
>   `z_20_d` (register-custom) is replaced by a by-name deregister/re-register
>   round trip. The self-brick and the `CreateExecutionProposal` gate are
>   unchanged and still mutation-verified (z_19_f, z_20_b). The body below is
>   kept for the record.

## Context

`prxxxx_commondao_open_kinds_pkg.md` (M1) and `prxxxx_commondao_open_kinds_realm.md`
(M2) shipped, unreleased and quarantined, two overlapping governance kinds that
both mutate a DAO's kind registry, split only by input encoding and asymmetric
in their operations:

- **`proposal-kinds`** (realm `proposalKindsKind`, wrapper
  `CreateSetProposalKindProposal(daoID, kindName, enabled)`): enable/disable a
  **catalog** kind **by name**; the disable path also handled removing any
  registered kind.
- **`register-kind`** (the /p/ `RegisterKindKind`, wrapper
  `CreateRegisterKindProposal(daoID, kind)`): register a **foreign** kind **by
  value**.

Plus a /p/ self-brick (`ErrCannotDeregisterRegisterKind`) that made the
`register-kind` name un-removable — which produced the "disable register-kind is
accepted-but-doomed" wart and was questionable anyway (the realm's
`proposal-kinds` could re-enable it, so it was not truly unrecoverable).

Net: one idea — "manage which kinds a DAO accepts" — spread across two
differently-named kinds plus a /p/ policy lock, because names are tx-encodable
and kind-values are not.

## Decision

### 1. One symmetric governance kind: `manage-kinds`

A DAO's accepted kinds are a set; governance changes it with two symmetric ops,
**register** / **deregister**, over a kind identified either by **catalog name**
(curated, CLI-encodable) or **by value** (foreign, realm-authored callers only).

The realm ships a single permanent kind `manageKindsKind` (name const
`kindManageKinds = "manage-kinds"`) with a `manageKindsPropDefinition{dao
*commondao.CommonDAO, remove bool, name string, kind commondao.ProposalKind}`.
Supermajority (both merged kinds were supermajority; unchanged). The executor:
`remove` → `dao.DeregisterKind(name)`; else `dao.RegisterKind(kind if non-nil
else catalogKind(name))`, returning the registry error unchanged. It mutates via
the args-captured `*CommonDAO` (the unchanged trust boundary: only the trusted
wrapper populates `dao`/`kind`; `New` receives only the readonly view).

Three council-gated wrappers (via `mustPropose`) replace the two old ones:

- `CreateRegisterKindProposal(cur, daoID, kindName string)` — register a
  **catalog** kind by name (e.g. `"execution"`).
- `CreateDeregisterKindProposal(cur, daoID, kindName string)` — deregister by
  name.
- `CreateRegisterCustomKindProposal(cur, daoID, kind commondao.ProposalKind)` —
  register a **foreign** kind by value (realm-only; not CLI-encodable).

`New` validates: register-by-name → `catalogKind(name) != nil` &&
`!dao.HasKind(name)`; register-by-value → `kind != nil && kind.Name() != ""` &&
`!dao.HasKind(kind.Name())`; deregister → `dao.HasKind(name)` && the self-brick
below. No-ops are rejected so a council vote is always about a real change.

Deleted: realm `proposalKindsKind`, `setProposalKindArgs`,
`setProposalKindPropDefinition`, `CreateSetProposalKindProposal`, the old
value-based `CreateRegisterKindProposal`, consts `kindProposalKinds` /
`kindRegisterKind`. Kept: `CreateExecutionProposal` + its
`assertKindEnabled(dao, kindExecution)` gate, `IsProposalKindEnabled`, the
realm-global `executing` latch.

Seeding: `defaultProposalKinds` swaps `proposalKindsKind{}` →
`manageKindsKind{}` (still 9 default kinds). `optInProposalKinds` becomes just
`{ExecutionKind{}}` (foreign registration now goes through `manage-kinds`, not a
/p/ meta-kind). Catalog = defaults ∪ opt-in, as before.

### 2. Naming / semantic swap of `CreateRegisterKindProposal`

The name `CreateRegisterKindProposal` is kept but its meaning **changes from
by-value → by-name** (register a catalog kind). By-value registration moves to
the new `CreateRegisterCustomKindProposal`. This is safe: both are unreleased
and quarantined. The symmetric set reads cleanly: register-by-name /
deregister-by-name / register-custom-by-value.

### 3. Two-layer self-brick, moved from /p/ to realm policy

`manage-kinds` is the ONLY un-deregisterable kind (it is the recovery
mechanism). Everything else — including `execution` and the other built-in kinds
(the 9 defaults minus `manage-kinds`) — is freely register/deregister-able and,
for catalog kinds, re-registerable by name.

The guard is enforced in **two layers**, both realm-side:

- `manageKindsKind.New` rejects `remove && name == kindManageKinds` at creation.
- `manageKindsPropDefinition.Validate()` re-asserts the same, so the guard holds
  at Execute too (`Validate` reruns inside `Execute`). This is the
  defense-in-depth second layer that used to be provided (for the old
  `register-kind` name) by the /p/ lock.

### 4. /p/ becomes pure mechanism (reverses M1 R3)

The /p/ package drops the `ErrCannotDeregisterRegisterKind` self-brick from
`DeregisterKind` (now a plain primitive with no reserved names), and drops
`RegisterKindKind` / `RegisterKindArgs` / `WithDefaultKinds`. This **reverses M1
decision R3** ("/p/ ships a register-kind default kind"): governance
kind-management is now realm policy, not a /p/ meta-kind. A pure-/p/ consumer
that wants it authors its own managing kind + anti-brick on top of the
primitives.

/p/ now exports only: the `ProposalKind` interface, the
`RegisterKind`/`DeregisterKind`/`HasKind`/`KindNames` registry primitives, the
`WithProposalKind` option, and the one concrete kind `ExecutionKind` /
`ExecutionArgs` (reusable arbitrary execution). No "meta" kinds in /p/. There is
no `WithDefaultKinds` replacement — callers seed execution with
`WithProposalKind(ExecutionKind{})`.

### 5. Inert-to-propose (unchanged residual, stated plainly)

On THIS reference realm a by-value foreign kind registered via
`CreateRegisterCustomKindProposal` is registered but has **no propose path**:
every `Create*` wrapper hardcodes its own kind and `mustPropose` is unexported.
So a custom kind is inert-to-propose here and useful only to a downstream realm
that imports /p/ and authors its own propose wrapper (or to a future generic
propose path). This residual is unchanged by this refactor; the wrapper is kept
(not silently dropped) and the inertness is documented at the wrapper.

### 6. render.gno

The Create-Proposal table entry keyed on `kindProposalKinds` becomes one keyed
on `kindManageKinds`, and the single `setProposalKindLink` ("Set Proposal Kind"
→ `CreateSetProposalKindProposal`) is rewritten into two links — "Register
Proposal Kind" (`CreateRegisterKindProposal`) and "Deregister Proposal Kind"
(`CreateDeregisterKindProposal`), each keyed on `daoID` + `kindName`.
Register-custom stays link-less (not CLI-encodable). Goldens z_10_a, z_10_d,
z_18_c regenerated; the only changes are the manage-kinds link entries (and, in
z_10_d, the settings kinds-row name `proposal-kinds` → `manage-kinds`, same
alphabetical slot).

## Alternatives considered

- **Keep the two kinds, just rename.** Rejected: the smell is conceptual (two
  meta-kinds for one idea), not cosmetic; renaming leaves the asymmetry and the
  doomed-disable wart.
- **Keep the /p/ self-brick, add the realm one on top.** Rejected: with
  `register-kind` gone there is no /p/ name to lock, and a /p/ lock on
  `manage-kinds` would leak realm policy into the mechanism layer. Anti-brick
  belongs to the realm that defines the recovery kind.
- **Drop `CreateRegisterCustomKindProposal`** (by-value) since it is inert here.
  Rejected: it is a user-requested capability, useful to downstream realms, and
  the trust boundary is unchanged. Kept and documented.
- **Fold register+deregister into a single `enabled bool` wrapper** (as the old
  `CreateSetProposalKindProposal`). Rejected: two explicit verbs render as two
  explicit links, are clearer at the CLI, and drop the boolean.

## Consequences

- **Simplicity accounting (honest):** the win is conceptual — 2 governance
  meta-kinds → 1, symmetric ops, /p/ de-meta'd — **not** wrapper count, which
  goes 2 → 3. The circular "kind that manages kinds" reduces from two
  self-brick-protected meta-kinds to exactly one; the doomed-disable wart
  disappears.
- **Preserved capabilities:** enable/disable catalog kinds (now symmetric
  register/deregister by name); register foreign kinds by value (realm-only,
  same inert-to-propose residual as before); arbitrary execution (`execution`
  kind, opt-in). Governable capability set intact — deregister a built-in to
  renounce a power, re-register by name to restore.
- **Breaking (safe):** `CreateRegisterKindProposal` changes signature and
  meaning (by-value → by-name); `CreateSetProposalKindProposal` removed; /p/
  `RegisterKindKind`/`RegisterKindArgs`/`WithDefaultKinds`/
  `ErrCannotDeregisterRegisterKind` removed. All unreleased and quarantined.
- **Tests:** /p/ `default_kinds_test.gno` drops the `WithDefaultKinds` /
  `RegisterKindKind` / self-brick tests and adds a plain `DeregisterKind` test
  (deregister-any succeeds, unknown → `ErrProposalKindNotFound`). Realm
  `proposal_kinds_test.gno` rewritten to `manageKindsKind` (all New cases + the
  `Validate` self-brick layer). Filetests z_19/z_20 rewritten to the symmetric
  API (z_20_d → `CreateRegisterCustomKindProposal`; new z_20_e/z_20_f pin the
  no-op register/deregister rejections). z_10_a/z_10_d/z_18_c goldens
  regenerated (link-only).
- **Mutation-verified:** removing the `manage-kinds` self-brick from `New` fails
  the self-brick filetest (z_19_f); removing `assertKindEnabled` from
  `CreateExecutionProposal` fails its gate filetest (z_20_b).
- No new attack surface: the args-capture trust boundary, the readonly-`New`
  guarantee, and the re-entrancy latch are all unchanged.
