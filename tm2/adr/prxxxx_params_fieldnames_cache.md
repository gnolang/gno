# Cache param struct field names in the params keeper

## Status

Proposed.

## Context

`ParamsKeeper.GetStruct` and `SetStruct` need the JSON names of a param
struct's fields. They got them from `amino.GetTypeInfo`, which builds a new
codec on every call so that the global codec's `TypeInfo` is never handed out.
`GetStruct` called it twice.

The ante handler reads `auth`, `vm` and fee collector params on every tx, in
CheckTx and DeliverTx alike. Profiling a signed bank send through the gno.land
app's `CheckTx` showed `GetStruct` at 53% of CPU, 33% in `amino.NewCodec`
alone. CheckTx is also what the mempool reruns for every pending tx after each
block, under the mempool lock, while Commit waits.

## Decision

Keep the JSON field names per `reflect.Type` in a package-level `sync.Map`
(`fieldNames` in `amino_helper.go`). The first lookup of a type still goes
through `amino.GetTypeInfo`; later lookups reuse the names.

Only a `[]string` copy is cached, never the `TypeInfo`, so the reason
`GetTypeInfo` builds a fresh codec still holds. Errors are not cached: a type
amino rejects panics on every call, as before.

## Alternatives considered

- Keep a private sealed codec in params and call its `GetTypeInfo`, as the
  `amino.GetTypeInfo` comment suggests. That method takes the codec's write
  lock on every call, so it would serialize param reads on the CheckTx path.
  It also stores a partial `TypeInfo` before panicking on a type it rejects.
- Cache the whole `*amino.TypeInfo`. Same speed, but shares a mutable amino
  structure across callers, which `GetTypeInfo` deliberately avoids.
- Cache decoded param values per block. Removes the remaining JSON decoding,
  but needs invalidation when params change mid-block. Not worth the risk here.

## Consequences

- Same benchmark, same inputs, before and after (`benchstat`, Apple M5):
  - `BenchmarkGetStruct` (params package): 28.6µs → 2.69µs, 414 → 78 allocs.
  - gno.land app `CheckTx` on a signed bank send: 222.8µs → 128.8µs (-42%),
    2,134 → 1,061 allocs. Signature verification is now the largest cost.
- Gas is unchanged: store reads are the same. `sdk/vm -run Gas` and
  `integration -run TestTestdata` pass without changes.
- The cache holds one entry per param struct type, a fixed set of Go types.
