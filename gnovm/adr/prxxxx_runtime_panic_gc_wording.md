# Runtime panic text matches gc

## Context

GnoVM and gc panic on the same operations and report the same numbers, but
several runtime sentences differ. [Issue #6538](https://github.com/gnolang/gno/issues/6538)
lists the pairs a differential fuzzer (`misc/gnodiff`, PR #6539) classifies as
`panic-wording`: integer division by zero, a negative shift, an index, a nil
pointer, a nil-slice index, and a slice whose low bound exceeds its high bound.
The fuzzer keeps the numbers and folds only the sentence, so those reports
never mean a behavioral disagreement.

The values, the control flow, and whether an operation panics are already the
same. This change is the sentence, plus which of several failed bounds is
named, so the sentence can be gc's.

## Decision

Use gc's sentence (Go 1.25 `src/runtime/error.go` and `panic.go`) wherever the
event is the same one gc reports.

- Integer `/` and `%`, every width including untyped bigint:
  `runtime error: integer divide by zero`.
- Negative `<<` and `>>`: `runtime error: negative shift amount`, with no
  operand. gc does not print one.
- Nil pointer dereference, including a nil pointer to an array:
  `runtime error: invalid memory address or nil pointer dereference`.
- Index of a string, array, or slice, including a nil slice (length 0):
  `runtime error: index out of range [i] with length n`. A negative index
  omits the length, as gc does: `runtime error: index out of range [i]`.
- Two-index slice. The high bound is checked against len (string, array) or
  cap (slice) before the low bound, which is the order gc's panic functions
  run. A negative bound omits the other number.
  `s[2:1]` is `runtime error: slice bounds out of range [2:1]`.
  `s[1:5]` on a slice of cap 3 is
  `runtime error: slice bounds out of range [:5] with capacity 3`.
  A string or array says `with length`.
- Three-index slice, same order (max, then high, then low). An array says
  `with length`; a slice says `with capacity`.
  `s[0:1:10]` on a slice of cap 3 is
  `runtime error: slice bounds out of range [::10] with capacity 3`.

An in-range slice still succeeds. That includes `nilSlice[0:0]` and
`nilSlice[0:0:0]`. A nil slice has length 0 and capacity 0, so any other
bound fails the same check an empty slice fails. The old
`nil slice index out of range` path was unreachable once those checks ran.

Sentences that are not the same event stay as they are:

- `runtime error: call of nil function` and
  `runtime error: defer called a nil function`. gc prints the generic nil
  pointer sentence. The specific text is the #5711 precedent; folding it
  into the nil-pointer sentence would make two events indistinguishable.
- `runtime error: method selector on nil interface`. PR #5715 rejected gc's
  sentence for this event.
- `value method PKG.T.M called using nil *T pointer`. This is already gc's
  `panicwrap` text (no `runtime error:` prefix). Using it for direct calls
  as well is the #5732 decision.
- `runtime error: uninitialized map index`. A nil-map read does not panic in
  gc, and gc's assignment panic is `assignment to entry in nil map` with no
  `runtime error:` prefix.
- Untyped bigdec division by zero stays `runtime error: division by zero`.
  It is not an integer divide, and gc has no runtime decimal.
- The compile-time diagnostic `invalid operation: division by zero` is
  already gc's constant error.

## Alternatives considered

- **Keep Gno's sentences and teach the fuzzer every synonym.** Rejected: the
  issue asks for the strings to compare equal, and a new wording would need
  another fold.
- **Change only the six sentences in the issue, and leave other slice bounds
  (`s[-1:]`, `s[:5]`, three-index slices) in Gno's wording.** Rejected: those
  are the same bounds checks. A partial edit would still disagree with gc on
  the operation the issue is about, and the negative-high path currently
  prints the low bound.
- **Reorder nothing, and only rewrite the format string around the numbers
  Gno already prints.** Rejected for the high-bound failures: gc prints
  `[:5] with capacity 3`, not `[1:5] with capacity 3`. Keeping the extra
  index would not compare equal. When more than one bound is wrong, naming
  the bound gc names does not change which programs panic.

## Consequences

- Recovered panic text changes. A realm that stores the error, as
  `zrealm_runtimeerror.gno` does, persists the new bytes. Replaying a block
  that recovers one of these panics produces the new sentence on every node
  running this binary.
- Programs that panic still panic, and programs that do not still do not.
  Gas is unchanged: the checks do not meter.
- Filetests and the `gno.land/pkg/sdk/vm` query test that pinned the old
  sentences are updated to the new ones.
