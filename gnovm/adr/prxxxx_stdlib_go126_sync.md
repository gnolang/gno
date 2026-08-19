# Sync `gnovm/stdlibs` against modern Go

## Context

`gnovm/stdlibs` holds `.gno` copies of Go standard library packages. They were
imported at different times and never systematically refreshed: `unicode` still
carried a `README.md` reading *"Directly copied from go1.17.5"*, `hash`,
`encoding` and `hash/adler32` had tracked no upstream addition since they were
first written, and most other packages sat around the Go 1.17–1.18 line.

Every package with a Go counterpart was diffed against the locally installed Go
1.26.1 (`/usr/local/go/src`). The audit produced 163 actionable items across 30
packages, 25 of them high priority. Each behavioural claim was executed under
both the GnoVM and Go 1.26 and the outputs compared; the full report is in
[prxxxx_stdlib_go126_audit.md](./prxxxx_stdlib_go126_audit.md).

This ADR records the decisions taken while acting on that report, not the
mechanical ports. The straightforward cases — missing functions, missing
guards, corrected doc comments — are visible in the diff and need no
justification here.

## Decisions

### Unicode moves to 17.0.0, ahead of the local toolchain

Go 1.26 ships Unicode 15.0.0, which is what gno already had. Go 1.27 (currently
rc2) ships Unicode 17.0.0. The tables were taken from `go1.27rc2` rather than
1.26 so that gno is not immediately two Unicode versions behind. `tables.gno`,
`casetables.gno`, `letter.gno`, `graphic.gno`, `digit.gno` and
`strconv/isprint.gno` are verbatim upstream copies; `isprint.gno` is generated
from the same tables and must be refreshed in the same change or the two
disagree about which code points are printable.

`unicode.C` / `unicode.Other` consequently change meaning: Go 1.25 redefined the
category from `Cc|Cf|Co|Cs` to `Cc|Cf|Cn|Co|Cs`, so it now covers unassigned
code points. This is the only change in the batch where identical input yields a
different answer, and it alters results for already-deployed realms that call
`unicode.Is(unicode.C, r)` or use `\p{C}` in a regexp. It was adopted
deliberately. `IsControl`, `IsPrint` and `IsGraphic` are unaffected, as their
range sets never reference `C`.

### Test-time `unicode` native bindings removed

`gnovm/tests/stdlibs/unicode/natives.go` bound `IsPrint`, `IsGraphic`,
`SimpleFold` and `IsUpper` to the *host* Go `unicode` package as a test-time
optimization. Those bindings answered according to whichever Unicode version the
host toolchain shipped, so once gno's tables moved to 17.0.0 and the host stayed
on 15.0.0, `unicode.IsUpper` classified runes one way under `gno test` and the
other way on chain.

This was a latent correctness bug — a test oracle that disagrees with production
— that the version bump made visible. The bindings were removed so tests
exercise the same table lookups production does. `gnovm/tests/stdlibs/generated.go`
was hand-edited to match, because `go generate` is disallowed by `AGENTS.md`.

Note that keeping the bindings was not an option once the tables moved: with
gno on Unicode 17.0.0 and the host on 15.0.0,
`strconv/quote_test.gno`'s `TestIsPrint` compares gno's `isprint.gno` against
`unicode.IsPrint` and would simply fail.

**This costs real test time**, because these functions are called in hot
interpreted loops:

| package | before | after |
|---|---|---|
| `strconv` | — | 750s → strided (see below) |
| `regexp` | 40s | 275s |
| `regexp/syntax` | 17s | 332s (also see test-trimming below) |
| `unicode` | 1.0s | 2.4s |

`strconv`'s figure came almost entirely from two tests that sweep all 1.1M code
points calling `unicode.IsPrint`/`IsGraphic`; those now stride in `-short` mode
and sweep exhaustively otherwise.

The slowdown is a consequence of *version skew*, not of the removal itself.
`go.mod` currently pins `go 1.25.9`, so the host is two Unicode versions behind
gno. Once the repository's toolchain reaches 1.27 (Unicode 17.0.0), host and gno
tables agree again and the bindings can be restored — with a test asserting
`unicode.Version` matches the host's, so the oracle can never silently drift
again.

### `regexp/syntax` resource limits, recalibrated for the GnoVM

Go's parser budgets (`ErrNestingDepth`/`maxHeight` from 1.19, and
`ErrLarge`/`maxSize`/`maxRunes` from the CVE-2022-41715 fix) were absent. They
are ported, with one deliberate divergence.

Go sizes a rune at 4 bytes and budgets 128 MB, giving a 32M-rune ceiling. That
calibration does not transfer. Measuring the upstream `\pL` × 27000 regression
case under the GnoVM showed roughly 118 bytes per rune: Go's ceiling is ~3.8 GB
resident on a chain whose per-transaction allocation cap is 500 MB
(`gno.land/pkg/sdk/vm/keeper.go`, `maxAllocTx`). Worse, merely *reaching* a
multi-million-rune budget takes minutes of interpreted work, so a limit that
high guards nothing gas has not already stopped. `runeSize` is therefore
inflated from 4 to 1024, bounding the budget at ~131k runes (~15 MB) — still far
more than a realistic contract regexp.

