# Cache type-privacy checks across commits (`typeHasPrivateDep`)

## Context

Every time a realm object is created or updated, `saveUnsavedObjects`
calls `assertTypeIsPublic` on its type to make sure nothing reachable
from that type is declared in a private package (a realm may use its
own private types, but another realm may not persist a reference to
them). This check recursively walks the full type graph — struct
fields, interface methods, declared-type methods and their signatures —
computing `TypeID()` and looking up each encountered package's privacy
via the store.

The walk is repeated **in full, from scratch, on every single commit**,
even for a type that was already proven type-private-clean in an
earlier transaction. `saveUnsavedObjects` builds a fresh
`visited map[TypeID]struct{}` per call specifically to dedupe repeats
*within* one commit; there is no memory across commits. A realm whose
objects reuse a handful of **acyclic** struct types — plain data/config
shapes with no self- or mutual reference, which covers most declared
types — re-pays the same O(graph size) walk, and the same repeated
`store.GetPackage` lookups, on every object saved in every block — pure
node-side overhead with no offsetting gas charge (this whole path —
confirmed by inspection — never calls
`ConsumeGas`/`incrCPU`/`Allocator.Allocate`; it's uncharged CPU work,
not a fee-fairness concern, just wasted time).

**This does not cover every "commonly reused type," notably not the
single most common stateful collection shape in the ecosystem.**
`gno.land/p/nt/avl`'s `Node` (`leftNode *Node`, `rightNode *Node`) is
self-referential, and — as the Decision section below explains — any
type reached only through a cycle is deliberately excluded from the
permanent cache. Realms built on `avl.Tree` (an extremely common
pattern for ordered maps/sets) get zero benefit from this change; they
keep paying the full walk on every commit, exactly as before. See
"Limitations" under Consequences.

An `XXX JAE` comment on `assertTypeIsPublic` already named the fix:
precompute an `IsPrivate`-style flag per type, once, instead of
re-deriving it every commit.

## Decision

Add `typeHasPrivateDep(store, t) bool` — a fast-path check consulted at
the top of `assertTypeIsPublic` that returns whether *anything*
reachable from `t` belongs to a package with `Private == true`. If
false, `assertTypeIsPublic` returns immediately: there is nothing left
in the type graph that could ever trigger the panic below, for any
realm, so the exact walk is skipped entirely.

The result is memoized permanently, per type, via a new `privateDep
uint8` tristate field (0 = unknown, 1 = no private dep, 2 = has one) on
`StructType`, `InterfaceType`, and `DeclaredType` — the only three
kinds that carry a `pkgPath` — mirroring the existing `comparable`/
`effectiveFields` cache fields already on these structs. The cache is
not serialized; it's an in-memory, per-process optimization, populated
lazily on first use and rebuilt from scratch (cold) after every
process restart.

**Why this is safe to cache across commits, unlike the realm-exemption
check itself:** `assertTypeIsPublic`'s own logic additionally exempts
`pkgPath == rlm.Path` (a realm may always use its own types, private
or not) — that part genuinely depends on *which* realm is asking, so
it can't be folded into a permanent cache. But "does this type touch
*any* private package at all" does not depend on the caller: package
privacy (`PackageValue.Private`) is set once at package creation and
never mutated afterward (verified — no other write site exists). So a
`false` result is a fact that holds for every realm, forever, and is
safe to cache; `assertTypeIsPublic` keeps doing the exact, realm-aware
walk on top of that whenever the fast path can't rule out a
violation.

**Cycle safety (the actual hard part).** Resolving a node's
`privateDep` may need to walk through a self- or mutually-referential
type (e.g. a linked-list-shaped struct). A naive "cache a node's
result as soon as its own local recursion finishes" scheme is
provably wrong for a mutual cycle: given `A` referencing `B` and `B`
referencing back to `A`, where `A` *also* has an unrelated private-
package field explored *after* the `B` reference, walking `A` first
reaches `B` (and, through it, back to the still-open `A`) before `A`'s
own private field is ever examined — caching `B`'s result at that
point freezes it at "no private dependency," even though `B`
transitively holds a reference to `A`, which does have one. The fix:
track whether the walk crossed any cycle at all (via an `onStack` DFS
guard); if it did, discard every result computed during that walk
instead of committing it to the permanent cache — the type stays
"cold" and gets recomputed (correctly, just not memoized) on the next
call. Acyclic walks commit everything they touch; walks that touch a
cycle anywhere commit nothing.

This rule is deliberately coarse: it discards caching for *every* node
touched by a walk that crossed a cycle anywhere, not just the specific
node(s) at risk of the premature-caching hazard above. A type like
`avl.Node` (`leftNode *Node`, `rightNode *Node` — a linked-structure
self-reference, not the two-distinct-types mutual cycle in the example
above) may not actually need this exclusion in every case, but proving
a narrower rule safe against every adversarial shape (not just the one
counterexample this ADR happened to construct) was judged not worth
the risk for a correctness-critical realm-isolation check. So this
optimization gives no speedup at all for `avl.Node`-shaped self- or
mutually-referential types, regardless of how many times the same type
gets saved — see "Limitations" below.

