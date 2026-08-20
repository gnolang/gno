# Reject compiler and tooling directives in submitted packages

## Context

Gno has no conditional compilation, no inliner to suppress, and nothing for
`go generate` to drive: one target, one file set, and `gnomod.toml` carries the
language version. A directive in a `.gno` file — `//go:build`, `// +build`,
`//line`, `//go:generate`, `//go:noinline`, `//export` — therefore does nothing.
The VM compiles the file the same either way.

They are not inert to every consumer, though.

`go/types` honoured `ast.File.GoVersion`, which `go/parser` fills from a
`//go:build go1.N` line. Because the consensus type-check pins
`Config.GoVersion` to `go1.18`, a submitter could raise their own gate
(`//go:build go1.22` + `for range 10` type-checked under the pin), and a file
tagged above a validator's toolchain was accepted on one build and rejected on
another. That is fixed at the type-check layer by blanking the field
(`adr/pr5978_typecheck_strip_file_goversion.md`), but the fix is one consumer
deep: it makes the tag inert to `go/types`, and leaves it in the stored source.

Two costs remain once the tag is stored:

- **It misleads whoever reads the source.** A file whose first line is
  `//go:build ignore` reads as excluded while running unconditionally. On a
  chain whose premise is that anyone can audit deployed code, source that lies
  about its own inclusion is a hazard no type-check fix addresses.
- **Other consumers still honour them.** `gno tool transpile` emits the original
  constraint alongside its own `//go:build gno` header, and `go build` rejects
  the result with `multiple //go:build comments`. `//go:generate` becomes a
  shell command anyone running `go generate` over transpiled chain source would
  execute. Each future consumer of stored source is another chance to trip over
  the same dead configuration.
- **`//line` forges the position in a failed transaction's error.** It is
  honoured by `go/parser`, so it rewrites what the consensus type-check reports:

  ```
  //line ../../../etc/passwd:9999:1  →  gno.land/etc/passwd:9999:23: undefined: nope
  (without it)                       →  zz.gno:4:23: undefined: nope
  ```

  The verdict is unchanged and deterministic, so this is not a fork; it is
  spoofing, and it misleads any tooling that turns a position into a link. It
  cannot redirect a write: `WriteToMemPackage` derives its target from
  `fset.File(pos).Name()`, which is the registered name, not the remapped one.

## Decision

Reject build constraints in **user** packages, in `ValidateMemPackageAny`
(`gnovm/pkg/gnolang/mempackage.go`), the validation every chain entry point
already runs: `MsgAddPackage`, `MsgRun`, genesis, and `AddMemPackage`.

`FindDirectiveComment` returns the first directive in a file: a build constraint
(`constraint.IsGoBuild`, `constraint.IsPlusBuild`) or a directive comment —
`//line`, `//extern`, `//export`, and the `//tool:name` form that covers
`//go:generate`, `//go:embed`, and the pragmas.

The **whole file** is scanned, and both comment forms are matched. Build
constraints must precede the package clause, but pragmas attach to declarations
anywhere, and Go honours the *block* form of a line directive anywhere too:
go/scanner accepts a comment as a directive when `lit[1] == '*' || offs ==
s.lineOffset` and the text after the opener begins with `"line "`. A rule
matching only `//` comments leaves `/*line forged.gno:999:1*/` live, which was
caught in review.

The check also runs **before the file is parsed**. A line directive rewrites the
positions `go/parser` reports, so parsing first lets a file that is about to be
rejected choose the filename and line printed in its own error. The directive
name is quoted in that error for the same reason: it is submitted text and may
carry terminal control bytes, and the error reaches both a terminal (`gno lint`)
and a transaction result.

`//go:generate` is matched on **raw lines** as well, because `go generate` is
the one consumer that never parses: it scans lines for the prefix at column 1
followed by a space or tab (`cmd/go` `isGoGenerate`). A command hidden from the
token scan inside a raw string or a block comment therefore survives
transpilation and still runs — checked end to end before this was closed: the
package validated, `gno tool transpile` kept the line at column 1 of the
generated `.go`, and `go generate -tags gno -n` printed its command. The column
and separator rules mirror `cmd/go` so this does not reject text `go generate`
would ignore. No file in `examples/` contains such a line.

