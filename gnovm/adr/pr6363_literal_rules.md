# PR6363: declaration and literal rules at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts. A group of them are rules on `:=` redeclaration,
const specs, variadic parameters, composite literals, `for` post
statements and type switch variables.

## Decision

Checks in `check_decls.go`, called from preprocess on the preprocessed
tree:

- A name on the left of `:=` that the block already declares is
  assigned, not redeclared: the value must be assignable to its type and
  an untyped constant converts to it.
- A const spec with a type and no value does not repeat the previous
  spec.
- A variadic parameter must be last.
- An unkeyed struct literal gives every field a value; array literal
  indices stay below the length.
- A keyed struct literal's keys are field names, each bad one reported
  before it is resolved, and name fields of the type.
- A blank field or method is never found through a selector.
- A method's receiver names a type declared in the package: a basic or
  foreign type cannot be given new methods.
- A type declaration whose type is a name names a type, not a constant,
  variable or function, said before that value is predefined.
- A `for` post statement may not declare.
- A type switch variable with no clause is unused.

Two native filetests take the preprocess-time message as their golden.

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.

## Consequences

- 7 errorcheck corpus files flip to `Clean`.

Follow-up (2026-10-30): in accumulate mode each non-final variadic parameter is reported and kept as a slice, so the rest of an interface is still checked.
