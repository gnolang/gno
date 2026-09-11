# Bound or meter every native's variable-length input

## Context

`pr6154_gno2go_slice_gas.md` narrowed the Gno→Go conversion from `Maxcap` to
`Length`, so a native no longer receives more bytes than the Gno slice makes
visible. `Length` is still attacker-chosen, and the conversion still allocates
and copies all of it — `Gno2GoValue` runs inside the native dispatcher, after
`chargeNativeGas` has already settled the price. A native priced `SizeFlat`
therefore performs an O(len(param)) allocate-and-copy for a constant fee, and
the buffer can be reused across calls so the Gno-side allocation is paid once.

Three natives were priced flat while accepting unbounded `[]byte`:

| native | flat charge | measured cost | ratio |
|---|---:|---:|---:|
| `crypto/bn254.g1Add` (16 MiB input) | 14,883 | 3.59 ms | 241x |
| `crypto/bn254.g1Mul` (16 MiB input) | 44,465 | 2.47 ms | 55x |
| `crypto/merkle.innerHash` (1+1 MiB) | 7,513 | 2.25 ms | 299x |

(End-to-end through the dispatcher on an AMD Ryzen 7 7840U, with Data-backed
`[]byte` params as Gno's `make([]byte, n)` produces. 1 gas = 1 ns.)

The two cases differ in kind. `g1Add` and `g1Mul` cannot use bytes past their
fixed input size — EIP-196 ignores the excess for `g1Add`, and `g1Mul` rejects
any length but 96 — so the entire copy beyond that is waste. `innerHash` hashes
everything it is given, so its work is real; it was simply mispriced.

## Decision

Every native parameter of variable length must be either **bounded** before the
dispatcher copies it, or **metered** per byte.

Bounded, when the native cannot use the excess: the length check moves into the
`.gno` wrapper, ahead of the native call. `bn254.G1Add` truncates to
`G1AddInputSize`; `bn254.G1Mul` returns `nil` unless the input is exactly
`G1MulInputSize`. Both mirror what the Go side already did, so behaviour is
unchanged — only the point at which the excess is discarded moves. The Go-side
checks stay for direct Go callers and as defence in depth. This is what
`crypto/cometblszk` already does: it validates every length in `.gno` before
touching a native.

Metered, when the native genuinely consumes the input: the gas row carries a
per-byte slope for each such parameter. `crypto/merkle.innerHash` hashes
`0x01||left||right`, so both operands carry `leafHash`'s calibrated per-byte
rate (`Slope` on param 0, `Slope2` on param 1). One slope would not do: the
payload would just move to the unmetered operand.

## Alternatives considered

- Charging a per-byte slope on `g1Add`/`g1Mul` instead of bounding them. This is
  the cheaper option for honest callers — a slope at the measured copy rate costs
  a 128-byte call about 30 gas, against the 3,549 the wrapper check costs (see
  Consequences) — but it keeps performing work no caller can observe, and the
  rate would have to be measured on a harness the rest of the table does not use:
  every existing byte slope was fitted on List-backed `[]byte` from
  `Go2GnoValue`, which converts element-by-element and runs ~30x slower than the
  Data-backed slices realms actually pass.
- Metering the conversion generically inside `Gno2GoValue`. Every existing slope
  was calibrated end-to-end through the dispatcher, conversion included, so a
  separate conversion charge double-prices every byte-slice native and forces a
  full recalibration.
- Rejecting oversized native input at the dispatcher with a global byte cap. One
  cap cannot serve both a 96-byte precompile and `keccak256.sum256`, and it
  would not help `innerHash`, whose input is legitimately unbounded.
- Enforcing the bounds in the `X_` functions. Too late: the copy happens before
  the native body runs.

## Consequences

- Native authors get one rule: a variable-length parameter is bounded in the
  `.gno` wrapper or sloped in `gnovm/stdlibs/native_gas.go`. Where the bound
  lives in the wrapper, the native's doc comment says so, because deleting it
  reopens an unpriced copy rather than merely duplicating a check.
- Honest callers pay a little more, and no result changes. `innerHash` costs
  9,569 instead of 7,513 gas for the 32+32-byte inner node Merkle proofs
  actually use. The `bn254` rows are untouched, but the wrapper check is Gno
  code and so is metered: +3,549 gas per `G1Add`, +3,560 per `G1Mul` (measured
  end-to-end over 2000 exact-size calls). Nearly all of that is the `len()`
  builtin, charged the flat `OpCPUCallNativeBody` of 2,205 — an uverse-pricing
  artifact, not the cost of the comparison. It is noise next to the
  `pairingCheck` (457,574) that a realm doing EC verification also pays.
- `crypto/merkle.innerHash` is the first production row to use a pre-call
  `Slope2` on a second parameter. `gen_native_table.py` grew a `NATIVE_SPECS_PAIR`
  shape that fits a symmetric two-operand sweep and emits both slopes, and it
  exits non-zero rather than degrading such a row to flat — a regeneration must
  not silently reopen this.
- `innerHash`'s `Base` is still the 32+32-byte median rather than a regression
  intercept, so it double-counts ~2 µs of per-byte cost at that size. That is
  conservative; the pending reference recalibration (the whole IBC block is
  draft, measured on a Xeon Silver 4114) resolves it, and the new sweep gives it
  the data points to do so.

## Not covered

Same class, left as separate work — all measured through the dispatcher as
above, all reachable from any realm:

- `crypto/merkle.verifySimpleProof` slopes on `aunts` only. `leaf` is hashed
  twice and unmetered (157x at 1 MiB, 2708x at 16 MiB); `rootHash` is only
  compared against a 32-byte digest, so it is pure unpriced copy (248x at
  16 MiB). Partially filed as dora `8b9eedc6`, which covers `leaf` but not
  `rootHash`.
- `crypto/cometbls.verifyZKP` is flat at 2.63 ms with one `string` and three
  `[]byte` parameters, all length-validated after conversion. The large base
  makes the ratio small (~4x at 64 MiB), and the schema's two slopes cannot
  cover four parameters, so bounding in the wrapper is the likely answer.
- `chain.pubKeyAddress` is flat at 2,631 with an unbounded bech32 `string`
  (dora `aa552fdd`).

Also noticed, and not a security problem but worth knowing before the reference
recalibration: the bench harness feeds natives `[]byte` built by `Go2GnoValue`,
whose `reflect.Slice` path always produces a List-backed array — one
`TypedValue` per byte, converted element-by-element. Realms produce Data-backed
`[]byte` (`make`, `append`, `[]byte(string)` all go through `NewDataArray`), for
which `Gno2GoValue` does a single `make`+`copy`. The harness path is ~10-30x
slower, so every per-byte slope in the table overcharges the production path by
about that factor — `leafHash` charges 33.7M gas for a 1 MiB leaf that costs
1.18 ms. Overcharging is safe, and `innerHash` inherits the same bias by design
(it takes `leafHash`'s rate), so this is self-consistent as it stands. Fixing it
means feeding the harness Data-backed slices and refitting every byte slope
together.
