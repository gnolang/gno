# ADR: Bound the TOML decoder (fork), and charge gas for the gnomod.toml decode

## Status

Proposed (PR pending). Fixes two dora findings against
`gnovm/pkg/gnomod/toml.go`: `1b774a7a-560e-43ba-9c48-563b371b0701` (fatal stack
overflow from deeply nested TOML) and `cf80016b-206c-48c6-8847-e569c5d30bfb`
(ungas-metered CPU amplification).

## Context

`gnomod.toml` travels inside the `MemPackage` of a `MsgAddPackage` or `MsgRun`,
so its body is attacker-controlled with no size or shape restriction:
`MemFile.ValidateBasic` bounds the file *name* (256 chars) and nothing else,
`.toml` is an allowed extension, and `gnomod.toml` matches `reFileName`.

Every message decodes that body **twice**:

1. `gno.TypeCheckMemPackage` → `typeCheckMemPackage` → `ParseCheckGnoMod`
   (`gnovm/pkg/gnolang/gnomod.go:64`) → `gnomod.ParseMemPackage`;
2. the keeper's own checks — `AddPackage` at `keeper.go:749`, or, for `MsgRun`,
   the same type-check at `keeper.go:1131` before the generated `gnomod.toml`
   replaces the caller's at `keeper.go:1172`.

Both land in `parseTomlBytes` → `toml.Unmarshal` (pelletier/go-toml v1.9.5),
which has no depth limit, no size limit and no allocation accounting. Three
independent costs are unbounded in it.

**Nesting → goroutine stack.** `parseRvalue`, `parseArray` and
`parseInlineTable` are mutually recursive, one frame pair (~1.4KB) per nesting
level, and closing brackets are not needed to make it descend — `a = [[[[…`
recurses to EOF before reporting the syntax error. Measured on this checkout
(`StackSys`, so it includes the doubling overshoot):

| `a = ` + N × `[` | goroutine stack | decode |
|---|---|---|
| 4 KB | 9.0 MB | 7.4 ms |
| 16 KB | 33.6 MB | 27 ms |
| 64 KB | 132 MB | 155 ms |

Probing the decoder directly, `a = ` + N × `[` survives at N = 370,000 and dies
at N = 400,000 — so under 400KB of input, well inside `MaxTxBytes` (1MB), and
no balancing closers needed. Reproduced with dora's balanced 840KB payload:

```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
github.com/pelletier/go-toml.(*tomlParser).parseArray  parser.go:447
github.com/pelletier/go-toml.(*tomlParser).parseRvalue parser.go:385
[… repeating …]
```

A stack overflow is a fatal runtime error, not a panic: `baseapp.runTx`'s
`defer`/`recover` cannot catch it and neither can anything else. Every node that
processes the message dies, so this is a consensus halt from a single message,
and it is reachable at genesis too (`InitChain` → `AddPackage`).

**Table count → O(n²) CPU.** `parseGroup` rescans `seenTableKeys` once per
table (`parser.go:146`), and `parseGroupArray` does the same with a deletion
(`parser.go:120`), so cost is quadratic in the number of tables:

| `[tNNNNNN]` tables | body | decode (×1) |
|---|---|---|
| 1 K | 14 KB | 3.9 ms |
| 10 K | 140 KB | 203 ms |
| 50 K | 700 KB | 4.9 s |
| 70 K | 980 KB | 9.3 s |

At 980KB that is ~18.6s of CPU per message, against ~9.8M gas — the ante
handler's `TxSizeCostPerByte` (10 gas/byte) was the only charge, because
`chargePreprocessGas` counted `.gno` files only and skipped the mod file
outright. At 1 gas ≈ 1ns (the calibration `PreprocessGasPerByte` uses), that is
a ~1900x undercharge.

Note on scale: dora's report puts the aggregate at ~299 such transactions per
block from `MaxBlockMaxGas` alone. `MaxBlockDataBytes` (2MB) is the binding
limit, not gas, so the real aggregate is ~2 maximum-size transactions and ~37s
of CPU per block. Still far past any block time, and finding 1 needs only one
message either way.

**Key-path depth → O(n·d) CPU.** `parseAssign` calls `Tree.GetPath(currentTable)` for every assignment (`parser.go:186`), walking the whole current-table key path each time. A dotted table header `d` levels deep with `n` assignments under it therefore costs O(n·d), and both grow with the body, so this is quadratic in it — while spending a single `[`:

