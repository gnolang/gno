# PRxxxx: Version discovery by seek, not scan

## Status

Proposed.

## Context

Every state query pays a full scan of the bptree's version roots before it
reads a single key, so its cost grows with every block the node retains.

The path is shared by all three query entry points: `handleQueryCustom`
(VM queries, `auth/...`), `handleQueryStore` through
`multiStore.QueryImmutable` (`.store/...` queries), and `Simulate`. Each calls
`MultiImmutableCacheWrapWithVersion`, which builds a fresh read-only multistore
over the committed snapshot (PR5431). For the main store that means
`bptree.Store.loadImmutableView`, then `MutableTree.LoadReadonly`, then
`nodeDB.discoverVersions`, which iterates every `PrefixRoot` key to find the
first and latest versions. There is one root key per retained version, and the
default `syncable` strategy retains 705,600 of them. The comment on
`loadVersionDiscovered` already records that this scan sits on the per-query
path.

The cost is linear in retained versions. On a local gnoland node (pebble with
the options the node opens it with, synthetic root records under the main
store, an 18-core laptop):

| retained versions | state query (`.store/main/key`, `auth/accounts`, `vm/qfile`, `vm/qpkgmeta_json`) |
|---|---|
| ~30 | 0.03 to 0.14 ms |
| 100,000 | 8.3 ms |
| 500,000 | 49 ms |
| 900,000 | 89.5 ms |

`health`, `abci_info` and `.app/version`, which take the same RPC path but
build no store view, stay at about 0 ms throughout, and the query kind and
package size make no difference. A CPU profile of the node at 500,000 versions
under query load puts 99.6% of `BaseApp.Query` in `discoverVersions`, almost
all of it in `pread` reading root-key blocks: the node opens pebble with its
default 8 MB block cache, which cannot hold the root range.

On the public RPCs the same query costs about 280 ms on onyx-1 (~490,000
versions) and about 450 ms on gnoland-1 (~897,000 versions, on a release that
scanned twice per query). Every client pays it per request: gnoweb renders,
gnokey queries, indexers, and the package approver (gpao), which fetches an
import closure one `vm/qfile` per file and on onyx-1 overruns its one-minute
prepare budget on a 232-query closure.

## Decision

`discoverVersions` reads the two ends of the root-key range instead of all of
it. Root keys are `PrefixRoot` followed by the version in big-endian, so key
order is version order:

- `firstVersion` is the first root key of a forward iterator over
  `[PrefixRoot, PrefixRoot+1)`;
- `latestVersion` is the first root key of a reverse iterator over the same
  range.

Both seeks keep the scan's exact semantics:

- a key under the prefix whose length is not 9 bytes is skipped, as the scan
  skips it;
- a key encoding version 0, or a version with the top bit set, is not a root
  version and is skipped. No node path writes either (the node only sets an
  initial height above 1); through the bptree API, a `uint64` initial version
  above `math.MaxInt64` would write a top-bit root, which the scan also
  mishandled. The seek results equal the scan's on every key set whose root
  keys lie in `[1, math.MaxInt64]`;
- an iterator error is returned, as the scan returns it.

`LoadReadonly`, which every query view goes through, needs only the latest
version and does only the reverse seek. Pruning deletes the lowest versions, so
the top of the range is live and the seek returns at once whatever the pruning
state. A read-only tree leaves `firstVersion` unset; only `SaveVersion` and
`PruneVersionsTo` read it. `Load` and `LoadVersion` keep both seeks, and `Load`
no longer goes through `LoadReadonly`: the live store loads through `Load`, and
a restart without `firstVersion` would stop pruning every version retained
before it. On pebble their forward seek steps over uncompacted tombstones of
pruned roots (about 7 ms per 100,000 at 500,000 roots), paid when a node
starts. Pruning reads the cached `firstVersion` and is untouched.

The two seeks of `Load` and `LoadVersion` are two reads, so on backends without
snapshots they are not taken at one point in time. This is harmless: they run
when the live store loads, with no concurrent writer.

A guard ships with the change, so the scan cannot come back unnoticed: a test
that discovery over 10,000 roots reads a bounded number of root keys, and a
benchmark of `LoadReadonly` across root counts, including a pruned case, in
`tm2/pkg/bptree/nodedb_discover_bench_test.go`, with automatic compactions off
so tombstones outlive the run.