`assertTypeIsPublic`'s own traversal and `typeHasPrivateDep`'s walker
originally duplicated the same ~10-case type-kind switch
(`FuncType`/`FieldType`/slice·array·pointer/`tupleType`/`MapType`/
`InterfaceType`/`StructType`/`DeclaredType`/…). That's now factored
into one shared `typePkgPathAndChildren(t) (pkgPath string, children
[]Type)` helper both functions call, so a new `Type` kind (or a new
field on an existing one) only needs updating in one place.

## Alternatives considered

- **Eager, preprocess-time precomputation** (the `XXX JAE` suggestion)
  — compute the flag once when each type is declared, instead of
  lazily on first use. Rejected as unnecessary complexity: it would
  need a preprocessor hook reachable from every type constructor
  across every package, and offers no correctness advantage over lazy
  memoization once package-privacy immutability is established —
  "compute once, cache forever" is achieved either way; lazy is
  strictly less invasive.
- **Caching the node's result unconditionally, without cycle
  tracking** — the obvious first attempt, and provably incorrect (see
  `TestTypeHasPrivateDep_MutualCycleDoesNotPoisonPeerCache`, which
  fails against that version and passes against the shipped one).
- **Folding the `pkgPath == rlm.Path` exemption into the permanent
  cache** — would make the cache's meaning realm-dependent, defeating
  the whole point of a cross-commit, cross-realm cache. Kept as two
  layers: a realm-independent fast path (`typeHasPrivateDep`) in front
  of the pre-existing realm-aware exact check (`assertTypeIsPublic`).

## Consequences

- **Perf, for acyclic types**: for a type whose privacy has already
  been resolved once in the current process, `assertTypeIsPublic`
  drops from an O(graph size) walk (~2.1µs, 51 allocations for a
  representative mid-sized nested struct, benchmarked) to two map
  lookups (~25ns, 0 allocations) — see
  `BenchmarkAssertTypeIsPublic_RepeatedCommits` in
  `gnovm/pkg/gnolang/realm_assertpublic_bench_test.go`. The adversarial
  case (every call sees a brand-new type, so the cache never hits) is
  not slower than before — incidentally also faster, since the shared
  traversal helper stopped double-walking through `FieldType` wrappers
  along the way.
- **Limitation — no benefit for self- or mutually-referential types**:
  any type reachable only through a cycle never gets cached (see
  Decision), so this gives **zero speedup for `avl.Node`-shaped types**
  — concretely, `gno.land/p/nt/avl`'s `Node` (`leftNode`/`rightNode
  *Node`), and by extension every realm built on `avl.Tree`, one of
  the most common ordered-map/set patterns in the ecosystem.
  `BenchmarkAssertTypeIsPublic_RepeatedCommits_SelfReferential` in
  `gnovm/pkg/gnolang/realm_assertpublic_bench_test.go` exercises an
  `avl.Node`-shaped type and shows no improvement over repeated calls,
  in contrast to the acyclic benchmark above — making this gap visible
  to anyone re-measuring this optimization later, rather than only
  documented in prose.
- **No gas or consensus impact**: this whole path is unmetered before
  and after this change, so cache warmth (which varies by node uptime
  and is never part of consensus state) cannot affect billed gas.
  Verified end-to-end, across a real node restart, by
  `gno.land/pkg/integration/testdata/typecache_restart_gas.txtar`: the
  same call reports identical `GAS USED` whether the process's type
  cache is warm or was just reset by a restart.
- **A latent trap for future changes, now guarded by a test**: if gas
  metering is ever added to this path (e.g. mirroring how the GC's
  `gcVisitGas` charges per node visited), it must charge a flat,
  cache-independent cost — charging proportional to *actual* work done
  in a given call would make gas a function of node-local, restart-
  reset cache state instead of consensus state, which is exactly the
  kind of change that forks a chain across a validator set with mixed
  uptime. `typecache_restart_gas.txtar` fails immediately if that
  invariant is ever broken.
- Restarting a node cold-starts this cache; there is no persistence
  and none is needed — the type graph itself is reconstructed from the
  package's source/state during normal startup, and the first commit
  after a restart simply repopulates the cache the same way the very
  first commit ever did.

## Verification

```sh
go test ./gnovm/pkg/gnolang/ -run 'TestTypeHasPrivateDep|TestAssertTypeIsPublic' -v
go test ./gno.land/pkg/sdk/vm/ -run Gas
go test ./gno.land/pkg/integration/ -run TestTestdata
go test ./gnovm/pkg/gnolang/ -run Files -test.short
go test ./gnovm/pkg/gnolang/... 2>&1 | grep -v <pre-existing go/types message-wording failures, unrelated>
go test ./gnovm/pkg/gnolang/ -run '^$' -bench 'BenchmarkAssertTypeIsPublic' -benchmem
```

All pass. The full `gnovm/pkg/gnolang` package run has the same
pre-existing `go/types` error-message-wording failures present on
`master` with no changes at all (unrelated to this change — a Go
toolchain/type-checker message-drift issue).
