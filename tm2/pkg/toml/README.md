# tm2/pkg/toml

A fork of [pelletier/go-toml](https://github.com/pelletier/go-toml) **v1.9.5**
with resource bounds on decoding.

## Why this exists

`gnomod.toml` rides inside the `MemPackage` of a `MsgAddPackage` or `MsgRun`, so
its body is caller-controlled, and upstream go-toml v1 has no resource bounds
of any kind. Three axes were unbounded:

| axis | upstream behaviour |
|---|---|
| array / inline-table nesting | `parseRvalue`, `parseArray` and `parseInlineTable` are mutually recursive with no depth limit. Closing brackets are not needed to descend, so `a = [[[[…` recurses to EOF: ~400,000 openers (under 400KB) exhaust Go's 1GB goroutine stack. **A stack overflow is a fatal runtime error that no `recover()` can catch**, so every node processing the message dies. |
| key-path depth | `parseAssign` re-walks the current table's key path for every assignment under it, so depth × assignments is quadratic in the document. 4KB reached 2.8ms; the shape spends no bracket at all. |
| table count | `parseGroup` rescans `seenTableKeys` once per table, so a decode is O(n²) in the table count. 1MB of `[tNNNNNN]` headers took ~9s. |

Upstream is not an option for fixing these: go-toml's `SECURITY.md` lists **all
1.x as unsupported**, and its README declares v1 deprecated with "no active
development is expected on it".

Migrating to go-toml v2 does not fix the class either — v2 has the same
unbounded mutual recursion (it survives 900KB of `[` and fatally overflows at
~1MB, i.e. its only protection is that the payload barely does not fit in
`MaxBlockTxBytes`), and it is *slower* on the other two axes (19.0s for 50k
tables, vs ~9s on v1).

## The deviation from upstream

Kept small on purpose, and recorded in `gno.patch` so it can be reviewed as a
diff and re-applied if the vendored base ever moves. See the `Makefile`.

1. **`parser.go` — `maxNestingDepth` (256).** A depth counter incremented in
   `parseRvalue`, the single choke point every level of nesting passes through.
2. **`keysparsing.go` — `maxKeyDepth` (16).** Checked against `parseKey`'s
   returned path, so a dotted key, adjacent quoted segments (`["a""b""c"]` is
   the path `a.b.c`) and a quoted segment spanning newlines all count alike.
3. **`parser.go` — O(1) duplicate-table check.** A `map` kept alongside
   `seenTableKeys`, making the table count linear rather than quadratic. This is
   a fix, not a limit: no document is rejected for it.

Both limits are checked against the parser's **own state** rather than against
raw bytes. That is the point of fixing this here instead of in each caller: a
caller cannot approximate the lexer's notion of depth from the byte stream
without getting it wrong. A `[` may sit in a comment or a string and buy no
recursion; a closer may sit in a comment and unwind nothing; a quoted key
segment may span newlines and hide its depth from any per-line count. Counting
the parsed path and the actual recursion is exact.

Also carried, both unrelated to the bounds:

4. **Four `go vet` fixes** — `l.errorf(err.Error())` → `l.errorf("%s", err.Error())`
   in `lexer.go`, so `go test ./tm2/pkg/toml/` is clean without `-vet=off`.
5. **One test made timezone-robust** — `TestUnmarshalLocalDateTime` compared raw
   wall-clock components against a `time.Local` decode, so it failed on any zone
   putting `1979-05-27T00:32` in a historical DST gap. It now normalizes both
   sides identically. Verified under UTC, Europe/Paris, Asia/Kolkata,
   America/New_York and Australia/Lord_Howe.

Files are also `gofmt`ed and run through the `go fix` modernizers (`interface{}`
→ `any`, 3-clause loops → `for range n`, `min`, `reflect.TypeFor`), matching how
`gnovm/pkg/parser` vendors Go's own parser. That is not cosmetic here: the repo
gates on `go fix` being clean over every package (`.github/workflows/_ci-go.yml`),
and unlike the `golangci-lint` exclusion in `.github/golangci.yml`, that gate
takes no path exclusions. `-omitzero` stays disabled, as it is repo-wide.

`make import` and `make genpatch` apply the same treatment to both sides of the
diff, so `gno.patch` stays the semantic deviation alone rather than 2,500 lines of
mechanical rewrites. Two wrinkles are worth knowing before touching that flow:

- The fixers are applied **one at a time**. A combined `go fix ./...` over this
  package reports `4 of 416 fixes skipped (e.g. due to conflicts)`, then writes
  nothing at all and exits 1 — the conflict discards the other 412 fixes with it
  (go1.26.1). Applied individually they all land.
- `import` stages `bounds_test.go` out before running them. It is ours, so it
  survives `make clean`, and until `gno.patch` is applied it references
  `maxNestingDepth`/`maxKeyDepth`, which do not exist yet — leaving the test
  package failing to typecheck, whereupon `go fix` silently declines the whole
  package and the tree comes out unmodernized.

The rewrites are mechanical, but this is a parser on a consensus path, so they
were checked rather than assumed: 600,000 generated documents (tables, table
arrays, shared prefixes, quoted and dotted keys, every value shape, nesting)
decode to byte-identical output and byte-identical error text against upstream
`pelletier/go-toml`, and `Marshal` round-trips identically.

## Tests

Upstream's full suite is vendored with the source and passes. The bounds have
their own regression suite in `bounds_test.go`, which is the only test file not
from upstream.

## Choosing the limits

Both are far above any real document and far below what an attack needs:

- The deepest key path in any of the ~440 mod files in this repository is 1; a
  `tm2` `config.toml` reaches 2. The limit is 16.
- Real documents nest arrays 1–3 deep. The limit is 256, which costs ~90µs at
  the cap for a 4KB body.

`maxKeyDepth` is 16 rather than something larger because at 16 a deep key path
stops being the most expensive shape a size-capped document can hold — the worst
4KB body becomes an ordinary flat one, so decode cost is linear in length with no
shape premium. Measured worst-4KB decode: 1.17ms at depth 16, 2.37ms at 32,
8.56ms at 64 (past what a caller charging `PreprocessGasPerByte` over the same
bytes collects).

## Not a general-purpose hardening

These bounds make *decoding* safe. They say nothing about the encoder, and
nothing about how much memory a decoded `*Tree` can occupy — a caller feeding
this untrusted input still wants a size cap of its own (`gnovm/pkg/gnomod` uses
`maxFileSize`).
