# Recoverable panic on nil interface method call

## Context

Calling a method on a nil interface value in Gno raised a raw Go-level panic
that bypassed the VM's `recover()` machinery. The deferred `recover` in user
code therefore never observed the panic, and execution terminated with a
Go-level stack trace instead.

Minimal reproducer:

```go
type I interface{ M() }

func main() {
    defer func() {
        err := recover()
        if err == nil {
            panic("panic expected")
        }
        println("recovered:", err)
    }()
    var i I
    i.M() // before: program dies with raw Go panic; recover() never fires
}
```

The panic originates in `gnovm/pkg/gnolang/values.go`, in
`TypedValue.GetPointerToFromTV`, `VPInterface` case:

```go
case VPInterface:
    if dtv.IsUndefined() {
        panic("interface method call on undefined value")
    }
```

The recover loop in `Machine.Run` (`machine.go`) only re-pushes `*Exception`
panics into `m.Exception`; raw string panics are re-panicked as-is. This is
the same class of bug fixed for nil function calls in PR #5711 and for nil
map operations in PRs #5195, #4452, #3856.

## Decision

Replace the raw panic with a `*Exception` panic so it routes through the
Gno-level recover path, matching the style already used elsewhere in
`values.go` (e.g. `runtime error: nil pointer dereference` at lines 1813,
1829):

```go
case VPInterface:
    if dtv.IsUndefined() {
        panic(&Exception{Value: typedString("runtime error: method call on nil interface")})
    }
```

The message follows the `runtime error:` prefix convention established by
PR #5501 and the descriptive style of PR #5711's `runtime error: call of nil
function`. It distinguishes a nil interface (no underlying type) from a
typed-nil receiver, which is a different case and does not panic at call
time.

`GetPointerToFromTV` does not hold a `*Machine`, so `m.Panic()` is not
available here; panicking with `*Exception` directly is the established
pattern in this file (see lines 1659, 1686, 1813, 1829, 2021, 2049, etc.)
and is documented in `machine.go`:

> Some code in realm.go and values.go will panic(&Exception{...}) directly.
> Keep this code in sync with those calls.

## Alternatives considered

- **Raise the panic at the selector op site (`op_expressions.go:doOpSelector`)
  via `m.Panic`.** Rejected: `GetPointerToFromTV` is invoked from several
  call sites (`op_expressions.go`, `machine.go`, `debugger.go`, and
  recursively from itself); hoisting the nil check would duplicate logic.

- **Keep the original message `"interface method call on undefined value"`.**
  Rejected: it lacks the `runtime error:` prefix that PR #5501 standardized
  for recoverable runtime panics, and "undefined value" is internal VM
  terminology rather than Go-runtime language.

- **Use Go's exact message `"runtime error: invalid memory address or nil
  pointer dereference"`.** Rejected: less informative for debugging Gno
  programs; PR #5711 established the precedent of using a more specific
  message (`call of nil function`) for the structurally similar case.

## Consequences

- `recover()` now catches method calls on nil interface values, matching Go.
- The runtime panic message changes from
  `"interface method call on undefined value"` to
  `"runtime error: method call on nil interface"`.
  No existing tests or stdlib code grep for the old string.
- Aligns with the broader effort (PRs #5195, #4452, #3856, #5196, #5711) to
  convert raw VM panics into recoverable Gno-level exceptions.
