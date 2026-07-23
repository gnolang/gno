# Cache the interface-comparison verdict at preprocess (ATTR_IFACE_CMP)

## Status

Implemented on `perf/maxwell/eql_iface_preprocess_attr`. Follow-up to #5713
(runtime panic on comparing uncomparable types via interface). Supersedes the
opcode-specialization attempt, which a full-cycle benchmark showed to be a tie
with #5713 — see "Why not opcodes".

## Context

For `==`/`!=`, Go panics when the operands are interface-typed and the dynamic
type is uncomparable. The static verdict (is an operand interface-typed?) cannot
be replaced by a runtime check — `var s []int; s == nil` (legal) and
`any(s) == any(s)` (panics) reach `isEql` with the byte-identical operand pair
`{T: []int, V: nil}`. So the verdict must travel from the preprocessor.

#5713 carries it by recomputing it on **every evaluation**: `doOpEql`/`doOpNeq`
call `isInterfaceCmp(bx)`, which does two `ATTR_TYPEOF_VALUE` attribute-map
lookups (one per operand). In a loop, that runs every iteration and is the
dominant cost of the feature — far larger than the `.(*BinaryExpr)` assertion
the opcode attempt targeted.

## Decision

Compute the verdict **once at preprocess** and cache it as an attribute.

- In `preprocess.go`'s `*BinaryExpr` `TRANS_LEAVE`, where the operand static
  types `lt`/`rt` are already resolved, set `ATTR_IFACE_CMP=true` for `EQL`/`NEQ`
  when `isInterfaceStaticType(lt) || isInterfaceStaticType(rt)`. Set only when
  true, so a present-and-`true` attribute is the verdict.
- `doOpEql`/`doOpNeq` read `bx.GetAttribute(ATTR_IFACE_CMP) == true` — one map
  lookup instead of #5713's two-lookups-plus-`baseOf`.

The same cache covers the `switch`-tag path: a non-type-switch whose tag is
statically an interface compares each case like an interface equality, so its
`*SwitchStmt` gets `ATTR_IFACE_CMP` at preprocess and `op_exec` reads it per
clause comparison. With both call sites converted, the per-evaluation helpers
`isInterfaceCmp(*BinaryExpr)` and `hasInterfaceStaticType(Expr)` are removed;
`isInterfaceStaticType(Type)` is the single preprocess-time predicate.

### Why an attribute, not a struct field

A struct field on `BinaryExpr` read per-eval is ~0 ns and reaches feature-off
parity, but it mixes a runtime-perf cache into the AST node's syntactic
definition. There is precedent for preprocess-derived struct fields
(`NameExpr.Path`, `NameExpr.Type`), but that is the exception, not the pattern
to extend; the codebase's standard home for preprocess-derived metadata is the
attribute map (`ATTR_TYPEOF_VALUE`, `ATTR_SHIFT_RHS`, …). The attribute keeps
`BinaryExpr` syntactically clean and matches that idiom. The measured cost of
this choice is ~2.2% vs the struct field (one residual per-eval map lookup) —
accepted for the cleaner placement.

### Persistence

Attributes are not amino-persisted. Correctness across a node restart relies on
packages being re-preprocessed on load (store.go: "Upon restart, all packages
will be re-preprocessed") — the exact lifecycle the operands' `ATTR_TYPEOF_VALUE`
already depends on, which is how #5713's runtime reads work at all.
`TestBinaryExprIfaceCmp_SurvivesColdReload` guards this: it persists a realm,
reloads it into a cold store via the restart re-preprocess protocol, and asserts
the uncomparable-comparison panic still fires (teeth-checked: fails when the
preprocess set is disabled).

## Results (Apple M3, n=15 interleaved, benchstat, program-level RunMain loop)

vs `pre` (feature off, no check):

| file | #5713 | opcode | attr (this ADR) | struct field |
|---|---|---|---|---|
| concrete `==` | +8.5% | +7.1% | **+2.3%** | ~0% (n.s.) |
| interface `==` | +5.6% | +3.7% | **+2.2%** | ~0% (n.s.) |

The attribute beats both #5713 and the opcode design on both paths, with no new
opcodes and no struct change. The struct field would recover the last ~2.2% but
at the cost of AST-node cleanliness (see above).

### Why not opcodes

The opcode attempt (`OpEqlIface`/`OpNeqIface` selected in `doOpEval`) moved
`isInterfaceCmp` from the handler to `doOpEval` — but `doOpEval` also runs every
evaluation, so it relocated the cost rather than removing it (full-cycle geomean
tie with #5713). Its original "−22%" came from a benchmark that timed `doOpEql`
in isolation, hiding the relocated lookup in the untimed `doOpEval`. It also
doubled the `==`/`!=` op surface.

## Consequences

- Per-evaluation `==`/`!=` cost drops from two attribute lookups to one,
  cutting the feature's overhead by ~3-4× on the common path while preserving
  the interface-comparison panic.
- The `switch`-tag comparison path gets the same treatment in one change, so no
  per-evaluation interface-type recomputation remains at the top-level `==`/`!=`
  and switch-tag boundaries. (The recursive array-element/struct-field interface
  checks in `isEql` still resolve per-eval — they walk values, not Exprs, so they
  have no AST node to cache on; their cost is a `baseOf(t).Kind()`, not the
  attribute-map lookups this cache eliminates.)

## Verification

- 20 `tests/files/types/cmp_uncomp_*` filetests pass (behavior unchanged,
  including the function-call-returns-interface operand case).
- `TestBinaryExprIfaceCmp_SurvivesColdReload`.
- Gas, `TestTestdata` (txtar), and `Files -short` suites green.
- Benchmarks: `benchdata/cmp_concrete.gno`, `benchdata/cmp_iface.gno`.
