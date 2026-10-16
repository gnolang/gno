# PRxxxx: type rules at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts. A group of them are rules on types: map keys,
embedded fields, interface elements, array lengths, switch clauses, type
assertions and `copy`. Some errors preprocess did raise were positioned
at the whole declaration or statement rather than the offending node.

## Decision

Checks in `check_types.go`, called from preprocess on the preprocessed
tree:

- A map key type must be comparable, checked at the key and re-checked
  once a declared key type is complete.
- An embedded field may not be a pointer to an interface or a pointer to
  a pointer; an interface may not embed a non-interface; a func-typed
  field is not a method.
- An array length may not be negative; `[...]` is valid only in a
  composite literal.
- An expression switch may not repeat a constant case, nor a `default`;
  a type switch needs an interface operand and may not list an
  impossible case; an impossible type assertion is rejected.
- `copy` takes a slice destination and a slice or string source of the
  same element type.
- Errors can be positioned at a sub-node (`panicAt`). A declaration
  cycle is reported at the declaration that opens it, naming the cycle
  only, as go/types does; that commit overlaps #6254, which replaces the
  detection, and is dropped when this branch is rebased on it.
- Accumulate mode keeps a statement after a reported duplicate case and
  drops a failed dependency declaration only once.

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.
- Positioning every error at the statement: simpler, but gc's markers
  sit on the offending line, and users read positions.

## Consequences

- 24 errorcheck corpus files flip to `Clean`; seven native filetests take
  the positioned message as their golden.
- `fixedbugs/bug336.go` (run, #6254) keeps its verdict; its error now
  points at the dependency.
