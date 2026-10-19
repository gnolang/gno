# PR6369: method expressions through embedded fields

## Context

`T.m` where `m` is promoted through an embedded field failed at
preprocess with "unknown *DeclaredType method"; `(*T).m` with a value
receiver was refused too. Go accepts both (#5787). Six corpus files and
the harness's KnownIssue canary pinned the gap.

## Decision

- `T.m` for a promoted `m` lowers to a function literal that selects `m`
  on its receiver, so the call site is unchanged and the promoted method
  runs as in Go.
- `(*T).m` with a value receiver takes the same path and has type
  `func(*T, ...)`.
- `T.m` needs `*T` when `m` has a pointer receiver; an unknown method is
  reported as such, positioned.
- A pointer to a named pointer or interface type (`*P3` for
  `type P3 *struct{T}`, `*error`) exposes no field or method, as `**T`
  and `*I` already did; the walk used to reach through it and panic.
- An undefined selector is reported in go/types' words, `x.f undefined
  (type T has no field or method f)`, also on a type with no fields or
  methods at all (was an internal "unexpected selector expression type").
- The KnownIssue canary, which existed to show the verdict flow on a real
  bug, moves to the builtin `print` operand spacing (#6331), still open.

## Alternatives considered

- Synthesising a bound method value in the VM: more machinery for the
  same observable behaviour; the literal reuses selector lowering.

## Gas

The embed walk is metered as for any selector. Preprocessing the
literal evaluates its type expressions on the machine, so the lowering
is metered per op, O(params + results) (`TestPreprocessGas_PromotedMethodExpr`
pins the growth); no separate charge is added.

## Consequences

- Six `#5787` corpus files and one errorcheck file flip to `Clean`;
  `method7.go` keeps `#5787` for a method value on an interface-typed
  field, still open.
- Pointer methods on non-addressable values (`T{}.pm()`, `conv(x).pm()`)
  stay with #6083, which inserts the `&x` under Go's addressability rule.