Everything else scans **tokens**, via `go/scanner`, not raw lines. That is the mechanism
`go/parser` itself uses, so the two agree on inputs where a line scan does not:
a leading BOM, and a `package` line sitting inside a block comment ahead of the
real clause — `go/parser` honours a constraint in both (verified against
go1.25.9). Token scanning also gives the converse for free: a string literal
that merely spells a directive is not a comment token, so a linter or formatter
written in Gno stays valid.

The directive rule mirrors the unexported `go/ast.isDirective`, **copied rather
than called**: this decides whether a package is accepted on chain, so it must
not shift when the toolchain's own copy evolves. A test pins the copy against
Go's own behaviour (observable through `CommentGroup.Text()`, which drops
directives), so a future toolchain change surfaces as a failure to be decided on
rather than as silent drift. Requiring no space after `//`
is what keeps an ordinary `// see: below` comment from counting. The predicate
is a pure function of the file bytes, as a consensus check must be.

`gno lint` reports the same files through the same predicate, so the rule
surfaces at lint time instead of at a failed transaction. It scans the files from
disk *before* `ReadMemPackage` and skips a tagged package, mirroring the
validator's ordering: reading first would let a line directive rewrite the
positions in the package's own parse errors.

Scope is deliberate:

- **User packages only** (`mptype.IsUserlib()`). Those are the submitted,
  attacker-controlled ones. Stdlibs ship with the node binary and are reviewed
  as part of it.
- **Stdlibs and filetests are untouched.** The VM suite deliberately pins that
  constraints are inert (`gnovm/tests/files/build0.gno`, `extern/ct`); those
  fixtures test real VM semantics and keep working unchanged.
- **The whitelist holds exactly one entry: `//nolint`.** Everything else is
  refused. Scanning all 1479 `.gno` files in `examples/` yields zero directives
  of any kind; the only hits anywhere are 6 stdlib files carrying Go's own
  inherited ones (`//go:generate`, `//go:noinline`), which the user-package gate
  excludes. The entry is therefore for code not yet written — see Consequences.

  Entries live in the `allowedDirectives` list, deliberately *above*
  `isDirectiveText`:
  whether something is a directive is Go's question and is answered by the
  faithful copy, whether Gno accepts it is ours and is answered by policy. A
  test pins the two apart, so relaxing policy cannot quietly corrupt the mirror.

## Alternatives considered

- **Strip the directive at storage time.** Rejected: rewriting source a user
  signed is worse than refusing it. The stored bytes must be the submitted
  bytes.
- **Reject only build constraints, and handle the rest later.** The rationale is
  identical for every directive, and each round costs reviewers the same
  consensus-input argument. Closing the class once is cheaper than a blacklist
  that grows an entry per discovery.
