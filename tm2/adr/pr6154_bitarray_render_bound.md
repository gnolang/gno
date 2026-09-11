# ADR-6154: Bound BitArray rendering by backing data

## Context

`BitArray.stringIndented()` and `MarshalJSON()` iterated `bA.Bits`, the declared bit count,
instead of the actual backing slice `len(bA.Elems)`. `Bits` comes straight from a varint at
amino decode time with no upper bound, and `Bits`/`Elems` consistency is only checked later
by `ValidateBasic()`, so a malformed `BitArray` with a huge `Bits` and a nil or tiny `Elems`
can reach these methods.

## Decision

Bound both render paths by the actual backing data, with behavior appropriate to each contract:

* `stringIndented()` iterates over `min(len(Elems)*64, Bits)` and represents any
  remaining declared bits that have no backing data with a compact `<N unset>`
  token. `String()` implements `fmt.Stringer` and is used for diagnostics, so it
  cannot return an error and must always produce a bounded string. The token
  preserves visibility into the declared size without expanding the missing
  portion.
* `MarshalJSON()` returns an error when `numElements(Bits) != len(Elems)`. This
  mismatch can only result from corrupt or hostile decoding, not from normal use
  through `NewBitArray` and `SetIndex`. Such a value has no meaningful JSON
  representation, so serialization should fail explicitly.

For a well-formed `BitArray` where `len(Elems) == numElements(Bits)`, both paths
produce output that is identical to the previous behavior at the byte level.

## Alternatives considered

Render the `<N unset>` token in `MarshalJSON()` as well, as the originally
proposed patch did. Rejected because `UnmarshalJSON()` accepts only `[_x]*`, so
the output would not round-trip through its own parser. A value that should never
be serialized is better rejected than approximated.

## Consequences

`MarshalJSON()` now returns an error for inconsistent inputs. Callers that marshal only
well-formed `BitArray`s are unaffected.
