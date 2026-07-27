# ADR: commondao Charter (Purpose + Description)

## Context

The Common DAO Spec (`docs/CONSTITUTION.md:1485-1487`) makes one thing
mandatory at creation: *"Every DAO, upon creation, must have a Charter (which
is composed of Purpose and Description)."* Bylaws and Mandates are an explicit
*"may also have"* — optional. The realm modelled the Charter only half-way: it
had `description` but no `purpose`, so it was out of compliance on the single
mandatory clause. This PR closes that gap.

Scope was set by a three-reviewer design convergence (recorded in
`CHARTER_DESIGN.md`): ship the Charter now; defer the optional Bylaws/Mandates
document store and all document/Charter amendment to a follow-up, so this
change stays a mechanical, low-risk addition.

## Decision

- **Charter = Purpose + Description.** Add a `purpose` field to the package
  `CommonDAO` beside `description`, with a `Purpose()` accessor (on both
  `CommonDAO` and `ReadonlyCommonDAO`) and a `WithPurpose` option. The package
  stores the string verbatim; the realm enforces validation, matching how
  `name`/`description` already split.
- **Purpose is required (non-empty, ≤250); Description stays optional
  (≤250).** A Charter that must exist is only non-vacuous if Purpose carries
  content; Description was already optional and stays so. The realm's
  `assertDAOPurposeIsValid` enforces this at `New` and at
  `CreateSubDAOProposal` (via `newSubDAOPropDefinition`).
- **`name` stays a separate realm handle**, not part of the spec Charter — it
  is the short label used in links, the tree render, and sub-DAO name
  uniqueness, a realm-UI concern the spec's Charter does not cover.
- **Creation entry points gain a purpose parameter**, slotted after `name`
  (Charter order is "Purpose and Description"): `New(cur, name, purpose,
  description, members)` and `CreateSubDAOProposal(cur, daoID, name, purpose,
  description, members)`. Threaded through `createDAO`/`createSubDAO`/
  `newDAOOptions`/`subDAOPropDefinition`; the sub-DAO proposal body shows the
  purpose; the genesis DAO gains one.
- **Render** shows the Purpose on the DAO page (escaped with `md.EscapeText`,
  the file's existing convention for user strings).
- **No amendment in this PR.** With amendment deferred, a Charter is set at
  creation and not yet changeable — which satisfies the mandatory-at-creation
  clause. Charter amendment is ancestor-only (`:1491`, simple majority) and
  ships with the follow-up.

## Deferred to a follow-up PR (design recorded so it is not relitigated)

- **Bylaws and Mandates** as package-side, per-DAO **flat path-keyed maps**
  (`path → content`, slashes standing in for folders — empty folders are not
  represented, an accepted loss), with caps (content ≤ ~16 KB, ≤ ~64 documents
  per store, path ≤ ~256 B). Package-side because amendment executors mutate
  DAO state through narrow mutators (like `SetTreasuryFrozen`) and the
  Governing-Documents traversal (`:1494`) must walk parents.
- **Amendment proposals** (two): a self-Bylaws proposal (`:1501`) and a
  combined ancestor-document proposal (`kind ∈ {purpose, description, bylaw,
  mandate}`, `ThresholdSimpleMajority`, reusing `assertIsProperAncestor`,
  `:1491`). Charter and Mandates are not self-amendable — the conservative
  reading of `:1501`, which grants the council self-power over Bylaws only.
- **Open interpretation for the follow-up:** the self-Bylaws threshold.
  `:1491`'s Simple-Majority rule is scoped to *ancestor* changes ("from any of
  the DAO's ancestors"); the self-Bylaws power (`:1501`) states no threshold,
  so it falls to the default passage rule (supermajority). A 2/1 review lean
  favored supermajority-via-default; re-confirm at implementation.

## Consequences

- Breaking signature changes to `New` and `CreateSubDAOProposal` (quarantined
  realm, no live state). Package gains `WithPurpose`/`Purpose()`; existing
  package `New(options...)` calls are unaffected (variadic).
- Migration was mechanical: a purpose argument inserted into ~82 realm
  filetests and the treasury txtar, and every DAO-page golden gained a Purpose
  line (regenerated and reviewed — the only golden drift is the Purpose line,
  the sub-DAO body's Purpose, and surrounding spacing). New tests: `z_18_a`
  (empty purpose rejected) and a package `TestCharter`.
- Treasury and council/tally invariants are untouched: Purpose/Description are
  inert data read by no authority, treasury, or tally path.
