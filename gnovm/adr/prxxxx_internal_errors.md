# PRxxxx: internal errors become diagnostics

## Context

The errorcheck ledger (#6346) stopped counting an internal error as
coverage: a VM failure landing on a marked line (`should not happen`,
`unexpected selector expression type`, a name printed with its value
path) had made thirteen files look Clean. These are crashes the checker
reaches on invalid programs, where gc reports a positioned error.

## Decision

Each crash site raises a checking error with a code instead:

- `.(type)` outside a type switch: `use of .(type) outside type switch`
  (`CodeTypeSwitchTag`).
- A selector on a value with no fields or methods (a primitive, a func,
  a slice, a map, a channel), in expressions and in constant
  declarations: `x.m undefined (type T has no field or method m)`
  (`CodeMissingField`); a method expression on such a type says so.
- A pointer to an interface never implements an interface (Go spec,
  Method sets): `cannot use *I as I: *I is pointer to interface, not
  interface`, checked before the implementation walk that crashed.
- An init-expression count mismatch names the variable plainly.

Messages spell the expression as written (`exprString`), not with the
value paths of the preprocessed tree.

## Alternatives considered

- Keeping the panics and filtering them in the harness only: the
  programs are rejected either way, but a deploy would see an internal
  message and the ledger would still not know which rule applied.

## Consequences

- The internal-error lines the ledger lists drop from 26 to the cascades
  accumulate mode produces after dropping a declaration, which the
  accumulate-mode seam PR takes; those files keep `KnownIssue`.
- Native filetests that pinned a crash message take the diagnostic.
