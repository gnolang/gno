# Reject build constraints in submitted packages

## Context

Gno has no conditional compilation: one target, one file set, and `gnomod.toml`
carries the language version. A `//go:build` or `// +build` line in a `.gno`
file therefore selects nothing — the VM compiles the file either way.

The lines are not inert to every consumer, though.

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
- **Other consumers still honour it.** `gno tool transpile` emits the original
  constraint alongside its own `//go:build gno` header, and `go build` rejects
  the result with `multiple //go:build comments`. Each future consumer of
  stored source is another chance to trip over the same dead configuration.

## Decision

Reject build constraints in **user** packages, in `ValidateMemPackageAny`
(`gnovm/pkg/gnolang/mempackage.go`), the validation every chain entry point
already runs: `MsgAddPackage`, `MsgRun`, genesis, and `AddMemPackage`.

`HasBuildConstraint` reports `//go:build` (`constraint.IsGoBuild`) or
`// +build` (`constraint.IsPlusBuild`) in a file's header — the comments before
the first non-comment token, the only position where Go gives a constraint
meaning. Restricting it to the header keeps a package that merely *mentions* the
syntax (a formatter or linter written in Gno, a string literal) valid.

It scans **tokens**, via `go/scanner`, not raw lines. That is the mechanism
`go/parser` itself uses to fill `ast.File.GoVersion`, so the two agree on inputs
where a line scan does not: a leading BOM, and a `package` line sitting inside a
block comment ahead of the real clause. Both are honoured by `go/parser`
(verified against go1.25.9) and both slip past a line scan, which would leave a
live tag in a stored file. The predicate is a pure function of the file bytes,
as a consensus check must be.

`gno lint` reports the same files through the same predicate, so the rule
surfaces at lint time instead of at a failed transaction.

Scope is deliberate:

- **User packages only** (`mptype.IsUserlib()`). Those are the submitted,
  attacker-controlled ones. Stdlibs ship with the node binary and are reviewed
  as part of it.
- **Stdlibs and filetests are untouched.** The VM suite deliberately pins that
  constraints are inert (`gnovm/tests/files/build0.gno`, `extern/ct`); those
  fixtures test real VM semantics and keep working unchanged.
- **The whitelist is empty**, not absent. Nothing in the corpus needs an
  exception today (verified: zero constraint lines in `examples/`,
  `gnovm/stdlibs/`, `gno.land/`).

## Alternatives considered

- **Strip the constraint at storage time.** Rejected: rewriting source a user
  signed is worse than refusing it. The stored bytes must be the submitted
  bytes.
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

## Consequences

- A package carrying a constraint line is now rejected at `AddPackage`/`Run`
  with `invalid file %q: build constraints are not supported`, and flagged by
  `gno lint` as `gnoBuildConstraintError`.
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
- No corpus change: `examples/` passes unmodified, and the three VM filetests
  with constraint lines are out of scope by type.
