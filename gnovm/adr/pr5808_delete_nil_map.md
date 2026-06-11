# ADR: delete on a nil map is a no-op

## Status

Proposed (AI-assisted fix; found via differential testing against the Go
toolchain).

## Context

Per the Go spec ("Deletion of map elements"): "If the map m is nil or does
not contain such an element, delete is a no-op." GnoVM previously panicked
with an unrecoverable VM abort (`interface conversion: gnolang.Value is
nil, not *gnolang.MapValue`) — the delete builtin type-asserted the map
value without a nil guard. The guard itself landed in PR #5196 (early
return before the `*MapValue` type assertion and before the readonly-taint
check), with a basic filetest (`map48.gno`).

## Decision

Keep the guard exactly as #5196 placed it, and pin its full semantics with
tests, since two of its consequences are deliberate behavior decisions that
were not previously covered:

1. **Ordering relative to the readonly check is unobservable.** A nil
   value can never carry the readonly taint: `TypedValue.IsReadonly`
   requires `V != nil` and `SetReadonly` early-returns on nil values, so
   `Machine.IsReadonly` is definitionally false for a nil map regardless
   of origin. Readonly-tainted *non-nil* maps still panic on delete
   (pinned by the existing `zrealm_map1.gno`/`zrealm_map3.gno`).
   `zrealm_mapnil.gno` pins the cross-realm nil no-op path.

2. **Unhashable interface keys on a nil map no-op.** The gc runtime
   panics (`hash of unhashable type`) even when the map is nil; gno
   follows the spec text instead. This matches gno's pre-existing
   behavior for nil-map *reads* with unhashable keys, and the
   alternative — hashing the key before the nil return — would today
   turn the case into an unrecoverable VM abort, since gno's unhashable
   map-key panic does not go through the recoverable exception path.
   Pinned by the `delete(mi, []int{1})` case in `delete1.gno`. If gno's
   map-key hashing later gains a recoverable "hash of unhashable type"
   panic, gc parity here can be revisited.

## Alternatives considered

- Hashing/validating the key before the nil-map return for gc parity:
  rejected for now; see consequence 2.
- Placing the nil guard after the readonly check: behaviorally identical
  (see consequence 1); the earlier return is simpler.

## Consequences

- `delete` on nil maps matches the Go spec in all paths probed (local
  vars, package vars, struct fields, function returns, conversions,
  cross-realm values); key expressions are still evaluated exactly once.
- Covered by `gnovm/tests/files/delete1.gno` and
  `gnovm/tests/files/zrealm_mapnil.gno`, extending the basic coverage of
  `map48.gno` from #5196.
