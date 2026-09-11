# Bound predefine declaration lookup

## Context

`predefineRecursively2` resolves each undefined package name with
`FileSet.GetDeclFor`. That lookup scans files and declarations linearly. A
package-level dependency chain can perform one full scan per declaration, so
predefinition consumes O(N²) host work even though each declaration is finite.

`FileSet` is also mutable: `Machine.runFileDecls` appends files before calling
`PredefineFileSet` with only the newly added files. Multi-name value
declarations can then be split in place during predefinition. A persistent
index, or one storing declaration slice positions, can therefore become stale.

## Decision

`PredefineFileSet` eagerly builds a declaration index local to that call, immediately
after static blocks are initialized. It indexes the complete `pn.FileSet`, not
the new-file-only `fset`, and stores the concrete `Decl` pointer with its
`FileNode`.
The index retains that `FileSet`, so indexed hits and miss fallback cannot
observe different file sets.

Construction matches `FileSet.GetDeclForSafe`: files are visited in reverse,
declarations within a file in source order, imports are skipped, and only the
first entry for each name is retained. `PredefineFileSet` uses the indexed
recursive-predefinition wrapper, which passes the index through its call chain.
Standalone `Preprocess` uses the live-lookup wrapper and retains the existing
`FileSet.GetDeclFor` behavior, including its miss panic.

When a multi-name `ValueDecl` is split, each index entry is replaced with
the concrete split part only if the current winner still identifies the same
file and original declaration. This preserves later-file and earlier-declaration
winners while avoiding stale slice indices.

## Alternatives considered

- Add an index to `FileSet`, built eagerly or lazily. This extends mutable,
  serialized VM state with cache invalidation and synchronization concerns;
  incremental file addition and declaration splitting can stale the cache.
- Store `{fileIdx, declIdx}` positions. Splitting a declaration shifts later
  positions and can resolve a name to the wrong declaration or report a false
  cycle.
- Charge gas for the repeated scans. Metering can price the work but does not
  bound the host-side O(N²) algorithm, and changing gas costs requires broader
  consensus-version coordination.

## Consequences

Index construction is linear in files, declarations, and declared names;
dependency declaration lookup is an allocation-free map lookup. The temporary
map is released with `PredefineFileSet`, and no codec, copy, or persistent
`FileSet` representation changes.

`GetDeclForSafe` deliberately remains O(N) per call. Outside the indexed hot
path, its remaining production callers are the circular-dependency error path
at `preprocess.go:6290`, which performs one lookup immediately before panicking,
and `GetNameSourceForPath` at `nodes.go:2082`, an offline `gno fix` path already
marked "Too slow for runtime." Neither permits attacker-driven O(N²) work, so
changing the `FileSet` interface is unnecessary to close the exploitable path.

Other quadratic preprocess work remains outside this change: repeated work in
`tryPredefine` (around `preprocess.go:5383`) and `slices.Contains` in
`resolveEffectiveDeps` (around `preprocess.go:6193`).

## Appendix: comparison with the FileSet index alternative

Three implementations were benchmarked on the same workload: the original linear scan (baseline); a persistent index on `FileSet`, built lazily on first lookup, storing `{fileIdx, declIdx}` positions; and the chosen call-local index, built once per `PredefineFileSet` call.

The benchmark parses and runs `PredefineFileSet` over value declaration chains of length N. The reverse case makes each declaration depend on the next — the worst case for the linear scan. The forward case makes each declaration depend on a name that is already defined. Conditions: `N = 250, 500, 1000, 2000`, `-benchtime=10x -count=6`, darwin/arm64. Medians shown; `allocs/op` variance was 0% in every cell.

| Case         | baseline | FileSet index | chosen call-local index |
|--------------|---------:|--------------:|------------------------:|
| Reverse/250  |    49.5k |         18.4k |                   18.4k |
| Reverse/500  |   161.4k |         36.7k |                   36.7k |
| Reverse/1000 |   572.7k |         73.3k |                   73.3k |
| Reverse/2000 |    2.15M |        146.3k |                  146.3k |
| Forward/2000 |   146.2k |        146.2k |                  148.3k |

The doubling ratio from 250 to 2000 shows the change in complexity. The baseline reverse case grows by 3.26x, 3.55x, then 3.75x, approaching O(N²). Both the FileSet index and the chosen call-local index grow by roughly 2.0x, which is O(N). The forward case is O(N) for all three.

Findings:

- On the reverse case the two O(N) approaches are equivalent. Their `ns/op` difference was not statistically significant at any N (`p` between 0.13 and 0.94); each performs one full pass over the declarations.
- On the forward case the chosen call-local index is always built, costing about 1.4% more `allocs/op` and 1.5% more `B/op` than the FileSet index, which is lazy and never built when no undefined name is looked up.

The call-local index was chosen for the correctness and lifecycle reasons in *Alternatives considered*, not for a performance advantage: it keeps no cache on the mutable, shared `FileSet`, and holds no declaration slice positions that a split can invalidate. This benchmark does not measure those differences.
