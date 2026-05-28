# PR5712: explicit `MemFile.Kind` enum + drop the `_filetest.gno` suffix

## Status

Accepted.

## Context

A `MemFile` is the wire-format unit of a Gno package's file contents
(`tm2/pkg/std/memfile.go`). Until this PR, "what kind of file is this?"
was answered by string-sniffing `MemFile.Name`:

- `*_test.gno` → in-package test
- package-clause `xxx_test` → integration (xxx) test
- `*_filetest.gno` → standalone filetest

This worked, but encoded classification in a regex-friendly substring of
the filename. Two follow-on needs surfaced after PR #5704 introduced the
on-disk `filetests/` subdir convention:

- Authors wanted to drop the `_filetest.gno` suffix entirely once the
  directory carried the classification (every `.gno` in `filetests/` is
  a filetest by construction).
- Every classifier in the tree (gnoweb, doc, packages.Load, gnodev,
  file walkers, lint, test) had to learn whatever convention the suffix
  encoded; suffix sniffing scattered across ~25 files made future
  evolution fragile.

A first draft of this PR allowed `MemFile.Name` itself to carry a
`filetests/` prefix. That approach was rejected after review feedback
from @thehowl:

> IMO if we start having directories within mempackages it quickly turns
> into a mess (this needs to be handled on gnoweb, and there's no way to
> unambiguously get what the actual pkgpath is from a filename). I would
> avoid the chaos that is entailed with having subdirs in mempackages and
> simply have the filetest special-case when using WriteTo.

The objection lands cleanly: a path-shaped `Name` makes pkgpath
reconstruction ambiguous and forces every consumer to relearn the
prefix convention. But the alternative — keeping `_filetest.gno`
everywhere forever — abandons the suffix-drop goal. This ADR proposes
the third path: make classification a **field on `MemFile`**.

## Decision

Add `MemFile.Kind MemFileKind` carrying the classification explicitly.
`MemFile.Name` remains a flat basename.

```go
type MemFileKind uint8

const (
    KindUnknown       MemFileKind = iota // legacy / unset; fall back to suffix
    KindPackageSource                    // prod .gno file
    KindTest                             // *_test.gno, same package
    KindXTest                            // *_test.gno, xxx_test package
    KindFiletest                         // standalone filetest
    KindOther                            // non-.gno (md, toml, LICENSE, etc.)
)

type MemFile struct {
    Name string      `json:"name" yaml:"name"`
    Body string      `json:"body" yaml:"body"`
    Kind MemFileKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}

func (mfile *MemFile) IsFiletest() bool { ... }
```

Concretely:

- `MemPackage.WriteTo` routes by `Kind` (preferring it; falling back to
  the legacy `_filetest.gno` suffix when `Kind == KindUnknown`). Filetests
  land at `<dir>/filetests/<Name>`; everything else at `<dir>/<Name>`.
- `gnolang.ReadMemPackage` stamps `Kind` at read time: any `.gno` file
  under `<dir>/filetests/` is loaded with `Kind = KindFiletest` and its
  bare basename as `Name`. **The `_filetest.gno` suffix is no longer
  required** — the subdir IS the classification.
- `gnolang.ReadMemPackageFromList` reuses the same classification
  (`classifyMemFileKind`) from the disk path that produced each entry.
- `pkgdownload.Download` (used by `examplespkgfetcher`, `rpcpkgfetcher`)
  also routes by `Kind` so a downloaded package round-trips through
  `ReadMemPackage` cleanly.
- The legacy filename convention survives as a fallback. `IsFiletestName`
  is the suffix check; `(*MemFile).IsFiletest()` prefers `Kind`, falls
  back to `IsFiletestName(Name)` only when `Kind == KindUnknown`.
- The walker collapse in `gnovm/cmd/gno/util.go` (`<pkg>/filetests/` →
  `<pkg>`) is kept — it's still needed because the on-disk layout
  retains `filetests/`.
- The lint-side work in commit 4ff865d10 (per-filetest panic isolation,
  `// Error:` / `// TypeCheckError:` directive awareness) is kept.

## On-chain hash stability

This is the load-bearing constraint. A `MemPackage` is amino-encoded
and hashed to derive package addresses. Adding a field to `MemFile`
naively would change the canonical encoding for every package and break
every on-chain address.

Mitigation: **`MemFile.Kind` is intentionally NOT in the amino canonical
encoding.** The hand-rolled `MemFile.MarshalBinary2` / `SizeBinary2` /
`UnmarshalBinary2` in `tm2/pkg/std/pb3_gen.go` only encodes `Name` and
`Body` — they are untouched by this PR. As a result:

