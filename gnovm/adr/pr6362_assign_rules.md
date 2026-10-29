# PR6362: assignability rules at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts. A group of them are assignability rules: type identity
with struct tags, writes into map-held values, and selectors through
declared pointer types. The last also caused an internal "should not
happen" on two corpus files, which made one of them (`issue21934`) look
covered.

## Decision

Checks in `check_assign.go` and the type lookup, on the preprocessed
tree:

- Struct tags take part in type identity where assignability, static
  interface satisfaction (method types, `checkImplementedByStatic`) and
  `copy` element types are decided (`tagsIdentical` beside the tag-blind
  `TypeID`). The runtime assertion path keeps the tag-blind comparison
  and its gas. Left tag-blind for now: `checkSame` (slice, map and array
  element identity), `==` and channel element types; conversions ignore
  tags, as in Go.
- A field or element of a value held in a map is not assignable
  (`m[k].f = v`); the native filetest that relied on go/types refusing
  the statement while Gno ran it moves to its own error filetest.
- A declared pointer type has no method set: a selector reaches fields
  through it but no method, and a method at a shallower depth still
  shadows a deeper field, so the selector is undefined (Go spec,
  Selectors). This replaces the internal error on `issue43384` and
  `issue9017`.

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.
- Treating the shadowing method as selectable through the pointer type:
  diverges from Go and changes which function runs.

## Consequences

- 6 errorcheck corpus files flip to `Clean`, `issue21934` now for a real
  error rather than an internal one.
- Tagged struct types that were identical by accident are now distinct;
  conversions between them need an explicit conversion, as in Go.

Follow-up (2026-10-30): `assertAssignValues` at every one-value-per-target site (assignment, `:=`, `var`, `return`) with gc's "assignment mismatch" wording; fewer targets than values is checked before the walk; argument errors are positioned at the argument.