## Alternatives considered

**One shared query view per commit.** Build the read-only multistore once per
committed snapshot and hand it to every query at that height. With the seek in
place, building a view is about 5% of a query's remaining allocations; the rest
is HTTP handling, GC, and the VM decoding package blobs. Sharing would remove
that 5% and make tree nodes readable by concurrent queries for the first time,
a new concurrency surface for a gain too small to measure at the RPC. Not
pursued.

**Two seeks on every caller, query views included.** Simpler, but a pruning
node's queries would step over uncompacted tombstones at the low end until
compaction drops them.

**Pass the live tree's version bounds into each view.** Correct, but it needs
new plumbing through the store options or the `CommitStore` interface to
replace a computation two seeks already make cheap, and it leaves `Load` at
startup scanning.

**Cache the scan's result per snapshot.** Still one full scan per commit on the
query path (about 90 ms at 900,000 versions on the laptop above, more on node
hardware), plus a new cache lifetime to reason about.

**A larger pebble block cache.** Makes the scan cheaper, not constant, and
changes every node's memory footprint. It is a separate decision about the
node's database options.

## Consequences

**State-query cost does not depend on retained versions.** On the node and
data above, with the seek:

| retained versions | baseline | seek |
|---|---|---|
| 100,000 | 8.3 ms | 0.03 to 0.17 ms |
| 500,000 | 49 ms | 0.03 to 0.19 ms |
| 900,000 | 89.5 ms | 0.03 to 0.21 ms |

At 900,000 versions, gpao's `verify-one` on a package importing
`gno.land/r/sys/names` finishes in 0.65 s instead of 14.98 s, and 16 parallel clients get 13,700
queries per second at 1 ms instead of 64 at 253 ms.

**No consensus change, no coordinated upgrade.** Discovery returns the same
first and latest versions. They select the version to load and the start of the
prune range; neither feeds the app hash, gas, or stored data. A network mixing
both binaries keeps agreeing, and nodes upgrade one at a time in any order.

**Query cost depends on neither retained versions nor pruning.** Startup pays
the forward seek once.

**The public-node figures are not reproduced locally.** The laptop pays about
99 ns per retained version, while onyx-1's per-query cost implies several times
that. Disk reads behind the 8 MB block cache would explain the gap, but that is
not confirmed on a public node. The seek reads no root range at all, so the
height-dependent part goes regardless of hardware; the final per-query figure
on a public node is known only once a node runs the change.

## Validation

- **Equivalence.** `TestDiscoverVersionsMatchesScan` and
  `TestDiscoverVersionsKnownLayouts` compare the seek against the previous
  scan, kept as a reference implementation, over many root-key layouts: empty,
  single, contiguous, pruned prefix, a range starting at a hardfork height,
  gaps, random sets, the maximum version, version 0, and keys of other lengths
  at both ends of the range. They run on memdb and on the production
  composition (pebble behind a `PrefixDB` and a `SnapshotDB`, with base-store
  keys sharing the prefix). `TestDiscoverVersionsSkipsNonRootVersions` covers
  version 0 and top-bit keys, and `TestDiscoverVersionsPropagatesIteratorError`
  covers iterator errors at open and from `Error()`.
- **Guard.** `TestDiscoverVersionsReadsOnlyTheEnds`: discovery over 10,000
  roots steps the iterator at most twice. `BenchmarkLoadReadonlyVersionRoots`
  covers 1 to 500,000 roots, and 500,000 with 400,000 pruned.
- **Load paths.** `TestLoadReadonlyReadsOnlyTheLatestRoot`: a query view opens
  no forward iterator. `TestLoadKeepsPruningOlderVersions`: after `Load`,
  pruning still removes the versions retained before the restart.
- **Existing suites.** `tm2/pkg/bptree/...`, `tm2/pkg/store/...`,
  `tm2/pkg/sdk/...`, `TestAppHashCrossrealm38`, the consensus handshake and
  replay tests, `gno.land/pkg/gnoland/...`, and the integration testdata.
- **Mixed network.** A validator on the previous binary produced blocks with
  state-changing transactions while a full node on the new binary followed it
  with the `everything` prune strategy, exercising `firstVersion`, and was
  restarted mid-run, exercising `Load`. Across 204 heights, 21 with
  transactions, both nodes committed the same app hash at every height.
