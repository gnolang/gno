# Unmetered big-number DoS via the free `vm/qeval` query path

## Status

Implemented. AI-assisted (see AGENTS.md).

## Context

`vm/qeval` / `vm/qeval_json` evaluate an attacker-supplied expression under a
named package (`withQueryEvalMachine`, keeper.go:1379: `ParseExpr` then
`Eval`). These queries are **free**, **unauthenticated**, and have no timeout
that cancels the handler goroutine (`WriteTimeout: 30s` only bounds the
response write). A real node caps the request body at **1 MB** and concurrent
connections at **900**: `node.go:803-806` starts from `rpcserver.DefaultConfig()`
(`MaxBodyBytes: 5 MB`, `MaxOpenConnections: 0`) but immediately overwrites both
from `config.RPC` (`DefaultRPCConfig`: `MaxBodyBytes: 1 MB`,
`MaxOpenConnections: 900`). An earlier revision quoted the lib-server defaults
as if they were effective; the 1 MB figure is the one that matters, because it
is what bounds the largest literal an attacker can deliver (see Consequences).
Unlike `AddPackage`/`Run`, this path does **not** call
`chargePreprocessGas` (keeper.go:590), so the gas meter
(`maxGasQuery = 3e9`) is the only backstop.

Three operations were mis-metered:

1. **Integer literal parse.** `doOpEval` INT case calls
   `big.Int.SetString(s, 10)`, O(n²) in digits (measured M-series, parse only,
   order of magnitude: 700K ~0.3 s, 1M ~0.6 s). The only charge was the flat
   `OpCPUEval = 82`, and
   the parser bills a giant literal as one token, so
   `<realm>.<700K-digit-int>` cost ~84 gas for hundreds of ms of CPU.
2. **Float literal parse.** `doOpEval` FLOAT case → `parseBigdecLiteral`
   (op_binary.go:877) runs `big.ParseFloat` and then, unless the
   `MantExp > ratOverflowBits` guard fires, `big.Rat.SetString` as well, both
   O(n²). Appending `.0`/`e1` routes an integer here, so an INT-only fix is
   bypassable.

   **The guard bounds the parsed value's magnitude, not the literal's length.**
   For an *int-shaped* float (`999…9.0`) magnitude grows with length, so
   `MantExp ≈ digits·log₂10` crosses `ratOverflowBits = 4096` at ~1234 digits
   and the Rat parse is skipped above that. For a *frac-shaped* float
   (`1.999…9`) the magnitude stays near 1, so `MantExp` stays ~1–2 however long
   the literal is, the guard never fires, and `Rat.SetString` runs on the full
   mantissa. Measured
   (M1 Pro, go1.25.9, 200K digits): int-shaped 28 ms, frac-shaped **58 ms**,
   against `big.Int.SetString` 28 ms. An earlier revision of this ADR claimed
   the Rat parse only runs for literals ≤ ~1233 digits, which is true only of
   the int-shaped form; the frac-shaped form is the adversarial one and it is
   what `OpCPUSlopeBigDecParse` is calibrated against.
3. **Chained big-int shift.** `doOpShl` charged only the shift amount, blind
   to the operand size (contrast `doOpShr`, which uses `incrCPUBigUnary`).
   `maxBigintShift = 10000` caps a single shift, but `1<<10000<<10000<<…`
   grows the operand ~10000 bits/step at flat gas. (A flat
   `BigintValue.GetShallowSize` was noticed alongside this but is *not* fixed
   here; see Alternatives.)

**Metering runs during preprocessing.** `Preprocess` const-folds each
`*BasicLitExpr` via a throwaway machine with no explicit `GasMeter`; it
charges only because `NewMachineWithOptions` inherits the meter from the
store's preprocess allocator (machine.go:152), which `withQueryEvalMachine`
sets to `ctx.GasMeter()` (keeper.go:1384). This is non-obvious, so the
regression test must go through `QueryEval`, not a bare `doOpEval`.

