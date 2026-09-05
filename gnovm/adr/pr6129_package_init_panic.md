# Package-initialization panic without a call frame

## Context

Runtime panics raised by package-level variable initializers can happen before
the VM has pushed any function call frame. `pushPanic` converted the panic value
into a VM `Exception`, called `PopUntilLastCallFrame`, and then assumed the
returned frame was non-nil so it could link `fr.LastException`.

During package initialization that assumption is false. A normal VM runtime
panic, such as an out-of-range index or division by zero in a variable
initializer, therefore became a Go-host nil pointer dereference instead of the
intended Gno unhandled panic.

## Decision

When `pushPanic` finds no call frame to unwind, keep the constructed exception
in `m.Exception` and immediately use the existing `makeUnhandledPanicError`
terminal path.

The normal call-frame path is unchanged: if a call frame exists, `pushPanic`
continues to link the previous frame exception and pushes
`OpPanic2`/`OpReturnCallDefers` so defers and `recover` run as before.

## Alternatives Considered

- Return when no call frame exists. Rejected because it leaves the machine with
  an active exception but no panic-unwind ops; the remaining `OpHalt` could let
  package initialization continue and swallow the original panic.
- Broaden the exception subsystem. Rejected because `doOpPanic2` already
  defines the no-call-frame terminal behavior, and #6051 only needs
  `pushPanic` to use that same path when package initialization starts without
  a frame.

## Consequences

- Package-level initializer panics now surface as Gno unhandled panics with the
  original runtime error value.
- Ordinary function panics, defer unwinding, and `recover` behavior remain on
  the existing call-frame path.
- No gas constants, package initialization order, or public API are changed.
