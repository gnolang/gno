# SpanFromGo O(N²) on chain-shaped AST nodes

## Status

Implemented in `fix/spanfromgo-on2`. Generalized from a single
`*ast.BinaryExpr` fix to all 11 chain-shapeable AST types after
empirical confirmation that the same O(N²) mechanism applies.

## Context

A deeply nested left-leaning binary expression — for example a
`const x = 1 + 1 + 1 + ... + 1` declaration with N additions — takes
O(N²) wall time to parse, dominated by `go/ast.(*BinaryExpr).Pos()`
called from `gnolang.SpanFromGo`. For N = 54 000 the source is only
~216 KB (well under `MaxTxBytes = 1 MB`) but parsing takes ~13 s of
validator CPU on Apple M1 Pro, all of it **before** any gas meter is
installed by the preprocess pass. Preprocess-time gas accounting
therefore cannot bound this surface; the fix must live in the
parser/translator.

**The same mechanism applies to ten other AST types.** Stdlib
`go/ast/ast.go:495-569` (Go 1.24.1) defines `Pos()` or `End()` to
recurse through a child field for six Pos-recursive types
(`BinaryExpr`, `CallExpr`, `IndexExpr`, `SelectorExpr`, `SliceExpr`,
`TypeAssertExpr`) and four End-recursive types reachable from Go2Gno
(`StarExpr`, `UnaryExpr`, `ArrayType`, `MapType`). `ChanType` is
also stdlib-End-recursive but is rejected at parse time by Go2Gno
(`panicWithPos("channels are not permitted")`) — channels are not
permitted in Gno, so no chain is constructible. All ten reachable
shapes were empirically O(N²) on `main` pre-fix.

## Diagnosis

CPU profile of a single `MsgRun` carrying the 54 000-chain payload on
`upstream/jae/preprocess-dos`, captured with
`go test -cpuprofile`:

```
Duration: 16.94s, Total samples = 16.59s (97.93%)
Showing top 5 nodes
      flat  flat%   sum%        cum   cum%
    10.21s 61.54% 61.54%     10.34s 62.33%  go/ast.(*BinaryExpr).Pos
     1.25s  7.53% 69.08%      1.25s  7.53%  runtime.memclrNoHeapPointers
     1.02s  6.15% 75.23%      1.84s 11.09%  gnolang.Go2Gno
     0.53s  3.19% 78.42%      0.53s  3.19%  runtime.madvise
     0.49s  2.95% 81.37%      0.49s  2.95%  runtime.(*mspan).heapBitsSmallForAddr
```

`pprof -peek 'BinaryExpr\)\.Pos'`:

```
go/ast.(*BinaryExpr).Pos
  caller: gnolang.SpanFromGo  (99% of cumulative time)
```

**Mechanism (Pos-recursive case).** The Go standard library defines
`(*ast.BinaryExpr).Pos() = X.Pos()` — it recurses leftward through
the chain to return the leftmost leaf's position. `SpanFromGo`
(`gnovm/pkg/gnolang/nodes_location.go:107`) calls `gon.Pos()` on
every AST node it visits during `Go2Gno`'s translation walk via the
deferred `setSpan(fs, gon, n)` at `go2gno.go:291-295`. For a
left-leaning BinaryExpr tree of depth N, each `setSpan` call costs
O(N) and N setSpan calls sum to O(N²).

**Same mechanism, mirror image (End-recursive case).** Five other
types define `End() = <field>.End()` — e.g. `StarExpr.End() = X.End()`
recurses through `X` for `***...*int` chains. `SpanFromGo` calls
`gon.End()` per node, so deep chains are O(N²) on the End side.

## Why preprocess-side mitigations don't help

The cost lands in the parse step, which runs in the message handler
**before** the preprocess pass begins. By the time any of the
following would fire, the seconds have already been spent:

- **Preprocess-time gas accounting** (separate feature): the gas
  meter is installed on the context only by the preprocess pass
  itself.
- **Per-tx allocator hard cap** (`maxAllocTx = 500 MB` in
  `fix/jae/preprocess-dos`): this is a memory bound, but the
  adversarial source is CPU-bound. Per the profile, the actual heap
  allocation is dominated by Gno AST node construction
  (~1.5 GB cumulative across all iterations, but per-iteration
  allocation is kilobytes — far below the cap).
