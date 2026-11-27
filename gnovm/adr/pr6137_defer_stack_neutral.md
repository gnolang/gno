# PR6137: Keep `doOpDefer` value-stack neutral on every branch

## Context

`doOpDefer` runs with the deferred func value and its `NumArgs` evaluated
arguments on top of the operand stack. The `*FuncValue` and
`*BoundMethodValue` branches pop the arguments through `popCopyArgs`, then the
shared `m.PopValue()` pops the func value, so the statement leaves the stack as
it found it. Two branches did not:

- `case nil:` (a nil func value, e.g. `var nf func(int, int); defer nf(1, 2)`)
  popped only the func value and left the arguments.
- `default:` (a callee that is not a func, e.g. `defer int64(x)`) panicked and
  returned before the shared pop, leaving the func value and the arguments.
  The type checker rejects such a defer, so AddPackage and MsgRun never get
  here, but paths that skip it do, such as filetests and `vm/qeval` queries.

Frame truncation on return discards the leftovers, so Gno code cannot read
them. But GC walks the operand stack since #6164
(`pr6164_gc_visit_operand_stack.md`), so until the frame returns each leftover
slot counts as 40 B of live memory and is charged as a scanned slot. For the
`default:` branch that includes the deferred calls run by the panic, since
`pushPanic` truncates frames, not values.

## Decision

Pop on every branch:

- `case nil:` pops `numArgs`; the shared `m.PopValue()` pops the func value.
  #6504 landed this pop on master too; this PR adds the test that pins it.
- `default:` pops `numArgs + 1` before `pushPanic`, because it returns before
  the shared pop. `pushPanic` does not read the operand stack (`Stacktrace`
  walks frames) and `cv` is already held, so popping first changes nothing
  else.

## Alternatives considered

- `popCopyArgs` in `case nil:`: needs a usable `*FuncType` and copies
  arguments nobody reads. `PopValues` only drops them.
- Rely on frame truncation: correct for Gno code, but GC and its gas still see
  the leftovers until the frame returns.
- Truncate `m.Values` to the frame's `NumValues` in `pushPanic`: would also
  cover the `default:` branch and every other runtime panic, which can leave
  operands the same way (calling a nil func directly does). Broader, since it
  changes GC accounting for every panicking frame; out of scope here.

## Consequences

- Every branch of `doOpDefer` leaves the operand stack as it found it.
- GC no longer counts the func value and argument that `defer int64(x)` left
  behind, 80 B. The 240 B for three `defer nf(1, 2)` came with #6504's
  nil-branch pop.
- Not consensus-visible: AddPackage and MsgRun reject `defer int64(x)` at type
  check, so only filetests and `vm/qeval` reach `default:`.
- `gnovm/tests/files/alloc_defer_operand_stack.gno` pins both branches with a
  MemStats golden and fails if either pop is removed.
- `defer_multivalue_call.gno` (from #6164) also covers variadic, method and
  builtin callees, which failed the same way before #6164's `NumArgs` fix.
