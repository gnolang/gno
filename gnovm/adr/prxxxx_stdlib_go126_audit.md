# gnovm/stdlibs vs Go 1.26 — port audit

Every package in `gnovm/stdlibs` that has a Go counterpart, diffed against the locally installed
Go 1.26.1 (`/usr/local/go/src`), with a recommendation on what should be ported.

**Result: 163 actionable items across 30 packages, 25 of them high priority.**
A further 35 differences were examined and deliberately excluded (concurrency, generics, `reflect`,
`unsafe`, nondeterminism, Go-runtime internals), and 2 proposed findings were refuted outright.

---

## How to read this

Every finding was produced by a two-stage process: an analyzer diffed the trees, then an independent
verifier tried to **refute** each claim — re-checking the cited lines, grepping the whole
`gnovm/stdlibs` tree (plus `generated.go` and `native.go`) to make sure a "missing" symbol really is
missing, and, for every behavioural claim, **running the code under the GnoVM and under Go 1.26 and
comparing actual output**.

Verdicts: **CONFIRMED** (87) survived unchanged · **ADJUSTED** (57) survived with corrected
citations, severity or mechanism · **REJECTED** (2) did not survive.

That adjustment rate is the useful signal here — roughly a third of first-pass findings had something
materially wrong with them. Several examples are called out inline below, because they show what the
verification caught.

| Priority | Meaning |
|---|---|
| **H** | Correctness bug reachable from realm code, security-relevant, or a widely-needed missing API |
| M | Real gap worth scheduling |
| L | Polish: docs, examples, micro-optimisations |

---

## Coverage and baseline

21 gno packages have a Go 1.26 counterpart; they were audited in 16 groups. Packages with no Go
counterpart were out of scope: `chain/*`, `sys/params`, `math/overflow`, `internal/execctx`, and
`crypto/{bech32,bn254,chacha20,cometbls,cometblszk,keccak256,merkle,modexp}`.

Three mapping traps were resolved before diffing, since a naïve file-by-file comparison gets them
badly wrong:

- **Go 1.26 restructured `strconv`.** The number internals moved to `internal/strconv/`
  (`atof.go`, `atoi.go`, `ftoa.go`, `decimal.go`, …); the public package is now a thin wrapper.
  gno's `strconv/*.gno` correspond to the *internal* files.
- **gno's `math/rand` is a port of `math/rand/v2`**, not v1. Diffing against `math/rand` would
  report the whole package as divergent.
- **Architecture-specific files are noise.** `*_amd64.go`, `*.s` and the assembly fast paths are
  irrelevant to a tree-walking interpreter; only the `*_generic.go` / pure-Go bodies matter.

**Fork baseline.** The packages were not forked at one moment. `unicode` still carries a
`README.md` reading *"Directly copied from go1.17.5"*, yet its tables are Unicode 15.0.0 — matching
Go 1.26 — so it has been partially refreshed since. `io` is roughly Go 1.22-era (it has
`SectionReader.Outer`, added in 1.22). `errors` is pre-1.21 in its core with `Unwrap`/`Is`/`Join`
hand-added later by PR #5385. `hash`, `encoding` and `hash/adler32` have not tracked any addition
since they were first written. Most of the rest sits around the Go 1.17–1.18 line. Individual
estimates, with the marker that proves each one, are in the per-package data.

---

## Tier 1 — the 25 high-priority items

### Resource limits missing on attacker-supplied input

**`regexp/syntax` has none of Go's parser budgets.** This is the single most consequential cluster
in the audit, because a realm calling `regexp.Compile` on caller-supplied input reaches all of it.

- **`ErrNestingDepth` / `maxHeight` / `checkHeight` (Go 1.19)** — no parse-tree height cap.
  Verified: `regexp.Compile(strings.Repeat("(",1000)+strings.Repeat(")",1000))` returns `nil` error
  under gno; Go 1.26 returns *"expression nests too deeply"*. The boundary matches `maxHeight`
  exactly — 999 deep parses fine on both.
- **`ErrLarge` / `maxSize` / `maxRunes` (Go 1.19.2/1.18.7, CVE-2022-41715)** — no compiled-size or
  rune-count budget. gno's existing `repeatIsValid` is the older, narrower CVE-2022-24921 fix and
  does not catch the regression cases. Verified: both upstream cases parse cleanly under gno and are
  rejected by Go 1.26 with *"expression too large"*.

The verifier corrected the framing here, and the correction matters. The analyzer called this
unbounded on-chain recursion; it is not. gno-level recursion runs on heap-allocated GnoVM frames, and
every node allocation is gas-metered — the 1000-deep parse burned 215,937,524 gas. So the case for
porting is **upstream parity, matching error reporting, and defence-in-depth ahead of the unmetered
tree walks in `simplify.gno`/`compile.gno`/`String`** — not an unmetered halt.

One case did exhaust memory — `strings.Repeat("\\pL", 27000)` left the gno process running past two
minutes at >1 GB RSS and had to be `SIGKILL`ed, where Go 1.26 rejects it instantly — **but that is a
test-harness artifact, not an on-chain reachable OOM.** `gno test` builds its machine with
`gno.NewAllocator(math.MaxInt64)` and `store.NewInfiniteGasMeter()`
(`gnovm/pkg/test/test.go:519,549`), so gas is *counted* but never *enforced* and the allocation
budget is unlimited. On chain, `maxAllocTx = 500_000_000` (`gno.land/pkg/sdk/vm/keeper.go:50`) and a
finite `GasWanted` stop the same parse well before that. This strengthens rather than weakens the
"parity and defence-in-depth" framing above.

