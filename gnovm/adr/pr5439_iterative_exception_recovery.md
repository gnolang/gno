# PR5439: Iterative Exception Recovery in Machine.Run()

## Context

`Machine.Run()` contained a `defer/recover` handler that caught Go-level `*Exception`
panics and recursively called `m.Run(st)`. This design meant each panicking deferred
function added a new Go stack frame. An attacker could register enough deferred closures
that each trigger a nil pointer dereference (a Go-level `panic(&Exception{})`), causing
unbounded Go stack growth that exceeds the Go runtime's 1GB goroutine stack limit
(~500K defers are sufficient to crash the process).

The resulting `runtime.throw("stack overflow")` is a fatal error that bypasses all
`recover()` handlers in the call chain — including the VM keeper's `doRecover` and
`BaseApp.runTx()` — killing the node process.

The GnoVM has two panic mechanisms:
1. **Cooperative path** (`pushPanic`): pushes `OpReturnCallDefers` + `OpPanic2` onto the
   op stack and returns. The main `for` loop processes defers iteratively — no Go stack growth.
2. **Go-level path** (`panic(&Exception{...})`): used at ~19 call sites in `values.go`,
   `alloc.go`, and `realm.go`. These trigger real Go panics that unwind past the `for` loop
   and are caught by `Run()`'s defer/recover.

The old code converted Go-level exceptions back into the cooperative path via `pushPanic`,
but then re-entered the op loop by recursively calling `m.Run(st)` — accumulating one Go
stack frame per exception.

## Decision

Split `Machine.Run()` into two methods:

- **`Run(st Stage)`** — outer method, contains the benchmark defers and an iterative loop.
  When `runOnce()` returns a caught `*Exception`, it calls `pushPanic` and loops back.
  No recursion, O(1) Go stack frames regardless of the number of panicking defers.

- **`runOnce() *Exception`** — inner method with its own `defer/recover`. Runs the op loop
  until `OpHalt` (returns nil) or a Go-level `*Exception` panic is caught (returns the
  exception). Non-Exception panics are re-raised.

This preserves the existing semantics: Go-level `*Exception` panics are still converted
to the cooperative `pushPanic` path, and the op loop still processes `OpReturnCallDefers`
iteratively. The only change is that re-entering the op loop after catching an exception
no longer adds a Go stack frame.

### Alternatives considered

1. **Depth counter on recursive `Run()`**: Would limit recursion depth, but choosing the
   right limit is fragile and the recursive design is fundamentally unnecessary.

2. **Convert all 19 Go-level `panic(&Exception{})` sites to use `pushPanic`**: Would
   eliminate the problem at the source, but is a much larger change that touches many
   files and risks subtle behavioral differences. The iterative approach is a minimal,
   surgical fix.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/machine.go:1268` | `Run()` — outer iterative loop |
| `gnovm/pkg/gnolang/machine.go:1300` | `runOnce()` — inner op loop with defer/recover |
| `gnovm/tests/files/defer_panic_many.gno` | Regression filetest — 500K panicking defers |

## Testing

A dedicated filetest (`gnovm/tests/files/defer_panic_many.gno`) registers 500K deferred
closures that each trigger a nil pointer dereference — a Go-level `panic(&Exception{})`.
Before the fix, this would exhaust the Go goroutine stack via recursive `m.Run(st)` calls,
crashing the process with `runtime.throw("stack overflow")`. With the iterative recovery
loop, all 500K panicking defers complete in ~1s and the final panic is recovered normally.

The fix is also validated by the existing 96 panic/defer/recover file tests in
`gnovm/tests/files/`, which exercise the `Run()`/`runOnce()` iterative recovery path
on every run.

## Consequences

- Node processes can no longer be crashed by transactions with many panicking defers.
- The Gno-level panic semantics are preserved — all 96 panic/defer/recover file tests pass.
- `runOnce` is unexported, keeping the public API unchanged.
