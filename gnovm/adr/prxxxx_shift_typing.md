# PRxxxx: untyped shift operands take the type of their context

## Context

Spec, Operators: "If the left operand of a non-constant shift expression
is an untyped constant, it is first implicitly converted to the type it
would assume if the shift expression were replaced by its left operand
alone." The preprocessor typed `1 << s` at its own node, so `var x int64 =
('a' << s) + (1 << s)` summed two int32 shifts and `1.0 + 1<<s` was a
float shift refused too early (#6393). Independently, an untyped rune
constant op an untyped int constant folded as int, so `r := 'a' + 1`
declared an int (#6339).

## Decision

- A non-constant shift with an untyped constant operand, and `+ - * / %
  & | ^ &^` or unary `+ - ^` over such shifts, keeps an untyped kind as
  its static type (`ATTR_TYPEOF_VALUE`) until the context is known. A
  comparison types both operands to the kind's default at its own node,
  since its result does not depend on them.
- `bindUntypedOperands` combines two untyped operands by the spec's kind
  order, integer < rune < floating-point (`Specificity`: an operand
  converts to the operand type with the smaller number). A
  rune and an int constant that does not fit int32 fold as int instead,
  since Gno's untyped rune holds int32 (#6389).
- `checkOrConvertType` and `convertType`, given the context type (or
  the kind's default when there is none or it is an interface), hand an
  untyped numeric tree to `pushUntypedType`: go/types' `updateExprType`
  in one pass. Each constant leaf converts to the type with its range
  check, each shift asserts the type is an integer, and each node's
  static type becomes the type, so the shifts run typed. The walk is
  metered per node at `OpCPUSlopeCheckNode`, like the checker walks.
- `&&`/`||` of untyped bools and comparison results keep the series'
  rule: the context's boolean type, or bool (`checkOrConvertType`).
- Errors keep their codes: an unassignable context (`var s string = 1
  << s`) fails `CodeAssignability` before the shift-operand check, so
  goldens that named the shift now name the assignment.

## Upgrade note

Untyped rune and untyped int constants now fold as rune. Stdlib constants
such as `unicode.MaxRune + 1` and `'\n' - 1` persist as rune
`ConstExpr` bytes instead of int, so the genesis app hash moves
(`apphash_crossrealm38_test.go` re-derived); their values are unchanged.
An untyped rune constant still overflows past int32 (#6389).

## Consequences

- `gocorpus/testdata/rune.go` and `fixedbugs/bug445.go` compile clean
  and `goconst/shift_context.go` runs clean; five `types/shift_*`
  goldens name the typed shift (`Float64Kind`) or the assignment.
- `goconst/rune_arith.go` now wraps `'a' * 1000000000` at int32 (#6389,
  folded as rune); `goconst/rune_shift_const.go` pins `'a' << 40`.

Fixes #6393
Fixes #6339
