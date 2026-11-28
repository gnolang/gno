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

Three costs remain once the tag is stored:

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

Reject build constraints, and every other compiler or tooling directive, in
**user** packages, in `ValidateMemPackageAny`
(`gnovm/pkg/gnolang/mempackage.go`), the validation every chain entry point
already runs: `MsgAddPackage`, `MsgRun`, `MsgEnablePackage` (re-validating the
parked blob) and genesis-file loading.

It is an **admission** rule, so it is waived where admission is not being
decided. `ValidateOptions.AlreadyOnChain` skips it, mirroring
`TypeCheckOptions.AlreadyOnChain`: the keeper sets it from
`auth.IsGenesisReplay` on all three messages, because a replay re-delivers
deploys the source chain already accepted (gnoland-1's history included), and
refusing one would stop `InitChain` or, under `-skip-failing-genesis-txs`, bring
the fork up diverged. `IsGenesisReplay` is true for *every* `InitChain` tx, so
this also waives the rule for a tx an operator writes into a genesis file by
hand. That matches the type-check and code-policy waivers, and genesis content
is the chain's own; `LoadPackage`, which builds genesis txs from a directory,
still applies the rule. `AddMemPackage` always sets it: the store is not an
admission gate, it has no context to tell a replay from a live deploy, and it
panics on failure. A live `MsgEnablePackage` still applies the rule to a blob
parked by an older binary, as `validateParkedBlob` re-validates under the
current rules by design. Tools that load packages through the store without
the keeper (`gno test`, `contribs/gpao`) therefore do not apply the rule;
`gno lint` does.

`FindDirectiveComment` returns the first directive in a file: a build constraint
(`//go:build`, caught by the `//tool:name` form below, and legacy `// +build`,
caught by `constraint.IsPlusBuild` since it is not directive-shaped) or a
directive comment — `//line`, `//extern`, `//export`, and the `//tool:name`
form that covers `//go:generate`, `//go:embed`, and the pragmas.

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

`cmd/go` reads the **transpiled** `.go`, though, and the transpiler formats with
`go/printer`, which strips the common indentation of a multi-line block comment.
`/*\n\t//go:generate x\n\t*/` is therefore not at column 1 in the `.gno` but is
in the `.go` — found in review, where it validated and `go generate` ran it. Each
line of a block comment is therefore also checked after dropping the bytes
`go/printer`'s `commonPrefix` strips: every byte `<= ' '` (control bytes
included, which a second review round found) and `*`. Unicode spaces are left,
as `go/printer` leaves them. Trimming each line fully over-approximates the
common prefix, which only ever refuses more, inside a comment. Raw strings are
not re-indented, so an indented line inside one stays accepted; but go/scanner
strips CR from a raw string and go/printer prints the result, so
`` `\n\r//go:generate x\n` `` also reaches column 1, and the raw-line check runs
with every CR removed.

The first line of a block comment is checked too, after its `/*` (a third
review round found it). In doc position, above a declaration or the package
clause, `go/printer`'s `formatDocComment` re-emits a multi-line block comment
that is not all stars as `"/*\n" + text + "*/"` after `go/doc/comment`
unindents it, so `/* //go:generate x\n\n*/` comes out with the directive on a
column-1 line of its own.

Two tests pin this. A mirror test builds every byte `<= ' '` (alone, paired,
and beside `*`) and Unicode spaces into each comment and raw-string shape, in
declaration, package-doc and function-body position, formats the file with
`go/format`, and requires the scan to refuse every file whose output has a
column-1 `//go:generate`. And because each hand-written mirror so far missed a
shape (indentation, control bytes, the first line), a seeded random
differential test in `gnovm/pkg/transpiler` builds files from those pieces in
six positions and runs every file the rule accepts through the real
`Transpile`. It hits the first-line case within its 20,000 files when the fix is
reverted; a one-off run of 2,000,000 files (1.66M accepted and transpiled) over
nine positions found nothing. A mirror test must enumerate the consumer's input
space, not the implementation's.

