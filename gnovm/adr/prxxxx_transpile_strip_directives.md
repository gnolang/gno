# Strip inherited directives from transpiled output

## Context

`gno tool transpile` turns a `.gno` file into a `.go` file. It parses with
`ParseComments` and reprints with `format.Node`, so every comment in the source
is reproduced verbatim in the output.

A directive means nothing to the GnoVM — one target, no conditional
compilation. The transpiled file, however, is Go, and the Go toolchain acts on
what it finds there:

- **`//go:build` collides with the header.** The transpiler writes its own
  `//go:build gno`, so a source that carries any `//go:build` line produces a
  file `go build` rejects outright: `multiple //go:build comments`.
- **`//go:generate` is a command.** `go generate` does not parse; it scans lines
  for the prefix and runs what follows. A contract's source line becomes a
  command in whoever's shell runs it over transpiled output.
- **`//nolint` suppresses findings.** Where the transpiled package is valid Go —
  a pure `/p/` helper, though not anything importing a gno stdlib, since `std`
  is not an importable Go package — golangci-lint honours it, and a contract
  gets to hide diagnostics from whoever lints it.
- **`//line` forges positions.** The transpiler writes a `//line` header for
  exactly this purpose; an inherited one competes with it.

## Decision

Drop directive comments from the AST before printing, in
`TranspileWithResolver`. The header the transpiler writes is emitted as text
around the printed AST, so it is unaffected.

`IsDirectiveComment` in `gnovm/pkg/gnolang` decides what counts: `//line`,
`//extern`, `//export`, the `//tool:name` form, and the block `/*line ...*/`
form Go accepts anywhere. It mirrors the unexported `go/ast.isDirective`, copied
rather than called because go/ast does not export it; a test holds the copy to
Go's own behaviour through `CommentGroup.Text()`, so a toolchain change surfaces
as a failure rather than as drift.

`//nolint` is stripped as well, though Go's rule does not count it as a
directive (bare `//nolint` carries no colon). It steers a Go tool reading the
generated file, which is the reason the rest are stripped.

## Alternatives considered

- **Reject these at the chain instead.** Necessary but not sufficient, and a
  different layer: it cannot help a package already stored, a package that never
  goes on chain, or a stdlib. #6078 does that for submitted packages; this makes
  the generated file safe regardless of where its source came from.
- **Strip only `//go:build`.** Fixes the build breakage and leaves the command
  execution and the suppression, which are the parts that act on someone else's
  machine.
- **Neutralize rather than remove** (e.g. rewrite as `// go:build`). Keeps the
  text visible but makes the output diverge from the source in a way a reader
  must decode; removal is simpler and the source remains the record.

## Consequences

- Generated files carry only the directives the transpiler writes, so
  `go build`, `go generate` and golangci-lint see what it intends and nothing
  inherited.
- A directive that is not a comment is out of reach: `//go:generate` at column 1
  inside a raw string is string data to the parser, and `go generate` — which
  scans lines rather than parsing — would still act on it. Rewriting string
  literals is not something a transpiler may do, so that case belongs to
  whoever validates the source (#6078 rejects it).
- Round-tripping is not affected: `.gen.go` files are generated artifacts, and
  the `.gno` source keeps its comments.
