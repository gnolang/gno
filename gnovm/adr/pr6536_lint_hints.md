# ADR: Advisory lint hints for realm authors

## Context

`gno lint` only reports errors: parse, type check, preprocess, a wrong
`Render` signature. Some realm code is valid but easy to misjudge in cost or
behaviour, and nothing points the author at it before deployment:

- `Render` building its page with `+` or `+=` inside a loop copies the whole
  output on every iteration, so its cost grows with the square of the page.
- An exported function that sizes a `strings.Repeat` / `bytes.Repeat` from an
  integer argument, or appends a whole slice argument into package state,
  lets each caller choose the size of the result or of the stored state.
- A banker `SendCoins`, `IssueCoin` or `RemoveCoin` that fails on chain (for
  example on an insufficient balance) fails with a Go panic raised by the
  native binding (`gno.land/pkg/sdk/vm/builtins.go`, `SDKBanker`), not with a
  Gno panic. `Machine.Run` only turns `*Exception` values into Gno panics
  (`Machine.runOnce` re-raises anything else), so the transaction ends and a
  `defer func() { recover() }()` in the realm never runs on it. Authors who
  wrap a payment in `recover()` expecting a fallback get an aborted
  transaction instead.

## Decision

- Add a hint tier to `gno lint`, printed in the existing issue format
  (`<file:line:col>: <msg> (code=<code>)`) with confidence 0.8. Hints never
  set the error flag, so they never change the exit status.
- Three codes, declared next to the existing ones in `cmd/gno/common.go`:
  - `gnoRenderLoopConcatHint`: an exported `Render` returning a single
    `string` that, inside a `for` or `range` body, does `s += ...` or
    `s = s + ...` on a variable known to be a string (a string literal
    initialiser, `var s string`, a named string result) or with a string
    literal on the right. One hint per outermost loop.
  - `gnoUnboundedBulkCopyHint`: in an exported top-level function, a
    `strings.Repeat` / `bytes.Repeat` whose count is an integer parameter
    (optionally converted), or an `append(state, param...)` whose destination
    is a package-level variable and whose source is a slice or string
    parameter. Silent if that parameter appears in any `if`, `for` or `switch`
    condition of the function, which counts as a bound.
  - `gnoUnrecoverableNativeHint`: a `SendCoins`, `IssueCoin` or `RemoveCoin`
    call in a function that directly defers a function literal calling
    `recover()`, in a file that imports `chain/banker`.
- Hints run only on realms (`gno.IsRealmPath`), on production files, after
  type checking succeeded. They use `go/ast` only, no type information:
  every rule matches a shape it can see for certain, and stays silent when it
  would have to guess.

## Alternatives considered

- **Type-aware checks through `go/types.Info`.** More precise, but
  `TypeCheckMemPackage` does not expose its `Info`, and widening that API for
  advisory output was not worth it. The syntactic rules are narrow enough that
  the examples tree produces one hint.
- **Including `copy`.** `copy(dst, src)` never writes more than `dst`
  already holds, so a caller-sized source cannot grow anything. Reporting it
  produced a false positive on `examples/`; it is left out.
- **Following `Render` into helper functions.** Would find more, also more
  noise and a call graph to maintain. Left for later.
- **Errors instead of hints.** These are judgement calls, not correctness
  failures; failing lint on them would break valid packages.

## Consequences

- `gno lint` may print new lines on packages that used to print nothing, with
  an unchanged exit status. Tools that treat any stderr output as failure
  will see them.
- On `examples/gno.land/r/...` (61 realms) the hints report one line:
  `r/tests/vm/interfaces.gno`, a `Render` concatenating in a loop.
- `gnoUnrecoverableNativeHint` describes current behaviour. If the banker
  binding moves to `m.Panic`, the hint should be removed.
