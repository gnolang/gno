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
accumulator. Granularity is the statement, the switch or select clause,
the composite-literal element and the file declaration; what cannot be
typed keeps an invalid type instead of taking its dependents down:

- `Transcribe` gains an optional `PanicHandler`; the body, declaration,
  clause and element loops call it through `guarded`. In accumulate mode
  a `*PreprocessError` raised anywhere inside such a node is recorded, the
  node is dropped in place (an element becomes an `invalidType`
  placeholder, a clause is skipped by its switch), the block stack is
  rewound to the enclosing block, and the walk continues with the next.
- `predefineFileSet` guards each file-level declaration the same way; a
  dropped declaration is marked predefined and preprocessed so the walk
  skips it. The deploy path's package-level duplicate check runs first,
  reporting each redeclaration at the declaration that loses.
- `invalidType` (go/types' `Typ[Invalid]`) is the type of what was
  reported. A name declared nowhere becomes a placeholder at each use in
  a declaration (`undefinedName`), so a signature naming it still defines
  its function, body and callers checked. A dropped declaration or
  statement gives its names that type (`dropDecl`, `dropStmt`; a declared
  type keeps its identity with an invalid base). At `TRANS_LEAVE` an
  expression with an invalid operand is invalid itself and its rule does
  not run; assignability, conversion and the bool check pass it, and a
  rule that still mentions it is suppressed by message. A name with no
  slot to type is poisoned instead: later "not declared" errors about it
  are dropped. The same message reached through two paths is kept once.
- The dependency path (`predefineRecursively2`) carries the accumulator
  into its nested walk, so a package-level literal is checked element by
  element and a function literal statement by statement as well.
- Go runtime errors are never swallowed: they are VM bugs.

`Preprocess` passes a nil handler, so the on-chain path is unchanged:
first error, same gas, same text. `invalidType` never exists there; the
checks for it are type assertions on paths it cannot reach. The
accumulated result must not be executed.

## Consequences

- One pass gives every error with exact positions; the harness's
  multi-pass neutraliser can go, and `gno lint` can report all errors.
- Finer recovery (per expression inside a statement) can be added at the
  same seam: record, substitute an `invalidType` placeholder, go on.
- Suppression stays heuristic in two places: by name for what cannot be
  typed, by message for a rule that prints the invalid type.

## Limits

- A failing statement is still dropped whole: `x := undefined + 1`
  reports the name, not what else the statement does wrong.
- A type switch clause whose case is not a type drops with its body;
  go/types types the clause variable invalid and walks the body silently,
  which reports the same.
- An import with the default name and an unknown path fails in
  `initStaticBlocks`, before any guard: the run panics as on chain.
- A same-file redeclaration is reported by the parser (`ParseFileCollect`,
  `DeclarationErrors`) and again, in the deploy path's wording, by the
  package-level check.
- `PreprocessCollect` on a package node without a `FileSet` cannot index
  declarations: a declaration that fails there drops whole.

## Parser errors

`ParseFileCollect` is `ParseFile` for this mode: it parses with
`AllErrors`, returns every parser error located, and converts the partial
tree anyway; `Go2Gno` maps `ast.BadExpr` / `BadStmt` / `BadDecl` to a name
no program can declare (`!syntax`), poisoned up front, so the statement
holding it drops behind the parser's error for that line like any failed
statement. gc and go/types type-check after syntax errors; the old
`ParseFile` threw the tree away on the first one. On chain `ParseFile` is
unchanged.