- **Leave it at the type-check blanking (status quo after #5978).** Closes the
  version-override vector but keeps the misleading source and the transpiler
  residue, and relies on every future consumer remembering that constraints are
  meaningless here.
- **Allow a whitelist from the start.** No entry has a use case, and an unused
  extension point invites one to be added without the argument being made.
  Loosening later is backward-compatible; see below.
- **Reject for stdlibs too.** Would require editing VM fixtures that exist
  precisely to pin constraint inertness, for no security gain — stdlibs are not
  submitted.
- **Validate a line directive's payload before rejecting it.** Go only *remaps*
  positions when the text after `line ` parses as `filename:line[:col]`, so
  `/*line items are processed below*/` is inert and this rule rejects it anyway.
  Declined on three grounds. It matches Go's own classification: `ast.isDirective`
  is prefix-only and strips `//line items are processed below` from doc text
  without validating a payload either, and the differential test holds the copy
  to exactly that. Validating only the block form would then split the two
  spellings — `//line prose` rejected, `/*line prose*/` allowed — which is less
  coherent than treating both alike. And payload validation moves the rule toward
  accepting *more*, which is the direction where a mistake becomes a consensus
  break rather than a relaxable annoyance. No `.gno` file in the tree contains
  `/*line `.

### Cost of the scan

Validation now reads every `.gno` file end to end — `FindDirectiveComment` must,
since a pragma can sit on any declaration — where the checks before it stopped at
the package clause. Measured on a 1 MB source: **9.8 ms** for the scan against
**0.09 ms** for `PackageNameFromFileBody`. This rule, not the older checks, sets
the cost of validating a package.

Those bytes are already priced. The ante handler charges
`TxSizeCostPerByte` (10 gas/byte) on every transaction, in `CheckTx` as well as
`DeliverTx`, and `PreprocessGasPerByte`'s calibration note records that 1 gas is
1 ns on the reference machine. The existing charge therefore budgets ~10 ns/byte
and the scan spends ~9.4 ns/byte, so it fits inside a charge that already
applies — it is not unmetered work.

Moving the preprocess charge (1250 gas/byte) ahead of validation was considered
and rejected. It would price the scan 125× over the ante rate and, because it
lands before the reason for rejection is known, would turn a clean
`invalid package path` on a large package into an out-of-gas panic for any
submitter whose budget assumed early rejection.

What remains is headroom, not exposure: the scan consumes most of a per-byte
budget that validation previously barely touched, and the 9.8 ms figure comes
from an Apple-silicon machine, so a slower validator has less margin. If that
margin ever needs buying back, the cheap move is a prefilter — a directive
requires `//` or `/*` followed immediately by `[a-z0-9]`, and ordinary comments
start `// ` with a space, so honest files would skip tokenizing entirely.

## Consequences

- A package carrying a directive is now rejected at `AddPackage`/`Run` with
  `invalid file %q: directives are not supported: %s`, naming the directive, and
  flagged by `gno lint` as `gnoDirectiveError`.
- The verdict and its error text are functions of the submitted bytes alone. The
  predicate consults no version information, so unlike the bug in #5978 — whose
  message embedded the builder's Go release — it cannot differ between two
  validators built with different toolchains. Rejection also happens before the
  type-check, so a bad package is refused at the cheap end of AddPackage.
- **This narrows accepted consensus input.** A package already stored on a live
  network with a constraint line would fail to replay under a node running this
  change, so it needs the usual coordinated upgrade. Doing it now is cheap;
  after mainnet freezes the accepted-input set it is not. The direction matters:
  loosening later (adding a whitelist entry) is backward-compatible because it
  only accepts more, while tightening later is a break — so rejecting now and
  whitelisting on demonstrated need is the only ordering that never forces a
  second fork.
- No corpus change: `examples/` passes unmodified, and the VM filetests carrying
  constraint lines are out of scope by type.
- **`//nolint` is allowed**, and it is the only exception. Surveying every `.go`
  and `.gno` file in the repo (6453 files, 143 carrying a directive) puts the
  `//nolint:` family at ~65 occurrences across its variants — second only to
  `//go:build` at 55, and more than `//go:generate` and `//go:embed` combined.
  No `.gno` file uses it today, so nothing changes hands either way; the point is
  that Gno authors come from Go carrying the habit.

  It is **not** inert downstream, and an earlier draft of this ADR wrongly said
  so. `gno tool transpile` preserves `//nolint:gosec` verbatim, and where the
  transpiled package happens to be valid Go — true for a pure `/p/` helper,
  false for anything importing a gno stdlib, since `std` is not an importable
  Go package — golangci-lint honours it and the suppression is real. So the
  exception rests on cost, not on inertness: `//nolint` is the second most
  common directive in Go code, Gno authors arrive with the habit, and the
  workflow it could mislead (transpiling on-chain source and linting it) is not
  one anyone runs today, whereas the deploy failure is immediate and confusing.

  That is a weaker footing than the rest of this rule, and it points the wrong
  way on the ordering argument above: dropping the entry now would be
  backward-compatible to restore, while removing it later is a consensus break.
  It is kept deliberately rather than by oversight.
- That survey also bounds the false-positive risk: across those 6453
  human-written files, every hit was a real directive. No ordinary `//word:word`
  prose comment was caught, because a directive requires no space after `//`.
