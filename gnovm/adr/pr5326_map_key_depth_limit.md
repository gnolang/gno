# ADR: Map Key Nesting Depth Limit (NEWTENDG-93)

## Context

`ComputeMapKey` (`values.go`) and `isEql` (`op_binary.go`) both recurse
through composite values (arrays and structs) without any depth limit.
A sufficiently deep `[1]any{...}` chain causes `fatal error: stack overflow`
— unrecoverable, kills the node process. Because GnoVM execution is
deterministic, every validator crashes on the same block → chain halt.

The attack window is approximately [~1.88 M, ~2.02 M] nesting levels,
reachable within production limits (`maxAllocTx = 500 MB`,
`MaxGas = 3 B`) with a single `MsgRun --simulate skip`.

## Decision

### `ComputeMapKey` — depth-limited recursion

Added a `maxComputeMapKeyDepth = 10000` constant and split `ComputeMapKey`
into a public wrapper (same signature, no behaviour change for callers) and
a private `computeMapKey(store, omitType, depth)` that:

1. Panics with `"map key nesting depth limit exceeded"` when `depth > 10000`.
2. Passes `depth+1` into each recursive call (array elements, struct fields).

The panic is a plain string panic, which is caught by the GnoVM machine's
`defer/recover` in `runTx`. It surfaces as a normal transaction error — the
node process stays alive.

10 000 levels is far above any legitimate use case (typical nesting: 1–10)
and well below the Go goroutine stack limit (~1 900 000 frames at ~526 B
each).

### `isEql` — iterative rewrite

Converted `isEql` from recursive to iterative using an explicit LIFO pair
stack (`[]struct{ l, r TypedValue }`). This eliminates all stack growth for
arbitrarily nested arrays and structs used in equality comparisons.

No depth limit is needed: the value pairs already consumed allocator-tracked
memory proportional to depth, so `maxAllocTx` is the natural bound.

## Alternatives Considered

1. **`runtime/debug.SetMaxStack`** — sets a per-goroutine limit but
   terminates the goroutine (panic with `stack overflow` string), which is
   still unrecoverable inside `defer/recover`. Rejected.

2. **Single depth limit on `isEql` instead of iterative rewrite** — simpler,
   but a depth limit on `isEql` would be an artificial cap on legitimate
   equality comparisons. The iterative rewrite is strictly better.

3. **Gas-proportional cost for deep keys** — metering each recursion level
   with gas would bound depth via gas limits. Valid long-term, but requires
   changes to the gas model and doesn't address the immediate crash window.
   Deferred.

## Consequences

- `ComputeMapKey` panics (recoverable) for nesting > 10 000. Any Gno program
  that genuinely requires >10 000 levels of composite map keys would need the
  limit raised or a gas-based approach.
- `isEql` now uses O(depth) heap memory instead of O(depth) stack frames —
  bounded by `maxAllocTx`, which is the same constraint as building the
  values in the first place.
- All existing call sites of `ComputeMapKey` are unchanged (public API
  preserved).
- Integration test `mapkey_recursion_overflow.txtar` updated: Part 2 now
  uses 15 000-level depth and asserts a recoverable error instead of a node
  crash.
