# Pricing crypto/modexp on a counted operation, not a fitted product

## Context

`crypto/modexp` exposes EIP-198 modular exponentiation as a native. Its cost is
a product of two operands — one modular squaring per exponent bit, each
`O(words(modulus)^2)` — so the single-slope machinery the gas table is built
around cannot express it. The row originally charged on `len(modulus)` alone,
which priced the exponent at zero and left a consensus-halt vector: a large
exponent against a small modulus ran unbounded work for a near-flat fee.

The first fix in this PR introduced `SizeModExpWork`, folding the product into
one metric, `work = 8*len(exp) * (floor + words(mod)^2)`, and fitted a slope
over a grid of eleven bench points. That closed the original hole. A review of
it found the fit was not an upper bound, for reasons the grid could not see:

1. **Two routines, one coefficient.** `big.Int` takes the windowed Montgomery
   path only when the exponent exceeds one machine word (`nat.go`: `if len(y) >
   1 && !slow`). At eight bytes or fewer it runs the generic square-and-multiply
   loop, which reduces by full division every iteration instead — roughly 2.5x
   the cost per exponent bit. The grid sampled `exp ∈ {4, 32, 256, 1024}`,
   jumping straight over the band where the two regimes diverge.
2. **A non-monotonic charge.** The Montgomery loop walks whole exponent *words*
   (`for i := len(y)-1; i >= 0; i--`, all sixteen windows of each, including a
   partly filled top word), so nine bytes of exponent does the same work as
   sixteen. Charging `8*len(exp)` made an eight-byte exponent — measurably the
   *more* expensive of the two — cost less than a nine-byte one. An inversion is
   directly exploitable: a caller buys the dearer computation at the cheaper
   shape's price.
3. **A modulus-independent floor.** Converting operands across the dispatcher,
   allocating the result and filling it is linear in `len(modulus)` and runs
   even when the exponent is empty. Folding that into `Base` made a call with a
   1024-byte modulus and no exponent cost a flat `Base`.
4. **An unreproducible constant.** The committed calibration input contained no
   bench lines for this row, so the shipped slope could not be re-derived. The
   bench also filled the exponent with a patterned value rather than a
   worst-case one, measuring up to 1.75x cheaper than the shape being charged.

## Decision

**Count the operations instead of fitting the product.** Every branch `expNN`
takes is decided by the operand lengths alone, so `modExpWork` now returns the
number of modular multiplications `big.Int.Exp` will perform, in units of one
multiplication at `words(modulus)`:

- Montgomery path: `38 + 160*words(exp)` units — read off `expNNMontgomery`
  (sixteen powers-table steps, ~two for the RR setup, one final convert, then
  `words(exp) * 16` windows of four squarings plus one multiply), each doubled
  because one Montgomery step is a product followed by a REDC reduction.
- Generic path: `5` units per exponent bit. This one is calibrated rather than
  counted: `nat.div`'s cost per word is set by hardware divide latency, not by a
  multiplication count. The counted estimate is ~3.25.

The Montgomery constants are properties of a fixed algorithm, so they need no
re-derivation on new hardware; only the slope converting units to nanoseconds
does. Over the recorded grid the Montgomery points hold to 1.39x of each other
across a 128x range of exponent sizes, which is the evidence that the count
matches the algorithm rather than the benchmark.

**Give the dispatcher its own slope.** The row gains `Slope2` on `SizeLenBytes`
at the modulus parameter, fitted from the `expLen=0` grid points, which do no
exponentiation and therefore isolate that cost.

**Make the row a fitter output.** `gen_native_table.py` gains `fit_modexp`,
which takes maxima rather than least squares, and a `--hw-factor` flag for
projecting a non-reference run. Running it over the committed grid reproduces
the shipped row exactly, so the row is no longer a hand-edit the fitter would
silently undo. The grid itself — 26 points, the machine, the flags, and the four
anchors behind its projection factor — is committed in the bench file.

## Alternatives considered

