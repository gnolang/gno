# SpanFromGo O(N²) on left-leaning BinaryExpr chains

## Status

Implemented in `fix/spanfromgo-on2`.

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

**Mechanism.** The Go standard library defines
`(*ast.BinaryExpr).Pos() = X.Pos()` — it recurses leftward through
the chain to return the leftmost leaf's position. `SpanFromGo`
(`gnovm/pkg/gnolang/nodes_location.go:107`) calls `gon.Pos()` on
every AST node it visits during `Go2Gno`'s translation walk. For a
left-leaning BinaryExpr tree of depth N, each `setSpan` call costs
O(N) and N setSpan calls sum to O(N²).

## Why preprocess-side mitigations don't help

The cost lands in the parse step, which runs in the message handler
**before** the preprocess pass begins. By the time any of the
following would fire, the 13 seconds have already been spent:

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
  left-leaning value-expression chains.

## Decision

Patch `Go2Gno`'s `*ast.BinaryExpr` case to set the Span from the
already-translated Gno children's Spans when `gon.X` is also an
`*ast.BinaryExpr`. This avoids the recursive `gon.Pos()` walk.

The non-chain shapes (`gon.X` is `*ast.ParenExpr`, `*ast.Ident`,
`*ast.SelectorExpr`, etc.) keep the original behavior because their
`Pos()` is O(1) (returns `Lparen`, `NamePos`, etc.) and their
column-position semantics differ subtly from a child's translated
Span — `Go2Gno` unwraps `*ast.ParenExpr`, for instance, so reading
the unwrapped Gno child's `Span.Pos` would shift positions by one
column relative to what existing tests expect.

The gate `_, ok := gon.X.(*ast.BinaryExpr); ok` is what makes the
fix safe to land without modifying any test fixture.

## The patch

```go
case *ast.BinaryExpr:
    // … (see source comment for the full rationale) …
    bx := &BinaryExpr{
        Left:  toExpr(fs, gon.X),
        Op:    toWord(gon.Op),
        Right: toExpr(fs, gon.Y),
    }
    if _, chained := gon.X.(*ast.BinaryExpr); chained && fs != nil &&
        bx.Left != nil && bx.Right != nil {
        lspan := bx.Left.GetSpan()
        rspan := bx.Right.GetSpan()
        if !lspan.IsZero() && !rspan.IsZero() {
            bx.SetSpan(Span{Pos: lspan.Pos, End: rspan.End})
        }
    }
    return bx
```

## Benchmark proof

`BenchmarkParseFile_BinaryChain` is a new parameterized bench in
`gnovm/pkg/gnolang/bench_parse_test.go`. It builds a
`const x = 1 + 1 + … + 1` source of depth N and times one
`(*Machine).ParseFile` call. The ratio of ns/op between successive N
values is the diagnostic: linear (O(N)) means 4× N → ~4× ns;
quadratic (O(N²)) means 4× N → ~16× ns.

```
go test -run=NONE -bench=BenchmarkParseFile_BinaryChain \
    -benchtime=10x -count=3 ./gnovm/pkg/gnolang/
```

### Before fix (upstream/master @ 5111dbc22)

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro
BenchmarkParseFile_BinaryChain/N=1000-10           10     3950321 ns/op
BenchmarkParseFile_BinaryChain/N=1000-10           10     4030812 ns/op
BenchmarkParseFile_BinaryChain/N=1000-10           10     3966283 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10    63856033 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10    62466729 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10    62965483 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10   991612592 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10   993676375 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10   979176717 ns/op
```

Median ns/op:
| N       |        ns/op | ratio vs prev N (4×) |
|--------:|-------------:|---------------------:|
| 1 000   |    3 966 283 | —                    |
| 4 000   |   62 965 483 | **15.9×**            |
| 16 000  |  991 612 592 | **15.7×**            |

→ pure O(N²).

### After fix (fix/spanfromgo-on2)

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro
BenchmarkParseFile_BinaryChain/N=1000-10           10      545808 ns/op
BenchmarkParseFile_BinaryChain/N=1000-10           10      462279 ns/op
BenchmarkParseFile_BinaryChain/N=1000-10           10      666638 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10     2919604 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10     2441554 ns/op
BenchmarkParseFile_BinaryChain/N=4000-10           10     2574888 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10    11588533 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10    11804625 ns/op
BenchmarkParseFile_BinaryChain/N=16000-10          10    11371025 ns/op
```

Median ns/op:
| N       |       ns/op | ratio vs prev N (4×) | speedup vs before |
|--------:|------------:|---------------------:|------------------:|
| 1 000   |     545 808 | —                    | **7.3×**          |
| 4 000   |   2 574 888 | **4.7×**             | **24.5×**         |
| 16 000  |  11 588 533 | **4.5×**             | **85.6×**         |

→ O(N), as expected. The slight excess over a perfect 4× (4.5–4.7×)
is from the Go parser's own work and Gno AST node allocation, both
of which are genuinely O(N).

## Verification

```sh
go build ./...
go test ./gnovm/pkg/gnolang/ -run TestFiles -test.short -count=1 -timeout 600s
go test ./gnovm/pkg/gnolang/ -run TestParseFile_BinaryChain_SpanCorrect -count=1 -v
go test ./gno.land/pkg/integration/ \
    -run 'TestTestdata/(gc|ghverify|gnokey_gasfee|restart_gas|simulate_gas|issue_4983|restart_missing_type)$' \
    -count=1
```

All commands pass on the branch. `TestFiles` includes the
column-sensitive cases (`cmp_longtoken.gno`, `anon_convert1.gno`,
`shift_deep2.gno`, `zrealm_method_expr_unexported.gno`) that surfaced
during local development when an unconditional version of this fix
was tried; they all pass under the gated version because their `X`
nodes are `*ast.ParenExpr` or `*ast.Ident`, not `*ast.BinaryExpr`.

## Follow-up audit

Other Go AST types whose `Pos()` recurses through a child:

| Type                  | Pos()         | Shape that would chain     |
|-----------------------|---------------|----------------------------|
| `*ast.CallExpr`       | `Fun.Pos()`   | `f(g(h(...)))`             |
| `*ast.IndexExpr`      | `X.Pos()`     | `a[0][0][0]...`            |
| `*ast.SelectorExpr`   | `X.Pos()`     | `x.a.b.c.d...`             |
| `*ast.SliceExpr`      | `X.Pos()`     | `s[:][:][:]...`            |
| `*ast.TypeAssertExpr` | `X.Pos()`     | `x.(I).(I).(I)...`         |

In practice these are bounded:

- `CallExpr`: `Fun` is typically `*ast.Ident` or `*ast.SelectorExpr`
  wrapping an Ident — both are O(1) for `Pos()`. Even chained
  `f(g(h(...)))` is shallow in real Gno code (`Args` arity is
  small; the chain depth is the call nesting, also small).
- `IndexExpr`, `SliceExpr`: the parser caps composite type depth
  (see `validateTypeDepth` and `MaxTypeDepth = 8` in
  `fix/jae/preprocess-dos`), preventing the array `[1]...[1]int`
  shape that's required to make the chain meaningful.
- `SelectorExpr`: requires nested struct embedding or recursive
  pointer types to type-check at any depth; both have their own caps.
- `TypeAssertExpr`: requires a chain of `*Ident` interface types,
  bounded by the same parser limits.

If a concrete adversarial benchmark surfaces for any of these, the
same children-span pattern applies — but the gating condition needs
care because `Go2Gno` unwraps `*ast.ParenExpr`, which can shift
column positions relative to `gon.Pos()`. Add the fix incrementally
once a real triggering shape is benched.
