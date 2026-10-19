# PR6345: Preprocess in accumulate mode, off-chain

## Context

The preprocessor stops at its first error: every checking failure is a
`panic` wrapped by `doRecover` into a `*PreprocessError` with the node's
location. On chain that is right (the deploy is rejected, the gas spent is
minimal, the error text is one message). Off chain it is not: `gno lint`,
`gno test` and the Go-corpus harness need every error of a file, and the
harness had to blank one line per error and re-run, which cannot attribute
errors exactly. The future Gno type checker is measured against Go's
errorcheck corpora by the set of lines it reports, so one run must yield
all of them.

## Decision

`PreprocessCollect` / `PreprocessFilesCollect` run the same walk with an
accumulator. Granularity is the statement and the file declaration:

- `Transcribe` gains an optional `PanicHandler`; the ten body and
  declaration loops call it through `guarded`. In accumulate mode a
  `*PreprocessError` raised anywhere inside a statement is recorded, the
  statement is dropped in place, the block stack is rewound to the
  statement's enclosing block, and the walk continues with the next one.
- `predefineFileSet` guards each file-level declaration the same way; a
  dropped declaration is marked predefined and preprocessed so the walk
  skips it.
- Names declared by a dropped statement are poisoned: later "not declared"
  errors about them are suppressed (go/types' invalid-object rule). The
  same message reached through two paths is recorded once.
- Go runtime errors are never swallowed: they are VM bugs.

`Preprocess` passes a nil handler, so the on-chain path is unchanged:
first error, same gas, same text. The accumulated result must not be
executed.

## Consequences

- One pass gives every error with exact positions; the harness's
  multi-pass neutraliser can go, and `gno lint` can report all errors.
- Per-site "record and return an invalid type" can replace the
  statement-level rule incrementally where finer granularity is wanted;
  the accumulator and the guard are the seam.
- Suppression is heuristic (by name); a rule that reports a dropped
  name under another spelling shows up as noise and gets added to
  `reUndeclaredName`.

## Limits

- Poison is by bare name, package-wide, not by scope: in
  `func a() { x := "s" + 1 }; func b() { println(x) }` the genuine
  `name x not declared` in `b` is suppressed by the drop in `a`.
- A dropped declaration can leave partial state: `type T struct{ X Undef }`
  is dropped but `T` exists as an empty struct, so a method body's `t.X`
  reports `missing field X in main.T`, which no name poison covers.
- Dependency attribution (one error, at the dependency) needs the file-set
  index, i.e. `PreprocessFilesCollect`; `PreprocessCollect` on a lone file
  node may report a dropped dependency at each dependent too.
