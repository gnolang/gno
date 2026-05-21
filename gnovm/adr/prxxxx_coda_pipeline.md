# ADR: Coda Middleware Pipeline for Post-Preprocess Passes

## Status
Accepted

## Context

`Preprocess()` in `gnovm/pkg/gnolang/preprocess.go` runs several passes that
execute *after* the main `preprocess1` body. Collectively these are called the
"coda" passes (from the musical term for "conclusion"). Prior to this change
there were four:

1. `codaInitOrderDeps` — per-Decl walk that records package-level
   initialization dependencies (Go spec ordering).
2. `codaHeapDefinesByUse` — writes `ATTR_HEAP_USES` and `ATTR_PACKAGE_DECL`
   based on closure captures and ref expressions; mutates NameExpr paths for
   heap captures.
3. `codaHeapUsesDemoteDefines` — reads `ATTR_HEAP_USES` to promote define
   NameExprs (and FuncDecl/FuncLit parameters) to heap-define, or demote back
   to regular define when a name is not actually escape-captured.
4. `codaPackageSelectors` — rewrites bare package-level `NameExpr`s into
   `SelectorExpr{X: RefValue, Sel: name}` forms.

Passes 2–4 were dispatched from a single outer `Transcribe` loop that, for
each root `BlockNode`, ran all three as *separate* full-tree walks back-to-
back. This meant the same subtree was transcribed three times per root
BlockNode. The FuncLit-skip guard (`ATTR_PREPROCESS_SKIPPED ==
AttrPreprocessFuncLitExpr`) was duplicated inline at multiple sites. There
was also a long-standing `XXX` note in `transcribe.go` suggesting the passes
should migrate from `Transcribe` to `TranscribeB` for uniform block-stack
management.

Motivation:
- Adding a new coda pass today means another full-tree walk. The
  commented-out `codaGotoLoopDefines` hints at future ones.
- Inconsistent transcribe helpers (`Transcribe` + manual block-stack vs
  `TranscribeB`) make the passes harder to reason about.
- Three redundant transcribe walks per root BlockNode.

## Decision

Introduce a composable middleware pipeline over `TransformB`, and rewrite
passes 2–4 as middleware. The pipeline runs the middleware in order for each
`(n, stage)` pair; control values (`TRANS_CONTINUE / TRANS_SKIP / TRANS_EXIT`)
short-circuit subsequent middleware uniformly.

The post-preprocess dispatch now runs **two** TranscribeB walks per root
BlockNode (down from three), gated by an outer `Transcribe` that preserves
the original "only at the first BlockNode found" semantics:

```go
pass1 := codaPipeline(
    skipUnprocessedFuncLitMW,
    codaHeapDefinesByUseMW,
)
pass23 := codaPipeline(
    skipUnprocessedFuncLitMW,
    codaHeapUsesDemoteDefinesMW,
    codaPackageSelectorsMW,
)
Transcribe(n, func(..., n Node, stage TransStage) (Node, TransCtrl) {
    if stage != TRANS_ENTER { return n, TRANS_CONTINUE }
    if isSkippedFuncLit(n) { return n, TRANS_SKIP }
    if bn, ok := n.(BlockNode); ok {
        _ = TranscribeB(ctx, bn, pass1)
        _ = TranscribeB(ctx, bn, pass23)
        return n, TRANS_SKIP
    }
    return n, TRANS_CONTINUE
})
```

`codaInitOrderDeps` (pass 1 in the original ordering) is deliberately kept as
a separate per-Decl walk — see **Alternatives** below.

### Correctness fixes folded into the merge

Investigation surfaced three ENTER/LEAVE hazards that the prior three-walk
ordering masked and that the merge needed to fix:

**Blocker A — `*ValueDecl` stage mismatch.** `codaHeapDefinesByUse` writes
`ATTR_HEAP_USES` on `TRANS_LEAVE` of a top-level `*ValueDecl`.
`codaHeapUsesDemoteDefines` used to read it on `TRANS_ENTER` of the same
node and *panic* on miss (`"expected heap use for top level value decl"`).
This was safe only because the two passes ran as fully sequential tree
walks. Fix: the read moved to `TRANS_LEAVE`, making it symmetric with the
write.

**Blocker B — `*NameExpr` stage mismatch.** `codaHeapDefinesByUse` sets
`ATTR_PACKAGE_DECL` on `TRANS_LEAVE` of a `*NameExpr`;
`codaPackageSelectors` used to read it on `TRANS_ENTER` of the same node.
Fix: the read (and the `NameExpr → SelectorExpr` replacement) moved to
`TRANS_LEAVE`. `NameExpr` has no children the replacement cares about, so
LEAVE is semantically equivalent.

**Blocker C — `codaInitOrderDeps` must see `*NameExpr` before
`codaPackageSelectors` replaces them.** This is already handled by the
*existing* ordering: `codaInitOrderDeps` runs on the full file before the
pipeline walks. This blocker is the main reason `codaInitOrderDeps` was left
out of the merge.

### Why two walks, not one

