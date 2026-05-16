# Meter BigInt and BigDec comparison operators

## Context

After PR #5291 (gas-model calibration), per-N CPU gas was wired into the BigInt
arithmetic ops (`+ - * / % | ^ & &^ << >>`) and into the BigDec arithmetic ops
(`+ - * /`). Comparison operators were only partially covered.

In `gnovm/pkg/gnolang/op_binary.go` (master state):

| Op  | BigInt slope | BigDec slope |
|-----|--------------|--------------|
| `==` | `OpCPUSlopeBigIntEql` upfront in `doOpEql`        | none |
| `!=` | none — `doOpNeq` delegates to `isEql` only        | none |
| `<`  | `OpCPUSlopeBigIntLss` upfront in `doOpLss` (renamed to `…Cmp` here) | none |
| `<=` | none                                              | none |
| `>`  | none                                              | none |
| `>=` | none                                              | none |

Master also recently changed the `is*` helpers (`isEql`, `isLss`, `isLeq`,
`isGtr`, `isGeq`) to take `*Machine` and moved the per-N gas for
`string`/`Array`/`Struct` comparisons into those helpers.

### Why this metering is defensive (and dead today)

`UntypedBigintType` and `UntypedBigdecType` are internal preprocess types used
to represent arbitrary-precision integer/float literals before they get
assigned a concrete Go type. They are **not user-facing**:

- No `bigint`/`bigdec` keyword in Gno; users cannot declare variables, fields,
  or function params/returns of these types.
- Every binary expression with both operands constant gets folded into a
  `ConstExpr` at preprocess (`preprocess.go:1397-1425`).
- Any const-vs-non-const expression converts the constant to the non-const
  operand's type, losing the bigint-ness.
- The native bridge panics with "not yet implemented" on `UntypedBigintType`
  (`gonative.go:189-192`).

Empirical confirmation: `_ = (1 << 10) == (1 << 10)` and
`_ = (1 << 200) == (1 << 200)` consume identical gas (2250) — the entire
expression collapses to `ConstExpr{V: true}` before machine execution. Larger
shifts (`1 << 1000`+) are rejected by the typechecker. There is no `.gno`
source pattern that produces a runtime `BigintKind == BigintKind` compare.

The same is true for the **arithmetic** bigint metering already in master: it
charges no gas in any reachable user path. The metering is purely defensive
against:

1. A future user-facing `bigint`/`bigdec` type (proposed but not landed).
2. Crafted bytecode bypassing source-level type reconciliation.
3. Future preprocess refactors that defer folding to runtime.

## Decision

Mirror the master arithmetic pattern: charge per-N gas for BigInt/BigDec
operands inside the `is*` helpers. This closes the gap between arithmetic and
compare metering — both will be defensively wired even though neither is
reachable from `.gno` source today.

- `isEql` (`==` and `!=`): add `incrCPUBigInt(...OpCPUSlopeBigIntEql)` to the
  `BigintKind` case and `incrCPUBigDec(...OpCPUSlopeBigDecEql)` to the
  `BigdecKind` case. `doOpNeq` automatically inherits the charge because it
  calls `isEql`.
- `isLss/Leq/Gtr/Geq` (`<`, `<=`, `>`, `>=`): each gets
  `incrCPUBigInt(...OpCPUSlopeBigIntCmp)` and
  `incrCPUBigDec(...OpCPUSlopeBigDecCmp)` in the BigInt/BigDec branches.
  A single `Cmp` constant is shared by all four lex ops on each side because
  the underlying `Cmp()` work is identical regardless of the operator.
- Rename master's `OpCPUSlopeBigIntLss` → `OpCPUSlopeBigIntCmp`. The fit
  value (`= 9`) is unchanged — it was measured against `Lss` because that
  was the only caller in master, but `big.Int.Cmp()` is comparator-invariant
  so the same per-bit cost applies to `<=`, `>`, `>=` as well. Renaming
  avoids the misleading sight of `OpCPUSlopeBigIntLss` at `isGeq` call
  sites, and mirrors the new `OpCPUSlopeBigDecCmp` exactly.