- **Composite-type / embed-chain depth caps** (also in
  `fix/jae/preprocess-dos`): target deeply nested type expressions
  (`[][]...[]int`, `type T struct{T1; T2; …; T_{K-1}}`), not
  the chain-recursion shapes here.

**Note on type-check-based bounding.** An earlier draft of this ADR
argued that some shapes (e.g. `SelectorExpr` chains requiring nested
struct embedding to type-check) are bounded "in practice" by later
passes. **That argument is wrong**: the O(N²) cost lands during
`Go2Gno` translation, well before any type-checking runs. A chain of
undefined names is syntactically valid, accepted by the parser, and
burns the full O(N²) before the type checker has a chance to reject
it. The fix must address all chain shapes that the *parser* accepts.

## Decision

Add two helpers alongside `setSpan` and gate each of the 11 chain
cases in `Go2Gno`. The deferred `setSpan` stays as a safety net for
non-chain cases; its `IsZero` short-circuit makes the helpers'
explicit Span setting take precedence with zero cost for the defer.

The helpers:

```go
// setSpanFromLeftChild — for Pos-recursive types. Pos comes from
// the already-translated leftmost child's Span (O(1)); End from
// gon.End() which is O(1) for the six Pos-recursive types.
func setSpanFromLeftChild(fs *token.FileSet, gon ast.Node, n Node, leftChild Node) {
    if fs == nil {
        return
    }
    endn := fs.Position(gon.End())
    n.SetSpan(Span{
        Pos: leftChild.GetSpan().Pos,
        End: Pos{Line: endn.Line, Column: endn.Column},
    })
}

// setSpanFromRightChild — for End-recursive types. Mirror image:
// Pos from gon.Pos() (O(1)), End from translated right child's Span.
func setSpanFromRightChild(fs *token.FileSet, gon ast.Node, n Node, rightChild Node) {
    if fs == nil {
        return
    }
    posn := fs.Position(gon.Pos())
    n.SetSpan(Span{
        Pos: Pos{Line: posn.Line, Column: posn.Column},
        End: rightChild.GetSpan().End,
    })
}
```

The gate per case is `gon.<recursing-field> is *ast.<SameType>`. This
ensures by induction that the translated child's endpoint matches
`fs.Position(gon.Pos())` / `fs.Position(gon.End())` exactly — without
the gate, `ParenExpr`-unwrap (`go2gno.go:317`) and other column-
shifting translations would corrupt Span. Column-sensitive `TestFiles`
fixtures (`cmp_longtoken.gno`, `anon_convert1.gno`, `shift_deep2.gno`,
`zrealm_method_expr_unexported.gno`) all pass under the gated version.

### Why `gon.End()` instead of `bx.Right.GetSpan().End`

An earlier draft of this fix read `End` from the translated right
child's Span (`bx.Right.GetSpan().End`). That approach has a
correctness bug on `BinaryExpr` when `gon.Y` is `*ast.ParenExpr`:
`Go2Gno`'s ParenExpr case (line 317) unwraps, so `bx.Right` is the
unwrapped inner expression whose End is the column after the inner
expression — **not after `)`**. Reading from `gon.End()` instead
gives `ParenExpr.End() = Rparen+1` which is correct and O(1).

Test `binary/rightmost_paren` in `go2gno_span_test.go` pins this:
`const x = a + b + (c + d)` must have `Span.End` at the column after
`)`, not after `d`.

### BinaryExpr — symmetric Y-side gate

`BinaryExpr` is the only AST type with both Pos- and End-recursion
(`Pos = X.Pos`, `End = Y.End`). The Y-side gate covers precedence-
chained shapes like `1 + 2 * 3 * 4` where `gon.Y` is itself a
BinaryExpr:

```go
case *ast.BinaryExpr:
    bx := &BinaryExpr{...}
    if _, chained := gon.X.(*ast.BinaryExpr); chained {
        setSpanFromLeftChild(fs, gon, bx, bx.Left)
    } else if _, chained := gon.Y.(*ast.BinaryExpr); chained {
        setSpanFromRightChild(fs, gon, bx, bx.Right)
    }
    return bx
```

Precedence-only right chains are bounded by Go's 5 precedence tiers,
so the Y gate is defensive — not a DoS surface today, but keeps the
pattern symmetric and forecloses regressions.

