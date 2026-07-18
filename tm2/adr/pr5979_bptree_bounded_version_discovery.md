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

The same immutable query-height open can also write live state: when the fast
index stamp is behind the loaded version, `Load` runs `ensureFastIndex`, which
rebuilds and writes through the raw DB even though the open is nominally
read-only. It is latent (a current stamp makes the rebuild a no-op) and does not
affect the app hash (the fast index is outside the Merkle commitment), but a
correct fix must gate the immutable view's fast-read on the stamp so skipping
the rebuild cannot serve a stale value. That is a change to the fast-index
read-trust contract and is left as a separate follow-up.