`codaHeapDefinesByUseMW` writes `ATTR_HEAP_USES` on an *ancestor* BlockNode
(the definition scope) when it encounters a closure capture — i.e. the
attribute is set on a node the DFS already ENTER'd earlier.
`codaHeapUsesDemoteDefinesMW` reads that attribute on `TRANS_ENTER` of
Define/HeapDefine `*NameExpr`s that can appear before the capturing closure
in DFS order. Consequently pass 2 can only run correctly after pass 1 has
finished the whole subtree — merging them into one walk produces stale
reads.

Pass 3 (`codaPackageSelectorsMW`) *can* be merged with pass 2 because both
read per-node attributes written on the same node's `TRANS_LEAVE`, and the
pipeline's fixed middleware order guarantees pass 2 mutates the NameExpr
type before pass 3 (in the later LEAVE fire) replaces the NameExpr outright.

### Why preserve the outer `Transcribe` dispatcher

`Preprocess` is called recursively on sub-expressions that are not
`BlockNode`s (e.g. an `Expr` argument, a `*CallExpr`). The original code
used an outer `Transcribe` that only triggered the coda walks when it
encountered a `BlockNode`; for recursive calls on expressions with no nested
BlockNode, the coda walks did not run. Replacing the dispatcher with a
direct `TranscribeB(ctx, n, pipeline)` would fire the middleware on every
node of every recursive Preprocess call, corrupting attributes set by the
outer Preprocess call (e.g. re-deleting `ATTR_HEAP_USES` that were still
needed). This was caught by test failures in `heap_alloc_forloop*.gno`,
`var_initorder19.gno`, and others; the outer dispatcher was restored.

## Consequences

### Correctness
- The ENTER/LEAVE stage fixes (blockers A & B) remove two latent ordering
  hazards that existed in the previous code. They were masked by the
  three-separate-walks structure but would break any attempt to merge.
- Full test suite (`go test ./pkg/gnolang/...` — unit + `TestFiles` +
  `TestStdlibs` filetests covering the heap/initorder invariants) passes.

### Performance
- **Walk count**: 3 → 2 coda walks per root BlockNode.
- **End-to-end Preprocess benchmark** (`BenchmarkPreprocessCoda`): within
  noise of master on small/medium/large synthetic sources (~+1% geomean on
  `ns/op`, p-values not significant). `preprocess1` dominates total cost and
  one fewer coda walk does not move the needle at this level. Allocations
  are essentially unchanged (-0.03% geomean).

Realistically the measured end-to-end speedup is ~zero. The refactor is
justified primarily by the clarity/extensibility value below; anyone
reading the PR should not expect a meaningful perf win.

### Clarity / extensibility
- Adding a new coda pass now means writing a `codaMiddleware` function and
  inserting it into the pipeline — not adding another top-level
  `TranscribeB` call.
- The FuncLit skip guard is extracted to `isSkippedFuncLit` (shared between
  `codaInitOrderDeps` and the pipeline via `skipUnprocessedFuncLitMW`).
- `codaHeapUsesDemoteDefines` and `codaPackageSelectors` now use
  `TranscribeB` consistently (removing manual `pushInitBlock`/pop
  bookkeeping from the former and the `Transcribe` signature from the
  latter). Resolves the `XXX Replace usage of Transcribe() with
  TranscribeB()` note at `transcribe.go:121` for these passes.

## Alternatives considered

1. **Full three-into-one merge.** Rejected: pass 2 reads attributes that pass 1
   writes on *ancestor* BlockNodes during subtree descent; interleaving them
   produces stale reads (broken tests confirm this).

2. **Drop the outer `Transcribe` and make the pipeline the single
   entry-point.** Rejected: recursive `Preprocess()` calls on non-BlockNode
   expressions would trigger the pipeline on every node, overwriting
   attributes set by the parent call. Caught during verification.

3. **Merge `codaInitOrderDeps` (pass 1 of 4) into the pipeline.** Not done
   here. It maintains a per-Decl deps map and must run on original
   `*NameExpr`s *before* `codaPackageSelectorsMW` replaces them (Blocker C).
   A stateful middleware could encode this, but the invariants are subtle
   and the value is marginal — pass 1 already uses `TranscribeB` and the
   current separation keeps the pipeline's state-free property. Deferred.

4. **Keep three separate walks but factor middleware for clarity only.**
   Equivalent in structure to the final design but with one extra walk.
   Rejected because the two-walk split is free (pass 3 merges naturally with
   pass 2 at no correctness cost) and worth the small code saving.

## Files changed

- `gnovm/pkg/gnolang/preprocess.go` — coda function bodies rewritten as
  middleware (`*MW` suffix); dispatch block updated; `isSkippedFuncLit`
  helper added; Blocker A & B stage relocations.
- `gnovm/pkg/gnolang/coda_pipeline.go` — *(new)* `codaMiddleware` type,
  `codaPipeline` composer, `skipUnprocessedFuncLitMW`.
- `gnovm/pkg/gnolang/preprocess_bench_test.go` — *(new)* benchmark to
  measure Preprocess + coda cost on synthetic sources.