## Benchmark proof

`BenchmarkParseFile_*` benches in
`gnovm/pkg/gnolang/bench_parse_chains_test.go` build chains of
each shape and time `(*Machine).ParseFile`. Asymptotic diagnostic:
linear (O(N)) means 4× N → ~4× ns; quadratic (O(N²)) means 4× N →
~16× ns.

### Before fix (current `main`)

Apple M1 Pro, `-benchtime=3x -count=2`, median ns/op:

| Shape | N=1k | N=4k | N=16k | 4k/1k | 16k/4k |
|---|--:|--:|--:|--:|--:|
| BinaryExpr | 4.59M | 70.8M | 1033M | 15.4× | 14.6× |
| CallExpr (`f()()()…`) | 4.29M | 65.8M | 1216M | 15.3× | 18.5× |
| IndexExpr | 4.22M | 68.8M | 1030M | 16.3× | 15.0× |
| SelectorExpr | 4.17M | 62.1M | 953M | 14.9× | 15.4× |
| SliceExpr | 4.29M | 74.8M | 1083M | 17.4× | 14.5× |
| TypeAssertExpr | 4.38M | 68.6M | 1077M | 15.7× | 15.7× |
| StarExpr | 4.07M | 59.3M | 960M | 14.6× | 16.2× |
| UnaryExpr | 4.39M | 62.8M | 996M | 14.3× | 15.9× |
| ArrayType | 4.96M | 71.8M | 986M | 14.5× | 13.7× |
| MapType | 4.26M | 68.2M | 1001M | 16.0× | 14.7× |

All ratios in the 13.7×–18.5× band → **all 11 shapes are O(N²)**.

### After fix (this branch)

Median ns/op:

| Shape | N=1k | N=4k | N=16k | 4k/1k | 16k/4k | Speedup @ N=16k |
|---|--:|--:|--:|--:|--:|--:|
| BinaryExpr | 825K | 4.26M | 16.2M | 5.2× | 3.8× | **64×** |
| CallExpr | 408K | 2.60M | 13.5M | 6.4× | 5.2× | **90×** |
| IndexExpr | 633K | 3.65M | 17.9M | 5.8× | 4.9× | **58×** |
| SelectorExpr | 502K | 3.05M | 15.6M | 6.1× | 5.1× | **61×** |
| SliceExpr | 509K | 3.94M | 15.1M | 7.7× | 3.8× | **72×** |
| TypeAssertExpr | 619K | 4.17M | 20.2M | 6.7× | 4.8× | **53×** |
| StarExpr | 494K | 2.76M | 16.4M | 5.6× | 5.9× | **59×** |
| UnaryExpr | 437K | 2.28M | 15.4M | 5.2× | 6.8× | **65×** |
| ArrayType | 675K | 3.85M | 17.3M | 5.7× | 4.5× | **57×** |
| MapType | 647K | 3.82M | 19.2M | 5.9× | 5.0× | **52×** |

All ratios in 3.8×–7.7× band → **all 11 shapes are O(N)**. The
excess over perfect 4× is Go-parser and Gno-AST allocation cost, both
genuinely O(N). Speedups at N=16000 range **52×–90×**.

## Verification

```sh
go build ./gnovm/pkg/gnolang/
go test ./gnovm/pkg/gnolang/ -run TestSpan -count=1 -v
go test ./gnovm/pkg/gnolang/ -run TestFiles -test.short -count=1 -timeout 600s
go test -run=NONE -bench=BenchmarkParseFile_ -benchtime=3x -count=2 \
    -timeout=600s ./gnovm/pkg/gnolang/
go test ./gno.land/pkg/integration/ \
    -run 'TestTestdata/(gc|ghverify|gnokey_gasfee|restart_gas|simulate_gas|issue_4983|restart_missing_type)$' \
    -count=1
```

All pass on this branch. `TestSpan` has 23 sub-cases covering all 11
chain shapes with both gate-firing and gate-boundary (paren-wrapped
operand) cases plus the BinaryExpr Y-side gate (precedence chain) and
both-sides case.

## Audit — all chain-shapeable AST types

`$GOROOT/src/go/ast/ast.go:495-569` (Go 1.24.1) gives complete
Pos/End for every Expr / type expression. Status against this fix:

