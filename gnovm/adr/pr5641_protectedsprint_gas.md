# Writer-based streaming for `ProtectedSprint` / `ProtectedString`

Refactor of `gnovm/pkg/gnolang/values_string.go` to remove the
intermediate native allocations in `print` / `println` and the panic /
stacktrace formatters, and to bring the output bytes under per-tx gas
accounting.

## Context

`print` / `println` and the panic / stacktrace formatters call into
the `ProtectedString` / `ProtectedSprint` family on `TypedValue` and
every aggregate value type. The pre-refactor implementation builds
the formatted output as a Go string in three layers:

1. Leaf primitive formatting via `fmt.Sprintf` → one Go string per leaf.
2. Aggregate methods build `make([]string, N)`, recurse to fill it,
   then call `strings.Join`. One extra allocation per element plus the
   joined result.
3. The final string is handed to `uversePrint` / the exception
   formatter, which only charges gas *after* the full string is in
   memory.

A 1M-element `[]int` printed via `println` consumes ~887 MB of
cumulative native allocation (~87 MB peak heap) for a contract whose
GnoVM-accounted footprint is ~164 MB — a ~100× amplification factor
from `fmt.Sprintf` and string intermediates.

## Decision

Refactor stringification to a **writer-based streaming model** so the
formatter emits bytes incrementally into a small buffer rather than
building the whole string first, and so output is metered as it is
produced.

### `WriteProtected` is the real formatter

Every value type that participates in `ProtectedString` recursion gains
a method:

```go
func (X) WriteProtected(w *meteredWriter, seen *seenValues)
```

Receivers: `*ArrayValue`, `*SliceValue`, `*StructValue`, `*MapValue`,
`PointerValue`, `TypedValue`. A `writeProtectedSprint` internal helper
mirrors the bare-form recursion used by the existing
`(*TypedValue).ProtectedSprint`. The public entry point for streaming
callers is `func (tv *TypedValue) Fprint(w io.Writer, m *Machine)` —
the analog of the old `tv.Sprint(m) string`, same recursion and same
byte output, written directly into `w`.

The recursive formatters **return no error**. The only sink is the
underlying writer, which in every path is a `bytes.Buffer` /
`strings.Builder` / the machine output — none of which error in
practice; a parent write failure panics. Dropping the `error` return
removes the `if err != nil { return err }` ladder from every method.

### `meteredWriter` is the accounting boundary

`meteredWriter` is a `bufio.Writer`-style buffer: bytes accumulate in a
fixed `[1024]byte`, and a flush drains the buffer to the parent writer
and charges gas once:

```go
const meteredWriterBufSize = 1024

type meteredWriter struct {
    parent   io.Writer
    gasMeter store.GasMeter
    buf      [meteredWriterBufSize]byte
    n        int
}

func (mw *meteredWriter) Flush() {
    if mw.n == 0 {
        return
    }
    n := mw.n
    mw.n = 0 // reset before charging: a deferred re-Flush after an
             // OutOfGas panic in ConsumeGas is then a no-op (charged once).
    if mw.gasMeter != nil {
        mw.gasMeter.ConsumeGas(streamOutputGas(n), "stream output")
    }
    if _, err := mw.parent.Write(mw.buf[:n]); err != nil {
        panic(...) // parents are bytes.Buffer / strings.Builder; never error
    }
}
```

The numeric helpers (`WriteInt`/`WriteUint`/`WriteBool`/`WriteFloat`)
`reserve(worst-case)` then `strconv.Append*` straight into the buffer
tail, so there is **no separate scratch array** and no per-value heap
allocation. Writers are recycled through a `sync.Pool` so a Sprint
doesn't heap-allocate its kilobyte buffer each call.

**Why per-flush, and why charge the gas meter directly (not
`m.Alloc.Allocate`).** An earlier revision charged
`m.Alloc.Allocate(len(p))` on every `Write`. That was wrong on three
counts, all raised in review:

1. The hot path writes into a `bytes.Buffer` that was already grown —
   a `Write` is a memcpy, not a Go allocation, so charging "allocation
   cost" describes work that isn't happening.
2. The formatter emits many 1-byte writes (separators, digits). Each
   hit `allocGasTable[0]` (12 gas) — the per-allocation *floor* —
   making output cost ~12 gas/byte, roughly 50× what the table's own
   1 KB entry implies (≈ 0.24 gas/byte).