| body | dotted-table decode (×1) |
|---|---|
| 4 KB | 2.8 ms |
| 8 KB | 14 ms |
| 64 KB | 827 ms |
| 256 KB | 25.7 s |

Measured 3.94x / 4.06x / 5.07x per doubling. This axis is why `maxDelims` alone is not enough, and it is the one the first revision of this ADR got wrong: go-toml genuinely does not *recurse* on a dotted key — `Tree.SetPath`/`GetPath` are iterative loops, so no stack grows, and a 1M-level key really does decode in 549ms with 1MB of stack — but the per-assignment re-walk makes the *combination* of a deep header and many assignments superlinear, which measuring a lone dotted key does not show.

## Decision

Bound the decoder itself, in a fork, and keep only a size cap in the caller.

`tm2/pkg/toml` is pelletier/go-toml v1.9.5 vendored with three changes; `gnomod`
keeps `maxFileSize` and an error-text cap; `chargePreprocessGas` charges the mod
file. The bulk of the reasoning below is about *where* the bounds go, because
that is what four revisions of this branch got wrong.

### 1. The fork: `tm2/pkg/toml`, with `maxNestingDepth` and `maxKeyDepth`

Two limits and one outright fix, 33 lines of code across two files
(`parser.go`, `keysparsing.go`); the deviation is kept as `gno.patch` with a
`Makefile` that rebuilds the tree from the upstream module, mirroring how
`gnovm/pkg/parser` vendors Go's own parser.

- **`maxNestingDepth = 256`** — a counter incremented in `parseRvalue`, the
  single choke point every level of array/inline-table nesting passes through.
- **`maxKeyDepth = 16`** — checked against the path `parseKey` returns, so a
  dotted key, adjacent quoted segments and a quoted segment spanning newlines all
  count alike.
- **The O(n²) table rescan is fixed, not limited** — a `map` kept alongside
  `seenTableKeys` makes the duplicate-table check O(1) and the table count linear.
  No document is rejected for it. 700KB of headers went from 4.09s to 205ms.

Both limits are checked against the parser's **own state**. That is the whole
point, and the reason this is not done in the caller: a caller reading raw bytes
cannot tell structure from text, and every byte-level approximation this branch
tried was wrong in one direction or the other (see *Rejected* below). The parser
already knows how deep it has recursed and what key path it just parsed.

#### Why not upstream, and why not v2

Upstream is not available: go-toml's `SECURITY.md` lists **all 1.x as
unsupported**, and its README declares v1 deprecated with "no active development
is expected on it". There is nowhere to send this patch and no fix coming.

Migrating to go-toml v2 does not fix the class either — measured on this
checkout:

| axis | v1 | v2 |
|---|---|---|
| nesting (`a = ` + N × `[`) | fatal overflow at ~400KB | survives 900KB, **fatal overflow at ~1MB** |
| 50K tables (700KB) | ~4.1s | **19.0s** |
| 1015-level dotted key | 2.8ms | 9.6ms |

v2 has the same unbounded mutual recursion (`parseVal`/`parseValArray`/
`parseInlineTable`); its only protection on the chain-halt vector is that the
payload barely does not fit in `MaxBlockTxBytes` (1MB), which is not a safety
margin. It is also slower on both other axes, and larger (7462 vs 4781 non-test
LOC), so the patch surface grows. Forking v1 is the cheaper and safer option.

#### Choosing the limits

Both are far above anything real and far below what an attack needs. The deepest
key path in any of the ~440 mod files in this repository is 1; a tm2
`config.toml` reaches 2. Real documents nest arrays 1–3 deep.

`maxKeyDepth` is 16 rather than larger because of what it does to the worst case.
Worst admitted 4KB body, by limit:

| `maxKeyDepth` | worst 4KB decode | worst shape |
|---|---|---|
| 16 | 1.17ms | an ordinary **flat** body |
| 32 | 2.37ms | deep key path |
| 64 | 8.56ms | deep key path |

At 16 a deep key path stops being the most expensive thing a size-capped body can
hold — the worst case becomes a flat file, so cost is linear in length with no
shape premium. That is the forward-safety property this ADR wanted all along: a
raised `maxFileSize` raises the cost and the gas charge together. At 64 the
quadratic shape costs more (~8.56ms across the two decodes) than the 1250 gas/byte
charged over the same bytes collects (~5.12ms at 1 gas ≈ 1ns).