*Porting note:* `calcSize` uses the `max` builtin and `parse()` uses panic/recover. `recover` works
in gno, but there are no `min`/`max`/`clear` builtins, so `size = max(1, size)` must become an
explicit `if`.

**Uncapped split allocation in both `strings` and `bytes`.** Go clamps the requested field count
before allocating (`if n > len(s)+1 { n = len(s)+1 }` in `genSplit`, and the `explode` equivalent).
gno has neither guard in either package, so a caller-supplied `n` sizes the allocation regardless of
how short the input is.

Verified, and again the verifier tempered the claim usefully: gno's
`strings.SplitN("a", ",", 100000)` returns `len=1` but **`cap=100000`**, costing only ~21k extra gas
— because GnoVM allocation gas is a log2-interpolated table while the byte budget is linear, so the
allocation is nearly free in gas but consumes real budget. Go returns `cap=2` for the same call even
at `n = 1<<40`. The blow-up is *bounded* by the allocator reserving before the Go allocation, so this
is a budget-exhaustion and correctness-of-`cap` issue rather than an unbounded OOM.

**`crypto/subtle` has no constant-time primitives at all.** `ConstantTimeCompare`, `ConstantTimeEq`,
`ConstantTimeSelect`, `ConstantTimeByteEq`, `ConstantTimeCopy`, `ConstantTimeLessOrEq` — all absent;
the package contains only `XORBytes`/`XORBytesUnsafe`. Confirmed by a live `gno test` failing with
`undefined: subtle.ConstantTimeCompare`. Go 1.26 rewrote the internals as compiler intrinsics
(meaningless in an interpreter), but the pre-intrinsic bitwise formulas are trivially portable pure
integer code. For a chain where contracts compare tokens, MACs and secrets, this is a conspicuous
hole. Effort: trivial.

### Host parsing accepts what Go rejects

**`net/url.parseHost` does not validate bracketed hosts as IPv6 literals.** gno only checks
`HasPrefix(host, "[")` and never validates the contents, so anything can be smuggled inside brackets.
Verified: `http://[not-an-ip]/`, `http://[test.com]/` and
`hxxp://mathepqo[.]serveftp(.)com:9059` all parse successfully under gno. Any realm that renders or
trusts `u.Host`/`u.Hostname()` is handed a value that is not a host. Not godebug-gated upstream.

**`parseHost` uses the last colon instead of the first**, so `https://1:2:3:4:5:6:7:8` and
`https://example.com:80:` are accepted — the classic host/port confusion primitive. The verifier
caught an important nuance the analyzer had overstated: Go 1.26 only rejects these when
`urlstrictcolons=1`, and its first Go run had silently picked up the gno repo's `go 1.25.9` from
`go.mod`, which defaults the godebug off. Go also keeps a deliberate `postgres`/`postgresql`
exception for libpq multi-host URLs, which any port must preserve.

### Silent misparsing

**`time.Parse` accepts out-of-range and non-numeric zone offsets.** gno still uses `atoi` where Go
uses `getnum(x, true)` plus explicit range checks. Verified: `"+-500"` parses as a silently *negated*
offset, `"+9900"` is accepted as a 99-hour offset, and `+05:99` normalises to `+0639`. Go 1.26
rejects all three.

**`time.Parse` accepts a missing seconds field.** In the `stdSecond` case Go breaks immediately when
`getnum` fails; gno falls through into the fractional-second branch, which calls `parseNanoseconds`
and *overwrites `err` with nil*. Verified: `time.Parse("15:04:05", "03:04:.99")` returns
`03:04:00.99` with a nil error. Three-line fix. This one was found by the **verifier**, not the
analyzer — it appeared only as an aside in another finding and was never tested.

**`encoding/base64.NewEncoding` does not reject duplicate alphabet symbols** (Go 1.22). Verified: an
alphabet with `'A'` duplicated builds without panicking, then `Encode([]byte{0,1})` → `"AAE="` →
`Decode` → `[]byte{4,17}`. Silent round-trip corruption where Go now fails loudly at construction.