A note on framing, because the audit's first pass got it wrong: these limits are
**upstream parity and defence in depth, not a fix for an unmetered halt**.
`gno test` runs with `NewAllocator(math.MaxInt64)` and `NewInfiniteGasMeter()`
(`gnovm/pkg/test/test.go`), so pathological patterns can exhaust memory *in the
test harness*. On chain, gas and `maxAllocTx` bound them regardless.

### Upstream test material trimmed where it dominates CI

Go's `regexp/syntax/parse_test.go` was adopted for its new cases, but three
entries were adjusted:

- `mkCharClass` walks all 1.1M code points. Upstream calls it nine times in the
  table; the duplicate upper-case scans are hoisted into shared variables.
- The `\p{Assigned}` / `\p{^Assigned}` entries compared against two more such
  scans with an `unicode.In(r, Cn)` predicate. They are replaced by
  `TestAssignedClass`, which checks representative code points.
- The 999-deep nesting and 12345-way alternation entries are parsed, dumped and
  re-parsed by several tests each. They are reduced to a 100-deep case; the
  nesting limit itself is covered by the `invalidRegexps` entries.

Together these took the package from over 30 minutes to 47 seconds. It then rose
to 332s when the host-backed `SimpleFold` binding was removed (previous section);
the trimming is what keeps that figure from being far worse.

### `math/rand` loses its package-level source

Gno's global source was `&Rand{src: &PCG{}}` — a fixed `PCG(0,0)` returning the
same stream in every transaction forever, behind a doc comment that only warned
outputs "might be easily predictable". The top-level functions (`Int64`, `IntN`,
`Float64`, `Shuffle`, `Perm`, …) were removed outright rather than documented,
because `math/rand` is a pure package: a realm cannot seed or advance shared
state, so no package-level Source can ever be anything but a fixed seed.
Callers now construct `rand.New(rand.NewPCG(a, b))` and own the seed, making it
an auditable decision. The ChaCha8 generator was not added.

### `net/url` host parsing follows Go 1.26, unconditionally

`parseHost` now rejects a `[` that is not at index 0, validates bracket contents
as a real IPv6 address, and splits the port at the *first* colon rather than the
last. Go gates the colon change behind the `urlstrictcolons` GODEBUG for
compatibility; gno has no GODEBUG and no deployed-behaviour burden, so the
strict form is unconditional. Go's `postgres`/`postgresql` exception for libpq
multi-host URLs is preserved.

Go validates the bracket contents with `netip.ParseAddr`. `net/netip` is
non-deterministic host networking and absent from gno, so
`gnovm/stdlibs/net/url/ip.gno` implements the IPv6 grammar `net/url` actually
needs. Two test cases that asserted the old lenient behaviour were removed, as
they were upstream.

### The drift report's baseline is pinned, not inherited

`misc/stdlib_diff` renders the gno-vs-Go diff published to GitHub Pages. Its
baseline came from `go-version-file: go.mod`, which silently re-baselines the
report whenever the build toolchain is bumped — quietly erasing drift that had
been visible. `.github/workflows/deploy-pages.yml` now pins `go-version:
"1.26.1"` (the version this audit was performed against), asserts the resolved
version and fails loudly on a mismatch, then restores the repo toolchain for the
rest of the job so `misc/gendocs` is unaffected.

The report will now permanently show `unicode/*` and `strconv/isprint.gno` as
divergent, since those come from go1.27rc2. That is the report correctly
describing a deliberate divergence; it is documented in
`misc/stdlib_diff/README.md` rather than suppressed.

### Not done, deliberately

- **`crypto/subtle` constant-time primitives.** Everything a chain executes is
  public and replayable, so there is no secret whose comparison time could leak.
  Adding them would cost gas and imply a guarantee the execution model does not
  need.
- **`crypto/ed25519` `Sign`/`NewKeyFromSeed`, `crypto/sha256` incremental API.**
  Both would require new native bindings.
- **Anything needing new native functions**, per the same constraint.

## Consequences

- `unicode.C`/`Other` classification changes for unassigned code points. Realms
  relying on the old set change behaviour.
- `math/rand`'s top-level functions are a compile error rather than a silent
  fixed-seed stream. This is a breaking change for any code using them, and is
  intended to be.
- Regexps that previously compiled may now be rejected with `ErrNestingDepth` or
  `ErrLarge`, and gno's rune budget is tighter than Go's.
- `net/url` rejects multi-colon unbracketed hosts and unvalidated bracket
  contents that previously parsed.
- `unicode` classification under `gno test` now matches on-chain behaviour;
  tests that relied on host-Go answers may shift.

## Alternatives considered

- **Taking Unicode 15.0.0 from Go 1.26** rather than 17.0.0 from 1.27rc2. That
  matches the toolchain but leaves gno two versions behind on a table that is
  costly to refresh, and would still have required the `unicode.C` decision.
- **Keeping Go's `maxRunes` unchanged.** Rejected: the budget would exceed the
  chain's own allocation cap, so it could never fire before something else did.
- **Documenting `math/rand`'s fixed seed instead of removing the API.** Rejected:
  a documented footgun that silently returns a constant is worse than a compile
  error, and there is no correct way to use a fixed global seed on chain.