The fix makes the query *bounded*, not *cheap*: a correctly-metered literal
can still spend the full 3e9-gas budget (~seconds), like any expensive query.
That residual is a property of the free-query design, addressed
architecturally (see Consequences), not per-literal.

## Decision

Meter each under-charged operation at the VM op layer. No hard size caps,
because the
codebase meters bignum work with gas rather than forbidding it, and a
literal-only cap would not lower the free-query ceiling anyway (see
Alternatives).

### 1. Quadratic charge for literal parsing (INT and FLOAT)

`op_eval.go` + `machine.go`: before parsing, charge
`gas = (digits/10)² * slope / 10` (the existing `incrCPUBigDecQuad` idiom, via
`overflow.Mulp`):

```
OpCPUSlopeBigIntSetString = 2 // INT big.Int.SetString
OpCPUSlopeBigDecParse     = 4 // FLOAT: 2x INT, measured, not headroom
```

Charged before the parse, so a large literal OOGs first (above ~1.23M INT /
~866K FLOAT digits on the 3e9 budget). FLOAT is charged too (`.0`/`e1`
bypass). The floored `(digits/10)²` charges **0** for INT literals under 30
digits (FLOAT under 20), i.e. every ordinary int64/uint64 constant, so the common
case is unchanged.

Two things a reviewer will check:

