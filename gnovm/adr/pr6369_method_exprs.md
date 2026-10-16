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
- The KnownIssue canary, which existed to show the verdict flow on a real
  bug, moves to the builtin `print` operand spacing (#6331), still open.

## Alternatives considered

- Synthesising a bound method value in the VM: more machinery for the
  same observable behaviour; the literal reuses selector lowering.

## Consequences

- Six `#5787` corpus files and one errorcheck file flip to `Clean`;
  `method7.go` keeps `#5787` for a method value on an interface-typed
  field, still open.