`maxNestingDepth = 256` costs ~90µs at the cap for a 4KB body, so it is not the
binding cost on any axis; it is set for generosity to real documents rather than
tuned.

### 2. `gnomod.maxFileSize = 4KB`

The one bound the decoder cannot infer: how many bytes the caller is willing to
spend at all. With shape bounded, cost is linear in length, so length is what
`ParseBytes` caps.

`ParseBytes` is the single funnel — `ParseDir`, `ParseFilepath`,
`MustParseBytes`, `ParseMemPackage` and `ParseCheckGnoMod` all reach a decoder
through it, and it covers the deprecated `gno.mod` form too — so the bound also
covers genesis, `gno` CLI parsing of local files, gnoweb and gnodev, none of which
have a gas meter to fall back on.

4KB is ~33x the largest `gnomod.toml` in this repository (123 bytes) and fits
~100 `[[replace]]` entries.

### 3. `maxDecodeErrBytes = 1KB` on the decoder's own error text

go-toml renders the offending value into its error message, and for a nested
table that value is the whole `*Tree` formatted with `%v`. `Tree.String` indents
by depth, so the text is O(depth²).

`maxKeyDepth` takes most of this away — the worst text a 4KB body can now produce
is **26KB**, down from **2.7MB** — but 26KB is still ~6x the body that caused it,
every caller copies it whole (`sdk.clipLog` splits it line by line before capping,
so its cap does not save the allocation), and it reaches consensus-visible logs.
Nothing reads it programmatically, so `parseTomlBytes` truncates at the source.
The wrap becomes `%s` rather than `%w`: the wrapped text is the oversized part,
and no caller unwraps to a go-toml error.

### 4. `chargePreprocessGas` charges the mod file

`.gno` bytes pay `PreprocessGasPerByte` (1250); the mod file paid 10. It now pays
the same rate as source. The worst admitted body costs ~586ns/byte across the two
decodes a message pays, so 1250 is conservative for it.

With the decoder bounded, this charge is no longer what stands between a body and
a halted chain — it is what makes the remaining linear cost paid for. `gno.mod` is
charged alongside `gnomod.toml`: it is rejected outright at `keeper.go`, but only
*after* being decoded, and `ParseMemPackage` accepts either name.

### Rejected: bounding the shape from the caller's raw bytes

Four revisions of this branch tried to bound nesting and key depth in
`gnomod.ParseBytes` by counting bytes, with `maxDelims = 512`,
`maxKeyUnits = 32` per line, `maxTotalKeyUnits = 1024`, a backslash ban and a
per-line quote-parity rule. All of it is removed. The record matters because each
revision looked sound and had a passing worst-case test.

**A byte count cannot see where the decoder recurses.** A `[` may sit in a
comment or a string and buy no recursion; a closer may sit in a comment and unwind
nothing, so `a = [[[[[[[[[[#]]]]]]]]]` opens ten arrays and closes none —
4033 bytes of that shape reached 1900 real levels while a running depth counter
never passed 190. Counting monotonically fixed that, at the cost of counting
openers in comments and strings too.

**A dot count did not bound key depth.** `parseKey` appends a key group on every
quoted run and does not require a `.` between segments, so `["a""b""c"]` is the
path `a.b.c` and `[` + N × `""` + `]` is N levels for 2N bytes with no dot
anywhere. 4KB of that reached ~2000 levels at 2 dots in the whole body, decoding
in 2.85ms against the 609µs the worst-case test was pinning — and through
`MsgAddPackage` the body was not merely undercharged but **admitted**: the deploy
succeeded. Charging quote characters as well brought it back under a bound.

**A per-line count did not bound key depth either.** `lexInsideTableKey` copies
every byte through to `parseKey` until `]`, and `parseKey`'s quoted-key branch
consumes runes to the closing quote without checking for a newline, so a key path
can span lines inside a quoted segment, buying a level per 4 bytes while putting
one unit on each line. 4093 such bytes reached **1333** levels against a per-line
limit of 32 — 11ms and 40MB per message where ~5.2M gas was charged.

