# PRxxxx: programs Go accepts that the checker refused

## Context

Both corpora list over-strict lines: programs gc accepts that Gno's
preprocess rejects. Five causes covered them.

## Decision

- `&&` and `||` of untyped bools stay untyped, and a comparison result
  takes a named bool context (`type B bool; var b B = x < y`), which
  also closes #6338 (an untyped bool loop variable).
- A constant declaration's `iota` reaches a literal type preprocessed on
  its own.
- `fallthrough` may be followed by empty statements.
- A switch case across an interface boundary is checked for
  implementation, not converted.
- Calling a non-function is reported as such instead of an internal
  error.

## Alternatives considered

- Relaxing the rules at the corpus instead (tagging the files): the
  files are valid Go, and users hit the same rejections.

## Consequences

- Four errorcheck corpus files and one compile file (`issue63462`, #6338)
  flip to `Clean`.