- **Radix complexity is Go-version-dependent; the single slope is deliberately
  conservative for power-of-two bases.** On the toolchain this repo requires
  (`go.mod`: `go 1.25.9`), `nat.scan` packs bases **2, 4 and 16** directly into
  words (`case 2: n=_W; case 4: n=_W/2; case 16: n=_W/4`, then
  `append`+`slices.Reverse`); those parse **linearly**. Only the `default:`
  arm (`maxPow`+`mulAddWW`) is quadratic, which covers bases **8 and 10**,
  i.e. plain decimal and `0o`/legacy-octal literals. Measured on go1.25.9 at
  400K chars: base 10 ~1.0e2 ms, base 8 ~0.8e2 ms, base 16 and base 2 ~2 ms
  each.
  Base 8 is **cheaper per character** than base 10, not dearer (measured
  ~0.83×): it takes the same quadratic arm, but `maxPow(8)` packs 21 digits
  per word against base 10's 19, and it builds a narrower nat (3 bits/char
  against 3.32). Those two compound to 0.817. At equal *value* that
  per-character advantage is cancelled almost exactly by base 8 needing 1.107×
  as many characters, which on a quadratic costs 1.107² = 1.226×: 0.817 ×
  1.226 = **1.002**. So octal and decimal cost the same for the same value
  (measured ~1.0×).
  End-to-end through `QueryEval` at 1.2M chars, `0x`+`f`… costs ~1e1 ms
  against ~1 s for the same length in decimal, **~1e2×** for an identical
  charge (the parse-only ratio is higher, ~1.5e2×; quote whichever, but not
  one labelled as the other). (One significant figure: re-measurement moved
  every absolute figure
  here by 1.5–5×, though the ratios and the linear-vs-quadratic split are
  stable. An earlier revision quoted these to three figures and to ~215×.)

  An earlier revision of this ADR claimed the opposite ("no power-of-two fast
  path; 1.23M hex chars → ~1.6 s") and used it to dismiss reviewer reports of
  hex over-charging. That measurement was taken under go1.22.3, which does
  lack the packing path. A scratch module picks up the local toolchain
  (`GOTOOLCHAIN=local`), not the auto-switched one `go version` reports. The
  reviewers were right.

  We keep one base-independent slope anyway: it errs toward over-charging
  (never a silent undercharge), and base-aware slopes would add a second
  consensus-relevant constant for a case with no legitimate users. The cost is
  bounded but real: end-to-end, hex/binary literals are over-charged ~1e1× at
  100K chars and ~1e2× at 1.2M, and any literal **≥ ~1.23M hex chars is
  rejected** on the 3e9 query budget for single-digit ms of actual work.
  Revisit if a real
  workload ever
  embeds a large power-of-two literal, or if the minimum Go version moves
  backwards.

- **Huge-exponent floats are safe** (~5–40 µs each), but not by the route an
  earlier revision described. Verified on go1.25.9, the three regimes are:
  `1e2000000000`: `ParseFloat` returns **+Inf with `err == nil`** (it does not
  error), so `MantExp` is 0, the `> ratOverflowBits` guard does **not** fire,
  and it is `Rat.SetString` that rejects the oversized exponent.
  `1e100000000`: finite, `MantExp` = 332192810, so the guard *does* skip the
  Rat parse. `1e1000`: finite, `MantExp` = 3322, and the Rat parse
  legitimately accepts. Either way no cheap-length / expensive-parse escape
  exists, so the
  length-based charge is not bypassable this way.

#### The charge must not depend on evaluation history

`chargeBigLitParse` takes the **raw** literal length (`x.Value` as written in
source, blank-identifier (`_`) separators included) so that the strip's own
O(n) scan, which does not go through `m.Alloc`, is billed rather than run for
free ahead of any charge. (`strings.Replace` returns the input untouched when
there is no separator, so the common case is a scan, not a copy.)

That is only safe because `doOpEval` no longer writes the stripped text back
into `x.Value`; it previously did, as a small caching optimisation. Combining
the two would make gas depend on how many times a node had already been
evaluated: the first evaluation bills the separators, later ones cannot see
them. Measured on a 102-digit literal with 11 separators: **24 gas on the
first evaluation, 20 on every subsequent one.**

**This was latent, not reachable.** `preprocess.go:1404` replaces every
`*BasicLitExpr` with a `*ConstExpr`, and `transcribe.go` treats `*ConstExpr`
as a leaf, so no later pass re-visits the node; an instrumented build that
panics on a second evaluation of the same node saw none across the gnovm unit
tests and the full `gno.land/pkg/sdk/vm` suite. But a gas value derived from
mutated AST state is one refactor away from being reachable, and gas that
varies with evaluation history is the same failure shape as the mem-package
cache divergence, so the strip now returns a new string (`stripBlanks`), the
node is left as written, and `TestBigLitParseChargeIsEvalInvariant` fails if
the mutation is reintroduced.

Billing separators is a real gas change against the previous revision of this
branch. Gas can only differ when `floor(raw/10) != floor(stripped/10)` **and**
the raw length is ≥ 30 (INT) or ≥ 20 (FLOAT); below that, the `(digits/10)²`
floor charges 0 either way. The size of the delta is *not* bounded to one step
in general: legal Gno `9_9_9…` with 200 digits and 199 separators is raw 399 vs
stripped 200, i.e. 304 gas vs 80. It is always ≥ the old charge, so it can only
ever over-charge.

An exhaustive parse of every `.gno` file in the repo (~65,000 INT/FLOAT
literals across ~4,500 files; exact counts drift with the tree) found **zero**
whose charge changes, reproduced
independently. The longest separator-bearing literal on disk is
`gnovm/tests/files/const22.gno`'s `1_000_000.000_000` (raw 17); the longest in
the `.txtar` corpus is `1_000_000_000_000_000` (raw 21, an INT) in
`gno.land/pkg/integration/testdata/alloc_array.txtar`. Both charge 0 either
way. None of the 21 `// Gas:` fixtures contains a separator. `Files` and
`TestTestdata` are unchanged.

#### Calibration

The const block's rule for a quadratic slope is `slope = ns/digit² × 1000`
measured on the reference box (Intel Xeon Platinum 8168 @ 2.70 GHz), and every
neighbouring slope records its fit. These two did not, because the only
`doOpEval` literal benchmarks topped out at 100 digits, far inside the
constant-dominated regime, where no quadratic term can be recovered. That gap
is now closed: `BenchmarkOpEval_BigIntLit_{1000,4000,16000,64000}`,
`BenchmarkOpEval_BigDecLit_*` (frac-shaped) and the
`BenchmarkOpEval_BigIntLitHex_*` / `BenchmarkOpEval_BigDecLitInt_*` controls
are in `bench_ops_test.go`, so the next reference-HW calibration run produces a
native fit.

Local two-point fit (Apple M1 Pro, go1.25.9, 16000→64000 digits, base 10).
**One significant figure**: run-to-run spread on this hardware reaches tens of
percent, and an independent re-measurement reproduced the shapes but not the
third digit.

| series | ns/digit² (two-point) | local slope | note |
|---|---|---|---|
| `BigIntLit` (decimal INT) | ~0.8–1.1e-3 | ~0.8–1.1 | |
| `BigDecLit` (frac-shaped FLOAT) | ~1.6–1.9e-3 | ~1.6–1.9 | ~2× the INT fit |
| `BigDecLitInt` (int-shaped FLOAT) | ~0.7–0.95e-3 | ~0.7–0.95 | guard fires, Rat skipped |
| `BigIntLitHex` | n/a | n/a | ~4× for 4× size ⇒ **linear** |

Ranges, not points, and deliberately so: independent re-measurements of these
same quantities have differed by ~20%. **The basis matters and is part of the
number**: this is `ns/op(pure)` from the benchmarks above, which carries
allocation pressure with no GC between iterations. The same fit on GC-quiesced
standalone `SetString` reads up to ~25% lower at n=16000, converging to ~4% by
750K. Do not run these in a process with other large-allocation benchmarks;
GC pressure skewed one reviewer's anchor run by ~40% before they caught it.

The frac-shaped/int-shaped ratio is **~2×**, which isolates the
`big.Rat.SetString` pass, and it holds from 200K digits up to the ~750K
delivery ceiling. That, not headroom, is why `OpCPUSlopeBigDecParse` is twice
`OpCPUSlopeBigIntSetString`. **Do not lower it to match `big.ParseFloat` timed
on its own**; that measures only half the op.

Scaling to the reference box needs an M→Xeon factor. Measuring
`OpMul_BigInt_4096` and `OpQuo_BigInt_4096` locally against the recorded Xeon
run in `cmd/calibrate/op_bench_do_dedicated.txt` gives **1.8–1.95×**, stable
across `OpMul_BigInt` and `OpQuo_BigInt` at both 1024 and 4096 bits, so the
factor is insensitive to which you pick. Propagating the fit ranges honestly,
that implies a reference slope of roughly **1.4–2.1** for INT and **2.9–3.7**
for FLOAT. The shipped `4` sits above the FLOAT estimate. The shipped `2` sits
at the **top** of the INT estimate rather than above it, so INT should be read
as priced at roughly its measured cost, not comfortably conservative. An
earlier revision of this ADR claimed a wider INT margin; that came from quoting
a GC-quiesced fit in a table whose stated basis is `ns/op(pure)`, which is the
same basis-mixing this section warns about.

Two honest caveats on that estimate. First, it is *not* repo-paired data: the
repo's M-side file (`cmd/calibrate/bench_output_m2_arm64.txt`) contains **no
bignum benchmarks at all** (only `Alloc_*`, `OpAdd_String_*`, `OpCopy*`,
`OpEnterCrossing_*`, `OpUnrefCopy_*`), and it is an M2, not the M1 Pro used
here, so the M-side halves above are local measurements combined with the
repo's Xeon halves. Second, `BigDec` op ratios scatter from ~0.5× to ~27×
depending on op and size, so they are not a usable anchor in either
direction.

**Take the fit two-point, on one basis, and compare like with like.** The
neighbouring entries in the const block use naive `t/n²` at a single size. For
these ops a real linear term is still material in the benchmark range, so naive
`t/n²` reads high (~1.1e-3 at n=16000 on the benchmark basis, ~9e-4
GC-quiesced) against a ~0.8–1.1e-3 two-point fit on the benchmark basis. An earlier
revision of this ADR
mixed the two methods across sizes (naive at 16K against fitted at 750K) and
concluded the parse was "sub-quadratic at the sizes that matter", over-charging
~20% at the ceiling. That was an artifact of the mismatch. A later revision then
claimed the coefficient was "flat to ~2.5%, exponent ~2.0", which reproduces
only by committing the same naive-vs-fitted mismatch again, in the other
direction.

Measured consistently, the effective exponent over 16K–750K digits is a
little under 2 rather than exactly 2: **~1.9–2.0** for both INT and FLOAT
across sessions. Do not read a third significant figure into it, and do not
quote a coefficient drift: the two-point drift lands anywhere from −8% to
+10% depending on estimator and session, and the naive decline anywhere from
~11% to ~24%. The argument rests only on the exponent being below 2. That is
the safe direction: because the exponent is below 2, a coefficient fitted on
the small window over-charges at the ceiling rather than under-charging. The
op stays essentially quadratic, as the
Go source implies: base 10 takes `nat.scan`'s `maxPow`+`mulAddWW` arm with no
divide-and-conquer path.

Settle the constant with a native run on the reference box rather than by
argument. That is what the `TODO(calibration)` marks, and it is the one open
question a reviewer should not take on faith here.

### 2. Meter `doOpShl` by operand size

`op_binary.go`: add `incrCPUBigUnary(lv, OpCPUSlopeBigIntShl)` beside the
shift-amount charge, mirroring `doOpShr`, so a chained shift OOGs. Mirrored
into `doOpShlAssign` (op_assign.go) for symmetry; that branch is currently
unreachable (`<<=` needs a concrete-typed lvalue), so it is defense-in-depth.

### 3. Clamp the shift-amount gas charge

`op_binary.go` + `op_assign.go`: the pre-existing shift-amount charge
`int64(rv.GetUint()) * OpCPUSlopeBigIntShl / 1024` runs *before* `shlAssign`
enforces `maxBigintShift`, so `rv` is unvalidated. Above ~2.4e17 the multiply
wrapped int64 negative and `ConsumeGas` panicked `"gas must not be negative"`
instead of reporting `"shift amount exceeds maximum"`, e.g.
`const _ = 1 << 300000000000000000`. (`MaxUint64` happened to be safe:
`int64(MaxUint64)` is -1, which truncates to a 0 charge.)

Clamp the amount to `maxBigintShift` for charging. Anything above the cap
panics in `shlAssign` regardless, so this cannot under-charge an operation
that actually runs; it only changes the gas billed for a shift that is going
to be rejected either way. Covered by `TestShiftAmountGasOverflow`.

## Alternatives considered

- **Hard per-literal digit cap.** An earlier revision rejected literals over
  10000 digits. Dropped: it does not lower the free-query ceiling (the shift
  path stays gas-bounded, not capped, because arbitrary-precision arithmetic
  must not
  be forbidden), so capping literals buys nothing, while being the only
  Go-incompatible part of the change. The place to lower the ceiling is the
  query gas budget (see Consequences).
- **`chargePreprocessGas` on qeval (linear per-byte).** Insufficient: a linear
  charge can't cover a quadratic cost within the 5 MB body, and wouldn't meter
  the shift vector.
- **Dora's slope `60` with `÷32`.** Over-charges ~1000× (an arithmetic slip)
  and is inconsistent with the true-ns `OpCPUSlope*` family.
- **Global bignum size cap.** Too invasive; it breaks legitimate
  arbitrary-precision arithmetic.
- **Size-proportional `BigintValue.GetShallowSize`.** Considered and dropped.
  `GetShallowSize` returns a flat `allocBigint` regardless of magnitude, and
  `allocBigintByte` is defined but unused; making it
  `allocBigint + allocBigintByte*ceil(BitLen/8)` looked like a natural
  companion fix. Dropped for three reasons: it is independent of this DoS
  (bigints are never charged to `Alloc.Allocate` at creation, so decision 2's
  CPU charge is what bounds growth); it is consensus-affecting on its own, and
  a coordinated upgrade is the worst place to carry a change nothing depends
  on; and its reachability is unproven. The only call site that could dispatch
  to it is the GC walk (garbage_collector.go:203); `store.go:564` cannot,
  because it calls `GetShallowSize` on an `Object` and `BigintValue` is a bare
  `struct{ V *big.Int }` with no `ObjectInfo`. An instrumented build recorded
  **zero** calls with a non-nil `V` across chained bigint-const shifts, a
  5000-digit const, and stored int64/uint64 vars; untyped bigints are
  preprocess-only and appear not to reach persisted realm state. Worth
  revisiting as its own change if someone demonstrates a reachable path, and
  `BigdecValue.GetShallowSize` has the identical gap (`allocBigdecByte` also
  unused), so the two should move together.

## Consequences

- **Consensus-breaking → coordinated upgrade required.** The gas charges and
  the alloc size change tx-path values too (AddPackage/Run preprocess inherit
  the tx meter), so mixed-version validators would diverge on `AppHash`. No
  literal is *rejected* (gas/alloc accounting only), so nothing that compiled
  before fails to compile. Blast radius is small: only INT literals ≥ 30
  digits / FLOAT ≥ 20, untyped-bigint shifts, and persisted bignums change:
  `gc.txtar` and all filetests are unchanged (verified; filetests fold under a
  nil meter, so their `// Gas:` fixtures are unaffected).

- **The charge never refuses a literal a default node will accept.** This is
  the sharpest statement of what this PR does *not* fix, and it is now a
  checked fact (`TestVMKeeperQueryEvalBignumResidualCeiling`), not prose. The
  charge first refuses a decimal INT at **1,224,750 digits**, but a default
  node caps the request body at 1 MB, and `data` is a `[]byte` so
  `encoding/json` base64s it, so the largest literal that can be *delivered*
  is **749,886 digits**. Gas alone therefore rejects nothing reachable:

  | payload at the deliverable max | gas | % of 3e9 | wall, end-to-end (M1 Pro) | refused at | margin |
  |---|---|---|---|---|---|
  | `999…9` (INT, decimal) | 1.125e9 | 37.5 % | ~0.4–0.6 s | 1,224,750 | 1.63× |
  | **`1.999…9` (FLOAT, frac-shaped)** | **2.249e9** | **75.0 %** | **~0.8–1.7 s** | **866,030** | **1.15×** |
  | `0xfff…f` (INT, hex) | 1.125e9 | 37.5 % | ~6 ms | 1,224,750 | 1.63× |

  Gas figures are exact and deterministic; the wall column is indicative only.
  It has moved several-fold between reviewers on the same hardware depending
  on load and GC state (one FLOAT sample ranged 0.8 s to 4.7 s within a single
  session), which is why the column is a band and
  `TestVMKeeperQueryEvalBignumResidualCeiling` asserts the gas and merely logs
  the time.

  Note the margin on the *frac-shaped FLOAT*, the worst payload, is only
  **1.15×**, not the ~1.63× the INT row suggests: it is charged at twice the
  slope, so its refusal point is far closer to the delivery ceiling. A modest
  rise in `MaxBodyBytes`, or any drop in the FLOAT slope, closes that gap and
  flips the FLOAT case to refused.

  `1.999…9` is representative rather than strictly maximal: `1.000…0` reduces
  to `1/1` and stays in rat form, running the full `Rat` normalisation (GCD and
  reduction divisions). Repeated measurement put it anywhere from
  indistinguishable to ~8 % worse, so no figure is quoted. Same order, same
  bound; read the row as the shape class, not a proven maximum.

  Two tails on the same residual, both out of scope here: the FLOAT case
  *succeeds*, so a free unauthenticated query also returns a ~750 KB result
  string (response amplification); and the INT path builds its `big.Int` with a
  plain `big.NewInt`, not `m.Alloc`, so ~312 KB materialises with no allocator
  accounting. This PR's charge is CPU-only by design, consistent with
  `f14892babf` dropping the `BigintValue.GetShallowSize` change.

  The frac-shaped FLOAT is the worst admitted single-literal payload, ~2× the
  INT case, because it takes the `Rat.SetString` pass described above. Both
  earlier analyses of this residual priced only the INT form and so understated
  it ~2×. (The INT form additionally fails *after* parsing, converting an
  untyped bigint that large to a concrete kind; the CPU is already spent, so
  this does not help the node.)

  Note this residual is sensitive to `MaxBodyBytes`: at the lib-server default
  of 5 MB the deliverable max would be ~3.75M digits, comfortably past the
  refusal point, and the charge *would* bind. It is the node's 1 MB override
  that keeps the whole reachable range under the budget.

- **Residual free-query CPU ceiling (out of scope), and it is not bignum-bound.**
  `maxGasQuery = 3e9` still allows a multi-second metered burn per free query.
  For the chained-shift vector, measured on M-series: 6.3–9.5 s at 4000 shifts
  (36 KB) and **14.0 s at 40000 shifts (360 KB)**. An earlier revision quoted
  "~4 s for a ~26 KB body", which understated the ceiling ~3.5×; the Xeon 8168
  the gas model targets is slower still.

  More importantly, **bignum metering is not what sets this ceiling.** The
  generic `withQueryEvalMachine` path (`ParseExpr` + `Eval`) is quadratic in
  term count and under-priced ~4–5× against the `1 gas = 1 ns` definition, so a
  payload containing no bignum at all dominates. Measured for
  `1+1+1+…` on the eval path:

  | terms | bytes | gas charged | wall | gas/ns | % of 3e9 |
  |-------|-------|-------------|------|--------|----------|
  | 5000  | 10 KB | 27.2 M      | 138 ms | 0.20 | 0.9 %  |
  | 10000 | 20 KB | 104 M       | 633 ms | 0.17 | 3.5 %  |
  | 20000 | 40 KB | 409 M       | 1.82 s | 0.22 | 13.6 % |
  | 40000 | 80 KB | 1.62e9      | 7.06 s | 0.23 | **53.9 %** |

  So ~80 KB of plain arithmetic buys over half the query budget; a full budget
  on this path is ~13–18 s of validator CPU on M-series (3e9 divided by the
  table's flat 0.17–0.23 gas/ns), comparable to the
  chained-shift figure above, and reachable without a single big number. The
  work *is* metered, and metered with the right quadratic shape; it is simply
  under-priced, so this is a calibration gap, not a metering gap. Note that
  const-*folding* the same expression is linear and cheap (226 ms, 17.7 M gas
  at 40000 terms). The quadratic blow-up is specific to the qeval
  expression-eval path.

  All of this is pre-existing and unchanged by this PR (merge-base
  `d1a33f5747` shows the same shape: 8.2 s at 40000 terms). It is a property
  of the free-query design; mitigate architecturally with a lower gas budget for
  read-only queries, a handler-cancelling timeout, and/or a finite
  `MaxOpenConnections` / rate limit. Re-pricing the generic expression path
  would be a separate, broader consensus change.

- **Slopes still want a reference-hardware bench, but they are not outliers.**
  With the cap gone, these slopes are what bounds *literal size*, so they
  should be benched on the Xeon reference before leaving draft. Spot-checked on
  M-series, `OpCPUSlopeBigIntSetString = 2` yields gas/ns 0.19–1.03 for base 10,
  the same band as the already-calibrated `OpCPUSlopeBigIntMulQ = 4`, which
  measures 0.23–0.59 at genuinely schoolbook widths (1024 and 2048 bits;
  `karatsubaThreshold = 40` words = 2560 bits, so 2560 and 4096 bits are
  already Karatsuba and drift up to 1.22 as the quadratic charge outruns the
  subquadratic multiply). So the magnitude is in family with the existing
  quadratic bignum charges rather than an order-of-magnitude miss.
  `OpCPUSlopeBigIntShl` under-prices the chained shift ~3×, which over-runs
  into OOG rather than a silent undercharge. Note these
  slopes bound literal/shift size only; they are *not* what bounds total
  free-query CPU; see the residual bullet above.

- **Not a surface: the tx-path type-check pre-pass.** `go/constant.MakeFromLiteral`
  is O(n²), but go/types caps literals at 10000 chars and rejects a 1M-digit
  `const` in ~8 µs before reaching it (`Simulate` hits the same cap), so it
  needs no charge. That cap is go/types' own `const limit` in
  `go/types/literals.go`, **not** something the `GoVersion` pinned in
  `gotypecheck.go` governs, so a toolchain bump can move it silently and this
  paragraph would quietly stop being true. `TestTxPathLiteralCap` pins it:
  all three literal shapes at 1.3M chars, the exact 10000/10001 boundary, and
  the worst tx-deliverable literal priced against the query budget. Raised by
  @Villaquiranm in review.

## Verification

```sh
go build ./gnovm/... ./gno.land/...
go test ./gno.land/pkg/sdk/vm/ -run 'TestVMKeeperQueryEvalBignum|Gas' -count=1
go test ./gnovm/pkg/gnolang/ -run Files -test.short -count=1 -timeout 600s
go test ./gno.land/pkg/integration/ -run TestTestdata -count=1
```

- `TestVMKeeperQueryEvalBignumGas` (gno.land/pkg/sdk/vm) drives the real
  `QueryEval` path: a ~2M-digit int and the `.0`/`e1` float variants OOG
  before the parse; a small literal still evaluates.
- `TestVMKeeperQueryEvalBignumResidualCeiling` pins the residual above: it
  asserts the charge's refusal point stays *above* the largest literal a
  default node can receive, and that both the INT and frac-shaped FLOAT worst
  cases are still admitted. It fails if that ever flips, at which point the
  residual guard can be dropped. Gas is asserted (deterministic); wall time
  is logged, not asserted.
- `TestPreprocessGas_BignumLiterals` (gnovm/pkg/gnolang) const-folds under a small
  gas meter: the chained shift OOGs (and does *not* without the `doOpShl`
  charge), INT/FLOAT literals OOG, and a 20000-digit literal folds. Each gas
  charge was mutation-tested; reverting it fails its sub-test.

All commands pass on this branch.

## Files

- `gnovm/pkg/gnolang/op_eval.go`: INT/FLOAT quadratic parse charge
  (`chargeBigLitParse`)
- `gnovm/pkg/gnolang/machine.go`: `OpCPUSlopeBigIntSetString`,
  `OpCPUSlopeBigDecParse`
- `gnovm/pkg/gnolang/op_binary.go`: `doOpShl` operand-size charge;
  shift-amount clamp
- `gnovm/pkg/gnolang/op_assign.go`: `doOpShlAssign` operand-size charge
  (symmetry); shift-amount clamp
- `gnovm/pkg/gnolang/bench_ops_test.go`: calibration benchmarks for both
  slopes: `BenchmarkOpEval_BigIntLit_*`, `BenchmarkOpEval_BigDecLit_*`, plus
  the `BigIntLitHex_*` (linear-radix) and `BigDecLitInt_*` (guard-fires)
  controls
- `gnovm/pkg/gnolang/preprocess_alloc_test.go`: `TestPreprocessGas_BignumLiterals`,
  `TestShiftAmountGasOverflow`, `TestBigLitParseChargeIsEvalInvariant`,
  `TestTxPathLiteralCap`
- `gno.land/pkg/sdk/vm/keeper_test.go`: `TestVMKeeperQueryEvalBignumGas`,
  `TestVMKeeperQueryEvalBignumResidualCeiling`
