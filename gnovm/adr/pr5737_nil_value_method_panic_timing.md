# ADR-5737: Call-time dispatch for interface-bound method values

## Status

Proposed. **Hard-fork-class** (changes a persisted value's wire format).

## Context

In Go a method value formed through an interface (`g := i.M`, `defer i.M()`)
saves the operand at formation and materializes the concrete method + receiver
*inside the call*. GnoVM resolved it eagerly at bind, diverging from Go (vs
go1.25) on: nil-panic timing, value snapshot, embedded-method promotion, field
re-read through a boxed pointer, dynamic re-dispatch when an embedded interface
field is reassigned, and a VM crash for a nil embedded pointer-receiver (c5).
These last facets need the operand's *live* value at the call, so earlier narrow
fixes (a `nilReceiverPanic` flag, a `viaInterface` gate) couldn't reach them.

## Decision

Bind interface method values lazily; resolve at the call.

- `BoundMethodValue` gains `Method Name`; `Func == nil` (operand in `Receiver`)
  marks a lazy bind (`IsLazy()`).
- **Bind** (`VPInterface`): just save the operand — `&BoundMethodValue{Func: nil,
  Receiver: *dtv, Method: name}`, no deref/dispatch. Concrete/pointer binds
  (`pt.M`, `t.M`) are untouched.
- **Call** (`resolveLazyBound`, from `doOpPrecall` and the deferred call): walk
  the operand's *current* value (`findEmbeddedFieldType` + `resolveInterfaceTrail`)
  to the concrete method + receiver — the deref, embedded-field walk, copy, nil
  panic, field re-read and re-dispatch all happen here. A nested embedded
  interface yields another lazy bind; the loop unwraps it.
- **Defers** carry the callable on `Defer.Callable` (`*FuncValue` or
  `*BoundMethodValue`), type-switched at the deferred call like `doOpPrecall`.
- **nil derefs** are raised by the walk (`GetPointerToFromTV`) via the `Run`
  loop's panic path; a `VPSubrefField` nil guard makes c5 panic cleanly.
- **Reload**: `resolveLazyBound` fills the operand (`fillValueTV`) first, so a
  persisted bind re-reads live state — what makes re-read / re-dispatch correct
  across reload.
- `Func == nil` consumers audited (`IsCrossing`, `String`, GC, realm
  persist/copy/fill, `GetShallowSize`).

## Consequences

- **Hard fork**: `Method` is proto field 4 (`pb3_gen` regenerated);
  `_allocBoundMethodValue` 200→216. Persisted bound methods change bytes →
  different IAVL hashes / gas (`stdlib_restart_compare` pin moved). Ship only on
  fresh genesis or a coordinated upgrade (like ADR-5544); old state still decodes
  (new field defaults empty).
- **Behaviour**: interface method values match Go on every axis, across reload;
  concrete/pointer binds unchanged.
- **Performance / gas.** Concrete path unchanged (`OpCPUPrecallBoundMethod` 199
  still valid). The dispatch *walk* moved from bind to call, so the two consts
  were re-fit together (ratio-scaled; reference HW unavailable — see the
  `TODO(calibration)` on each):
  - bind `OpCPUSelectorInterface` 751 → **276** — the eager selector walked the
    trail here (~751); the lazy bind only does the method lookup + lazy-bind
    alloc (~140 ns ≈ 276), so leaving it at 751 would double-charge the walk.
  - call `OpCPULazyBoundResolve` **529** (new) — the walk, now at call time.
  Net ≈ +54 gas per interface method call (the genuine extra: a re-lookup +
  throwaway bound-method alloc), not the ~480 an un-reduced base would
  over-charge. `stdlib_restart_compare` pin → 2235788. A lean walk avoiding the
  throwaway alloc was rejected — it would duplicate `GetPointerToFromTV`'s
  dispatch/nil machinery on consensus code; the reuse form stays the single
  source of truth.

## Follow-ups

- **Calibration, before the fork ships** (both consts are ratio-scaled — the
  reference HW was unavailable): re-measure `OpCPUSelectorInterface` (276) and
  `OpCPULazyBoundResolve` (529) on the gas-table reference HW; and consider a
  per-trail-step slope on the lazy resolve so deep/nested dispatch is metered
  per hop, matching the eager path (currently flat, a bounded under-charge).
- Orthogonal, pre-existing (not caused or addressed here): interface method
  *expressions* `I.M` rejected at preprocess (#5787); a method call on a *nil
  interface* panics uncatchably (#5850).
- Cache a value-operand lazy bind's resolution (it is stable) to skip the
  re-walk on repeated calls — **consensus-affecting, not just perf**: skipping
  the walk drops its gas charge (`OpCPULazyBoundResolve` + alloc-gas), and
  mutating a persisted bmv in place rewrites its bytes (merkle). So it is itself
  a hard-fork change — fold into this fork's window or skip; it cannot ship as a
  later rolling upgrade. Marginal benefit (only the call-a-stored-value-operand-
  method-value-N-times pattern; pointer operands must never be cached, they
  re-read live state), so deferred.
