# Skip the decode in params.GetStruct while the stored bytes are unchanged

## Status

Proposed, in PR #6315. Follows #6308.

## Context

`ParamsKeeper.GetStruct` (`tm2/pkg/sdk/params`) assembles a module's param
struct from one store entry per field: `getStructFieldsFromStore` reads each
field's bytes, `decodeStructFields` amino-JSON-decodes each into the struct.
`auth.AccountKeeper.GetParams` does this once a block from `EndBlocker`, for
ten fields with bech32 addresses among them, and the ante handler, `bank` and
the VM read their params the same way.

Profiled on a gno.land mainnet node (v1.5.0, 8 vCPU, 6 validators) syncing
from genesis, the read under `EndBlocker` was 7–9% of all CPU, 8–13% of what
remains once #6308 removes the validator-set replay. About 60% of it was the
decode (`decodeStructFields`, `amino.JSONUnmarshal`, `Address.UnmarshalAmino`)
and about 40% the per-field store reads. The params change only by
governance, so block after block the same bytes were decoded to the same
values.

The keeper is used from several stores at once: the deliver state, the
check state, a transaction's cache wrap that may be reverted, and query
contexts at old heights (queries run in parallel since #6082). Any cache
has to be right for all of them.

## Decision

The keeper remembers, per struct key, the field bytes it last decoded and
the values they gave (`structCache`, `struct_cache.go`). `GetStruct` still
reads the fields from the context's store, with the same gas; when the bytes
are the same as the remembered ones, it copies the remembered values into the
caller's struct and skips the decode; otherwise it decodes as before and
remembers the result.

Keying on the bytes just read, and on nothing else, is what makes the cache
correct everywhere: it never trusts a write-side signal, so a reverted
transaction, a cache wrap, a historical query or the check state each get the
values of the bytes their own store holds. There is no invalidation.

What the caller gets is a deep copy (`copyValue`), for the kinds the param
structs are made of: numbers, strings, bools, and arrays, slices and structs
of those. A struct with a field of any other kind (pointer, map, interface,
function, channel) or with an unexported field is never cached, so it is
decoded every time, exactly as before. Fields absent from the store are left
as the caller had them, as the decode leaves them. The cache is one per
keeper, holds one entry per registered struct key, and is mutex-guarded.

The existing decode writes struct field `i` for amino field `i`; the cache
mirrors that convention so a fill never lands in a different field than the
decode did.

## Alternatives considered

- Invalidate on writes (a counter bumped by every `Set*`, the cached value
  served without reading the store). Removes the reads too, but a keeper-side
  counter cannot tell a reverted transaction's write from a committed one,
  the check state from the deliver state, or a historical query's store from
  the live one. Under-invalidation there means wrong params; rejected.
- Cache in `auth.AccountKeeper` (or in `EndBlocker`) only. Same invalidation
  problem, and it leaves `bank` and the VM reads out.
- Read all fields with one prefix iteration instead of one `Get` per field,
  to cut the remaining ~40%. Changes the store access pattern; worth its own
  change with its own measurement.

## Consequences

- The decode disappears from every `GetStruct` whose bytes are unchanged
  since the last call: on the node profiled, ~60% of the read's cost, an
  estimated 5–8% of a sync from genesis after #6308. `BenchmarkGetStruct`
  (an `auth.Params`-shaped struct, the same ten reads in both arms): cached
  121 µs/op, 200 allocs; decoded 237 µs/op, 417 allocs (medians of 5).
- No change to what is stored, to gas, or to the values any caller sees.
- A struct param type that gains a pointer, map or interface field silently
  loses the cache for that type, never correctness.

## Context the AI agent worked from

Prepared with an AI coding agent: the profiles from the node (pprof with
`-focus` on `GetStruct` and `MustUnmarshal` to attribute the cost), the
keeper, its helpers and tests, the three `Params` structs (`auth`, `bank`,
`vm`) and their callers, `sdk.Context`'s store handling for the cache-wrap
test, and amino's `FieldInfo`. The gain is derived from the profile shares and
the package benchmark, not from a timed resync.
