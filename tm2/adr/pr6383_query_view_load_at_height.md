# PR6383: Load query views at their height without discovering versions

## Status

Proposed.

## Context

Every state query scans the bptree's version roots before it reads a single
key, so its cost grows with every block the node retains.

The path is shared by all three query entry points: `handleQueryCustom`
(VM queries, `auth/...`), `handleQueryStore` through
`multiStore.QueryImmutable` (`.store/...` queries), and `Simulate`. Each builds
a fresh read-only multistore over the committed snapshot (PR5431) at an
explicit height: BaseApp sets the height to the latest written version when the
client gives none, and `QueryImmutable` refuses a height of 0 or below. For the
main store, rootmulti calls `bptree.Store.LoadVersion(height)`, which calls
`loadImmutableView(height)`. That calls `MutableTree.LoadReadonly`, whose
`nodeDB.discoverVersions` iterates every `PrefixRoot` key to find the first and
latest versions, and then loads the latest root. There is one root key per
retained version, and the default `syncable` strategy retains 705,600 of them.

A view at an explicit height uses none of that work. The latest version
replaces the requested one only when that is 0, and the latest root goes into a working tree the read-only
store never reads. The view itself comes from `getImmutable`, which reads the
requested version's root by key (`GetRoot`) and gates the fast index on its
stamp.

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
prepare budget on a 232-query closure. The laptop pays about 99 ns per
retained version, while onyx-1's per-query cost implies several times that.
Disk reads behind the 8 MB block cache would explain the gap, but that is not
confirmed on a public node.

## Decision

`bptree.Store.loadImmutableView` calls `LoadReadonly` only when `ver == 0`. At
an explicit height it goes straight to `GetImmutableUnregistered(ver)`, which
reads that version's root key and root node and, with the fast index on, the
index stamp. Query views load at their height without discovering versions.

`ver == 0`, the latest version, still loads read-only to find it, and pays the
scan. Only `LoadLatestVersion` on an immutable store passes 0, and no query
path calls it: rootmulti builds query views through `Store.LoadVersion`, which
returns on 0 before loading anything and passes every other height through. The
store layer never calls `loadImmutableView(0)` on the query path.

`Load`, which the live store runs at startup, still scans the root keys once:
it needs the first version, where the prune range starts, as well as the
latest.

Version discovery in `tm2/pkg/bptree` and the database iterators are
unchanged.

## Alternatives considered

**Seek both ends of the root range.** Root keys are `PrefixRoot` followed by
the version in big-endian, so the first key of a forward iterator gives the
first version and the first key of a reverse iterator the latest. An
implementation gives the same `vm/qfile` and `verify-one` times at the RPC as
this change. It rewrites bptree version discovery, needs its own handling of
the first version when a read-only loaded tree saves, and is fast on goleveldb
and pebble only once their iterators are bounded to the requested range:
without the bound, a reverse seek steps over the deleted keys that sort after
the root range, the values of pruned versions among them, until compaction
drops them. That is a larger change to code next to consensus than the query
path needs. Bounding the LSM iterators to their range is an improvement of its
own, independent of this change.

**One shared query view per commit.** Build the read-only multistore once per
committed snapshot and hand it to every query at that height. With the scan
gone, building a view is a few point reads at that height. Sharing would save
those and make tree nodes readable by concurrent queries for the first time, a
new concurrency surface for little gain.

**Cache the scan's result per snapshot.** Still one full scan per commit on the
query path (about 90 ms at 900,000 versions on the laptop above, more on node
hardware), plus a new cache lifetime to reason about, to compute a value the
view does not use.

**A larger pebble block cache.** Makes the scan cheaper, not constant, and
changes every node's memory footprint. It is a separate decision about the
node's database options.

## Consequences

**Query cost depends on neither retained versions nor pruning.** Loading a view
at a height opens no iterator: it reads that height's records by key, so it
walks no root range and steps over no tombstones of pruned versions.

**Startup cost is unchanged.** `Load` scans the root keys once when the live
store loads, as before.

**No consensus change, no coordinated upgrade.** Only read-only query views
change. A view at a height loads the same version and reads the same records
for the query; it skips the scan and the load of a latest root it never reads. Block execution,
commit, pruning, bptree discovery and the database backends are untouched, so a
network mixing both binaries keeps agreeing, and nodes upgrade one at a time in
any order.

**A view at a height reads only that height's records.** A corrupt storage
block elsewhere in the root range, or a corrupt latest root, does not fail a
query at another height. Startup still scans every root key and loads the
latest root.

## Validation

- **Query view load.** `TestQueryViewAtHeightOpensNoIterator`
  (`tm2/pkg/store/bptree`) commits three versions that each change one key,
  then loads a fresh immutable store over the same DB at version 2, as rootmulti
  builds a query view. The load opens no iterator, and the view returns the
  version-2 value. Without the change, the load opens one iterator.
- **Racing commit.** `TestFastIndex_NoStaleRebuildOnRacingQueryLoad` lands a
  commit between the query view's root read and its stamp read; the view still
  returns its version's value, and the fast index still matches the tree.
- **End to end.** On a node with 900,221 retained versions:

  | | baseline | this change |
  |---|---|---|
  | `vm/qfile r/sys/names/gnomod.toml` | 96 to 130 ms | 0.6 to 0.9 ms |
  | gpao `verify-one`, realm importing `r/sys/names` | 15.4 s | 0.57 to 0.70 s |

- **Existing suites.** `tm2/pkg/bptree/...`, `tm2/pkg/store/...`,
  `tm2/pkg/sdk/...`, `tm2/pkg/db/...`, `TestAppHashCrossrealm38`, and the
  integration testdata.
