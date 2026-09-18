# PR5732: Error-Typed Runtime Panics (`.runtimeError`)

## Context

In Go, panics raised by the runtime (nil pointer dereference, division by zero,
index out of range, ...) carry a value implementing `error` (`runtime.Error`),
so `recover().(error)` works. GnoVM raised the same panics with plain string
values (`typedString("runtime error: ...")`), so the assertion failed and Gno
code ported from Go broke (#5667).

## Decision

Add a hidden uverse type `.runtimeError` — a sealed `DeclaredType` over
`StringType` with a native `Error() string` method, the same shape as Go's
`runtime.plainError` — and switch every VM runtime-panic site from
`typedString` to `typedRuntimeError`, which wraps the message in a
`.runtimeError` value.

- The leading dot makes the type unnameable from user code (same convention as
  `.grealm`), so there is no identifier collision; user code reaches it only
  through the `error` interface.
- It follows the exact pattern of the existing `address` uverse type: sealed
  string-based literal + `defNativeMethod` in `makeUverseNode`. `InitStoreCaches`
  caches all uverse `DeclaredType`s, so persisted values resolve by TypeID on
  reload; the `StringValue` payload persists inline, engaging no object-store
  machinery.
- Nil value-receiver method calls (`VPDerefValMethod`) now panic with Go's
  wrapper message `value method PKG.T.M called using nil *T pointer` when the
  pointee is a declared type. Go only emits this text on the non-inlined
  interface-wrapper path and emits the generic
  `runtime error: invalid memory address or nil pointer dereference` for
  direct calls; Gno intentionally uses the more informative message for both.
- Internal allocator guards in `alloc.go` (`NewListArray` etc., unreachable
  from user code because uverse `make` validates first) are also error-typed
  for uniformity, keeping their minimal messages.

## Alternatives considered

- **A `runtime` stdlib package exposing `runtime.Error`** — closer to Go, but
  requires stdlib plumbing for a VM-internal concern; the interface-shaped
  contract (`recover().(error)`) is what user code actually relies on.
- **Making `typedString` values implement `error`** — would change the type of
  every user `panic("...")` too, a much larger behavior change.
- **A `struct { msg string }` base** (the shape Go uses for its richer
  `boundsError`) — works identically at the language surface, but the
  `StructValue` is a heap Object (ObjectInfo, GC visit, object-store
  persistence) and the structured fields would be unobservable anyway since
  the type is unnameable; the string base is strictly leaner.

## Consequences

- `recover().(error)` works on VM runtime panics, as in Go
  (`gnovm/tests/files/recover26.gno` pins ten sites;
  `ptr11c.gno` mirrors Go's `test/fixedbugs/issue19040.go`).
- **Breaking:** `recover().(string)` on a VM runtime panic now fails the
  assertion (panics on the non-comma-ok form). User `panic("...")` values are
  unaffected.
- Unrecovered-panic output is unchanged in shape: `Exception.Sprint` detects
  the `error` interface and prints the message via `Error()`.