| Type | Pos() | End() | Chain shape | Status |
|---|---|---|---|---|
| `BinaryExpr` | `X.Pos()` ← | `Y.End()` | `1+1+…` (left), `1+2*3*4` (right via precedence) | ✓ both gates |
| `CallExpr` | `Fun.Pos()` ← | `Rparen+1` | `f()()()…` | ✓ left-Pos |
| `IndexExpr` | `X.Pos()` ← | `Rbrack+1` | `a[0][0]…` | ✓ left-Pos |
| `SelectorExpr` | `X.Pos()` ← | `Sel.End()` | `a.b.b…` | ✓ left-Pos |
| `SliceExpr` | `X.Pos()` ← | `Rbrack+1` | `s[:][:]…` | ✓ left-Pos |
| `TypeAssertExpr` | `X.Pos()` ← | `Rparen+1` | `x.(I).(I)…` | ✓ left-Pos |
| `StarExpr` | `Star` | `X.End()` ← | `***…T` | ✓ right-End |
| `UnaryExpr` | `OpPos` | `X.End()` ← | `!!!x` | ✓ right-End |
| `ArrayType` | `Lbrack` | `Elt.End()` ← | `[1][1]…int` | ✓ right-End |
| `ChanType` | `Begin` | `Value.End()` ← | `chan chan…T` | n/a — Go2Gno rejects via `panicWithPos("channels are not permitted")`; Gno disallows channels |
| `MapType` | `Map` | `Value.End()` ← | `map[K]map[K]…V` | ✓ right-End |

### Shapes that look chainable but aren't a DoS surface

- **`ParenExpr`** — `Go2Gno` unwraps it (`go2gno.go:317`), returning
  the inner translated node verbatim. The inner already has its Span
  set; the deferred `setSpan` sees `!IsZero()` and short-circuits.
  `ParenExpr.Pos()`/`End()` are both O(1) anyway.

- **`LabeledStmt`** (`L1: L2: L3: ;`) — `End() = Stmt.End()` is
  recursive in stdlib, but `Go2Gno`'s case (line 597-600) returns
  `toStmt(fs, gon.Stmt)` (delegating, like `ParenExpr`). The inner
  stmt's Span is preserved through the recursion; the deferred
  `setSpan` short-circuits via `IsZero`. Empirically confirmed O(N)
  via `BenchmarkParseFile_LabeledStmtChain` (ratios 5.9× / 5.4× per
  4× N — linear band).

- **Right-leaning `BinaryExpr`** (`1 + (1 + (1 + …))`) — every level
  requires explicit parens, so `gon.Y` is *ast.ParenExpr at each
  level, not *ast.BinaryExpr. `ParenExpr.End() = Rparen+1` is O(1),
  so the End() recursion never goes deep. Empirically confirmed O(N)
  via `BenchmarkParseFile_BinaryChainRightLeaning` (ratios 6.2× /
  4.9× per 4× N — linear band). **The earlier draft of this ADR
  incorrectly listed right-leaning BinaryExpr as a potential O(N²)
  surface; benchmarks prove otherwise.**

- **Precedence-chained BinaryExpr** (`1 + 2 * 3 * 4`) — Y can be a
  BinaryExpr via precedence, but Go has only 5 precedence tiers,
  capping the right-chain depth at 5. Not a DoS surface, but the Y
  gate keeps the pattern symmetric.

- **`*ast.IndexListExpr`** — `Go2Gno` panics on this (line 691) since
  Gno disallows generics. Not reachable.

- **`*ast.KeyValueExpr`** — `Pos() = Key.Pos()` is recursive in
  stdlib. `Key`/`Value` are typed `Expr`, but the Go grammar
  disallows `KeyValueExpr` as a Key or Value: the colon in `key:
  value` is only meaningful inside a `CompositeLit` `Elts` list, not
  inside an arbitrary expression. `{a: b: c}` is a parse error
  (unexpected `:` after `b`). So a `KeyValueExpr` chain is not
  syntactically expressible — not a DoS surface.

- **All `Stmt` types** (`AssignStmt`, `BlockStmt`, `IfStmt`, etc.) —
  their `Pos()`/`End()` delegate to child *expressions* or
  *statement lists*, not to the same statement type. No self-
  chaining; not a DoS surface.