3. The output is a transient sink (`bytes.Buffer` / `strings.Builder` /
   `io.Discard`) that never becomes a gno object, so it is never
   GC-walked — routing it through the GC-managed allocator's byte
   counter (which `Allocate` bumps, and which a collection resets) is
   the wrong meter entirely.

Buffering fixes (1) and (2): gas is charged in buffer-sized chunks, one
charge per flush regardless of how many `WriteByte`/`WriteString` calls
fed it.

Charging `gasMeter.ConsumeGas(streamOutputGas(n), "stream output")`
directly — rather than through `m.Alloc` — fixes (3): the writer holds
only a `store.GasMeter`, never an `*Allocator`, and never allocates.
Output cannot count against, nor trigger, the per-tx allocator/GC
budget, because the allocator does not own it. The per-byte cost is
`streamOutputGas`, a constant calibrated **separately** from `allocGas`
(see "Consequences for gas economics"): output cost is ~linear in bytes
(formatting + copy), whereas `allocGas` is the concave malloc curve, so
reusing it under-charged the formatter ~16×. The writer's gas meter is
taken from `m.GasMeter`, which on every machine-construction path is the
same meter `m.Alloc` would have used.

### Migration of streaming-capable callers

- **`uversePrint`** (`uverse.go`) streams each argument into
  `newMeteredWriter(m.Output, m)` via `ev.Fprint(mw, m)`, then
  `mw.Flush()`. The placeholder per-char gas tier
  (`NativeCPUUversePrintCharsPerGas`) is gone; output cost is now the
  per-flush `streamOutputGas` charge. `formatUverseOutput` is kept because
  benchmark instrumentation (`bm.NativeEnabled`, off in production)
  still references it.
- **`Exception.Fprint(w io.Writer, m *Machine)`** is added in
  `frame.go`; the string-returning helpers (`Sprint`,
  `StringWithStacktrace`) become thin `strings.Builder` wrappers.
- **`Machine.makeUnhandledPanicError`** (`op_call.go`) and the debug
  stacktrace exception line (`machine.go`) route through `Fprint`,
  avoiding the format-then-copy double allocation a `bytes.Buffer`
  wrapper would impose on large panic values.

### Preserving the public string-returning API

`String`, `ProtectedString`, `Sprint`, `ProtectedSprint` across all
value types are preserved as thin wrappers that flush before reading
the buffer:

```go
func protectedStringOf(v protectedWriter, seen *seenValues) string {
    var b bytes.Buffer
    mw := newMeteredWriter(&b, nil)
    defer mw.Release()
    v.WriteProtected(mw, seen)
    mw.Flush()
    return b.String()
}
```

`m == nil` means the debug paths charge no gas — they remain free to
call. Output bytes are byte-identical to the pre-refactor
implementation; only the dispatch path changes. No per-shape buffer
hint is needed: the writer buffers internally and flushes a single
chunk into the `bytes.Buffer`.

## What this addresses

- **`print` / `println` output is now metered.** Wide-aggregate output
  flushes through `gasMeter.ConsumeGas(streamOutputGas(n), "stream output")`,
  one charge per flushed chunk. Gas charges are monotonic, so the
  per-tx gas budget caps cumulative output and trips `OutOfGasError`
  mid-traversal for wide values. Empirical demonstration:
  `println(make([]int, 1_000_000))` (~8 MB of output) — on
  `upstream/master` the tx succeeds (no metering on the output path);
  on this branch, with `-gas-wanted` set above the slice-make cost
  (9 M), the print itself exhausts the budget:
  `out of gas, gasUsed: 9003167 location: stream output`. Regression
  test:
  `gno.land/pkg/integration/testdata/print_wide_value_gas_metering.txtar`.
- **Panic / stacktrace formatting output is metered too**, via the same
  `meteredWriter` route.
- **`make([]string, N)` + `strings.Join`** removed from every aggregate
  formatter; per-leaf result strings removed; most `fmt.Sprintf` calls
  in the recursion replaced by `strconv.Append*`. Native-heap
  amplification per print drops from ~100× (the ~887 MB measurement)
  toward output-size (one growing buffer, no N intermediate strings).

## What this does NOT address

