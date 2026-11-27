# Decode stored gnovm objects through genproto2, and stop paying reflect's interface check per Any

## Status

Proposed.

## Context

`defaultStore.loadObjectSafe` decodes every object a transaction touches with
`amino.MustUnmarshal(bz, &oo)`, where `oo` is a `gnolang.Object` interface.
`*Object` is not a `PBMessager2`, so `Codec.Unmarshal` skips the generated
`UnmarshalBinary2` methods that `gnovm/pkg/gnolang/pb3_gen.go` already has for
every object type, and decodes the whole value tree by reflection.

On a read-modify-write workload over an avl tree (pebbledb, 2000 measured
transactions), a CPU profile put object loading at 66% of DeliverTx wall time
and amino decode at 75% of that. Decode was bound by reflection, not by
allocation: `reflect.implements` alone was 28% of decode CPU, `mallocgc` 20%.

Switching to the generated path alone helps less than expected, because the
generated path is reflection-bound too. Each Any field goes through
`Codec.unmarshalAnyBinary2Depth`, which calls `reflect.Type.AssignableTo`
twice (pre and post decode) and then `reflect.Value.Set`. With an interface
destination, each of the three walks the method sets of both types. On a
`StructValue` that was 64% of decode CPU.

## Decision

1. `loadObjectSafe` decodes with a new `amino.UnmarshalAnyBinary2(bz, &oo)`,
   the package-level form of `Codec.UnmarshalAnyBinary2`. It reads the same
   bytes as `MustUnmarshal` on an interface pointer and falls back to
   reflection for a concrete type without generated code.
2. Both assignability checks in `unmarshalAnyBinary2Depth` go through
   `cachedAssignableTo`, a `sync.Map` keyed by the (from, to) type pair.
   Assignability is a static property of two Go types, so the cache cannot go
   stale. Both checks stay, as `TestUnmarshalAnyBinary2_PostDecodeAssignabilityCheck`
   requires.
3. `amino.RegisterInterfaceSetter[I]()` registers a typed setter for
   interface `I`. When the destination is a `*I` with a registered setter, the
   decoded value is stored with a type assertion (`*p = v.(I)`), which uses the
   runtime's itab cache, instead of `reflect.Value.Set`. gnolang registers its
   eight interface types (every interface reachable from its amino package,
   plus `Object`). An unregistered interface takes the unchanged reflect path.

## Alternatives considered

- Route every interface pointer in `Codec.Unmarshal` through genproto2. It
  would speed up every caller, but it changes behaviour for callers that rely
  on reflect's error semantics on malformed input, and it is a larger review.
  Left for a follow-up; this change opts in one call site.
- Cache the itab word and write the interface with `unsafe`. Generic, no
  registration, and a maintenance hazard. The typed setter costs one line per
  interface.
- Keep decoded objects across transactions. Measured at 1.8x to 2.9x on the
  same workload, but it changes gas (decode gas is only charged on a miss) and
  shares mutable objects across transactions. It needs a design, not a patch,
  and is out of scope here.

## Consequences

- Same objects, same bytes: tests compare reflect decode, genproto2 decode
  and genproto2 with setters, value for value and re-encoded byte for byte,
  on amino's own interface-heavy fixtures and on every object kind a realm
  persists. Gas is untouched: decode gas is charged on the stored byte length
  before decoding, and nothing else in the path changes.
- The assignability cache grows with the number of distinct (concrete,
  interface) pairs ever decoded, which is bounded by the registered types.
- A new exported generic function in amino. It is an optimization hook only:
  removing every registration changes speed, not results.
