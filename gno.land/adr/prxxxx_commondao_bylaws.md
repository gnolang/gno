# ADR: bylaws package + commondao amend-bylaws kind

## Context

The Common DAO Spec (`docs/CONSTITUTION.md:1485-1508`) gives every DAO a
Charter (Purpose + Description — already two commondao fields) and,
optionally, **Bylaws and Mandates**: "named plaintext files or folders of
plaintext files," changeable by a **Simple Majority** vote. Nothing in the
tree stored those documents or amended them. Three candidate designs were
drafted (whole-document replace; line-hunk patches; edit-script replay)
and the edit-script design was chosen: it literally applies a diff patch
and reuses the repo's existing tested Myers diff (`p/onbloc/diff`) instead
of introducing a second diff implementation.

## Decisions

### 1. A standalone, governance-agnostic `/p/` package

`gno.land/p/nt/bylaws/v0` stores one DAO's documents as an `avl.Tree` of
`path -> text` and amends them only through verifiable patches. It knows
nothing about DAOs, councils or votes — the consuming realm decides *who*
may amend; the package makes amendments *verifiable and deterministic*.

- **Documents** are plaintext keyed by a slash path
  (`"mandates/treasury.md"`). Folders are a naming convention — a folder
  exists exactly when a document path has it as a prefix; listing one is
  a sorted `avl` prefix scan. No folder objects (an *empty* folder is not
  representable; the requirement's "files or folders of files" does not
  need one).
- **Path charset is restricted** to `[a-zA-Z0-9._-/]` with non-empty
  segments, no `.`/`..`. This keeps paths render/link-safe and the patch
  wire encoding delimiter-free.
- **A stored document is never empty**: an amendment whose result is
  empty removes the document. So the `""` base sentinel ("document
  absent") can never collide with `sha256("")`.
- Size caps: `MaxDocLen` 64 KiB, `MaxPathLen` 200, `MaxOps` 16 Ki.

### 2. The patch is an edit script pinned by a base hash

`Patch{Path, Base, Ops}`: `Base` is the hex sha256 of the text the patch
was diffed against (`""` = create), `Ops` is the Myers edit script
coalesced to run-length ops — `Keep(n)`/`Delete(n)` address the base
positionally by rune count, only `Insert(text)` carries bytes, so a small
edit to a large document is a small patch.

- **Produce**: `Diff` runs `onbloc/diff.MyersDiff(current, proposed)` and
  coalesces. Char-level (the library as-is); a line-level producer would
  be a drop-in upgrade behind the same Patch/replay shape if payloads
  ever prove large.
- **Apply**: verify `hash(current) == Base`, then replay — Keep emits
  base runes and advances, Delete advances, Insert emits — and require
  the script to consume the base exactly. Content verification is the
  hash's job; replay verifies *fit* (counts), so a script that does not
  fit fails (`ErrInvalidPatch`) instead of producing garbage. On any
  error the store is untouched.
- **Optimistic concurrency**: two amendments diffed against the same
  base — the first applied re-pins the hash, the second returns
  `ErrStalePatch`. No clobbering, no merge machinery; the loser re-diffs
  and re-proposes.
- **Wire format** (`Encode`/`DecodePatch`):
  `v0:<path>:<base>:K<n>;D<n>;I<len>:<bytes>;…` — inserts are
  length-prefixed by byte length, so no escaping exists to get wrong.
  Plain strings, CLI/qeval-encodable (unlike the execution kind's
  closure). `Format(base)` renders the change (kept-run markers, deleted
  and inserted text) for proposal bodies; it is raw text the renderer
  escapes.

### 3. Realm wiring: a tenth default kind, `amend-bylaws`

The reference realm holds `bylawsSets` (daoID → `*bylaws.Bylaws`, created
lazily; the mutable handle never leaves the realm — public reads return
strings). `amendBylawsKind` joins `defaultProposalKinds` (10 defaults):
amending governing documents is a constitutional power, so it is seeded
on every DAO, not opt-in.

- `amendBylawsProposal` is args+definition collapsed (the manage-kinds
  pattern) carrying `{daoID, set, patch, display}`. `New` pins
  `daoID == dao.ID()` (host-identity, like manage-kinds), fails fast on a
  stale base, rejects no-op amendments, and renders `display` via
  `Format` — which also validates the script, so a malformed patch never
  becomes a proposal.
- **Threshold: SimpleMajority** (`CONSTITUTION.md:1491`) — the first
  built-in kind below supermajority.
- `Validate` re-asserts freshness (runs again inside Execute): an
  amendment that raced a concurrent change fails cleanly
  (`StatusFailed`), matching the treasury/manage-kinds convention. The
  executor is `set.Apply(patch)` with errors returned, not panicked; it
  moves no funds and ignores `sub` (not `Funded`).
- Public surface: `CreateAmendBylawsProposal(daoID, payload)` (council
  gated), `AmendBylawsPayload(daoID, path, proposed)` (read-only payload
  builder for vm/qeval — uses the non-creating `bylawsView`, so queries
  never write), `GetBylawsDoc`, `ListBylawsDocs`.
- Render: a `{daoID}/bylaws` page (escaped paths and text + per-document
  sha256, since document content is council-controlled input), a DAO-menu
  link when documents exist, and a create-proposal entry.

### 4. What is deliberately out

History/undo (the proposal archive is the audit trail), 3-way merge and
conflict resolution, rename/move ops, cross-document atomic patches,
ancestor-bylaws aggregation in render (ancestors' documents bind per the
Constitution, but aggregating them is a display concern deferred until
wanted), and any `Set` back door — mutation is only through a verified
patch, which is what makes on-chain review meaningful.

## Alternatives considered

- **Whole-document replace with a display-only diff** — simplest
  (~150 LOC), but every amendment ships and re-stores the full text, and
  the proposal payload is not itself the reviewed change. Rejected in
  favor of genuinely applying a patch.
- **Line-based hunk patches (`@@ -3,2 +3,3 @@`)** — most human-readable
  payloads, but requires a second diff implementation (line LCS) plus a
  patch-text parser; the same strict base-hash gate does the real safety
  work in both designs. Rejected for now; noted as the natural upgrade
  for the *producer* side if char-level payloads read poorly in practice.
- **Storing bylaws inside `/p/nt/commondao`** — rejected; the document
  store is generally useful and commondao stays focused. The realm maps
  DAO → document set.

## Consequences

- New package `p/nt/bylaws/v0` (~330 LOC + tests): store, patch engine,
  wire codec; only mutation is `Apply`.
- Realm: `proposal_bylaws.gno` (kind + state), wrapper + payload helper,
  two public reads, render page/link/entry, `kindAmendBylaws` in the
  10-kind default set.
- Goldens: z_10_a, z_10_d, z_18_c regenerate (create-proposal entry +
  kinds row). New filetests z_22_a (governance lifecycle:
  create/amend/list/render/remove), z_22_b (stale race →
  `StatusFailed`, no clobber), z_22_c (stale at create → rejected,
  terminal). Unit tests: package (path rules, diff/apply lifecycle,
  stale, bad scripts, unicode, codec round-trip + malformed, format,
  noop) and realm (`TestAmendBylawsKindNew` guards,
  `TestAmendBylawsLifecycle` executor + race + removal).
- A dissolved DAO's documents remain readable (the DAO record itself is
  soft-deleted); Propose is rejected on deleted DAOs, so they are frozen
  in place.
