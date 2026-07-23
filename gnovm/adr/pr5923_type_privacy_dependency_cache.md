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
`visited map[TypeID]struct{}` per call to dedupe repeats *within* one
commit; there is no memory across commits. A realm whose objects reuse
a handful of types (the common case) re-pays the same O(graph size)
walk, and the same repeated `store.GetPackage` lookups, on every object
saved in every block — pure node-side overhead with no offsetting gas
charge (this whole path — confirmed by inspection — never calls
`ConsumeGas`/`incrCPU`/`Allocator.Allocate`; it's uncharged CPU work,
not a fee-fairness concern, just wasted time).

An `XXX JAE` comment on `assertTypeIsPublic` already named the fix:
precompute an `IsPrivate`-style flag per type, once, instead of
re-deriving it every commit.

### A first design that didn't work, and why

The initial version of this change stored the memo as a `privateDep`
field on the `Type` object itself (`StructType`/`InterfaceType`/
`DeclaredType`), and refused to cache any type reached through a cycle.
Code review (PR #5923) showed, with reproducing tests, that this failed
to help in production for two independent reasons:

1. **The memo never survived a transaction boundary.** Every
   transaction gets a fresh `cacheTypes` map, and `GetTypeSafe` decodes
   or `copyTypeWithRefs`-copies a **fresh** `Type` object per
   transaction. The field set while checking one commit's object lived
   on an object the next commit never sees — so the "cache across
   commits" this optimization exists for never happened. The repeated-
   save path always missed.
2. **Any type with a method was never memoized.** A `DeclaredType`'s
   method carries a receiver of that same type, forming a self-cycle;
   the cycle-discard rule then threw away the whole walk's results.
   Since most real Gno types carry methods, the cache was dead for them
   even within a single transaction.

The verdict computed was always *correct* — it just almost never got
cached. The redesign below fixes both by changing *where* the memo
lives and *what* gets cached.

## Decision

Add `typeHasPrivateDep(store, t) bool` — a fast-path check consulted at
the top of `assertTypeIsPublic` that returns whether *anything*
reachable from `t` belongs to a package with `Private == true`. If
false, `assertTypeIsPublic` returns immediately: nothing left in the
type graph could ever trigger the panic below, for any realm, so the
exact walk is skipped entirely.

**Cache location: store-level, keyed by `TypeID`.** The verdict is
memoized in a new `typePrivacyCache` (`map[TypeID]bool` behind an
`RWMutex`) created once on the root `defaultStore` and shared by
reference into every `BeginTransaction` child — the exact pattern the
process-global `aminoCache` and `stdlibKeyBytes` already use. Keying on
`TypeID` rather than on the `Type` object is what makes the memo survive
the transaction boundary that broke the first design: `TypeID`
identifies a type's full structure, so the fresh object each
transaction reloads maps to the same key and the same cached verdict.
The cache is process-lived (no eviction; the value set is one bool per
distinct `TypeID` ever seen — tiny and bounded), not serialized, and
reset to cold on restart. It is created per root store, so tests don't
cross-contaminate.

**What is safe to cache, unlike the realm-exemption check:**
`assertTypeIsPublic` additionally exempts `pkgPath == rlm.Path` (a realm
may always use its own types, private or not) — that part depends on
*which* realm is asking and can't be cached realm-independently. But
"does this type touch *any* private package at all" does not depend on
the caller: package privacy (`PackageValue.Private`) is set once at
package creation and never mutated (verified — no other write site). So
the verdict is a pure, immutable function of `TypeID` for the life of
the process, safe to cache and safe to share across the concurrent-query
path (a race can only recompute the same value, never disagree; the
mutex is for memory-safety, not value-correctness).

**Cache only the queried root — this dissolves the cycle problem.**
Only the verdict for `t` itself (the type passed in, i.e. the object's
type at the save site) is cached — never the intermediate nodes reached
during the walk. This is the key correctness insight: a root's DFS
visits its entire reachable closure before returning, so *the root's
answer is always correct regardless of cycles*. The fixed-point hazard
that forced the first design to discard cyclic walks only ever affected
*intermediate* nodes, whose frame could close before their
strongly-connected component was fully explored. By never caching
intermediate nodes, there is nothing to poison:

- The walk (`computeTypeHasPrivateDep`) uses a local `visiting`
  map for the single call. A `TypeID` currently on the DFS stack maps
  to `false` — a back-edge contributes nothing to the OR-reachability.
  An intermediate node's value in `visiting` can be an under-
  approximation (false when the node, considered standalone, actually
  reaches a private package only via a cut back-edge), but that value
  is used only to finish computing the root and is discarded when the
  walk ends.
- Because the DFS visits every reachable node exactly once, it visits
  any private node in the closure, which returns `true` and propagates
  up the live stack to the root. So the root is `true` iff any
  reachable package is private — correct in every cycle/diamond shape
  (hand-verified; `TestTypeHasPrivateDep_MutualCycleResolvesCorrectly`
  and `TestTypeHasPrivateDep_SelfReferentialIsCached` pin the tricky
  cases).

This is why the redesign is *simpler* than the first attempt — no
`onStack`/`sawCycle`/`pending` bookkeeping, no per-node cache field —
and why it now caches method-bearing and self-referential types
(e.g. `avl.Node`) like any other, which the first design could not.

`assertTypeIsPublic`'s own traversal and the walk share one
`typePkgPathAndChildren(t) (pkgPath string, children []Type)` helper, so
a new `Type` kind (or a new field on an existing one) only needs
updating in one place.

## Alternatives considered

- **Field on the `Type` object** (the first design) — rejected: does
  not survive the per-transaction reload of `Type` objects, so it never
  caches across commits (the whole point). See "A first design that
  didn't work" above.
- **Per-node caching with cycle discard** (also the first design) —
  rejected: correct but never caches types reached through a cycle,
  which includes every method-bearing type. Root-only caching is both
  simpler and strictly more effective.
- **Eager, preprocess-time precomputation** (the `XXX JAE` suggestion)
  — compute the flag when each type is declared. Rejected as
  unnecessary complexity: it needs a preprocessor hook reachable from
  every type constructor across every package, with no correctness
  advantage over lazy memoization once package-privacy immutability is
  established.
- **Process-global cache** (like `aminoCache`) — rejected in favor of
  per-root-store: `TypeID` does not encode a package's privacy flag, so
  two test binaries (or a test suite) reusing the same path with
  different privacy could cross-contaminate a global cache. Per-root-
  store matches production (one node = one root store, persisting for
  the node's life) while isolating tests. In production the distinction
  is nil — there is one root store.
- **Folding the `pkgPath == rlm.Path` exemption into the cache** —
  would make the cached verdict realm-dependent, defeating a cross-
  realm cache. Kept as two layers: a realm-independent fast path in
  front of the realm-aware exact check.

## Consequences

- **Perf**: for a type whose verdict is already cached (every save
  after the first, per TypeID, per process — now including method-
  bearing and self-referential types, across commits), `assertTypeIsPublic`
  drops to a single cache lookup. Isolated walk-vs-hit cost, from
  `BenchmarkAssertTypeIsPublic_*` in
  `gnovm/pkg/gnolang/realm_assertpublic_bench_test.go` (Apple M4 Pro):

  | Type shape | cold (miss, full walk) | warm (hit) |
  |---|--:|--:|
  | acyclic (20-field nested struct) | ~546 ns, 4 allocs | ~33 ns, 0 allocs |
  | self-referential (`avl.Node`) | ~688 ns, 17 allocs | ~177 ns, 6 allocs |

  These isolate `assertTypeIsPublic` only; they deliberately exclude the
  per-commit cost of materializing the `Type` object (`GetType` / amino
  decode), which this cache does not change. The pre-fix design's
  "warm" column was unreachable in production — every commit paid the
  cold cost because the memo never survived the transaction boundary.
- **No gas or consensus impact**: this whole path is unmetered before
  and after, so cache warmth (which varies by node uptime and is never
  part of consensus state) cannot affect billed gas. Verified end-to-
  end, across a real node restart, by
  `gno.land/pkg/integration/testdata/typecache_restart_gas.txtar`: the
  same call reports identical `GAS USED` whether the type cache is warm
  or was just reset by a restart.
- **A latent trap for future changes, guarded by that test**: if gas
  metering is ever added here (e.g. mirroring the GC's `gcVisitGas`,
  which charges per node visited), it must charge a flat, cache-
  independent cost. Charging proportional to actual work done in a
  given call would make gas a function of node-local, restart-reset
  cache state instead of consensus state — forking a chain across a
  validator set with mixed uptime. `typecache_restart_gas.txtar` fails
  immediately if that invariant breaks.
- **Correctness rests on `TypeID` → privacy being immutable per
  process**: `TypeID` uniquely identifies a type's structure, and a
  package path's privacy never changes once created. A code reviewer
  independently validated this over 5000 random type graphs with random
  private subsets, querying in shuffled rounds so earlier cached answers
  feed later ones: the memoized answer never disagreed with a cache-free
  walk.
- Restarting a node cold-starts this cache; no persistence is needed —
  the type graph is reconstructed during normal startup, and the first
  post-restart commit repopulates the cache exactly as the first commit
  ever did.

## Verification

```sh
go test ./gnovm/pkg/gnolang/ -run 'TestTypeHasPrivateDep|TestAssertTypeIsPublic' -v
go test ./gno.land/pkg/sdk/vm/ -run Gas
go test ./gno.land/pkg/integration/ -run TestTestdata
go test ./gnovm/pkg/gnolang/ -run Files -test.short
go test ./gnovm/pkg/gnolang/ -run '^$' -bench 'BenchmarkAssertTypeIsPublic' -benchmem
```

All pass. The full `gnovm/pkg/gnolang` package run has the same
pre-existing `go/types` error-message-wording failures present on
`master` with no changes at all (unrelated to this change — a Go
toolchain/type-checker message-drift issue).