**Raise the fitted slope until it covers the worst point.** Simple, and wrong in
a way that matters: one linear coefficient over a two-regime cost either
underprices sub-word exponents or overcharges ordinary ones. Following the
documented procedure on the extended grid yields ~7500 against the shipped 4000
and overcharges every 32-byte-exponent call roughly 3x. It also leaves the
monotonicity inversion in place, since that is a property of the metric's shape
rather than its scale.

**Round the exponent up to whole words and stop there.** Fixes the inversion but
badly overcharges one-byte exponents (measured 14.8x model spread versus 3.7x
for the status quo), because below the crossover the generic loop really does
walk bits. Only the two changes together work: 1.90x.

**Implement `math/big` in Gno and drop the native.** Attractive — it would
delete this whole class of problem, since VM opcodes are metered directly and no
calibration is needed. Measured and rejected. Gno's `math/bits.Mul64` is the
portable 32-bit-limb fallback, ~15 interpreted operations where Go emits one
`MULQ`. Benchmarking `p/onbloc/uint256` against `math/big` at the same width:
a 256-bit `MulMod` costs 186ns in Go and 1.495ms in the VM — **8,020x** — at
2,476,054 gas. Extrapolating, one RSA-2048 exponentiation would need ~31.7
billion gas against a 3B block limit: it could not execute at all, let alone
serve the IBC light-client path these natives exist for. A Gno `math/big` is
still worth having as a general-purpose library; it cannot be the implementation
of a consensus-critical precompile.

## Consequences

- **Consensus-breaking.** Gas for every `crypto/modexp` call changes.
  `EXACT_GAS` in `stdlib_ibc_crypto_determinism.txtar` moves from develop's
  2547293 to 1850271; the txtar comment reconciles the delta from the row
  constants. Deployment needs a version or height gate, or coordination — see
  the open question below.
- Across the recorded grid, projected onto reference hardware, every point lands
  between 1.24x and 2.57x of measured cost, with no undercharge and a monotonic
  charge at every modulus.
- The model's residual error is a ~1.35x bow: it over-counts at both 32-byte and
  1024-byte moduli and under-counts at 256. That is Go crossing its Karatsuba
  threshold at 40 words, inherent to any `words^2` unit, visible in both paths
  equally, and in the safe direction at the ends.
- Re-calibration is now a single number — ns per modular multiplication —
  measured on the reference Xeon, rather than a shape to re-fit.
- `modExpWorkSaturation` drops from `1<<50` to `1<<44`. The binding constraint
  is `chargeNativeGas`'s unguarded `gi.Slope * N`, not the metric;
  `TestModExpWorkSlopeProductFitsInt64` pins the headroom.
- Guards added: `TestModExpRowCoversRecordedGrid` checks all 26 points rather
  than one anchor, `TestModExpChargeMonotonic` and
  `TestModExpWorkNoCrossoverInversion` pin the inversion shut, and
  `TestModExpConstantsMatchFitter` pins the five constants the Python copy
  duplicates.

### Note on the dispatcher slope

`Slope2` was first fitted at 170000 (~166 ns per modulus byte) against a grid
recorded before #97 data-backed byte slices. That commit changed
`Go2GnoValue` to build a flat `Data`-backed array instead of one `TypedValue`
per byte, which is the path the calibration harness feeds operands through, so
the measured per-byte dispatcher cost fell from ~134 ns to ~2.9 and the slope
with it — 3500, a 49x reduction. The grid was re-recorded after the merge and
the two must not be mixed; the bench file says so at the recorded run.

This is worth knowing for the next re-calibration: the dispatcher term is
sensitive to VM-internal representation changes that have nothing to do with
modular exponentiation, whereas the work metric is not.

## Open

- **The committed grid was measured on an AMD Ryzen 7 7840U, not the reference
  Xeon 8168**, and projected by 2.3 — the high end of four independent anchors
  (2.19–2.33). High, not low, is the safe end: a larger factor projects a larger
  cost and so demands a larger slope. The row is marked draft. Re-run on
  reference hardware before deployment; the procedure and the anchors are
  recorded in the bench file.
- **No version gate.** Raised in review: identical transactions cost different
  gas before and after this change, so mixed-version nodes diverge. Whether this
  needs a height gate or is covered by the release process is a deployment
  decision, not a metric one.