- **There is no hard *memory* ceiling on output — only a gas bound.**
  The writer charges gas but deliberately does not count output against
  `MaxAllocBytes` (the 500 MB per-tx allocator budget): output is not
  GC-owned, so it does not belong in that budget. Output size is
  therefore bounded by gas alone (~`streamOutputGas` rate). For the
  block-producing validator path this is moot — `vm.Output` is
  `io.Discard`, so output is streamed away and never accumulates (peak
  is the 1 KB buffer). The one path that *retains* output is
  `MsgRun` / `MsgCall`, which buffers it into the `bytes.Buffer`
  returned as the tx result (`keeper.go`); there, at the block gas
  ceiling, the retained buffer is bounded by gas (~GBs) rather than a
  hard cap. This is strictly better than master (which charged output
  even less — 0.1 gas/byte — *and* amplified ~100×), but it is not a
  hard ceiling. Adding one — capping the retained `MsgRun` buffer where
  it is held — is a deliberate, separable follow-up: it is a
  consensus-semantics change in a different code path, out of scope for
  this formatting refactor.
- **CPU pricing of deep-small-output trees.** With per-flush `streamOutputGas`
  as the only charge on the recursive path, a workload that recurses
  many times while producing few bytes is under-priced relative to the
  validator CPU it costs. A per-call CPU gas charge for `Protected*`
  recursion is deferred to a follow-up.
