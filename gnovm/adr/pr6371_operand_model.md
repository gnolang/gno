# PRxxxx: the operand model for the checker rules

## Context

The rule files reached for an expression's nature case by case: nil-ness,
builtin-ness, addressability, constant-ness, each dug out of the tree at
the call site. go/types answers "what is this expression" once, with an
operand mode. `type_check.go` also still held rules that belong with
their chapters.

## Decision

- `operand{mode, typ, expr, val}` is one answer to "what is this
  expression": constant, nil, type, built-in, variable (addressable), map
  index, value, tuple, no value, package, or invalid. `operandOf` builds
  it at a rule's call site from what the preprocessor already cached: no
  allocation, nothing stored on the tree.
- `assertComparableTo`, `assertSwitchTag`, `assertCopyArgs`,
  `assertExprStmtUsed`, `assertValidAssignLhs`, `assertLenCapArg` and the
  literal key rule read an operand's mode and type. The assignment target
  rule is "a variable or a map index expression", so an element of a call
  result is refused too (`bug278`), and an unassignable string element
  says why.
- The remaining rules move out of `type_check.go` into the chapter files
  (`check_assign.go`, `check_exprs.go`, `check_consts.go`); it keeps what
  the VM needs at runtime. A move, nothing else: goldens and both corpora
  are byte-identical.

## Alternatives considered

- Storing the mode on the expression node: costs memory on every tree
  and persists what is only needed while checking.

## Consequences

- Rules read like go/types' and new ones have one vocabulary.
- Goldens follow the assignment target wording; the corpora do not move.
