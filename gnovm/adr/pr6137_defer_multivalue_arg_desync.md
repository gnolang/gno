# PR6137: Fix `doOpDefer` value-stack desync for `defer f(g())` multi-value arguments

## Context

Deferring a call whose single syntactic argument is a multi-value function
call — the `f(g())` spread form, where `g` returns exactly the arguments `f`
expects — behaved incorrectly and diverged from Go.

Minimal reproductions (verified against `go run`):

```go
// Case 1: multi-value args spread into an ordinary function.
func twovals() (int, int) { return 7, 8 }
func h(a, b int)          { println("h called:", a+b) }
func main()               { defer h(twovals()) } // Go: "h called: 15"
```

```go
// Case 2: the first spread result is itself a func value.
func marker()            { println("WRONG") }
func g() (func(), int)   { return marker, 99 }
func f(fn func(), n int) { println("CORRECT n=", n) } // does NOT call fn
func main()              { defer f(g()) } // Go: "CORRECT n= 99"
```

Before the fix the GnoVM ran case 1 into a spurious
`runtime error: defer called a nil function` panic, and case 2 executed the
*wrong* function — it ran `marker` (a return value of `g`) and never ran `f`.

### Root cause

At preprocess time a `CallExpr` records `NumArgs`, "the number of argument
*values*", which is `len(Args)` in the common case but `len(Args[0].Results)`
for the `x(f())` spread form (`nodes.go`, `CallExpr.NumArgs`). The `DeferStmt`
executor (`op_exec.go`) evaluates the func expression and then each arg
expression, so for `defer f(g())` the value stack ends up as

```
[ ..., f, r0, r1 ]      // g() expanded to its two results r0, r1
```

`doOpDefer` (`op_call.go`) located the func value by peeking past the arguments:

```go
numArgs := len(ds.Call.Args)   // == 1 for f(g())  — WRONG
ftv := m.PeekValue(numArgs + 1) // peeks r0, not f
```

With `len(Args) == 1` it peeked `PeekValue(2)` = `r0` (an argument value)
instead of `PeekValue(3)` = `f`, then handed the wrong `numArgs` to
`popCopyArgs`, popping the wrong number of values. Depending on the runtime
kind of `r0` this manifested as either the nil-func panic (case 1: `r0` is an
`int`, whose `.V` is nil, so the `case nil` branch fired) or as running an
unrelated function value (case 2: `r0` is `marker`).

### Secondary defect: the `case nil` branch leaked stack values

The `case nil` branch (deferring a nil func value, e.g.
`var g func(int,int); defer g(1,2)`) pushed the deferred entry and then relied
on the single trailing `m.PopValue()` to clean the stack — but never consumed
the evaluated argument values, unlike the `*FuncValue` and `*BoundMethodValue`
branches which pop them via `popCopyArgs`. This left the argument values on the
value stack. The leak was *latent*: a function's frame truncates
`m.Values` back to its recorded height on return (`PopFrameAndReturn`), and
function results are read from the top of the stack, so the stray values below
were discarded before they could be observed — `CheckEmpty` (only run for
no-output/no-error filetests) therefore did not catch it either. It is still an
invariant violation (`doOpDefer` must be value-stack-neutral) and would grow the
stack unboundedly if such a defer ran in a loop, so it is fixed in the same
change.

## Decision

Use `ds.Call.NumArgs` — the expanded argument-value count the preprocessor
already computed — for both the `PeekValue` offset and the `popCopyArgs` count:

```go
numArgs := ds.Call.NumArgs
ftv := m.PeekValue(numArgs + 1)
```

`NumArgs` equals `len(Args)` for every non-spread call (including all
deferrable builtins and methods), so ordinary defers are unaffected; it differs
only for the `x(f())` spread form, which is exactly the broken case.

In the `case nil` branch, discard the evaluated arguments explicitly so the
branch is value-stack-neutral like the others:

```go
m.PopValues(numArgs)
cfr.PushDefer(Defer{Source: ds})
```

## Alternatives considered

- **Re-derive the count in `doOpDefer`** (e.g. `len(Args[0].Results)` when
  `len(Args)==1`): redundant — the preprocessor already encodes this in
  `NumArgs`, which is the single source of truth used by the ordinary call path
  (`op_call.go` precall) too. Reusing it keeps defer and call consistent.
- **Call `popCopyArgs` in the `case nil` branch** to drop the args: works only
  if `ftv.T` is a usable `*FuncType`; a plain `m.PopValues(numArgs)` is simpler,
  needs no type assertion, and matches the branch's intent (the copied args are
  never used — the deferred call raises call-of-nil regardless).

## Consequences

- `defer f(g())` with multi-value spread now matches Go for both an ordinary
  callee (case 1) and a callee whose spread arguments include func values
  (case 2). No change to any non-spread defer.
- The `case nil` branch no longer leaks argument values onto the value stack.
- Regression coverage added under `gnovm/tests/files/`:
  - `defer_multivalue_arg.gno` — cases 1 and 2; panics/misbehaves before the
    fix, passes after.
  - `defer_nil_func_args.gno` — nil-func-with-args branch; asserts the
    call-of-nil is recovered and execution continues correctly.

## Verification

- New filetest `defer_multivalue_arg.gno` fails on the pre-fix VM (panics in
  `doOpReturnCallDefers`), passes after the fix.
- `go test ./gnovm/pkg/gnolang/ -run Files -test.short` — 0 failures.
- `go test ./gno.land/pkg/sdk/vm/ -run Gas` — ok.
- `go test ./gno.land/pkg/integration/ -run TestTestdata` — ok.
- `cd examples && go run ../gnovm/cmd/gno test ./...` — ok.
