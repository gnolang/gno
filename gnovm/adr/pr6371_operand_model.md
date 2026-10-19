# PR6371: the operand model for the checker rules

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
  allocation, nothing stored on the tree. Its only walk, down the
  selector/index chain of the target, is unmetered: it is bounded by the
  expression's size, which the preprocess walk already charged for.
- A rule that takes an operand rejects a type first (`%s (type) is not
  an expression`, `CodeNotExpression`): an operand's `typ` is the type a
  type expression denotes, so a rule reading `typ` alone would accept
  `len([3]int)`.
- `assertComparableTo`, `assertSwitchTag`, `assertCopyArgs`,
  `assertExprStmtUsed`, `assertValidAssignLhs`, `assertLenCapArg` and the
  literal key rule read an operand's mode and type. The assignment target
  rule is "a variable or a map index expression", so an element of a call
  result is refused too (`bug278`), and an unassignable string element
  says why.
- The remaining rules move out of `type_check.go` into the chapter files
  (`check_assign.go`, `check_exprs.go`, `check_consts.go`); it keeps what
  the VM needs at runtime. The move commit is byte-identical (goldens and
  both corpora untouched); the rule commits flip five goldens for the
  assignment target wording: `types/assign_literal7c.gno`,
  `types/assign_op_b.gno`, `gocorpus/testdata/cannotassign.go`,
  `gocorpus/testdata/errorcheck/cannotassign.go`,
  `gocorpus/testdata/fixedbugs/bug278.go`.

## Alternatives considered

- Storing the mode on the expression node: costs memory on every tree
  and persists what is only needed while checking.

## Consequences

- Rules read like go/types' and new ones have one vocabulary.
- Goldens follow the assignment target wording; a method value and a
  call result's field are refused as targets, as gc does.