**And the fix for that was itself unsound.** Requiring both quote counts to be
even per line was justified as "quotes delimit pairwise, so an open string leaves
odd parity". They do not, for the counted character: a `"` inside a
single-quoted segment is not a `"` delimiter, and neither is one in a trailing
comment after `]`. So

```toml
['"'"
""
""
"] #'"'
```

keeps **even parity of both characters on every line** while the key path spans
all of them. Measured, 4094 such bytes reached **503** levels of key-path depth
and cost **2.58ms** per decode against the 611µs pinned ceiling; rooted at
`replace` so the Tree-to-struct step fails, **5.91ms**, 12.4MB of churn and a
1.31MB error string. Through `AddPackage` it consumed 5.19M gas for 5.52ms of
work — ~1.06x headroom, not the ~4x claimed. It scaled ~4x per doubling of body
size, so the "cost is linear in `maxFileSize`" property did not hold either.

The parity rule also **rejected legitimate content**: an apostrophe anywhere in a
comment (`# don't bump this`) is odd parity, and the backslash ban rejected a
Windows path in a `replace` directive — which `File.WriteString` itself emits, so
`gnomod`'s own write/read round trip was broken. Both parse fine in TOML and both
are things a human writes.

The pattern across all four is the same: a bound on a *proxy* for what the parser
does is only as good as the enumeration of what else can buy the thing being
bounded, and TOML has more ways than are obvious. `len(groups) > maxKeyDepth`
inside `parseKey` has no enumeration to get wrong.

## Consequences

The worst body the bounds jointly admit is now an ordinary flat 4KB one at
**~1.2ms** per decode (~586ns/byte across the two decodes a message pays, against
1250 gas/byte charged), down from 9.3s and a fatal unrecoverable overflow. A real
58-byte `gnomod.toml` decodes in ~24µs, fixed lexer setup dominating at that size.

Because a deep key path is no longer the worst shape, cost is linear in
`maxFileSize` with no shape premium: raising the size bound raises the cost and
the per-byte charge alike. `TestParseBytesWorstCase` still builds the deep-path
shape, since that is the one whose cost grows if `maxKeyDepth` is raised.

Rejection happens at the decode, so the transaction is rejected at `DeliverTx`
having paid its gas — the same shape as any other invalid package. The error
reaches the user through `ErrTypeCheck`'s msg trace (`Result.Log`), since the
hashed `Result.Error` carries only the sentinel.

Gas cost rises for every deploy by `PreprocessGasPerByte × len(gnomod.toml)` —
about 137,500 gas for a typical 110-byte file, ~1.4% of a small realm's deploy.
Gas is not part of `LastResultsHash` or the committed multistore root, so no
apphash moves; the `PreprocessGasPerByte` *default* is unchanged, so no new fork
repricing fingerprint is needed either. Four gas-pinning txtars were re-pinned
when the charge was added and are unchanged by the fork.

A `gnomod.toml` over 4KB, nesting past 256 levels, or with a key path past 16
segments is a hard error everywhere — including `gno` CLI use on local files and
any genesis containing one. Nothing in the repository is close: the worst of ~440
mod files is 123 bytes at key-path depth 1.

Unlike the byte-level rules this replaces, the bounds now reject **nothing that
TOML allows and a mod file might plausibly contain**. Comments with apostrophes,
escapes, multi-line strings and Windows paths in `replace` directives all parse.

Two other consumers of TOML in the tree — `tm2/pkg/bft/config` (node config) and
`gno.land/pkg/gnoland` (genesis params) — still use upstream pelletier/go-toml.
Both read operator-local files rather than attacker-controlled bodies, so they are
left alone to keep this change's blast radius at the one reachable call site.
Switching them to the fork would let the external dependency be dropped entirely
and is worth doing separately.

### What this does not do

- No general `MemFile` body limit. Other file bodies are not fed to an
  unbounded parser: `.gno` bytes are charged at `PreprocessGasPerByte`, and
  anything else is stored, not decoded. A body limit in
  `MemFile.ValidateBasic` would be worth having on its own merits — it would
  reject at `CheckTx`, before a block — but it is a separate consensus change
  with a much wider blast radius, and at any size generous enough for stdlib
  sources (~210KB) it would not have prevented the stack overflow.
- The fork bounds *decoding*. It says nothing about the encoder, and nothing
  about how much memory a decoded `*Tree` may occupy, so a caller feeding it
  untrusted input still wants a size cap of its own.
