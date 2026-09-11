# VM-owned string backing identity: a shared backing pointer in StringValue

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
type stringBacking struct {
	Extent int64 // full backing byte length
}

type StringValue struct {
	Str string
	B   *stringBacking // one per NewString mint; nil = untracked
}
```

- `NewString` allocates a fresh `stringBacking{Extent: len(s)}`. Value
  copies and `GetSlice` copy the pointer, so every value over one backing
  shares it; `GetSlice` still charges header-only at creation.
- Identity is the pointer itself. Go's GC cannot free or reuse a
  `stringBacking` while any `StringValue` still points at it, so two live
  mints can never compare equal: uniqueness holds by construction, with no
  counter, no wrap-around and no lifetime argument.
- The GC visitor charges `allocStringByte * B.Extent` once per distinct
  backing per run, from a **visitor-local** `map[*stringBacking]struct{}`.
  Dedup never writes into the value — deliberately unlike object
  `LastGCCycle` stamping, which is unsafe on values shared across machines
  (see the `.uverse` carve-out in `GCVisitorFn`); cached literals are
  exactly such shared values.
- `B == nil` (VM panic text via `typedString`/`typedRuntimeError`, uverse
  init) recounts as header only: a deterministic, bounded undercount —
  never a misattribution, since accounting never consults addresses.
- The pointer never reaches consensus: amino/pb3 persist `Str` alone
  (`MarshalAmino` repr keeps the wire format byte-identical), and loads
  re-mint through `fillTypesOfValue`. Only the partition "which values
  came from the same mint" matters, and that is a pure function of VM
  execution — identical on every node regardless of pointer values.
- **Loads charge strings atomically.** `fillTypesOfValue` re-mints each
  loaded string's identity with `mintString` and allocates nothing; the
  bytes are summed by `internalStringSize` (a walk that mirrors the fill)
  and folded into `loadObjectSafe`'s single `Allocate(ss + rs + strs)`,
  which runs before the object enters `cacheObjects`. An `Allocate` after
  that insert can trigger GC while the object is cached but not yet
  reachable; `GarbageCollectObjectCache` then evicts it and the next
  `GetObject` for the same ObjectID decodes a second copy, breaking
  one-object-per-ID. `gc_load_evicts_object.txtar` reproduces the
  eviction on chain and pins the fix.
- The allocator keeps **no string state**: the treap
  (`string_ranges.go`), `trackString` clone-on-overlap, the pin,
  `CleanupTrackedStrings`, and the between-messages `clearStringTracking`
  are deleted. `Fork`/`Reset`/`ClearObjectCache` need no string handling.

Both #4885 hazard classes are unrepresentable for *identity*: toolchain
backing sharing is irrelevant (two mints get two backings even if Go
shares the bytes), and address recycling cannot misattribute (there are
no addresses). `Extent` is still `len(s)` at mint, not Go's retained
array: `NewString` does not copy, so `sub + ""` (Go returns the operand)
mints `Extent == len(sub)` while Go retains the whole source. Same as
master, and a bounded undercount, not a misattribution.

### Struct-equality hazard

`StringValue` was an alias with value equality; as a struct, `==`
compares the backing pointer, so equal strings from different mints
differ. String comparison paths were audited: `isEql` and
`ComputeMapKey`/`MapKeyBytes` go through `GetString()` (safe); the one
direct-equality site — map-literal duplicate-key detection in
`preprocess.go` (`kset`) — now keys by `ComputeMapKey`, the runtime map's
own notion of key equality. New direct `==` on `TypedValue`/`.V` holding
strings must not be introduced; compare `Str`.

### Behavior deltas vs #4885

- Literal backings now recount post-GC. Under #4885, literals minted by
  the *preprocess* allocator were tracked in that allocator's ranges, so
  the runtime GC undercounted them as headers; here identity travels in
  the value, so live literal bytes are counted once wherever the mint
  happened. `alloc_0.gno` is the one pre-existing golden that moves
  (+724 B, which includes two new vars added to the test); `alloc_13*`
  and `alloc_14*` are new.
- **Charges that go up, not just down.** Two consequences of faithful
  retention accounting that realms near the alloc cap can feel: a slice
  `s[i:j]` now keeps the source's full `Extent` in the recount even after
  the source dies (a 32-byte slice of a 1 MB string recounts as
  header + 1 MB, not header + 32 B; `alloc_13a`/`alloc_13b` pin this), and
  a cold load now charges header + len for every string in the object,
  where master charged nothing. Both match what Go actually retains.
- `.grealm.Address()`/`.PkgPath()` and `address.String()` return the
  receiver's own StringValue (sharing its mint) instead of an untracked
  copy; `.grealm.String()`/`.Subpath()` and `.runtimeError.Error()` mint
  their fresh text through `newTypedString`, the tracked counterpart of
  `typedString`.

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
3. **Registry (identity → extent) in the allocator** instead of carrying
   the backing in the value — reintroduces allocator state, pruning, and
   cross-message lifecycle for no benefit.
4. **`ID uint64` from a process-global atomic counter, plus `Extent` in
   the value** — the first shape of this PR. Cheaper per mint (one atomic
   add, no allocation: 1.7 ns vs 5.1 ns to re-mint an existing string,
   12.9 vs 17.4 ns on a 32-byte concat+mint, 2.63 vs 2.87 ms for a 100k
   live-string GC dedup pass) but correct only by argument: uniqueness
   rests on "no live string outlives 2^64 mints", which is true by physics
   (centuries at a mint per nanosecond) yet is an assumption, not an
   invariant. A panic on wrap would turn it into a checked invariant but
   halts every tx on that node for the rest of its uptime; a silent wrap
   risks a single-node undercount and app-hash divergence. Neither
   failure is reachable in practice, but a review of consensus-adjacent
   accounting should not need a lifetime argument. The pointer removes
   the argument for a per-mint cost well below the surrounding VM op, and
   shrinks `StringValue` from 32 to 24 bytes.

## Consequences

- `StringValue` grows 16 → 24 bytes and is a representation change:
  every raw `StringValue(...)` conversion site was migrated (~21 in
  production code); `pb3_gen.go` is regenerated by `misc/genproto2` (the
  repr form for `MarshalAmino` types; the backing pointer is not on the
  wire).
- Persisted format unchanged (plain string); no store migration.
- CPU (median of 3, matched pre-boxed harnesses, vs #4885's pin):
  GC string pass 1k/10k/100k live: 31.5 µs/448 µs/6.0 ms →
  17.3 µs/182 µs/2.2 ms (measured with the counter shape; the pointer
  shape adds ~9% to the dedup pass and one 8-byte allocation per
  `NewString`, see alternative 4); `NewString` still cheaper than #4885's
  treap insert and clone at every size.
- Goldens: `alloc_0.gno` (+724 B, see above) and six gas txtars move
  (`compute_map_key_restore_gas`, `gc`, `gnokey_gasfee`, `simulate_gas`,
  `stdlib_ibc_crypto_determinism`, `stdlib_restart_compare`), from string
  charges on load and from folding those charges into one `Allocate`
  call per object (allocation gas is a per-call table). Full gnolang,
  sdk/vm, integration, and examples suites pass.