- **Stringer / Error intermediate-string accounting.** When a gno
  value's declared type implements `String()` / `Error()`, the gno-side
  result is allocated as a Go string (by the `m.Eval` call, under the
  gno call's own gas/alloc budget) before flowing through the writer.
  Tighter accounting is possible but out of scope here.

## Format-equivalence verification

Output is verified byte-identical via an inline golden corpus in
`gnovm/pkg/gnolang/values_string_stream_test.go`:

- ~70 curated fixtures in `fixtureCorpus()`, each paired with an
  expected output string in the `sprintGoldens` map.
- Coverage includes every primitive type, float edge cases (NaN, ±Inf,
  denormals, smallest / largest normal for both `float32` and
  `float64`), nil / undefined values, byte arrays and byte slices under
  and over the 256-byte hex-path cap, empty / small / nested aggregates
  of each kind, maps (zero + populated), pointers (typed-nil +
  non-nil), untyped bigint / bigdec, a recursive cycle (`ref@N`), a
  >1024-byte slice (the multi-flush boundary), nestedLimit `...`
  truncation, nil-base slice, and string-escape cases.
- `TestSprintMatchesGolden` asserts both the preserved
  `ProtectedString` API and the new `WriteProtected` path produce
  byte-identical output against the inline goldens, and that the
  fixture corpus and golden map stay in sync.

## Performance verification

A 12-benchmark suite (`BenchmarkProtectedString_*`) lives alongside the
golden corpus. It exercises `TypedValue.String()` on slice / struct /
byte-array / nested shapes at sizes N=10/100/1000, plus primitives and
bigint. Because it goes through the public receiver method, the same
file runs unmodified against the pre-refactor implementation on
`upstream/master` — direct apples-to-apples comparison via `benchstat`.
Compared to master, geomean **−84% allocs/op, −54% ns/op, −23% B/op**
across the 12 benchmarks; the recursive aggregate shapes that drive the
refactor see 100-400× fewer allocations and ~5× faster end-to-end at
N=1000 (e.g. `IntSlice_1000` 195 µs → 35 µs, 4744 → 12 allocs). The
buffered writer plus the `sync.Pool` is what turns the per-call
kilobyte buffer from a B/op regression into a net improvement.

## Gas-metering regression

`gno.land/pkg/integration/testdata/print_wide_value_gas_metering.txtar`
runs `println(make([]int, 1_000_000))` through a real `gnoland start` +
`gnokey maketx run` with `-gas-wanted 9000000` (deliberately above the
slice-make cost so the make succeeds and the print is what trips) and
asserts `out of gas` (a gas trip on the output path, not an
allocation-limit trip). On
`upstream/master` the print is unmetered and completes; on this branch
the ~8 MB of output flushes through `streamOutputGas` and the tx aborts
mid-print (`gasUsed: 9003167 location: stream output`).

## Commit structure

Commits, each independently reviewable (the branch is squash-merged, so the
split is for review, not bisect):

| Commit | Action | Behavior change? |
|---|---|---|
| 1 | Add `WriteProtected` / `Fprint` / `meteredWriter` + inline golden corpus + bench. | No — new code only; nothing calls it yet. |
| 2 | Collapse old `String` / `Sprint` / `ProtectedString` / `ProtectedSprint` to wrappers around `WriteProtected`. | No — golden-verified. |
| 3 | Migrate `uversePrint` + panic-path callers to stream via `Fprint`; add the gas-metering regression test. | **Yes — output bytes are now metered per flushed chunk via `streamOutputGas`.** |
| 4 | This ADR. | No — documentation. |
| 5 | Address maintainer review: rename `SprintTo`→`Fprint` and `free`→`Release`; fix the OOG-path output-gas double-charge (reset `n` before charging) and the benchmark-build double-charge; calibrate a dedicated `streamOutputGas` (4 gas/byte) replacing the `allocGas` reuse; `defer mw.Flush()`; test-helper dedup. | **Yes — output gas rate is now `streamOutputGas` (≈4 gas/byte) instead of `allocGas` (≈0.24); OOG paths charge each chunk once.** |

## Consequences for gas economics

`uversePrint`'s output-byte rate changes from the placeholder
`NativeCPUUversePrintCharsPerGas = 10` (0.1 gas/byte) to
`streamOutputGas(n) = n × streamOutputGasPerByte` charged once per flushed
chunk, with **`streamOutputGasPerByte = 4`**. This per-output-byte cost is
calibrated **separately** from `allocGas` (the maintainer review asked for a
dedicated benchmark rather than reusing the allocation table): output cost is
~linear in the bytes produced — `strconv.Append*` formatting plus a copy, in
native Go that no gno-opcode gas covers — whereas `allocGas` is the concave
malloc curve. Reusing `allocGas` (its 1 KB chunk ≈ 0.24 gas/byte) under-charged
the formatter ~16×.

Calibration (1 gas = 1 ns on the reference box, per
[PR #5629](https://github.com/gnolang/gno/pull/5629)): the format+flush path
measures ~3.0 ns/byte asymptotically (`BenchmarkStreamOutputProduce_*` in
`values_string_gas_test.go`), scaled to reference hardware via the
1 KiB-alloc anchor (`allocGasTable[10] = 241` gas / ~196 ns local ≈ 1.23×) →
~3.7 gas/byte, rounded up to 4. The result is a deterministic gas-schedule
change — output is now ~16× dearer than the interim `allocGas` reuse and ~40×
dearer than master's placeholder — consensus-safe but worth a maintainer's eye.
The same charge applies to panic / stacktrace output, which previously paid no
per-byte cost at all.

The fixed per-call base, `NativeCPUUversePrintInit` (1376), is left unchanged in
this PR — only the per-byte rate was in scope. It is now a conservative
over-estimate (it was calibrated against the pre-refactor `fmt.Sprintf` code,
which folded an element's formatting into the base; formatting now lives in
`streamOutputGas`). Re-calibrating the base is a deliberate follow-up — see
below.

## Follow-up work

- **A hard memory ceiling on retained output**, if the
  `MsgRun` / `MsgCall` result buffer at the block gas ceiling is deemed
  a concern. The cleanest shape is to cap that `bytes.Buffer` where it
  is held (`keeper.go`), rather than coupling output to the gno-value
  `MaxAllocBytes` budget. Deliberately out of scope here. See "What
  this does NOT address".
- **Re-introduce per-call CPU gas charging** for `Protected*` recursion
  to defend against the deep-recursion-small-output cost shape.
  Calibration via `gnovm/cmd/calibrate`.
- **Replace remaining `fmt.Sprintf` uses** in non-recursive `String()`
  methods (`*FuncType`, `TypeValue`, etc.) with the same
  `strconv.Append*` pattern. Performance cleanup.
- **Re-calibrate `NativeCPUUversePrintInit`** (the fixed per-call print base).
  It is still 1376 from the pre-refactor implementation, which over-estimates
  the streaming base (now just pool get/put + `GetLength` + empty flush); the
  per-byte cost moved to `streamOutputGas`. The re-calibration should use the
  in-VM `benchmarkingnative` pipeline on the reference hardware (the basis for
  the original 1376), not an off-box microbenchmark, so the base+slope split is
  measured consistently. Conservative until then (over-, never under-charges).

## See also

- Calibration framework: [PR #5629](https://github.com/gnolang/gno/pull/5629)