Formatting the file and scanning the output was considered and declined: the
verdict would then follow `go/printer`'s behaviour in whichever toolchain built
the validator, which is a consensus input. The byte sets above are pinned in
this repository and checked against the toolchain by the test instead.

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
rather than as silent drift. The copy pins the *classification*; tokenizing still
uses the toolchain's `go/scanner`, and legacy constraints its
`go/build/constraint.IsPlusBuild`. Consensus already depends on `go/scanner`
through `go/parser`, so this adds no new toolchain input, but "copied rather
than called" covers the rule, not every byte of the path. Requiring no space after `//`
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
- **Stdlibs and the VM's own fixtures are untouched** (the `MPFiletests` type;
  a `_filetest.gno` inside a user package is scanned like any other file). The VM suite deliberately pins that
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
- **A wider whitelist from the start** (staticcheck's `//lint:ignore`, cgo, or
  any `//tool:` form). Only `//nolint` has a demonstrated use (see
  Consequences); an entry without one invites the next to be added without the
  argument being made. `//lint:ignore` and `//lint:file-ignore` are therefore
  refused today, like any lowercase `//word:word` without a space, which is
  `go/ast`'s own classification. Loosening later is backward-compatible; see
  below.
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

That figure is for ordinary source. Review measured adversarial shapes at about
1 MB: ~62 ms for a file of newlines (of which ~30 ms is `PackageNameFromFileBody`
on master already), ~44 ms for 340k `//` lines, ~15 ms for 262k `/**/`. The
worst adds ~30 ns/byte over the ante rate. It stays bounded by
`MaxBlockTxBytes` and runs once per transaction, so it is not a DoS, but the
prefilter below is the lever if it ever matters.

Moving the preprocess charge (1250 gas/byte) ahead of validation was considered
and rejected. It would price the scan 125× over the ante rate and, because it
lands before the reason for rejection is known, would turn a clean
`invalid package path` on a large package into an out-of-gas panic for any
submitter whose budget assumed early rejection.

One input needed handling: bytes `go/scanner` cannot lex. It formats a message
per bad byte — `scanner.errorf` calls `fmt.Sprintf` before consulting the error
handler, so a nil handler does not avoid the cost — which turned a megabyte of
them into ~330 ms and three million allocations, against ~13 ms for ordinary
source of the same size. Since this scan runs *before* `chargePreprocessGas`,
that time was unmetered: measured end to end, a rejected 1 MB package of illegal
bytes bought roughly 240× more validator CPU per gas than a valid one. The scan
now stops at the first scanner error and reports it, which costs ~0.3 ms
instead.

Reporting it, rather than treating it as "no directive found", keeps the rule
self-contained. A file that cannot be lexed cannot compile, so refusing it here
costs nothing real — and the alternative leans on a later gate: `go/types` and
the VM's own parser both reject such a file, but `ValidateMemPackageAny` alone
would not, because `PackageNameFromFileBody` parses with `PackageClauseOnly` and
never reaches bytes deep in the file. The error carries no position, since the
file may hold a line directive ahead of the bad bytes.

What remains is headroom, not exposure: the scan consumes most of a per-byte
budget that validation previously barely touched, and the 9.8 ms figure comes
from an Apple-silicon machine, so a slower validator has less margin. If that
margin ever needs buying back, the cheap move is a prefilter — a directive
requires `//` or `/*` followed immediately by `[a-z0-9]`, and ordinary comments
start `// ` with a space, so honest files would skip tokenizing entirely.

## Consequences

- A package carrying a directive is now rejected at `AddPackage`/`Run` with
  `invalid file %q: directives are not supported: %+q`, naming the directive, and
  flagged by `gno lint` as `gnoDirectiveError`. The ABCI error type is
  `ErrInvalidPkgPath`, as for every other file-level validation failure
  (`errInvalidMemPackage`); that type is hashed into `ABCIResult.Error`, so
  giving file errors their own type is a separate, consensus-visible change.
- The verdict and its error text are functions of the submitted bytes alone. The
  predicate consults no version information, so unlike the bug in #5978 — whose
  message embedded the builder's Go release — it cannot differ between two
  validators built with different toolchains. Rejection also happens before the
  type-check, so a bad package is refused at the cheap end of AddPackage.
- **This narrows accepted consensus input** for new submissions only. History
  is unaffected: a genesis replay waives the rule (see Decision), so a chain
  whose past deploys carry a directive still replays. A package parked under the
  inert policy before the upgrade can no longer be enabled live if it carries
  one. The live change needs the usual coordinated upgrade. Doing it now is cheap;
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
