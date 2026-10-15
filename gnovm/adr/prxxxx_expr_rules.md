# PRxxxx: expression rules at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts. A group of them are expression rules on slices, maps,
functions, `len`/`cap`, constant indices and composite literals. Gno
reached some of them only at run time (a runtime "unexpected type for
len", a duplicate field at construction), unpositioned, and the rest not
at all.

## Decision

Checks in `check_exprs.go`, called from preprocess on the preprocessed
tree, allocating nothing (gas is unchanged):

- A slice, map or func compares only to nil, also against an interface
  and in an expression switch's cases; a typed nil (`S(nil)`) is not the
  nil literal.
- `len` and `cap` take a string, array, pointer to array, slice, map or
  channel.
- A constant index or slice bound is checked against an array's or a
  constant string's length (`high` may equal the length); slice bounds
  must be ordered.
- An array or slice literal may not repeat an index, a struct literal a
  field name, and a map literal needs every key.

Ten native filetests that reached a runtime error now fail at preprocess
with the positioned message; their goldens follow.

## Alternatives considered

- Keeping the runtime checks: they are unpositioned and never reach
  errorcheck files, which are not executed.
- Leaving the rules to the go/types guard: temporary, and errorcheck
  files never reach it.

## Consequences

- 7 errorcheck corpus files flip to `Clean`.
- Programs that relied on a runtime panic for these shapes are rejected
  at deploy with the same cause, positioned.
