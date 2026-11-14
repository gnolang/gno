# ADR: Convert multi-value assignment results at run time, not by splicing

## Context

Go lets `m, ok = fetch()` assign a result of type `NamedMap` to a variable of
type `map[string]int` (or the reverse): the types differ only by name. Gno
represents a value's type on the value, so the result has to be converted to
the Lhs type, or `m` carries `NamedMap` as its dynamic type. A single-value
assignment gets this from `checkOrConvertType`, which wraps the Rhs in a
conversion. A call's results cannot each be wrapped, so the preprocessor
used to "decompose" the statement instead:

```go
.decompose_3_0, .decompose_3_1 := fetch()  // in place of the original
m, ok = .decompose_3_0, .decompose_3_1     // spliced in after it, converting
```

The splice went into `last.GetBody()` after `body[index]`. That is the
statement's home only in a plain block body, and the second statement grows
the body by one. Every place that assumption fails was a bug on legal,
type-checked code:

| Position | Before |
|---|---|
| `for m, ok = f(); cond; post` | Silent miscompile: the converting half lands in the loop body, so `cond` and the first iteration see zero values. |
| `for ; cond; m, ok = f()` | The spliced half reads `.decompose_*` slots before they are defined. |
| `if`/`switch` init | Internal panic at preprocess ("IfStmt has no body"). |
| Any method body | Silent miscompile: the assignment never runs. Method `FuncValue`s keep the body slice taken before the splice, and `PrepareNewValues` refreshes only `pn.Values`. |
| Forward `goto` over it | Lands one statement early: its `BodyIndex` was resolved before the body grew. |
| A labeled statement | The label is lost (`n.Attributes = dsx.Attributes`), so `goto` to it fails to compile. |
| REPL top level | Internal panic ("PackageNode has no body"). |
| `arr[i()], ok = f()` | `f()` runs before `i()`: the call moved into the first statement, ahead of the Lhs operands. |

The first version of this change refused the statement in control-flow
init/post positions. Review found that this did not close the class (the
method, goto and label rows are plain body positions) and that the refusal
would also fire in `VMKeeper.Initialize`, which re-preprocesses every stored
package at node start: one stored realm with the for-init shape, which the
current chain accepts, would stop an upgraded node from booting.

## Decision

Keep the statement as it is and move the conversion to run time.

- **Preprocess** (`AssignStmt`, ASSIGN, `len(Lhs) > len(Rhs)`, call Rhs):
  for each result whose type `isNamedConversion` to its Lhs type, record the
  Lhs type in `ATTR_ASSIGN_CONVERT` (`[]Type`, one entry per Lhs, nil for
  none). As in `checkOrConvertType`, an interface Lhs, named or not, gets no
  conversion: the result is boxed. Nothing is spliced, and the
  `.decompose_*` names are gone.
- **Run time** (`doOpAssign`, multi-value path): before writing any Lhs,
  convert each flagged result with `convertValue`. The conversion happens
  in place on `m.Values`, before the results are popped. The copy it makes
  can allocate and so trigger GC, and GC counts live bytes from `m.Values`.
  The spliced form likewise converted while the earlier results were still
  on the stack.
- **`convertValue`** is the body of `doOpConvert`, extracted unchanged, so
  the implicit conversion is charged the same gas and runs the same
  inter-realm checks as an explicit one ("illegal conversion of readonly or
  externally stored value", "illegal conversion to external realm type").
  The spliced form also went through `doOpConvert`, so neither the checks nor
  the charge change.
- **Sealing**: `sealExprTypes` seals the `ATTR_ASSIGN_CONVERT` types, as it
  does the other attribute-held types concurrent queries read. The spliced
  form held them in `constTypeExpr`s, which were sealed the same way.

Block nodes are cache-only (`setBlockNode` never writes them to the backend),
and `Attributes.data` is not amino-encoded, so the attribute adds nothing to
persisted state. It follows `ATTR_REF_ELEM_TYPE` and `ATTR_TYPEOF_VALUE`,
which the runtime already reads off nodes.

### Alternatives considered

- **Refuse init/post positions** (this PR's first version). It leaves the
  method, goto and label miscompiles in plain bodies, gives the REPL a remedy
  it cannot apply, and halts boot on stored code unless boot is made lenient.
  A lenient boot would keep running the miscompiled lowering for stored
  realms.
- **Patch the splice**: refresh method bodies, renumber resolved gotos,
  carry the label over. That is three fixes for three symptoms of one shape,
  and still leaves init/post with no body to splice into.
- **Replace the statement with a `BlockStmt` holding both halves**: keeps
  indexes stable and works in headers, but the operands, already resolved
  against `last`, would be re-preprocessed one block deeper. Anything bound
  to block depth (static parents, a function literal's captures among the
  call's arguments) would have to be re-derived. It is a larger and riskier
  rewrite than recording the types.

## Consequences

- The statement converts correctly in every position, including those that
  were miscompiled or rejected. Code that deployed before runs with Go
  semantics after the upgrade. The only code whose behavior changes is code
  that was miscompiled.
- Nothing is refused, so `VMKeeper.Initialize` has nothing new to panic on.
  `addpkg_multiassign_named_conversion.txtar` restarts the node with the
  for-init shape stored and queries it again.
- Gas goes down for the converting statement and is unchanged elsewhere.
  One statement, one two-name `OpDefine`, and the evaluation of the
  temporaries are gone; the conversion is charged as before. Measured
  through the keeper with a crossing call running the statement in a loop
  (same txtar and loop counts on both binaries; a call counter confirms the
  iterations ran):

  | Call | Base | This change |
  |---|---|---|
  | `Body(0)` | 1,044,390 | 1,044,390 |
  | `Body(1000)`, `m, ok = fetch()` with a conversion | 5,992,014 | 5,264,014 |
  | `Plain(1000)`, the same loop with no conversion | 4,779,538 | 4,779,538 |

  That is 728 gas less per converting assignment: the conversion overhead
  per assignment drops from about 1,212 to 485 gas over the no-conversion
  control. Deploying the package costs 1,074 gas less (3,421,346 to
  3,420,272). The package, deployed with `addpkg` and called with
  `maketx call -func Body -args 1000` (and `Plain`):

  ```go
  type NamedMap map[string]int

  var calls int

  func fetch() (NamedMap, bool)        { calls++; return NamedMap{"a": 1}, true }
  func fetchU() (map[string]int, bool) { calls++; return map[string]int{"a": 1}, true }

  func Body(cur realm, n int) {
      var m map[string]int
      var ok bool
      for i := 0; i < n; i++ {
          m, ok = fetch()
      }
      _, _ = m, ok
  }

  func Plain(cur realm, n int) { /* same loop calling fetchU */ }

  func Count() int { return calls } // 2000 after both 1000-iteration calls
  ```
- The comma-ok forms `v, ok = m[k]` and `v, ok = x.(T)` skipped the
  conversion when `v`'s type differed from the element or asserted type only
  by name. `prxxxx_return_call_conversion.md` closes that with the same
  attribute.