**`strings.explode` corrupts invalid UTF-8** — and inconsistently. It rewrites undecodable bytes to
U+FFFD, but the loop runs only to `n-1`, so the final element escapes the rewrite. Verified:
`Split("a\xffb","")` → `["a", "�", "b"]` (no round-trip), while `Split("a\xff","")` → `["a",
"\xff"]` — the same byte treated differently by position. Go 1.26 slices invalid bytes individually
and every case round-trips. Confirmed not a deliberate divergence: it arrived with the original
import (98cc986cb, PR #585) and carries none of gno's `// Custom code:` markers.

### Wrong results

**`io.NewOffsetWriter` never sets `base`.** The keyed literal `&OffsetWriter{w: w, off: off}` omits
the field, so it stays zero and `WriteAt` computes the wrong absolute position for any nonzero
constructor offset. Verified on both sides: `NewOffsetWriter(buf,10).WriteAt([]byte("HELLO"),0)`
writes at `buf[0:5]` under gno, `buf[10:15]` under Go. gno's own tests miss it because the one
retained test uses `off=0`, where a buggy `base=0` is indistinguishable from a correct one — the
`WriteAt` tests were dropped as `os.CreateTemp`-dependent.

The verifier explicitly **refuted part of this finding**: the analyzer claimed
`Seek(0, SeekStart)` returns 10 in Go vs 0 in gno. It returns 0 in both, because `Seek` returns
`offset - o.base`, which cancels `base` algebraically. The bug is real, but only observable through
`WriteAt` and through `Write`-after-`Seek`. Fix: `&OffsetWriter{w: w, base: off, off: off}`.

**Three `utf8.RuneError` bugs in the regexp engine**, all the same family and all to be ported
together:
- `regexp.minInputLen` counts `RuneError` as 3 bytes, but an invalid input byte decodes to
  `RuneError` with width 1, so the minimum-length precheck rejects inputs that do match.
  Verified: `MustCompile("�").FindIndex([]byte("\xff"))` → `[]` in gno, `[0 1]` in Go. Also
  affects single-rune character classes, since `parser.push` rewrites those to `OpLiteral`.
- `regexp.onePassPrefix` gathers `RuneError` into the literal prefix. Verified:
  `MustCompile("a\\x{fffd}b").LiteralPrefix()` → `("a�b", true)` in gno vs `("a", false)` in Go
  — both the prefix *and* the completeness flag are wrong.
- `syntax.Prog.Prefix` has the identical defect on the non-one-pass path. Verified:
  `MustCompile("�").FindIndex([]byte("hello\xffworld"))` → `[]` in gno, `[5 6]` in Go.

**`syntax.Regexp.Equal` ignores `FoldCase`.** Not a niche API nit: `Equal` is called from
`parser.factor` to decide whether two alternation branches share a leading class, so gno factors
branches that must not be factored and the surviving prefix inherits whichever fold bit came first —
changing what the whole regexp matches. Here the verifier found the analyzer's example was wrong
(`Equal(Parse("(?i)a"), Parse("a"))` already returns false, because `minFoldRune` maps the runes
differently) **but the underlying bug is worse than claimed**: `Equal("(?i)1","1")` and
`Equal("(?i)[0-9]","[0-9]")` both return `true` in gno and `false` in Go.

**`math.FMA` loses the sign of zero.** gno folds `z == 0.0` into the zero/Inf/NaN early return and
evaluates `x*y + z`; when the exact product underflows to −0 and `z` is +0, IEEE addition yields +0.
Verified under the GnoVM: `FMA(0x1p-1022, -0x1p-1022, +0)` gives bits `0x0000…`, where Go 1.26 and
the upstream test table want `0x8000…`. Also verified the fix is realisable under gno's softfloat.
Go issue #73757. Three-line fix.

**`bufio.Reader.Discard`/`WriteTo` do not invalidate `lastByte`/`lastRuneSize`.** Verified:
`ReadByte(); Discard(1); UnreadByte()` succeeds and rewinds into the middle of the discarded region
instead of returning `ErrInvalidUnreadByte`.

**`bufio` zero-value `Reset` leaves `buf` nil** (Go 1.18 made the zero value usable). The verifier
substantially rewrote this one: the analyzer claimed the reader "never delivers a byte" via an
`ErrNoProgress` latch. That mechanism is wrong. Plain `Read(p)` actually works fine (it always takes
the large-read fast path and bypasses `fill()`); what really happens is that `fill()`-dependent
methods — `ReadByte`, `ReadRune`, `Peek`, `ReadSlice`/`ReadBytes`/`ReadString`/`ReadLine` — **panic
immediately** with *"bufio: tried to fill full buffer"*. Symmetrically, bulk `Write` degrades to
unbuffered but `WriteByte` panics on a nil-slice index.

### Missing API worth having

- **`bytes.Cut` / `CutPrefix` / `CutSuffix`** (Go 1.18/1.20) — absent, while `strings` already has
  all three. Confirmed by compilation, not just grep. `bytes` is marked `full` in the compat doc, so
  this is undocumented drift. Trivial.
- **`crypto/ed25519` has only `Verify`** — `Sign`, `NewKeyFromSeed` and the `PrivateKey`/`PublicKey`
  types are all absent. Both are deterministic pure math with no randomness, and gno's existing
  strategy (a native forward to real Go `crypto/ed25519`) extends cheaply. Worth flagging in docs
  that on-chain `Sign` means the private key is a public transaction argument, so realistic uses are
  test fixtures and deterministic derivation.
- **`unicode.Cn`, `unicode.LC`, `unicode.CategoryAliases`** (Go 1.25). gno has 36 categories, Go has
  38, and Go's 42-entry `CategoryAliases` has no gno equivalent. Not cosmetic: gno's own regexp reads
  this map, so `\p{Cn}`, `\p{LC}`, `\p{Cased_Letter}` and `\p{cntrl}` are all unsupported patterns
  today.
- **`unicode/utf8.RuneCount`** still hand-decodes. Go's `for range s { n++ }` form is a large win
  under the GnoVM, because `range` over a string is executed natively. Independently re-measured:
  **52,263,205 gas → 3,344,489 gas** for the same 200-iteration workload. `for range` with no
  iteration variables is Go 1.4 syntax, well inside the Go 1.17 target.

### Needs a decision, not a patch

**`unicode.C`/`unicode.Other` changed meaning in Go 1.25** — redefined from `Cc|Cf|Co|Cs` to
`Cc|Cf|Cn|Co|Cs`, so it now covers unassigned code points. This is the one place in the audit where
identical input yields a different answer: verified `Is(C, 0x0378)` is `false` under gno and `true`
under Go 1.26. Blast radius is narrower than it looks — `IsControl`/`IsPrint`/`IsGraphic` never
reference `C` — so the only reachable surfaces are direct `unicode.Is(unicode.C, r)` calls and
regexp `\p{C}`. Because these tables feed a deterministic VM, adopting the change alters results for
already-deployed realms. **This one warrants a conscious call rather than a silent table refresh.**

**`math/rand`'s global source is a fixed-seed `PCG(0,0)`.** Verified: `rand.Int64()` returns the same
three values in every transaction forever, byte-identical to an explicit `New(NewPCG(0,0))`. That is
arguably correct for a deterministic VM, but the package doc only says outputs "might be easily
predictable regardless of how it's seeded" — a generic PRNG caveat that never states the
load-bearing fact of total, permanent determinism. Footnote 12 covers only the v1→v2 renames. Needs
an explicit doc warning.

---

## Tier 2 and 3 — full per-package inventory

Priorities: **H** high · M medium · L low. "Go" is the release that introduced the change, where it
could be pinned.

### `regexp/syntax` — 10 items (4 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | trivial | — | Prog.Prefix gathers utf8.RuneError into the literal prefix |
| **H** | bug-fix | trivial | — | Regexp.Equal ignores FoldCase, corrupting alternation prefix factoring (wrong matches) |
| **H** | bug-fix | medium | 1.19 | Missing nesting-depth limit: ErrNestingDepth / maxHeight / checkHeight |
| **H** | bug-fix | medium | 1.20 | Missing compiled-size and rune-count limits: ErrLarge / maxSize / maxRunes (CVE-2022-41715) |
| M | behavior-change | large | 1.21 | Regexp.String() still uses the pre-Go-1.21 per-node flag printer (printFlags rewrite missing) |
| M | missing-api | small | 1.22 | (?<name>re) named capture syntax unsupported |
| M | missing-api | medium | 1.25/1.26 | unicodeTable lacks \p{Assigned}, \p{ASCII}, category aliases and inexact name matching |
| M | test | small | 1.19-1.25 | Test tables missing all upstream regression cases (limits, named captures, unicode aliases, TestString) |
| L | other | trivial | — | Stale hand-expanded fallthrough workaround in parseEscape octal handling |
| L | perf | trivial | — | IsWordChar operand order (lowercase tested first) |

### `regexp` — 7 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | trivial | — | minInputLen counts utf8.RuneError as 3 bytes, causing missed matches on invalid UTF-8 |
| **H** | bug-fix | trivial | — | onePassPrefix gathers utf8.RuneError into the literal prefix |
| M | missing-api | trivial | 1.21 | Regexp.MarshalText / UnmarshalText / AppendText missing |
| M | perf | small | — | One-pass engine effectively disabled: compileOnePass missing the hasAlt relaxation |
| M | test | trivial | — | Test gap: invalid-UTF-8 and alternation cases in findTests |
| L | doc | trivial | — | Package doc: wrong submatch index range and missing invalid-UTF-8 paragraph |
| L | perf | trivial | — | Inst copied by value in the backtracker and one-pass hot loops |

### `strings` — 16 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | behavior-change | trivial | — | explode() rewrites invalid UTF-8 bytes to U+FFFD (and skips the last element), making Split(s, "") lossy |
| **H** | bug-fix | trivial | — | genSplit missing `n > len(s)+1` clamp — caller-controlled n drives an unbounded make([]string, n) |
| M | bug-fix | small | — | Repeat: no count==1 fast path, stale panic message, no common-literal fast paths (do not port the chunk loop) |
| M | missing-api | trivial | 1.21 | strings.ContainsFunc missing |
| M | missing-api | trivial | 1.18 | strings.Clone missing |
| M | perf | medium | — | Trim/TrimLeft/TrimRight allocate a closure and call it per rune — but Go 1.26's replacement must NOT be ported verbatim |
| M | perf | trivial | — | ToUpper/ToLower ASCII fast path writes one byte at a time instead of batching unchanged runs |
| M | perf | small | — | EqualFold has no ASCII fast path (and re-slices both inputs once per character) |
| M | perf | medium | — | lastIndexFunc re-slices the string on every iteration, making all backwards scans ~4.7x more expensive and quadratic in allocation budget |
| M | test | large | — | No strings_test.gno / search_test.gno / compare_test.gno — the core API is entirely untested |
| L | bug-fix | trivial | — | Join has no output-length overflow guard |
| L | doc | trivial | 1.18 | Title is missing the Deprecated marker |
| L | doc | trivial | — | go-gno-compatibility.md lists strings as `full`, which is not true |
| L | doc | trivial | — | Doc-comment drift beyond Title: missing behavioural guarantees, cross-references and doc links |
| L | perf | small | — | IndexRune falls back to Index(s, string(r)) for multi-byte runes instead of last-byte search |
| L | test | small | 1.18/1.20 | Examples missing for Cut, CutPrefix, CutSuffix and ToValidUTF8 |

### `bytes` — 13 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | trivial | — | SplitN/SplitAfterN/explode allocate n slice headers without capping n to len(s)+1 |
| **H** | missing-api | trivial | 1.18 | bytes.Cut / CutPrefix / CutSuffix missing (strings already has all three) |
| M | missing-api | trivial | 1.20 | bytes.Clone missing |
| M | perf | medium | — | Trim/TrimLeft/TrimRight still build a closure per call instead of specialized byte/ASCII/Unicode loops |
| M | perf | small | — | EqualFold lacks the ASCII fast path |
| M | test | medium | — | Regression tests for the missing APIs and for the SplitN allocation cap are absent |
| L | behavior-change | small | — | Repeat: exact bits.Mul overflow check, renamed panic message, and chunk-limited copy |
| L | doc | trivial | 1.18 | Doc comments state behaviour that no longer matches Go: Reader.Size and Title |
| L | missing-api | trivial | 1.21 | bytes.ContainsFunc missing |
| L | missing-api | trivial | 1.21 | Buffer.Available and Buffer.AvailableBuffer missing |
| L | missing-api | trivial | 1.26 | Buffer.Peek missing (new in Go 1.26) |
| L | perf | medium | — | IndexRune still delegates to Index instead of scanning on the last UTF-8 byte |
| L | perf | trivial | — | Map still uses manual maxbytes/nbytes bookkeeping instead of append + utf8.AppendRune |

### `net/url` — 11 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | large | 1.26 | parseHost does not validate bracketed hosts as IPv6 literals |
| **H** | bug-fix | small | 1.26 | parseHost uses last colon instead of first, accepting multi-colon hosts |
| M | bug-fix | trivial | 1.23 | ResolveReference drops the base URL's Opaque field |
| M | bug-fix | trivial | — | JoinPath swallows the setPath error and returns a wrong result |
| M | perf | medium | 1.26 | shouldEscape replaced by a 256-entry lookup table |
| M | test | small | — | TestInvalidUserPassword and TestRejectControlCharacters are disabled by stale "not yet supported" markers |
| L | behavior-change | small | 1.26 | ParseQuery has no limit on the number of query parameters |
| L | doc | trivial | — | URL struct field documentation is the pre-restructure version |
| L | missing-api | trivial | 1.24 | (*URL).AppendBinary missing |
| L | perf | trivial | 1.26 | unescape re-tests the encoding mode on every '+' byte |
| L | perf | trivial | 1.26 | Values.Encode short-circuits only on nil, and grows the key slice by append |

### `time` — 13 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | small | 1.23 for the… | Parse accepts out-of-range and non-numeric time-zone offsets |
| **H** | bug-fix | trivial | <=1.22 | stdSecond parse silently accepts a missing seconds field when a fractional second follows |
| M | bug-fix | trivial | <=1.22 | Location.lookup passes `end` (omega) instead of the last transition time to tzset |
| M | bug-fix | small | 1.20 | MarshalJSON/MarshalText emit invalid RFC 3339 for zone offsets outside [-23:59, +23:59] |
| M | bug-fix | trivial | <=1.22 | stdFracSecond9 parse caps trailing fractional digits at 9, causing spurious "extra text" errors |
| M | missing-api | trivial | 1.20 | Time.Compare missing |
| M | missing-api | medium | 1.20 | format_rfc3339.go absent: no strict RFC 3339 parse, no RFC3339 fast paths, weak UnmarshalJSON errors |
| M | test | medium | — | No _test.gno coverage for time; Go's regression tests for the bugs above are unported |
| L | bug-fix | trivial | 1.23 | Parse rejects 'Z' for the Z070000 and Z07:00:00 layouts |
| L | bug-fix | small | <=1.22 | ParseError.ValueElem reports the post-consumption remainder instead of the failing element |
| L | missing-api | small | 1.24 | Time.AppendBinary and Time.AppendText missing |
| L | perf | large | 1.24 | Neri-Schneider calendar algorithms not adopted (date/clock decomposition) |
| L | perf | small | — | appendInt and formatNano/appendNano still use scratch arrays |

### `io` — 4 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | trivial | — | NewOffsetWriter never sets base, breaking WriteAt/Seek(SeekStart) for any nonzero initial offset |
| M | missing-api | small | believed Go … | multiReader missing WriteTo (Go 1.20): io.Copy(dst, io.MultiReader(...)) can't take the zero-copy WriterTo path |
| M | perf | small | — | ReadAll still uses the pre-chunked doubling algorithm; Go's version returns an exactly-sized final slice instead of a doubling-grown one |
| L | test | trivial | — | TestSectionReader_Max missing: no regression test for NewSectionReader's maxint64 overflow guard |

### `bufio` — 11 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | behavior-change | trivial | 1.18 | Reset on a zero-value Reader/Writer leaves buf nil, causing an immediate panic on common methods (not a silent hang) |
| **H** | bug-fix | trivial | — | Reader.Discard and Reader.WriteTo do not invalidate lastByte/lastRuneSize |
| M | behavior-change | small | — | Writer.ReadFrom never delegates to the underlying io.ReaderFrom when the buffer is non-empty |
| M | bug-fix | trivial | — | Scanner.Scan returns true for a nil token delivered with ErrFinalToken |
| M | bug-fix | trivial | — | Reader.WriteTo issues a zero-length Write when the buffer is empty, masking the read error |
| M | bug-fix | trivial | — | Reader.Reset/Writer.Reset lack the self-reset guard (b == r), causing unbounded recursion |
| M | missing-api | trivial | 1.18 | Writer.AvailableBuffer missing (Go 1.18 API) |
| M | perf | small | — | Writer.WriteString does not forward large writes to an underlying io.StringWriter |
| M | test | medium | — | No bufio_test.gno at all -- Reader/Writer are entirely untested |
| M | test | small | — | example_test.gno missing AvailableBuffer, ReadFrom and Scanner early-stop examples |
| L | doc | trivial | — | Reader.Read doc claims a zero count at EOF, contradicting actual behaviour |

### `crypto/subtle` — 2 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | missing-api | trivial | 1.0 | Constant-time comparison primitives (ConstantTimeCompare/Select/ByteEq/Eq/Copy/LessOrEq) entirely absent |
| L | missing-api | trivial | 1.24 | WithDataIndependentTiming missing (and footnote 6 doesn't mention it) |

### `crypto/ed25519` — 2 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | missing-api | small | 1.13 | Sign, NewKeyFromSeed, and the PrivateKey/PublicKey types are entirely absent -- only Verify exists |
| L | missing-api | medium | 1.20 | VerifyWithOptions/Options (Ed25519ph/Ed25519ctx variants) missing |

### `crypto/cipher` — 4 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | 1.0 | StreamReader/StreamWriter are field-only structs with no Read/Write/Close methods |
| M | missing-api | medium | 1.0 | None of the block-mode constructors (NewCBCEncrypter/Decrypter, NewCFBEncrypter/Decrypter, NewCTR, NewOFB) exist |
| M | missing-api | large | 1.2 | GCM (NewGCM, NewGCMWithNonceSize, NewGCMWithTagSize, NewGCMWithRandomNonce) missing |
| L | doc | trivial | — | AEAD.Seal/Open doc gained a 'dst and additionalData may not overlap' constraint gno's interface doc lacks |

### `crypto/sha256` — 1 item

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | large | 1.0/1.2 | Incremental hash.Hash API (New, New224, Sum224, BlockSize, Size224) missing -- only Sum256 exists |

### `encoding/base64` — 7 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | behavior-change | trivial | 1.22 | NewEncoding does not reject duplicate alphabet symbols |
| M | bug-fix | trivial | 1.22 | WithPadding accepts negative padding runes other than NoPadding |
| M | missing-api | small | 1.22 | base64 Encoding.AppendEncode/AppendDecode missing |
| M | test | small | 1.21 | Strict-mode and decode-bounds regression tests are commented out |
| L | bug-fix | trivial | 1.22 | EncodedLen/DecodedLen integer-overflow fix not applied |
| L | other | trivial | — | Stale "XXX fallthrough not yet implemented" hand-unrolled switch in decodeQuantum |
| L | perf | trivial | 1.22 | NewEncoding fills decodeMap with a 256-iteration interpreted loop |

### `encoding/hex` — 4 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | behavior-change | small | — | InvalidByteError message prints a decimal byte instead of Go's %#U form |
| M | missing-api | small | 1.22 | hex.AppendEncode/AppendDecode missing |
| M | perf | small | — | Decode still uses the fromHexChar switch instead of reverseHexTable |
| L | behavior-change | trivial | — | DecodeString decodes in place and returns a slice aliasing a 2x-sized buffer |

### `encoding/csv` — 2 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | behavior-change | trivial | — | ParseError.Error uses ':' where Go uses ';' between record line and parse line |
| L | other | trivial | — | errInvalidDelim is exported as ErrInvalidDelim in gno |

### `encoding/binary` — 1 item

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | small | 1.23 for App… | Read/Write/Size/Append/Encode/Decode fast path (intDataSize+encodeFast/decodeFast) is portable without reflect -- the blanket reflect/skip framing under-scopes what's actually portable |

### `encoding` — 1 item

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | 1.24 | encoding.BinaryAppender / encoding.TextAppender missing (Go 1.24) |

### `errors` — 3 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | doc | trivial | — | go-gno-compatibility.md footnote 8 and the io/hash/hash-adler32/encoding status-table rows are stale |
| M | missing-api | trivial | 1.21 | errors.ErrUnsupported missing (Go 1.21) |
| L | perf | trivial | — | errors.Is skips Go's err==nil fast path, always paying the recover-based comparable() check |

### `hash` — 2 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | 1.25 | hash.Cloner missing (Go 1.25) |
| L | missing-api | trivial | 1.25 | hash.XOF missing (Go 1.25) |

### `hash/adler32` — 2 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | AppendBinary… | digest is missing AppendBinary and Clone, blocked only by the two interfaces above |
| M | test | small | — | hash/adler32 has zero tests, including the RFC 1950 golden vectors that exercise the nmax=5552 chunking boundary |

### `math` — 7 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | bug-fix | trivial | — | FMA(x, y, ±0) returns +0 where the product underflows to -0 (Go issue #73757) |
| M | perf | small | — | Trunc still implemented via Modf; port Go's direct-bit trunc but NOT Go's Modf inversion |
| M | perf | trivial | — | Exp/Exp2 pay two redundant IsInf calls that Go removed (~13% of every Exp call under GnoVM) |
| M | test | small | — | Ceil/Floor/Trunc special-case tests never widened to the 1<<52 boundary; gno reuses ceilSC for all three |
| M | test | small | 1.11 | huge_test.go was never ported: no coverage for Payne-Hanek trig reduction on huge arguments |
| L | test | small | — | example_test.go (34 runnable examples) not ported although gno supports Example tests |
| L | test | trivial | not-applicab… | all_test.gno error strings say "trigmath.Reduce" - search-and-replace artifact from the fork's math.-prefixing pass |

### `math/bits` — 1 item

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| L | test | small | 1.9 | Go's runnable examples not ported |

### `math/rand` — 6 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | doc | trivial | 1.22 | Global rand source is a fixed-seed PCG(0,0); package doc never says so |
| M | test | medium | 1.22 | TestRegress is entirely commented out -- the golden table is dead code |
| M | test | small | 1.19 | TestAuto is a no-op safeguard in gno: it cannot detect (and does not flag) the deterministic global seed |
| L | missing-api | trivial | 1.23 | Rand.Uint() and package-level Uint() missing |
| L | missing-api | large | 1.22 | ChaCha8 generator missing entirely |
| L | missing-api | trivial | 1.24 | PCG.AppendBinary missing (encoding.BinaryAppender) |

### `sort` — 10 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | 1.19 | sort.Find missing; the compat doc excludes it on a rationale that does not apply |
| M | perf | medium | 1.19 | sort.Sort still uses the pre-Go-1.19 Bentley–McIlroy quicksort instead of pdqsort |
| M | test | small | — | All float64 sort coverage is commented out, leaving Float64Slice.Less NaN ordering untested |
| M | test | small | — | Interface-level Sort tests disabled under a false "reflect" justification |
| M | test | trivial | 1.19 | TestFind and TestFindExhaustive have no gno counterpart |
| M | test | trivial | 1.19 | pdqsort's own tests (TestBreakPatterns, TestReverseRange) and the Go 1.19 sorted/reversed/mod8 benchmarks are missing |
| L | doc | trivial | 1.21 | Interface.Less documented as "transitive ordering" rather than a strict weak ordering |
| L | doc | trivial | — | Sort/Stable doc comments predate Go's "ascending order as determined by the Less method" clarification |
| L | other | trivial | — | lessSwap is dead code in gno, alongside a go:generate line pointing at a nonexistent file |
| L | test | trivial | — | Stale XXX in search_test.gno: keyed composite literals now work |

### `strconv` — 9 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | perf | large | 1.26 | FormatFloat with 'f' and fixed precision always falls back to the multiprecision bigFtoa path |
| M | perf | small | 1.26 | Base-10 integer formatting still uses the 65-byte formatBits scratch buffer instead of a 24-byte formatBase10 |
| M | test | medium | 1.26 | Missing the new testbase fast-vs-slow-path regression corpora (atof1k.txt / ftoa1k.txt) |
| M | test | trivial | 1.26 | Missing new ftoa test cases for %f rounding and the David Chase shortest-representation cases |
| L | doc | trivial | 1.23-1.25 | Unquote doc omits the empty-character-literal clause |
| L | doc | trivial | 1.23-1.25 | FormatFloat doc omits the exponent-width guarantee |
| L | doc | trivial | 1.19-1.25 | Doc comments throughout gno's strconv lack Go's [Symbol] doc links and bulleted fmt list |
| L | perf | trivial | 1.26 | genericFtoa lacks the mant == 0 early return |
| L | perf | large | 1.26 | Shortest-form ftoa: Go 1.26 replaced Ryu with Dragonbox and deleted ftoaryu.go upstream |

### `unicode` — 5 items (2 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | behavior-change | medium | 1.25 | unicode.C / unicode.Other no longer excludes unassigned code points |
| **H** | missing-api | medium | 1.25 | unicode.Cn, unicode.LC and unicode.CategoryAliases missing |
| M | perf | small | — | SimpleFold does two CaseRanges binary searches where Go now does one |
| M | test | trivial | 1.25 | script_test missing the Cn/LC/C category regression cases |
| L | doc | trivial | 1.19 | Doc-comment drift: Go 1.19 doc links and reworded RuneLen comment |

### `unicode/utf8` — 3 items (1 high)

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| **H** | perf | small | — | RuneCount/RuneCountInString still hand-decode UTF-8 instead of ranging over the string |
| M | perf | small | — | ValidString's ASCII scan is 2.8x more expensive than a range-based equivalent under the GnoVM |
| M | test | small | — | Missing alignment-sweep regression cases for Valid/ValidString |

### `unicode/utf16` — 2 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| M | missing-api | trivial | 1.23 | utf16.RuneLen missing |
| M | test | trivial | 1.20 | utf16.AppendRune ships with no test |

### `html` — 3 items

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| L | doc | trivial | — | EscapeString/UnescapeString doc comments missing Go doc-link syntax |
| L | missing-api | large | — | html/template subpackage entirely absent |
| L | test | trivial | — | example_test.go (ExampleEscapeString / ExampleUnescapeString) not ported |

### `internal/bytealg` — 1 item

| P | Kind | Effort | Go | Item |
|---|---|---|---|---|
| L | missing-api | small | 1.21 | LastIndexRabinKarp missing; gno still carries the pre-generics duplicated Rabin-Karp helpers |

---

## Confirmed deliberate exclusions (35)

These were examined and are correctly absent; no action needed. They are listed so the same ground
is not re-covered next time.

**Concurrency** — `strings.Replacer` eager build vs `sync.Once`; `time.FixedZone` memoisation;
Timer/Ticker/Sleep/After/AfterFunc/Tick; `regexp`'s `sync.Pool` machine reuse (verified correctly
stubbed); `html` entity maps behind `sync.OnceValues`.

**Generics** — the whole `iter.Seq` family in `strings` and `bytes` (`Lines`, `SplitSeq`,
`SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq`); `strconv`'s generic `bsearch`; `sort`'s delegation to
`slices`; `errors.AsType`; `math/rand`'s top-level `N[Int]`; `utf8.Valid`'s word-at-a-time rewrite.

**reflect / unsafe / nondeterminism** — `sort.Slice`/`SliceStable`/`SliceIsSorted`;
`binary.NativeEndian`; `ed25519.GenerateKey`; `regexp`'s file-I/O conformance suites.

**Go-runtime-specific** — `internal/bytealg`/`internal/abi`/`unsafe` delegation in `strings`;
`Buffer.grow`'s size-class awareness; inlineability-driven fast-path splits in `utf8`; `math/bits`
bounds-check-elimination conversions.

**Do NOT port even though they look portable** — several upstream rewrites are behaviour-identical
refactors that would be *worse* under the GnoVM. The clearest: Go 1.26's `strings.TrimSpace`
rewrite uses `range []byte(s)`, which allocates a full copy in gno. Likewise `bytes`'s `indexFunc`,
`FieldsFunc`, `Buffer.WriteRune`, `Replace` and `LastIndex` refactors, `strconv`'s dropped ASCII fast
path in `appendQuotedWith`, and `math`'s `IsInf`/`Signbit` micro-rewrites.

---

## Refuted (2)

Recorded so they are not re-proposed.

- **`math`: the `dim.go` doc note contrasting `Max`/`Min` with the builtin `max`/`min`** — not
  applicable, since gno has no such builtins (`grep` over `uverse.go` confirms). Bodies are otherwise
  byte-identical.
- **`net/url`: `URL.String` should pre-size its `strings.Builder`** — refuted **by measurement**. The
  code difference is real but the optimisation inverts under the GnoVM. Four A/B gas arms doing
  identical work: no-`Grow` 287,116,360 gas; `Grow(constant)` 345,091,067 (**+20%**);
  `Grow(computed)` 388,810,470 (**+35%**); `Grow(loop-computed)` 420,117,222 (**+46%**).

---

## Documentation fixes this audit turned up

`docs/resources/go-gno-compatibility.md` has drifted and should be corrected alongside any porting:

- **Footnote 8 is wrong.** It says `errors` ships `New` only; `Unwrap`, `Is` and `Join` all landed in
  PR #5385. `As` remains correctly excluded (reflect), and `AsType` is correctly excluded (generics).
- **`bytes` is marked `full`** but lacks `Cut`/`CutPrefix`/`CutSuffix`, `Clone`, `ContainsFunc`,
  `Buffer.Available`/`AvailableBuffer`/`Peek`.
- **`strings` is marked `full`** but lacks `Clone` and `ContainsFunc`.
- **`unicode` is marked `full`** with no footnote, but lacks `Cn`, `LC` and `CategoryAliases`.
- **Footnote 7 (`encoding/binary`) is over-broad.** It excludes `Read`/`Write`/`Size` as
  reflect-dependent, but the `intDataSize` + `encodeFast`/`decodeFast` path is portable without
  reflect, and it omits `Append`/`Encode`/`Decode` entirely.
- **Footnote 13 (`sort`) excludes `sort.Find` on a rationale that does not apply** — `Find` is not
  reflect-dependent.
- **Footnote 6 (`crypto/subtle`) omits `WithDataIndependentTiming`.**
- **Footnote 12 (`math/rand`) should state the fixed-seed determinism** of the global source.
- The `io`/`hash`/`hash/adler32`/`encoding` status rows are stale.

---

## Caveats

- The audit compares against **Go 1.26.1 as installed locally**. Where a change is godebug-gated
  upstream (notably `urlstrictcolons`), the gate is noted in the finding.
- **Go version attributions are best-effort.** No Go git history was available locally, so versions
  were pinned from `/usr/local/go/api/*.txt` where the symbol is tracked there, and left as `—`
  otherwise. Method changes on unexported types are not tracked in `api/*.txt` at all.
- **Porting is not mechanical.** Gno targets the Go 1.17 language spec; several upstream fixes are
  delivered via generics or builtins that must be hand-lowered, and a few upstream "optimisations"
  are pessimisations under a tree-walking interpreter (see the refuted `URL.String` entry and the
  "do NOT port" list). Each item's `recommendation` field distinguishes `port` from `port-adapted`.
- Findings are machine-generated and adversarially reviewed, not human-audited. Every behavioural
  claim was executed on both sides, but **the fixes themselves have not been written or tested.**
