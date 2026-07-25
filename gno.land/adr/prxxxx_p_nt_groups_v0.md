# ADR: `p/nt/groups/v0` — group/role membership package with readonly views

## Context

DAOs and permissioned realms repeatedly need the same shape: a base member
set plus named roles, each role a subset of addresses with per-role metadata.
`p/nt/groups/v0` provides that container (`Group`, `Role`) on top of
`p/moul/addrset` (member sets) and `p/nt/bptree/v0` (role registry).

The dominant design constraint is interrealm safety: a `/p/` type with
exported mutator methods is a live write-authority handle under borrow rule
#2, so any pointer that crosses a realm boundary must expose no mutators
(see `docs/resources/gno-ai-contract-review.md`, case #8).

## Decision

1. **Typed readonly views** — `Group.Readonly() *ReadonlyGroup`,
   `Role.Readonly() *ReadonlyRole`, and a new `addrset.ReadonlySet` — are the
   only intended realm-boundary handles. Each holds the mutable object in an
   unexported field and exposes only read methods, so cross-package callers
   can neither reach the underlying object nor dispatch a mutator. Views are
   live handles (not snapshots); the docs state three boundary rules: never
   accept a mutable handle, never return one, and never *trust* a readonly
   view received from an untrusted caller (its contents are sender-controlled
   and mutable between reads).
2. **Explicit API split** — base-only ops (`Add`/`Remove`/`Has`/`Size`/
   `Iterate`), role-registry ops (`AddRole`/`GetRole`/`HasRole`/`RemoveRole`/
   `RoleCount`/`IterateRoles`), and aggregations (`HasAny`/`TotalSize`/
   `IterateAll`/`RolesContaining`/`RemoveFromAll`) are separate so each call
   site picks a semantic deliberately; `Has` never silently consults roles.
3. **Pagination everywhere** — every iterator takes `offset, count`
   (`IterateRoles` included) so `Render` pages stay gas-bounded.
   `Group.IterateRoles` yields mutable `*Role` (owner-side, same capability
   as `GetRole`); `ReadonlyGroup.IterateRoles` wraps each as `*ReadonlyRole`.
4. **Streaming dedup** — `IterateAll`/`TotalSize` share `visitDistinct`,
   which walks base then roles in name order, dedups via an internal addrset,
   and stops as soon as the requested window is served (no upfront
   materialization; no `offset+count` overflow).
5. **`meta any` stays free-form** — the role metadata slot cannot be
   type-constrained without losing generality, so the mutator-method-dispatch
   hazard (storing e.g. `*avl.Tree` retrievable via `ReadonlyRole.Meta()`) is
   documented rather than prevented: store value types only.

## Alternatives considered

- **Interface-based readonly views** (a `ReadonlyGroup` interface implemented
  by `*Group`): rejected — an interface value still carries the concrete
  pointer, recoverable by type assertion.
- **Snapshot copies at the boundary**: rejected — O(N) gas per call and stale
  data; live views with a documented trust rule are cheaper and honest.
- **Restricting meta to a value type or `/r/`-declared interface**: rejected
  for v0 — kills legitimate uses; revisit if misuse shows up in review.
- **Compile-error filetest pinning field privacy**: not possible — the
  `TypeCheckError` directive is restricted to gnovm-internal test files, so
  the property rests on compiler visibility rules plus unit tests.

## Consequences

- Consuming realms get safe-by-construction handles to expose; the remaining
  security obligations (the three boundary rules, meta discipline) are prose
  contracts in the package doc.
- `addrset` gains `ReadonlySet` (`IterateByOffset`/`ReverseIterateByOffset`
  naming matches the `Set` methods it wraps, keeping `Iterate` free for a
  future range iterator).
- Aggregations cost O(scanned) with an O(distinct-scanned) temporary set;
  documented, acceptable for moderate group sizes.
- The zero value of `Group` is unusable (`NewGroup` required), unlike
  `addrset.Set`; documented on the type.
