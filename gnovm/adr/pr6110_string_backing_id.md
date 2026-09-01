# VM-owned string backing identity: mint IDs in StringValue

> **Status: experimental.** Standalone alternative to #4885 — same
> accounting semantics (charge each live backing once per GC cycle;
> slices share), different identity mechanism. One of the two should
> merge, not both.

## Context

#4885 fixes string accounting (loaded strings uncharged, shared backings
double-counted, slices overcharged) by tracking backing **addresses** in
the allocator: a treap of `[start, end)` extents, containment lookup, and
— after two determinism hazards surfaced in review — clone-on-overlap and
a pin that holds each tracked backing alive. Both hazards stem from one
fact: address identity is rented from the Go runtime, which shares and
recycles memory on its own schedule. See #4885 (and its ADR,
`pr4885_string_alloc_reuse.md` on that branch) for that design and its
full alternatives catalog.

## Decision

Move identity into the value. `StringValue` becomes:

```go
type StringValue struct {
	Str    string
	ID     uint64 // mint-event serial; 0 = untracked
	Extent int64  // full backing byte length
}
```

- `NewString` assigns a fresh `ID` (global atomic counter) and
  `Extent = len(s)`. Value copies and `GetSlice` inherit both, so every
  value over one backing shares the ID; `GetSlice` still charges
  header-only at creation.
- The GC visitor charges `allocStringByte * Extent` once per distinct
  nonzero ID per run, from a **visitor-local** `map[uint64]struct{}`
  (pre-sized by the allocator's mint count). Dedup never writes into the
  value — deliberately unlike object `LastGCCycle` stamping, which is
  unsafe on values shared across machines (see the `.uverse` carve-out in
  `GCVisitorFn`); cached literals are exactly such shared values.
- `ID == 0` (VM panic text via `typedString`/`typedRuntimeError`, uverse
  init) recounts as header only: a deterministic, bounded undercount —
  never a misattribution, since accounting never consults addresses.
- The numeric ID never reaches consensus: amino/pb3 persist `Str` alone
  (`MarshalAmino` repr keeps the wire format byte-identical), and loads
  re-mint through `fillTypesOfValue`. Only the partition "which values
  came from the same mint" matters, and that is a pure function of VM
  execution — identical on every node regardless of counter values.
- The allocator keeps **no string state**: the treap
  (`string_ranges.go`), `trackString` clone-on-overlap, the pin,
  `CleanupTrackedStrings`, and the between-messages `clearStringTracking`
  are deleted. `Fork`/`Reset`/`ClearObjectCache` need no string handling.

Both #4885 hazard classes are unrepresentable here: toolchain backing
sharing is irrelevant (two mints get two IDs even if Go shares the
bytes), and address recycling cannot misattribute (there are no
addresses).

### Struct-equality hazard

`StringValue` was an alias with value equality; as a struct, `==`
compares ID, so equal strings from different mints differ. String
comparison paths were audited: `isEql` and `ComputeMapKey`/`MapKeyBytes`
go through `GetString()` (safe); the one direct-equality site —
map-literal duplicate-key detection in `preprocess.go` (`kset`) — now
normalizes string keys by content. New direct `==` on `TypedValue`/`.V`
holding strings must not be introduced; compare `Str`.

### Behavior deltas vs #4885

- Literal backings now recount post-GC (`alloc_0/1/7` +26/+16/+11 B).
  Under #4885, literals minted by the *preprocess* allocator were tracked
  in that allocator's ranges, so the runtime GC undercounted them as
  headers; here identity travels in the value, so live literal bytes are
  counted once wherever the mint happened. `alloc_13`/`alloc_13a`
  (shared-backing dedup, slice-outlives-source) are unchanged.
- `.grealm.Address()` returns the handle field's own StringValue (shares
  its mint) instead of a raw untracked value.

## Alternatives considered

The full catalog lives in #4885's ADR; deltas here:

1. **#4885's address ranges + pin** — works, and confines the change to
   the allocator (no representation migration); carries the treap, the
   clone-on-overlap, the pin (dead backings held ≤1 cycle), and
   per-message cleanup, all to compensate for rented identity. Measured
   slower on both hot paths (below).
2. **`StringValue{*ArrayValue}`** — the backing as a real VM Object;
   most faithful to Go's `{ptr, len}` representation and reuses object
   machinery, but adds a heap object + pointer chase per string on the
   VM's hottest path, a deeper amino change, and GC dedup by stamping
   `LastGCCycle` onto backings shared across machines (cached literals)
   reintroduces the `.uverse`-class nondeterminism this design avoids by
   keeping dedup visitor-local.
3. **Registry (ID → extent) in the allocator** instead of carrying
   `Extent` in the value — reintroduces allocator state, pruning, and
   cross-message lifecycle for no benefit.

## Consequences

- `StringValue` grows 16 → 32 bytes and is a representation change:
  every raw `StringValue(...)` conversion site was migrated (~21 in
  production code); `pb3_gen.go` was hand-adjusted to the repr form the
  generator emits for `MarshalAmino` types (regenerate to verify).
- Persisted format unchanged (plain string); no store migration.
- CPU (median of 3, matched pre-boxed harnesses, vs #4885's pin):
  GC string pass 1k/10k/100k live: 31.5 µs/448 µs/6.0 ms →
  17.3 µs/182 µs/2.2 ms; `NewString` faster at every size (one atomic
  add, no treap insert, no clone).
- Three alloc filetest goldens updated (+11…26 B, literals); gas txtars
  unchanged; full gnolang, sdk/vm, integration, and examples suites pass.
