# Memoize the amino type info of param structs

## Status

Proposed.

## Context

`ParamsKeeper.GetStruct` and `SetStruct` need the field list of the param
struct (each field's JSON name is a store key suffix). They get it from
`amino.GetTypeInfo`, which is:

```go
func GetTypeInfo(rt reflect.Type) (info *TypeInfo, err error) {
	cdc := NewCodec().WithPBBindings().Autoseal()
	return cdc.GetTypeInfo(rt)
}
```

so every call builds a new codec, `registerWellKnownTypes` included, and
throws it away. `GetStruct` calls it twice (`getStructFieldsFromStore`,
`decodeStructFields`), and every transaction reads param structs several
times: the auth ante handler reads the auth params, the VM keeper reads the VM
params, and `FeeCollectorAddress` reads the auth params again.

A CPU profile of a gno.land node running signed `MsgCall` transactions
attributed 12.8% of sampled CPU to `GetStruct`, almost all of it in
`NewCodec`. `params` is the only caller of `amino.GetTypeInfo` in the repo.

## Decision

`tm2/pkg/sdk/params/amino_helper.go` routes its three `amino.GetTypeInfo`
calls through `structTypeInfo`, which memoizes the result in a package-level
`sync.Map` keyed by `reflect.Type`.

Three properties were checked:

- **Concurrency.** `sync.Map` is safe for concurrent use. Two goroutines that
  miss on the same type both compute it; `LoadOrStore` keeps the first and
  both callers get that one pointer. The `*amino.TypeInfo` is never written
  after `GetTypeInfo` returns (the codec that built it is unreachable), and
  the three call sites only read `Fields[i].JSONName`, so sharing it is a
  read-only share. `TestStructTypeInfoConcurrent` runs 32 goroutines doing
  `SetStruct`/`GetStruct` on two types under `-race`; replacing the
  `sync.Map` with a plain map makes it report data races.
- **Growth.** Keys are the Go types passed to `GetStruct`/`SetStruct`, a set
  fixed by the binary (three in gno.land today: auth, bank and VM params). Nothing
  keys on input data. The memo cannot grow past the number of param struct
  types.
- **Staleness.** The result depends only on `rt`. `GetTypeInfo` uses a fresh
  codec with nothing registered, not the global codec or any codec a caller
  can mutate, so no later registration can change what it would return. There
  is nothing to invalidate. `TestStructTypeInfoMatchesAmino` checks the
  memoized result against a fresh `amino.GetTypeInfo`, and that two types of
  identical shape but different JSON tags get distinct entries.

Errors are not memoized. In practice amino panics on unsupported types rather
than returning an error, and the callers panic on an error either way.

## Alternatives considered

- **Cache inside `amino.GetTypeInfo`.** Same effect for the only caller
  today, but it changes a public amino function's contract (fresh vs shared
  `*TypeInfo`) for every future caller. Keeping the memo next to the read-only
  call sites keeps the sharing argument local.
- **Hold a sealed package-level codec in `params`.** Equivalent, but a codec
  is a larger mutable object than a map of finished type infos.
- **Cache per keeper.** Keepers are values copied around (`ForModule`), so a
  per-keeper cache needs a pointer field and still gets rebuilt per keeper;
  the type info is a property of the type, not of the keeper.

## Consequences

- No gas or state change: no gas is charged inside `GetTypeInfo`, and the same
  keys are read and written. Measured: identical gas per transaction and
  identical app hash after 32 blocks, three paired runs.
- `BenchmarkGetStruct`: 182 us to 8.4 us per call, 38 KiB to 1.4 KiB and 336
  to 30 allocations.
- End to end, signed `MsgCall` read-modify-write over pebble with 10,000 avl
  keys, three paired runs: DeliverTx p50 down 4.8 to 8.3%, ante handler self
  time 536 to 151 us per transaction.
- Remaining cost in the same path: `decodeField` (amino JSON per field per
  read) is now the bulk of `GetStruct`. Caching decoded param values would
  need invalidation on `SetStruct` and on transaction rollback, so it is left
  out of this change.
