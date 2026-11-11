# Decoded-object cache across transactions, gas-neutral by construction

## Context

Every transaction starts with an empty object cache (`BeginTransaction`), so
it re-reads and re-decodes every realm object it touches, even when the
previous transaction decoded the very same bytes. On a read-modify-write
workload over `p/nt/avl/v0` (10,000 keys), object loading was 66% of DeliverTx
wall time, amino decode alone 50%, and 56% of the loads were objects an
earlier transaction had already decoded.

Simply keeping objects across transactions is not acceptable: I/O and decode
gas are charged on a miss only, so gas would depend on what a node happens to
have cached (an experiment moved gas per tx from 9.72 M to 3.68 M). gnoland1
already halted once over that class of bug (#4983, #5400). Design discussion:
#6461.

## Decision

Design A of #6461: a process-wide cache of decoded objects keyed by the
SHA-256 of their stored amino bytes, consulted inside `loadObjectSafe` only
after the backend `Get` (I/O gas) and the decode gas charge, replacing only the
`amino.MustUnmarshal` call. On a hit it returns `copyValueWithRefs` of a
pristine decoded copy; on a miss it decodes and inserts a pristine copy taken
before anything touches the object. Everything after (allocator charge,
`SetHash`, `fillPackage`, `fillTypesOfValue`) is unchanged and runs on the
returned copy. This is the pattern `GetTypeSafe` already uses for types
(`aminoCache`, #5133).

Rule that makes it safe: the state machine cannot tell a hit from a miss. Then
cache contents may differ between nodes, restarts, CheckTx and queries.

| hazard | how it is closed | test |
|---|---|---|
| gas depends on cache state | same `Get`, same gas calls in the same order; only CPU work changes | `TestObjectDecodeCacheGasDeterminism`: cold, warming and warm runs give identical gas and result for every tx and identical app hash at every block |
| failed tx leaves dirty objects cached | a tx only ever holds a copy; entries come from a pristine copy taken at decode | same test (8 panicking and 8 out-of-gas writes, then every key read back against a model); `TestObjectDecodeCacheCopyIsolation` |
| a realm mutates an object other txs cached | content-keyed: new bytes, new key, the old entry is never hit again | same test (rewrites then reads); `TestObjectDecodeCacheContentKeyed` |
| copy differs from a fresh decode | `copyValueWithRefs` now preserves nil slices (`sliceLike`); with `-tags debugAssert` every hit is compared to a fresh decode (re-encoding, shallow size, ref size, ObjectInfo, `reflect.DeepEqual`) and panics on divergence | `go test -tags debugAssert` |
| unbounded memory | ristretto, `MaxCost` 64 MiB, each entry costed at the larger of its encoded size and the allocator's decoded-size estimate, TinyLFU admission and sampled LFU eviction; eviction is unobservable so it may differ by node | `TestObjectDecodeCacheBounded` |
| concurrency (CheckTx, queries) | entries are immutable and only copied out; ristretto is concurrent | n/a |

`sliceLike` changes `copyValueWithRefs`/`copyTypeWithRefs` to return nil
instead of an empty slice for nil input. Amino encodes both the same way, so
persisted bytes and hashes do not change; it only makes in-memory copies match
a decode (for example `FuncType.IsZero` distinguishes nil from empty).

## Alternatives considered

- Keep live objects across transactions (oid-keyed): fastest (2.95x measured)
  but moves gas and leaks failed-tx mutations. Rejected.
- Oid-keyed committed-state cache that skips the backend read and emulates its
  gas (design B in #6461): also saves the fetch, but its correctness depends on
  reproducing tm2 cache-store gas rules in a second place. Left for later.
- Price cache hits (design C): a gas-schedule change, an economics decision.

## Consequences

- CPU-only win, roughly 1.3x on the in-tree benchmark; the backend read, fill
  and type loads are still paid. Combines with a faster decode path.
- Up to 64 MiB of cost held per process; real decoded memory is above the
  allocator estimate (Go headers, type objects), by how much is not measured.
- `copyValueWithRefs` keeps nil-ness of empty slices.