- **`FuncType` / `InterfaceType` / `StructType`** — Pos/End from
  surrounding token positions, O(1). Not affected.

## Completeness claim

After this fix lands, no AST shape that the Go parser accepts can
cause O(N²) parse-time cost in `SpanFromGo`. The argument is:

1. The stdlib `go/ast` audit (Go 1.24.1, `ast.go:495-569`)
   enumerated every type whose `Pos()` or `End()` method recurses
   through a child field. There are exactly 11 such types.
2. All 11 are gated in this fix and empirically confirmed O(N) by
   `BenchmarkParseFile_*` (52×–90× speedups at N=16000).
3. Every other AST type is safe by one of three structural reasons:

   **(a) Constant token positions — no recursion at all.** The
   method returns a stored `token.Pos` field directly, O(1) per call
   regardless of subtree depth. Examples where both `Pos()` and
   `End()` are constants:
   - `BlockStmt.Pos() = s.Lbrace`, `End() = s.Rbrace+1`
   - `BasicLit.Pos() = x.ValuePos`, `End() = ValuePos+len(Value)`
   - `Ident.Pos() = x.NamePos`, `End() = NamePos+len(Name)`

   Some types have a constant `Pos()` but a recursive `End()` that
   delegates to a non-self-chainable child — also safe:
   - `ReturnStmt.Pos() = s.Return`, `End() = Results[n-1].End()`
     (results are arbitrary Exprs, not ReturnStmt)
   - `FuncDecl.Pos() = d.Type.Func`, `End() = Body.End()` or
     `Type.End()` (FuncType, not FuncDecl)
   - `StructType.Pos() = t.Struct`, `End() = Fields.End()` (FieldList)
   - `FuncType.Pos() = t.Func`, `End()` from Results/Params FieldList

   **(b) Delegates to a *different* AST type — chains can't form.**
   The recursing field can never be the same type as the parent, so
   depth-N self-chains aren't expressible.
   - `AssignStmt.Pos() = s.Lhs[0].Pos()` — `Lhs[0]` is an `Expr`,
     never another `AssignStmt`. So `x = y = z` doesn't exist as a
     chain — Go disallows assignment-as-expression.
   - `IncDecStmt.Pos() = s.X.Pos()` — `X` is `Expr`, not
     `IncDecStmt` (`x++++` is a parse error).
   - `ExprStmt.Pos() = s.X.Pos()` — `X` is `Expr`. Can't nest
     `ExprStmt` inside `ExprStmt`.
   - `KeyValueExpr.Pos() = e.Key.Pos()` — `Key` is `Expr`, typically
     `Ident` or `BasicLit`. Composite literal keys aren't themselves
     `KeyValueExpr` (`{a: b: c}` is a parse error).
   - `SendStmt.Pos() = s.Chan.Pos()` — `Chan` is `Expr`, not
     `SendStmt`.

   **(c) Structurally delegating in `Go2Gno` — `IsZero` short-
   circuits the defer.** The case returns an *inner* translated
   node verbatim instead of constructing a new one. The inner's Span
   was set during its own recursive `Go2Gno` call, so the outer's
   deferred `setSpan(fs, gon, n)` sees `!n.GetSpan().IsZero()` and
   does nothing — `SpanFromGo(fs, gon)` is never called.
   - `case *ast.ParenExpr: return toExpr(fs, gon.X)` (`go2gno.go:317`)
     — `(((x)))` chains are O(N), not O(N²).
   - `case *ast.LabeledStmt: stmt := toStmt(fs, gon.Stmt); ...; return stmt`
     (`go2gno.go:597-600`) — `L1: L2: L3: ;` chains are O(N),
     confirmed by `BenchmarkParseFile_LabeledStmtChain` (ratios 5.9× /
     5.4× per 4× N).

The 23 `TestSpan` sub-cases lock in correctness; the 13 benchmarks
lock in the asymptote; the audit above documents the closure
argument. Adding a new AST type to a future Go release would require
re-running the audit, but no current shape regresses.

## Out of scope

Separate concerns, separate PRs:

- The depth caps and per-tx allocator from `fix/jae/preprocess-dos`
  (#5642) — orthogonal, addresses preprocess-time surfaces.
- Other parse-time DoS surfaces not involving `SpanFromGo` (none
  currently known).
