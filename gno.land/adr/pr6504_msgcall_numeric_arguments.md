# ADR-pr6504: Canonicalize MsgCall numeric arguments

## Status

Implemented.

## Context

`MsgCall` accepts numeric arguments as strings, but non-canonical integers and
special or signed-zero floats can give ambiguous values or bypass ordinary
range checks at a consensus boundary.

## Decision

Accept integers only in their canonical decimal form, reject non-finite floats,
and normalize zero-valued floats to positive zero in the shared argument
conversion path.

## Alternatives considered

Handling only the reported spellings or leaving validation to each realm would
preserve more existing inputs, but would leave equivalent bypasses or repeat a
consensus-critical rule across callers.

## Consequences

Some previously accepted argument strings now fail, so this is a hardfork change
that requires a coordinated upgrade. No transaction on gnoland-1 up to height
1,361,900 is affected: its 140,323 `MsgCall` messages pass no NaN, Inf or
negative zero, and their 918 leading-zero arguments all reach `string`
parameters. Ordinary numeric values remain reachable through canonical integers
and finite floats.