- Switch the type gate inside `incrCPUBigInt`/`incrCPUBigDec` and friends
  (the `Quad` and `Unary` variants too) from `lv.T == UntypedBig{int,dec}Type`
  to `baseOf(lv.T) == UntypedBig{int,dec}Type`. Strict-equality gating
  silently skipped the charge on any `*DeclaredType` wrapping the underlying
  primitive — which is exactly the future scenario the metering claims to
  defend against. `baseOf` is what the corresponding `*Assign` helpers
  already use to dispatch the computation, so the metering surface now
  matches. Also applied to the inline gate in `doOpShl`.
- Add two new constants in `machine.go`:
  - `OpCPUSlopeBigDecEql` — used by `==` / `!=`.
  - `OpCPUSlopeBigDecCmp` — used by `<` / `<=` / `>` / `>=`.

  Both are placeholders set to `20`, reusing the BigDec `Sub` fit. The split
  mirrors the BigInt `Eql`/`Lss` split so the two sides can be calibrated
  independently if/when the runtime path becomes reachable.

## Alternatives considered

1. **Charge upfront in each `doOp*` handler** (the original plan, before the
   `is*` signature change). Rejected: duplicates the type switch between
   `doOpEql`/`doOpNeq` and forces parallel maintenance with the per-element
   charges that already live inside `isEql`. Once `is*` takes `*Machine`, a
   single charge inside the helper is the natural place.

2. **Single shared `OpCPUSlopeBigDecCmp` for all six ops** (no separate
   `Eql`). Rejected for symmetry with BigInt: existing code already has
   distinct `OpCPUSlopeBigIntEql` (= 10) and (post-rename)
   `OpCPUSlopeBigIntCmp` (= 9) constants, and `==`/`!=` may want a different
   calibration once benchmarks land.

3. **Define separate `OpCPUSlopeBigIntLeq/Gtr/Geq` constants** instead of
   sharing one `Cmp`. Rejected: `Cmp()` is the same function call regardless
   of comparator, so the per-bit cost is the same.

4. **Skip the BigDec metering** since runtime paths are unreachable today.
   Rejected for parity: master already meters BigDec arithmetic in the same
   unreachable manner. Asymmetry between arithmetic and compare would create
   confusion without solving the underlying "is dead defensive metering worth
   keeping" question — see Consequences.

5. **Remove all bigint/bigdec metering as dead code** (this PR plus deletion
   of master's arithmetic metering). Out of scope. If the project decides
   defensive metering is not worth keeping, it should be removed everywhere
   in a separate PR; this PR aligns with the existing pattern.

## Consequences

- All six comparison operators (`== != < <= > >=`) now have per-N gas wired
  for BigInt and BigDec operands, matching the metering pattern of the
  arithmetic ops.
- Charge lives in one place (`is*` helpers); `doOpNeq` is metered for free
  via `isEql`.
- Side benefit: `switch x { case y: ... }` on a runtime BigInt/BigDec value
  is also covered, since `doOpSwitchClauseCase` calls `isEql`.
- **No automated test for the new charges.** A `.gno` filetest cannot
  exercise the path (every BigInt/BigDec compare is folded at preprocess, as
  verified empirically). A Go-level unit test that constructs
  `TypedValue{T: UntypedBigintType, ...}` directly would work — that is the
  approach used by the calibration benchmarks for the arithmetic side
  (`bench_ops_test.go`) — but is not added here, in line with the existing
  arithmetic metering which also has no runtime test.
- **Placeholder slopes are not calibrated.** `OpCPUSlopeBigDecEql` and
  `OpCPUSlopeBigDecCmp` are set to `20` (BigDec `Sub` fit). If/when the
  runtime path becomes reachable (e.g., a `bigint` user type lands), these
  values must be calibrated against a dedicated `BenchmarkOpCmp` family in
  `gnovm/cmd/calibrate/`. The same caveat applies to the existing arithmetic
  bigint slopes when they first become reachable.
- **Open question for triage:** is dead defensive metering worth keeping
  across the codebase? Either remove all bigint/bigdec metering (arithmetic
  + compare) as dead code, or accept it as future-proofing and commit to
  recalibration when the runtime path opens. This PR takes the second path
  by extending the existing pattern; the broader question is out of scope.
