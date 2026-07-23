# Strip per-file Go version before the consensus type-check

## Context

The consensus type-check builds its `go/types` config with a pinned
`GoVersion: "go1.18"` in `TypeCheckMemPackage` (`gnovm/pkg/gnolang/gotypecheck.go`).
The pin exists so the accept/reject verdict for a submitted package is a function
of the package alone, not of the Go toolchain each validator binary happened to
be built with. Two honest validators must reach the same verdict.

The pin only governs files that carry no per-file language version. `go/types`
resolves each file's version in `initFiles`: a `//go:build go1.N` line upgrades
that file's version above the Config pin, and a file whose version exceeds the
building toolchain's version is rejected with
`file requires newer Go version goX (application built with goY)`, where `goY` is
the version the binary was compiled with.

Package bodies are attacker-supplied and reach `go/types` through gno's
`go/parser`, which populates `ast.File.GoVersion` from `//go:build` lines.
`prepareGoGno0p9` did not strip it. So a submitter could raise the gate on their
own file (`//go:build go1.22` + `for range 10` type-checks under the go1.18 pin),
and a file tagged above one validator's toolchain was accepted on a newer build
and rejected on an older one. Both are state forks, not merely results-hash
divergence.

## Decision

Blank `ast.File.GoVersion` on every parsed `.gno` file in `GoParseMemPackage`,
immediately after a successful parse. Build constraints have no meaning in Gno,
so no per-file version can legitimately be set from inside a submitted package.
With the field empty, `Config.GoVersion` is the sole version authority for every
file and every imported package.

## Alternatives considered

- **Reject files that carry a `//go:build` line.** Heavier, changes accepted
  input, and needs its own consensus-stable error. Blanking the field is a strict
  subset: build tags become inert rather than fatal.
- **Strip only `go1.N` version tags, keep other constraints.** No gain. Gno
  honours no build constraint, so the whole field is meaningless; blanking it
  is simpler and closes the version axis unconditionally.
- **Raise the Config pin.** Does not help: the per-file upgrade direction is
  always allowed, so any pin remains raisable from inside the package.

## Consequences

- The pin now holds for tagged and untagged files alike; the verdict no longer
  depends on the builder's toolchain.
- No behaviour change for the existing corpus: a `.gno` file with no `//go:build`
  line already has an empty `GoVersion`, so blanking is a no-op there.
- `TestTypeCheckMemPackage_BuildTagCannotRaisePin` pins both halves: a build tag
  cannot raise the pinned version, and the verdict never references the building
  toolchain.
