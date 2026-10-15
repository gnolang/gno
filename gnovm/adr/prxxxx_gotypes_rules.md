# PRxxxx: rules the go/types suite asked for

## Context

The go/types corpus (#6346's second suite) named rules gc's corpus did
not: type parameters silently dropped at conversion, `:=` left sides,
switch tags, `...` in conversions, conversion cycles, pointer
assignability, and an invalid path escaping accumulate mode as an
internal error.

## Decision

Checks in `check_types.go` and preprocess, on the preprocessed tree:

- Type parameters are rejected at conversion instead of silently
  dropped, so generic files route to `Unsupported generics` on their own
  error.
- The left side of `:=` holds names, none repeated.
- A switch tag is comparable and not untyped nil (a slice, map or func
  tag is allowed).
- No `...` in a conversion; a conversion to a type still being declared
  through that constant is a cycle.
- Pointer types are assignable only with identical elements.
- A dropped declaration's invalid path no longer escapes accumulate
  mode.

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.

## Consequences

- 3 errorcheck corpus files flip to `Clean`, 7 to `Unsupported`;
  `fixedbugs/issue4458.go` was Clean only through the internal error and
  is now an honest `KnownIssue`.