- amino-encoded bytes are byte-identical for any MemPackage before and
  after this PR; on-chain hashes are stable.
- An amino-decoded MemFile comes back with `Kind = KindUnknown` (the
  iota zero value). The legacy suffix-based fallback in
  `(*MemFile).IsFiletest()` and `IsFiletestName(Name)` handles it
  correctly: `_filetest.gno` files keep being recognized.
- The field IS serialized to JSON and YAML (with `omitempty`) so
  off-chain tooling and persistence layers can see it.

Trade-off: new-style filetests with a bare `.gno` name **cannot survive
a round-trip through amino canonical encoding** — the `Kind` field is
lost, and there is no `_filetest.gno` suffix to recover it. New-style
filetests therefore live only in tooling contexts (on-disk under
`filetests/`, in-memory MemPackages, JSON/YAML transit) and never enter
the on-chain canonical form. Filetests are a development/test artifact
and don't ship on-chain anyway, so this is acceptable.

## Migration of consumers

Code that previously called `std.IsFiletestName(mfile.Name)` falls into
two groups:

1. **Has a MemFile in hand.** Switch to `mfile.IsFiletest()` — picks up
   new-style filetests via `Kind`, falls back to the suffix for legacy
   data. Updated in this PR: `gnovm/pkg/gnolang/mempackage.go`
   (`FilterGno`, `ExcludeGno`), `gnovm/pkg/gnolang/gotypecheck.go`,
   `gnovm/pkg/doc/pkg.go`, `gnovm/cmd/gno/common.go`, `gnovm/cmd/gno/lint.go`,
   `contribs/gnodev/pkg/packages/{loader_base,package}.go`.

2. **Operates on a path string only.** Use the path-aware `IsTestFile`
   updated in this PR — it recognizes `_test.gno`/`_filetest.gno` by
   suffix AND any `.gno` inside a `filetests/` parent directory. Used by
   walkers like `gnovm/cmd/gno/util.go`.

3. **Operates on a basename only (no path, no MemFile).** Still uses
   `IsFiletestName(Name)`. This misses new-style filetests, but those
   callers (`gnovm/cmd/gno/{run,mod,tool_transpile}.go`,
   `gnovm/pkg/transpiler/transpiler.go`, `gnovm/pkg/gnofmt/utils.go`,
   `gno.land/pkg/gnoweb/components/view_source.go`) operate on
   contexts where the legacy layout is still safe to assume — a
   targeted follow-up will migrate them to receive Kind information
   when needed.

`gnovm/pkg/packages/filekind.go` gains a parallel `GetMemFileKind`
helper that prefers `Kind` and delegates to `GetFileKind` when unset.
`load_matches.go` and `imports.go` switch to it; the import-graph
correctly includes filetest imports for new-style filetests.

## Validation rules

- `MemFile.Name` is enforced flat (no `/`) by the existing `reFileName`
  regex — restored unchanged from pre-PR-5704. Any caller that produces
  `MemFile{Name: "filetests/..."}` is invalid by validation.
- `ReadMemPackage` only loads `.gno` files (not arbitrary kinds) from
  `filetests/`. The `Kind` stamping is unconditional for `.gno` in that
  subdir; for files at the package root it follows the suffix
  convention.
- `ValidateMemPackageAny` uses `mfile.IsFiletest()` for filetest-aware
  branches; otherwise unchanged.

## Consequences

- The on-disk `filetests/foo.gno` layout (introduced in commit 5038f9d2f
  on this branch) is preserved. The suffix-drop goal is achieved.
- No wire format change. Amino-marshaled MemPackages stored on chain
  before this PR remain bit-for-bit identical; hashes/addresses stable.
- Most classifiers across the codebase now prefer the explicit field;
  the suffix shim survives for legacy data and path-only callers.
- The `MemFile.Kind` field has minimal but real cost: a `uint8` per
  in-memory file and the discipline of stamping it at construction
  sites (`ReadMemPackage`, fetchers). New construction sites must do
  the same.
- The follow-up migration in (3) above is tracked but not blocking:
  every consumer that still suffix-sniffs continues to handle legacy
  data correctly and misses only the new no-suffix case in code paths
  outside lint/test/build.

## Alternative considered: morgan's flat-Name + WriteTo special-case

The middle path was to keep `MemFile.Name` flat and let `WriteTo`
special-case the `_filetest.gno` suffix — preserving the suffix as
the canonical filetest identifier. It was rejected because it
abandons the suffix-drop goal: every filetest stays named
`*_filetest.gno` forever, the directory `filetests/` becomes
redundant, and classifier code keeps sniffing strings. The Kind enum
generalizes the same insight (classification is data, not a substring)
while finishing the suffix-drop work.
