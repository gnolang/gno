# Bound Go2Gno const-inheritance materialisation

## Context

Go permits a const spec with no expression list to inherit the preceding
non-empty list. `toDecls` preserves the required per-spec AST ownership by
deep-copying that list. An expression containing M nodes followed by N empty
specs therefore made Go2Gno allocate O(N*M) nodes from O(M+N) source bytes.

These are ordinary Go allocations made while translating the parser AST. They
do not pass through `Machine.Alloc`, the transaction allocation limit, the gas
meter, or `TypeCheckMemPackage`, so those limits cannot bound this work.

The copies cannot share nodes. Preprocessing mutates each declaration's value
tree, including rewriting `iota` using that declaration's `ATTR_IOTA`.
`ATTR_PREPROCESSED` would then make later declarations skip a shared subtree
and retain the first declaration's value.

## Decision

Each const `GenDecl` receives an inherited-node budget equal to twice its own
source span in bytes. The span is `gd.End()-gd.Pos()`: both positions are
offsets from the same file base, so the difference is the declaration's byte
length whether or not the caller kept the parser's `token.FileSet`
(`ParseExpr` does not). Every explicit value list is walked once with an identity
`Transcribe` to record its node cost. Each empty spec adds the current list's
cost before `copyExprs`; exceeding the cumulative budget rejects the file with
an actionable parse error.

Factor two leaves ample room for ordinary inherited expressions: every
additional spec contributes its identifier and formatting to the source span,
while common expressions such as literals, `iota` enums, and unit multipliers
contain only a handful of nodes. Across the 4,376 `.gno` files in `examples/`,
`stdlibs/`, `tests/stdlibs/` and `tests/files/`, the highest ratio of inherited
nodes to declaration bytes is 0.266 (`stdlibs/time/time.gno`).

A node budget only bounds memory if nodes have a bounded size. Expression nodes
are at most about 200 bytes, but block nodes embed a `StaticBlock` and are
several times larger (`BlockStmt` is 896 bytes, `IfStmt` 2,696). The only way
to put block nodes into a constant expression is a function literal, for
example inside `len([1]func(){...})`. Inheriting such an expression is
therefore rejected outright, with a message asking to repeat it explicitly; no
file in the corpus above does this. Without the rule, a 600 KB MsgRun of such
declarations peaked at 5.9 GB of heap while type-checking and running within
1.26B gas; with it, the same payload is rejected at under 400 MB.

## Alternatives considered

- Share the preceding nodes or use copy-on-write. Sharing is incorrect because
  preprocessing mutates and marks subtrees per declaration. Copy-on-write
  would require invasive mutation tracking across the AST and preprocessor.
- Route copies through `Machine.Alloc` or charge gas. Go2Gno nodes are plain Go
  objects, and the translation helpers do not carry a machine allocator.
  Threading one through would be substantially broader and would still need a
  deterministic work model for offline parser callers.
- Use a fixed per-file or per-block node ceiling. A source-relative per-decl
  budget composes linearly across multiple declarations and does not penalize a
  legitimately large declaration merely for being large.

## Consequences

Inherited const expressions remain independent, so preprocessing and `iota`
semantics are unchanged. Adversarial const blocks now fail during Go2Gno before
the next over-budget deep copy. Counting adds one linear walk per explicit
value list.

The bound still permits up to two inherited expression nodes per source byte
for each const declaration. This is intentional linear amplification; the
quadratic amplification is removed.
