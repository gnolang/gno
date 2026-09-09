# Bound bptree version discovery to two seeks

## Context

`nodeDB.discoverVersions` finds the first and latest retained versions before a
tree loads. It scanned every root key (`PrefixRoot‖version-BE`) to take the min
and max. `MutableTree.Load` calls it, and the rootmulti immutable store opens a
tree per ABCI query at a height, so every custom query reran the full scan. The
cost is linear in the retained-version count, and the gno.land default prune
strategy keeps 705,600 versions, so the scan grows without bound as the chain
runs and sits on the RPC path under the mutex the query handler shares with
block production.

## Decision

Replace the scan with two backend seeks. Root keys are `PrefixRoot‖version-BE`,
so key order is version order: the first forward key is the smallest version and
the first reverse key is the largest. `edgeRootVersion` opens a forward or
reverse iterator over `[PrefixRoot, PrefixRoot+1)` and returns the first 9-byte
key's version, skipping any non-9-byte key at the edge so a stray key cannot
stop discovery. `discoverVersions` calls it once per direction.

`rootDBKey` is the only writer under `PrefixRoot` and always emits 9 bytes, so
the seek result is identical to the old scan's min/max, at O(log n) per edge
instead of O(retained versions).

## Alternatives considered

- **Cache first/latest on the nodeDB and skip discovery on immutable opens.**
  Larger surface, needs invalidation on prune and commit, and the seek is cheap
  enough that caching earns little.
- **One `db.Has` probe per version in `[first, latest]`.** Still linear in the
  retained range.

## Consequences

- Immutable query-height opens no longer scan all retained roots; discovery is
  two seeks regardless of retention depth.
- Behaviour is unchanged for every reachable input: the discovered first/latest
  match the prior scan, including after pruning opens a gap at the low end
  (`TestDiscoverVersionsSeeksEdgesAfterPruneGap`). The one divergence is a root
  at version 0, which the old scan's `first == 0` sentinel could not tell from
  "unset" and so reported as the second-smallest version; the seek reports 0.
  That case is unreachable (`WorkingVersion` never emits 0 and
  `SetInitialVersion(0)` falls through to version 1), and where it would apply
  the seek is the more faithful answer.
- `AvailableVersions` still scans, by design: it must return the full list.

## Out of scope

The same immutable query-height open could also write live state: when the fast
index stamp was behind the loaded version, `Load` ran `ensureFastIndex`, which
rebuilds and writes through the raw DB even though the open is nominally
read-only. [#6018](https://github.com/gnolang/gno/pull/6018) fixed it with
`MutableTree.LoadReadonly`, which skips `ensureFastIndex` and gates the
snapshot's fast reads on the stamp. This branch keeps that path and only makes
the discovery it still runs bounded.
