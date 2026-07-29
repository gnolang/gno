# ComputeMapKey: omit the TypeID prefix for concrete map key types

## Status

Implemented on branch `claude/using-superpowers-eebc09`. Written against
master `d14a03770`.

This is a gas *reduction*: it removes metered work outright, so charged gas
drops immediately without a recalibration pass through `cmd/calibrate`.

## Context

**Measured after implementation, this ADR's own premise turned out to be the
smaller of two wins, and the framing below is ordered accordingly.** The
prefix omission proposed here saves ~6 of the 98 gas charged per access on
the `map[address]T` shape that motivated it — the prefix is ~29% of the bytes
hashed, but only ~6% of the `ComputeMapKey` charge and **0.48% of program
gas** of the `compute_map_key_concrete_key` golden's total (see below),
because the 80-unit `OpCPUComputeMapKey`
per-call constant (`machine.go:1661`) dominates at realistic key sizes. While
implementing it, a second and much larger issue surfaced: every map *write*
was computing `ComputeMapKey` on its key **twice** — once by the caller to
look up a possibly-preexisting key, again inside `GetPointerForKey`.
Deduplicating that costs nothing to fix once it's noticed, and it dwarfs the
prefix removal: **-36.4%** on `compute_map_key_big_bytes`
(36827018 → 23405158 gas) and **-35.2%** on `compute_map_key_big_struct`
(186401382 → 120708884 gas), against the prefix's own **-0.48%** on
`compute_map_key_concrete_key` (125449 → 124849 gas). See
[Measured Results](#measured-results) for the full table. An ADR that oversells
its own headline is worse than useless to the next reader, so: the dominant
result in this document is the incidental one, not the one in the title.

`TypedValue.ComputeMapKey` (`gnovm/pkg/gnolang/values.go:1815`) serializes a
map key into a `MapKey` string used as the Go-level index into
`MapValue.vmap`. Unless `omitType` is set, it prepends the key's full
`TypeID` plus a `':'` separator:

```go
if !omitType {
    bz = append(bz, tv.T.TypeID().Bytes()...)
    bz = append(bz, ':') // type/value separator
}
```

Every byte appended to `bz` is charged via `OpCPUSlopeComputeMapKeyByte`
(`values.go:1841-1845`).

The prefix exists to discriminate keys whose *dynamic* type varies — for a
`map[any]int`, `int(1)` and `int64(1)` both serialize to the same eight
value bytes and would otherwise collide. When the map's *static* key type is
concrete, every key in the map necessarily carries the same `tv.T`, so the
prefix is identical for every key: it discriminates nothing and is charged
on every get, set and delete.

The recursion already knows this. Array elements and struct fields pass
`omitTypes` derived from the element/field type
(`values.go:1891`, `values.go:1918`):

```go
omitTypes := bt.Elem().Kind() != InterfaceKind
```

Only the top-level call never does. All three entry points hardcode `false`:
`GetPointerForKey` (`values.go:1011`), `GetValueForKey` (`values.go:1039`),
`DeleteForKey` (`values.go:1056`).

`map[address]T` and `map[NamedType]T` are ubiquitous in gno.land realms, and
their key types are `DeclaredType`s, whose `TypeID()` (`types.go:1931`) is
the most expensive implementation of that method — it is what produces the
discarded bytes.

## Diagnosis: why a flag on `MapValue` does not work

The natural first design is to store a `keyTypeIsConcrete` flag on
`MapValue`, computed once at construction and when types are filled on load.

**`MapValue` has no access to its own static type on the load path.**
The struct carries only ownership and ordering state (`values.go:906`):

```go
type MapValue struct {
    ObjectInfo
    List *MapList

    vmap map[MapKey]*MapListItem // nil if uninitialized
}
```

`vmap` is rebuilt from `List` on load, inside the `*MapValue` case of
`fillTypesOfValue(store Store, val Value)` (`realm.go:1934-1948`):

```go
case *MapValue:
    cv.vmap = make(map[MapKey]*MapListItem, cv.List.Size)
    for cur := cv.List.Head; cur != nil; cur = cur.Next {
        // ...
        mk, isNaN := cur.Key.ComputeMapKey(nil, store, false)
```

That function receives a bare `Value`. A `MapValue` reached through
`store.GetObject` — the normal path for a map held in realm state — has no
`*MapType` anywhere in scope. The static key type is simply not available
there.

This is not a cosmetic problem. `omitType` must be **identical** between the
`vmap` build and every subsequent lookup; if the load path guesses `false`
while lookups pass `true`, every key silently misses and map reads return
zero values. Deriving the flag from the stored keys' own dynamic types cannot
work either: an `any`-keyed map whose current entries happen to all be `int`
is indistinguishable from an `int`-keyed map, and mislabelling it as concrete
reintroduces the `int`/`int64` collision on the next insert.

Two further constraints bound any fix in this area:

**The per-byte charge on the prefix is deliberate.** `machine.go:1662-1669`
ties `OpCPUSlopeComputeMapKeyByte` to GHSA-m7rp-96x5-hvpx, explicitly naming
"a type with a long TypeID" as the DoS vector it closes. So the fix must stop
*appending* the prefix, not merely stop charging for it — exempting those
bytes from the meter while still copying them would partially revert a
security fix.

**`MapValue`'s size is a hardcoded gas constant.** `alloc.go:83` declares
`_allocMapValue = 168 // unsafe.Sizeof(MapValue{})`, asserted at runtime by
`check("_allocMapValue", _allocMapValue, unsafe.Sizeof(MapValue{}))`
(`alloc.go:157`). Adding a `bool` field pushes `Sizeof` to 176, which breaks
that assertion and raises allocation gas for **every map in every realm** —
cancelling part of the win this change exists to deliver.

## Decision

Omit the top-level TypeID prefix when the map's static key type is not an
interface, deriving the predicate at each call site and building `vmap`
lazily on first keyed access.

### 1. `ComputeMapKey` is unchanged

The `omitType` parameter, the `nilStr` early return, the `isComparable`
gate and the per-byte charge all stay exactly as they are. Only the
arguments passed to it change. The GHSA-m7rp-96x5-hvpx metering remains
intact: bytes that are appended are still charged, and the prefix bytes are
no longer appended at all.

### 2. The predicate

```go
omitKeyType := baseOf(mt.Key).Kind() != InterfaceKind
```

`baseOf` unwraps a `DeclaredType` around the key (`type M map[any]int`), and
`mt.Key` — **not** `mt.Elem()` — selects the key type.

This matters: the stale TODO at `op_expressions.go:582` reads

```go
// omitType := baseOf(mt).Elem().Kind() != InterfaceKind
```

and `(*MapType).Elem()` returns `mt.Value` (`types.go:1380`). Uncommented as
written it would have keyed the decision off the map's **value** type:
silently wrong for `map[any]int` (prefix dropped, `int(1)` collides with
`int64(1)`) and needlessly pessimal for `map[address]any` (prefix retained
for no reason). The TODO is replaced, not revived.

**RESOLVED: can `omitKeyType` ever reach `ComputeMapKey` true with `tv.T ==
nil`?** `ComputeMapKey` carries a `debug`-build assertion for exactly this
(`values.go:1874-1880`: `omitType` with `tv.T == nil` is "should not
happen"). This was an open question when the ADR was written; it is not
reachable, and not for the reason the original filetest plan assumed.

`gnovm/tests/files/map53.gno`'s `ptrs[nil]` case does **not** establish it —
tracing it shows why. `ptrs` is `map[*node]string`, a concrete key type, so
`mapKeyOmitType` returns `true` for every access against it. The key value is
a nil `*node`, but its `tv.T` is still the `*PointerType` for `*node`, not
nil — Gno's nil is a typed nil, and `iv.T` is unaffected by `iv.V == nil`. So
this case never reaches the `tv.T == nil` branch at all. Worse: even if it
somehow did, the assertion's premise — "if `tv.T` were nil, would probing
with `omitType=true` diverge from probing with `omitType=false`?" — collapses
identically either way, because `tv.T == nil` returns `nilStr` in both cases
(`values.go:1874-1880`, the special case runs before `omitType` is even
consulted). A test built on this path would pass whether the assertion held
or not, telling us nothing.

The real reason is structural, not test-shaped: `defaultTypedValue`
(`values.go:3110-3115`) returns a `TypedValue` with `T == nil` **only** for
`*InterfaceType` — every other branch (array, struct, slice, map, ...) sets
`T: t`. A nil `T` can therefore only occupy an interface-typed slot. And
`mapKeyOmitType` (like the nested `omitTypes` rule it composes with) returns
`false` — i.e. keep the prefix — precisely for interface-kinded slots. The
one case that can produce a nil `T` is exactly the case the predicate always
routes to `omitType=false`. The panic branch is dead code by construction,
not by having been tested into confidence.

This can only be established by argument today, not by running the
assertion: `go test -tags debug ./gnovm/pkg/gnolang/` panics in package init
on pristine master (`sealUverseTypes` → `(*InterfaceType).TypeID`: "generic
type has no TypeID", `uverse.go:1966`) — a pre-existing, unrelated bug, so
the `debug` build tag cannot currently be used to exercise this assertion at
all, confirmed on both `master` and this branch. Do not delete the assertion
on the strength of this argument; it is cheap, correct, and guards an
invariant that could silently break under a future change to
`defaultTypedValue` or `mapKeyOmitType`.

### 3. No new field on `MapValue`

`omitKeyType` is recomputed at each call site — one `Kind()` comparison,
cheaper than the `TypeID()` call it replaces — and `vmap == nil` serves as
the "not yet built" sentinel. `MapValue`'s layout is untouched, so
`_allocMapValue` stays at 168 and allocation gas is unaffected.

The sentinel does not collide with the existing "nil if uninitialized"
meaning. An uninitialized *Gno* map is represented by `tv.V == nil` and is
checked before any accessor runs (`op_expressions.go:17`, `:47`,
`values.go:2380`, `uverse.go:1239`); `alloc.NewMap` always calls `MakeMap`
(`alloc.go:640-646`), so a live map's `vmap` is non-nil from birth. The three
constructors that leave it nil are amino decode (`pb3_gen.go:1446`), the
persistence copy (`realm.go:1729`) and the JSON export copy
(`values_export.go:266`) — the first is exactly the case the lazy build must
handle, and the other two produce values that are serialized, never executed
against. The field comment is updated to say so.

### 4. Accessors take the predicate and ensure the map

`GetPointerForKey`, `GetValueForKey` and `DeleteForKey` gain an
`omitKeyType bool` parameter and open by ensuring `vmap` exists, building it
from `List` with that same flag when nil. Build and lookup therefore agree
by construction — one value feeds both, so they cannot diverge.

The build passes `nil` for the `*Machine`, preserving today's gas semantics
exactly: this work is unmetered on the load path today, and stays unmetered
after the move. Consequence: gas for existing realms can only decrease.

### 5. The eager build is removed

The `*MapValue` case at `realm.go:1934` keeps its `fillTypesTV` /
`fillValueTV` loop over `List` and loses the `ComputeMapKey` and `vmap`
lines.

Secondary win: `GetLength` reads `List.Size` and map iteration walks `List`,
neither of which touches `vmap`. A stored map that is only ranged over or
measured avoids the serialization and `TypeID` work `ComputeMapKey` would
have done per entry — but not the walk itself. See §6: a cheap
`fillMapKeyRefs` pass still visits every key on load, because that walk does
double duty as the deep ref-resolver composite keys depend on.

### 6. `fillMapKeyRefs`: the ref-resolver `ComputeMapKey` was quietly providing

This is the single most surprising consequence of moving `vmap` off the load
path, so it gets its own subsection rather than a footnote.

`ComputeMapKey` resolves refs as a **side effect** of serializing a composite
key: for array elements and struct fields it calls `fillValueTV` on each
child before appending that child's bytes (`values.go:1949`, `:1973`). The
eager `vmap` build in `fillTypesOfValue` therefore was not just an index
build — it was also the deep ref-resolver for every composite map key loaded
from the store. `gnovm/tests/files/map39b.gno` depends on this: it stores,
reloads and prints a map with a struct key holding a `RefValue` field, and
printing walks the *key*, not just the value.

Removing the eager build's `ComputeMapKey` call (§5) removes that side
effect. `fillMapKeyRefs` (`realm.go:1901-1923`) replaces it with an explicit
walk that mirrors `ComputeMapKey`'s traversal exactly — array elements only
when `av.Data == nil` (a byte array has no `List`), struct fields except the
blank identifier, recursing through both, and deliberately *not* following
pointer keys, since `ComputeMapKey` hashes a pointer by address without
dereferencing it.

**The same coupling bit a second time**, later in this same implementation.
Task 3 (§8) folds the caller's pre-lookup `ComputeMapKey` call into
`GetPointerForKey`, so that call — which also used to run before
`iv.Copy(alloc)` in `GetPointerAtIndex` and, as a side effect, resolve the
key's refs before the copy — now runs *after* the copy, inside the accessor.
`TypedValue.Copy`'s default branch shallow-copies an unresolved `RefValue`
child, so without a fix the stored key aliased the caller's live object
graph until it was next resolved: `gnovm/tests/files/zrealm_map8.gno` pins
this by mutating the source key variable after assignment and asserting the
stored key does not change. The fix is `fillMapKeyRefs(store, iv)` called
explicitly at `values.go:2449`, immediately before `iv.Copy(alloc)` at
`values.go:2450`, so the resolve happens before the copy again — this time
as a deliberate call instead of an accident of ordering.

### 7. Call sites

Five sites supply `omitKeyType` from a `*MapType` they already hold:

| Site | Current state |
|---|---|
| `doOpIndex1` — `op_expressions.go:15` | has `ct := baseOf(xv.T).(*MapType)` |
| `doOpIndex2` — `op_expressions.go:45` | has `ct` |
| `doOpMapLit` — `op_expressions.go:586` | has `mt`; replaces the stale TODO at `:582` |
| `GetPointerAtIndex` — `values.go:2379` | has `bt` |
| `delete` builtin — `uverse.go:1237` | type switch currently discards the `*MapType`; bind it |

### 8. The direct `vmap` read at `values.go:2389-2395` moves into the accessor

`GetPointerAtIndex` reads `mv.vmap` outside any accessor, so post-change it
would observe a nil map on a freshly loaded value. It is folded into
`GetPointerForKey`.

Doing so also removes a redundant computation that is independent of this
lead — and, per the measurements in Context, is the larger of the two gas
wins in this change: the block computes `iv.ComputeMapKey(...)` to look up
the pre-existing key, then calls `GetPointerForKey`, which computes **the
identical key again** at `values.go:1011`. Every map *write* currently pays
`ComputeMapKey` twice. `GetPointerForKey` already has the old
`*MapListItem` in hand at `values.go:1013` — the same item the caller was
reaching for — so it can return the displaced key's object to the caller
instead of the caller pre-computing it. Folding the read in is unavoidable
regardless of the gas win, because §5 already removed the load-time build
this code was implicitly relying on; §6 covers the ref-resolution ordering
hazard this reshuffle introduced.

## Determinism and state compatibility

- `omitKeyType` is a pure function of the map's static key type, which is
  fixed at map construction and identical on every node.
- Nothing observable depends on *when* `vmap` is built, because the build is
  unmetered. Store-cache residency therefore cannot leak into gas — the
  property that would otherwise make lazy construction a consensus hazard.
- `MapKey` is never persisted: `vmap` is unexported, absent from the amino
  image, and already reconstructed from `List` on load. `copyValueWithRefs`
  omits it when writing a `MapValue` (`realm.go:1729`). The encoding is free
  to change with no state migration and no genesis implications.
- Interface-keyed maps keep byte-identical `MapKey`s. Concrete-keyed maps get
  shorter ones, used only for in-memory equality within a single map.

## Consequences

- Charged gas drops on every get/set/delete against a concrete-keyed map, by
  `(len(TypeID) + 1) * 4 / 10` per access — the change this ADR is named
  for — plus a second, larger and unnamed reduction from deduplicating the
  map-write key computation (§8). See Measured Results below: the second
  effect outweighs the first by roughly two orders of magnitude on
  byte/struct-heavy keys.
- Loading a stored map no longer pays `ComputeMapKey`'s serialization and
  `TypeID` work per entry unless the map is actually indexed. It is **not**
  true that no work happens at load: `fillMapKeyRefs` (§6) still walks every
  key, because that walk is also the deep ref-resolver composite keys need.
  What is avoided is specifically the encode-and-hash work, not the traversal.
- Three accessor signatures change: each gains an `omitKeyType` parameter, and
  `GetPointerForKey` additionally returns the displaced key's object (per
  §8). All callers are in-tree (`gnovm/pkg/gnolang`); `MapValue`'s exported
  surface is otherwise untouched.
- No other code reads `vmap` directly. Iteration (`op_exec.go:391`), GC
  (`garbage_collector.go:320`), allocation accounting (`alloc.go:852`),
  printing (`values_string_stream.go:548`), realm save (`realm.go:1205`,
  `:1395`, `:1724`) and export (`values_export.go:261`) all walk `List`, so
  the lazy build is invisible to them.
- `MapKey` values for concrete-keyed maps lose their type prefix, so
  `colors.ColoredBytes()` debug output for those maps shows the value bytes
  without a leading type name. The `values.go:1846-1849` comment about human
  readability is narrowed accordingly.
- Gas goldens under `gnovm/tests/files/gas/` that exercise map access must be
  regenerated.

### Measured Results

Both changes are gas *reductions* with no calibration pass needed. Test
bodies (loop counts, key sizes) are byte-identical before and after in every
golden below — only the `// Gas:` comment changed in each commit — so the
percentages are directly comparable.

**Deduplicating the double `ComputeMapKey` call on map writes** (§8, commit
`c2bb1d0a5`) — the larger, unplanned win:

| Golden | Before | After | Δ |
|---|---:|---:|---:|
| `compute_map_key_big_bytes` | 36827018 | 23405158 | **-36.4%** |
| `compute_map_key_big_struct` | 186401382 | 120708884 | **-35.2%** |
| `compute_map_key_small_bytes` | 7974 | 7877 | -1.2% |
| `compute_map_key_small_struct` | 6109 | 5677 | -7.1% |

**Omitting the TypeID prefix for concrete key types** (this ADR's named
decision, commit `08024d145`) — the smaller, planned win:

| Golden | Before | After | Δ |
|---|---:|---:|---:|
| `compute_map_key_concrete_key` | 125449 | 124849 | -600 total (100 writes × 6/access; `TypeID` `"main.Address"`, 12+1 bytes) |
| `compute_map_key_big_bytes` | 23405158 | 23405151 | -7 (`TypeID` `"[33554432]uint8"`, 15+1 bytes) |
| `compute_map_key_small_bytes` | 7877 | 7873 | -4 (`TypeID` `"[32]uint8"`, 9+1 bytes) |
| `compute_map_key_big_struct` | 120708884 | 120708884 | unchanged — `map[any]int`, interface-keyed, prefix correctly retained |
| `compute_map_key_small_struct` | 5677 | 5677 | unchanged — same reason |

Per-access arithmetic for the `map[address]T` shape, with
`OpCPUComputeMapKey = 80` (`machine.go:1661`) and the per-byte slope charged
as `bytes*4/10`: **80 + 45*4/10 = 98** gas before this change, **80 + 32*4/10
= 92** gas after. The 13-byte prefix is ~29% of the 45 bytes hashed, but only
~6% of the 98-gas `ComputeMapKey` charge and 0.48% of program gas
(600 of 125449, the `compute_map_key_concrete_key` golden's total before this
change) — confirming the Context section's framing: the constant per-call
charge, not the per-byte one, dominates at realistic key sizes.

## Alternatives considered

**Store the flag on `MapValue`.** Rejected: no static type on the
`GetObject` load path (see Diagnosis), and the added field breaks
`_allocMapValue`, raising allocation gas for every map.

**Exempt the prefix bytes from the per-byte charge, keep appending them.**
One-line change, no encoding or state impact, and it captures the gas drop.
Rejected: `machine.go:1667-1669` documents that charge as the mitigation for
GHSA-m7rp-96x5-hvpx, naming long TypeIDs specifically. Making a
source-controlled-length byte string free to hash reopens that surface.

**Replace the TypeID prefix with a compact interned type index.** Needs no
static type, works uniformly for interface keys too, and shrinks the prefix
to a fixed 8 bytes. Rejected: the interning table must be process-wide and
its assignment order varies by execution history, so a variable-length
encoding would make gas node-dependent; pinning it to a fixed width to avoid
that discards most of the saving while adding an unbounded global table and
sacrificing `MapKey` readability entirely.

**Truncated hash of the TypeID.** Rejected: a collision between two distinct
types makes two unequal keys compare equal, which is consensus-visible
incorrect behavior. Determinism-critical code should not rest on collision
resistance where an exact encoding is available.

**Keep the eager `vmap` build and infer concreteness from stored keys.**
Rejected: an `any`-keyed map holding only `int` entries is indistinguishable
from an `int`-keyed one, and mislabelling it reintroduces the collision on
the next insert.

## Testing

Filetests targeting the cases that break if the predicate is wrong:

- `map[any]K` holding `int(1)`, `int64(1)` and `uint64(1)` — must remain
  three distinct entries. This is the case the stale TODO's `Elem()` version
  would have failed.
- `map[*T]V` with a nil pointer key (`gnovm/tests/files/map53.gno`,
  `ptrs[nil]`).
- NaN keys, covering the `isNaN` skip in both the lazy build and lookups.
- Struct- and array-keyed maps with interface-typed fields/elements, checking
  the nested `omitTypes` rule still composes with the new top-level one.
- A persistence round-trip: store a map, reload it, confirm lookups hit now
  that `vmap` is built lazily rather than during `fillTypesOfValue`.

Suites required by `AGENTS.md` for gas-affecting changes:

```
go test ./gno.land/pkg/sdk/vm/ -run Gas
go test ./gno.land/pkg/integration/ -run TestTestdata
go test ./gnovm/pkg/gnolang/ -run Files -test.short
```

Plus `TestComputeMapKey` and `TestComputeMapKey_collisions`
(`values_test.go:266`, `:377`), the `BenchmarkComputeMapKey_*` family, and
regenerated `gnovm/tests/files/gas/compute_map_key_*.gno` goldens.

Before/after gas numbers must be compared with the test bodies held
identical — same loop counts, same key sizes — per `AGENTS.md`.

### Maintainer note: gas-golden fragility

Two traps a future editor of `gnovm/tests/files/gas/*.gno` will hit, neither
specific to this change but both encountered while producing it:

- **A `// Gas:` golden is coupled to the source file's line count, not its
  content.** Every line the parser walks — including a blank line or a
  comment — advances position-tracking state that is itself metered.
  Measured on this branch: inserting one blank line into a gas filetest costs
  ~226 gas; adding one comment *token* (regardless of the comment's length)
  costs ~1 gas. Comment length itself is free. Consequence: editing a
  comment's wording in a gas filetest is safe; changing its line count, or
  adding/removing blank lines, requires either preserving the line count
  exactly or resyncing the golden with `-update-golden-tests` (see below) —
  do not hand-adjust the number to compensate.
- **Gas subtests only pass in a FULL `-run Files` run.** Running the `gas/`
  subset in isolation (`-run 'Files/^gas'` or similar) warms the filetest
  harness's caches differently than a full run does, and nine goldens
  inflate by a constant ~113 gas as a result (one, `slice_alloc`, by ~461) —
  reproducible on unmodified `master`, unrelated to this change. The
  committed goldens are the full-run numbers. **Never judge a gas delta, and
  never invoke `-update-golden-tests`, from a subset run** — it would bake
  the inflated subset numbers into the goldens and break the full run for
  everyone else. Always use the complete
  `go test ./gnovm/pkg/gnolang/ -run Files -test.short` invocation for both
  checking and updating.

## Out of scope

- `DeclaredType.TypeID()`'s redundant self-check (`types.go:1931-1941`). When
  `dt.typeid` is already memoized, the `else` branch still recomputes
  `DeclaredTypeID(...)` — three `Sprintf`s and their string allocations —
  purely to assert equality and `panic("should not happen")`, under a
  standing `// XXX delete this if tests pass`. It sits directly on this
  path and amplifies the win, but lands separately.
- Metering the load-time `vmap` build. Today's build is unmetered and this
  change preserves that; charging for it would convert a pure gas reduction
  into a mixed change with a first-touch spike, and needs its own ADR.
