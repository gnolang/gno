# ADR: migrate gnodev to the gnovm native package loader

Status: Accepted
PR: [#4957](https://github.com/gnolang/gno/pull/4957)
Date: 2026-04-20

## Context

`contribs/gnodev/pkg/packages/` implements a package loading/resolving
subsystem that predates `gnovm/pkg/packages/` (the native loader introduced
by noiznoiz). The two systems duplicate concerns:

| Concern | gnodev | gnovm |
|---|---|---|
| Pattern expansion (glob) | `GlobLoader` + `glob.go` | `expandPatterns` in `Load` |
| Dep walking | `BaseLoader.Load` + import recursion | `loadMatches` + `Deps: true` |
| Stdlib handling | `FilterStdlibs` middleware | Built in |
| Remote fetching | `resolver_remote.go` (RPC via `vm/qfile`) | `rpcpkgfetcher` |
| Mock/test packages | `resolver_mock.go` + `MockLoader` | (absent) |
| Syntax pre-check | `PackageCheckerMiddleware` | Built into parsing |
| Per-path lookup | `Resolver.Resolve(fset, path)` via chain | (absent — workspace/pattern only) |

gnodev's loader shape: `Loader` interface + `BaseLoader` + `Resolver`
interface with Local/Remote/Root/Mock implementations, chained and wrapped in
middleware (Log, Cache, FilterStdlibs, PackageChecker).

gnodev's lazy mode puts an HTTP proxy (`pkg/proxy/path_interceptor.go`) in
front of the RPC node. The proxy parses incoming tx/query bodies, extracts
referenced package paths, calls `Loader.Resolve(path)` per path, and
triggers a reload when a new resolvable path is seen.

### The shape mismatch

gnovm's `Load(LoadConfig, patterns...) PkgList` is eager and pattern-based:
a call returns the fully resolved graph for a workspace. It has **no public
single-path entry point** — you cannot ask "resolve this one import path."

gnodev's proxy-driven lazy mode needs exactly that: given an arbitrary path
from an incoming RPC request, find that one package (or fail).

The migration must resolve this mismatch.

## Decision

Replace gnodev's loader/resolver subsystem with a single `Loader` struct in
`contribs/gnodev/pkg/packages/` that delegates to `gnovm.Load` for bulk work
and implements per-path lookup locally. Reshape `gnodev`'s internals around
it.

### Shape

```go
type Config struct {
    Workspace       string                     // "" if none (detected from CWD)
    Examples        bool                       // include $GNOROOT/examples in the lazy set
    ExtraRoots      []string                   // user-supplied additional roots
    ExcludeDirs     []string                   // dir paths the FS scanner must SkipDir (matched against walker output)
    GnoRoot         string
    RemoteOverrides map[string]string          // domain → RPC URL (ignored when Fetcher set)
    Fetcher         pkgdownload.PackageFetcher // override the default rpcpkgfetcher (tests)
    Logger          *slog.Logger
}

type Loader struct { /* cfg, fetcher, index, tracked, rootIdx, mu */ }

func New(cfg Config) *Loader

func (l *Loader) LoadWorkspace() ([]*Package, error)        // eager, workspace only
func (l *Loader) Reload()        ([]*Package, error)        // workspace + tracked
func (l *Loader) LoadAll()       ([]*Package, error)        // workspace + examples + roots, all eager
func (l *Loader) Resolve(path string) (*Package, error)     // per-path lookup for the proxy
func (l *Loader) LookupFS(path string) bool                 // FS-only, non-mutating lookup for diagnostics
```

`LoadWorkspace` and `LoadAll` call `gnovm.Load` directly — they pass
workspace / root patterns that `expandPatterns` understands.

`Reload` is hybrid, because `gnovm`'s pattern expander treats bare import
paths (what the proxy accumulates via `Resolve`) as modcache lookups, not
filesystem scans. So `Reload` calls `gnovm.Load` once for the workspace
pattern, clears tracked entries from `index` (so they are re-derived from
disk on the next `Resolve`), and expands every tracked path into its
transitive import closure: each package's chain imports (read via
`Package.Imports`, which reuses gnovm's import parser and excludes test
files and stdlibs) are resolved recursively and emitted dependency-first,
so genesis deploys them in order. Without the closure, a lazily-loaded
realm's addpkg fails chain-side type-checking with "unknown import path" —
the proxy only ever sees the queried path, never its imports. An
unresolvable import warns and is skipped rather than failing the reload;
the chain then reports the precise type-check error for that one package.
`rootIdx` is preserved across Reload — see the Root scan caching section.

`Resolve` does **not** call `gnovm.Load`. It:
1. Hits the internal index if already loaded.
2. Shallow-scans `ExtraRoots` (and `$GNOROOT/examples` if `cfg.Examples`)
   for a directory whose `gnomod.toml` `module` field matches the path.
3. Falls back to `rpcpkgfetcher.FetchPackage` for remote-shaped paths.
4. Returns `ErrPackageNotFound` otherwise.

Hits populate the index and a `tracked` set used by `Reload`.

`LookupFS` is the read-only sibling of `Resolve`: it tells the caller
whether a path is reachable on the filesystem (workspace + extra roots
+ `$GNOROOT/examples` when enabled) without consulting the rpc fetcher
and without mutating the index or tracked sets. Used by pre-flight
diagnostics (e.g. the `-no-examples` import-graph check) that must
neither block startup on the network nor pollute loader state.

### Package kind rule

Every `*Package` carries a `Kind`:

- `Kind = KindUnknown` is the iota=0 zero value. A `Package{}` literal
  constructed without setting `Kind` reads as `KindUnknown` rather than
  silently registering as filesystem-backed; the watcher's `Kind == KindFS`
  filter then excludes it loudly instead of trying to watch a nonexistent
  dir.
- `Kind = KindRemote` iff the package's directory lives under
  `gnomod.ModCachePath()` (RPC-fetched packages) or was constructed from an
  in-memory `MemPackage` (tests, RPC fallback). The watcher skips these —
  modcache dirs are transient and not user-editable.
- `Kind = KindFS` for everything else (workspace, extra roots,
  `$GNOROOT/examples`).

### Root scan caching

`Resolve`'s FS walk scans each root (`ExtraRoots` + `$GNOROOT/examples` if
enabled) at most once per `Loader` lifetime. Any directory whose path
(as emitted by `filepath.WalkDir`) matches an entry in
`Config.ExcludeDirs` is skipped via `fs.SkipDir`. Results are cached in
a per-loader `root → (importPath → dir)` map. Invalidation is coarse:
a new `Loader` is constructed on gnodev restart. Mid-session, the file
watcher's reload already picks up newly added packages via
`gnovm.Load` — the root cache only serves `Resolve` miss lookups.

rootIdx is **not** invalidated on Reload — directories are stable
mid-session, and re-walking large extra roots on every watcher tick is
too expensive. New directories require a gnodev restart. The same
property has a corollary cost: deletion of an extra-root directory
mid-session is not detected; `Resolve` returns the stale dir from the
cached index until restart, and the subsequent `ReadMemPackage` then
fails at use time. Acceptable for a dev tool.

### User-facing changes

| Removed (hard) | Added |
|---|---|
| `-resolver <name>=<loc>` | `-extra-root <dir>` (repeatable) |
| `-lazy-loader` | `-no-examples` |
|  | `-remote-override <domain>=<rpc>` (repeatable) |
|  | `-without-quarantined-examples` (skip `$GNOROOT/examples/quarantined`) |

`-remote-override` closes the migration gap left by removing
`-resolver remote=<rpc>`: it populates `Config.RemoteOverrides`, which
`rpcpkgfetcher.New` consumes when resolving paths outside the workspace.
Workspace packages hit the FS lookup first, so the override only ever
applies to cross-workspace and unresolved imports.

When `gnodev local` runs in a directory without `gnomod.toml`, the
fallback import path is derived from the directory basename. The
basename is sanitized to match gno's `Re_name` rule (lowercase
`[a-z0-9_]`, must start with a letter or `_<letter>`); inputs with no
letters fall back to `app`.

Modes are not exposed. Behavior is derived from filesystem state
(workspace detected via `gnowork.toml` / `gnomod.toml`) plus the three
loader flags (`-no-examples`, `-extra-root`, `-remote-override`).

| CWD state | Flags | Behavior |
|---|---|---|
| In workspace | default | Eager load workspace; examples lazy via proxy |
| In workspace | `-no-examples` | Eager load workspace; no proxy. Workspace imports of `gno.land/*` paths unreachable via FS roots are warned at startup. |
| No workspace | default | Discovery-mode warning in the `Loader` log group; examples lazy via proxy. `gnodev local` registers CWD as an extra-root automatically so loose-realm dirs still resolve. |
| No workspace | `-no-examples`, no `-extra-root` | **Fatal**: "nothing to load". Explicit combination of flags asks gnodev to do nothing. |
| Any | `-extra-root <dir>` (nonexistent) | Warning logged; invalid root skipped |
| Any | `-extra-root <dir>` (valid) | `<dir>` added to the lazy set |
| Any | `-remote-override gno.land=<rpc>` | Cross-workspace fetches for the given domain go to `<rpc>` instead of the default |
| Any | `-without-quarantined-examples` | `$GNOROOT/examples/quarantined` is skipped during root scans and eager load; the rest of `examples/` is still indexed. Default `true` in `gnodev staging`. |

`gnodev staging` eager-loads the workspace, every `-extra-root`, and
`$GNOROOT/examples` (skipping `examples/quarantined/` by default — staging
deploys go to genesis, quarantined realms are excluded from the test-13
genesis safe-list). It does not start the proxy. Internally it calls
`loader.LoadAll()`.

### Error model

Fatal only in two cases:

1. Malformed `gnomod.toml` or `gnowork.toml` inside the workspace — gnovm's
   parse error bubbles up unchanged.
2. The user's flag combination asks gnodev to load nothing at all
   (`-no-examples` + no workspace + no `-extra-root`). Gnodev refuses
   rather than silently running an empty chain.

Everything else is a warning and gnodev proceeds with whatever it managed
to assemble:

- Missing workspace: a multiline Warn in the `Loader` log group; the
  column logger renders it as one bordered block, and routing through the
  logger keeps raw-mode line endings correct (a raw stderr write bypasses
  rawterm's `\n` → `\r\n` translation and staircases). gnodev proceeds.
- Nonexistent `-extra-root`: warn, skip that root.
- `Resolve` miss in the proxy: debug log, skip — normal in lazy mode.
- `rpcpkgfetcher` failure: warn, skip — remote not reachable or path
  absent.
- `-no-examples` + workspace pkg imports unresolvable `gno.land/*` paths:
  one-shot pre-flight via `LookupFS` warns with the missing import paths
  and a hint. Non-fatal — the user may be intentionally stubbing.
- Reload failure after startup: error log; node keeps the previous state
  live so the user can fix and re-save.

The rule: if there is any way gnodev can still serve something useful, it
stays up. Fatal exits are reserved for malformed config that gnovm itself
refuses to accept.

### Consumers

`contribs/gnodev/pkg/dev/node.go` no longer imports `packages.Loader`. Its
`NodeConfig` takes a `Reload func() ([]*Package, error)` closure, called
once on first `Reset()` to produce the initial package set and again on
every watcher-triggered reload. `app.go` wires the closure to
`loader.Reload` (lazy mode) or `loader.LoadAll` (`gnodev staging`).

The node also owns the genesis bootstrap tx: the `r/sys/users/init.Bootstrap`
call (mirroring the chain's genesis whitelist of the user-registration
controller) is injected only when the realm is part of the genesis package
set, re-evaluated on every rebuild so a later lazy load of the realm picks
it up. Genesis txs carry one empty signature slot per signer to pass
validation. Previously the tx was injected unconditionally at the app layer
with nil signatures, so it failed `ValidateBasic` at every startup and could
never succeed in lazy mode anyway.

The proxy (`pkg/proxy/path_interceptor.go`) calls the bound
`loader.Resolve` directly.

The watcher (`pkg/watcher/watch.go`) watches exactly what's currently in
the index — workspace pkgs from startup plus any lazily-resolved pkgs
added by proxy hits.

### Upstream

One small addition to `gnovm/pkg/packages/pkgdownload/`:

```go
type InMemoryFetcher struct { pkgs map[string][]*std.MemFile }
func NewInMemoryFetcher(pkgs ...*std.MemPackage) *InMemoryFetcher
func (f *InMemoryFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error)
```

Follows the existing `NewNoopFetcher` pattern. Replaces gnodev's
`MockLoader` / `resolver_mock.go`. Ships in the same PR.

If `gnovm` does not already expose a public helper for workspace
discovery, this PR adds one so gnodev does not re-implement the walk.

## Alternatives considered

### A. Wrap gnovm's loader behind gnodev's existing `Loader` interface

Rejected. Would keep the `Resolver` chain, middleware, and `BaseLoader`
scaffolding. Doesn't address the single-path lookup gap — the wrapper would
still need a parallel resolver path for the proxy. Achieves code reuse but
not simplification.

### B. Two separate operations, no unified type

A bulk-loader function (calls `gnovm.Load`) plus a standalone per-path
resolver. Rejected because `Package` construction, `Kind` determination
(FS vs Remote via modcache prefix), and path handling would live in both
paths and drift over time. The chosen shape (single `Loader` struct with
both methods) consolidates that shared logic in one place.

### C. Extend gnovm with a public `ResolvePath(conf, path) *Package` API

Moves the per-path logic upstream. Cleanest long-term but adds a new
public gnovm surface area we may not need — only gnodev's proxy needs it.
The local `Resolve` implementation in gnodev is small (~50 lines) and
avoids coordinating an upstream API change.

### D. Earlier WIP branch approach: pre-walk workspaces into an index at startup

Rejected. Breaks true-lazy: a proxy hit on a path not seen during the
pre-walk cannot be resolved until the walk is redone. Makes `lazy` slower
to start and misses the original UX. The chosen approach makes `Resolve`
do its work on demand.

### Modes vs flags

An earlier draft used `-load auto|lazy|full`. Rejected in favor of deriving
behavior from filesystem state + `-no-examples`. Reasoning:

- "Is there a workspace?" is answered by the filesystem, not the user.
- "Do I want examples?" is a real user choice.
- Pure-lazy-including-workspace (today's `-lazy-loader` behavior) has no
  compelling use case — workspaces are small and preloading them is
  cheap.

Removing the mode enum removes a branching config without removing any
real capability.

## Consequences

### Positive

- One loader. No parallel implementation to keep in sync.
- ~1000 lines removed from `contribs/gnodev/pkg/packages/`.
- Simpler user UX: three flags (no modes, no resolver chain), one subcommand.
- Testing surface shrinks: no middleware chain to cover.
- Mock/test fixture support moves upstream where other tools can reuse it.

### Negative / costs

- Hard flag removal breaks scripts using `-resolver` / `-lazy-loader`.
  Migration table is published in the PR.
- `gnovm.Load`'s error messages become user-facing. If they are too terse
  for dev UX, the fix is upstream, not a gnodev wrapper.
- Validation drops (`validateMemPackage`, `isMemPackageEmpty`). gnovm's
  parse errors take over. If stricter validation is wanted later, it
  belongs in `gnolint` or upstream, not in gnodev's load path.
- Reload semantics change subtly: `Reload()` replays the full workspace +
  tracked set through `gnovm.Load` every time. Benchmarked to be
  acceptable for typical dev workspaces; if it becomes a hotspot,
  incremental reload is a future optimization.

### Deferred

- Whether `gnodev staging` should grow a distinct name (`sim`, `genesis`,
  etc.). Keeping `staging` for now; rename if intent diverges.
- A `gnodev packages list` subcommand that dumps the loader's index,
  tracked set, and rootIdx for diagnosing "why isn't my package
  loading?" without grep-walking debug logs.

## References

- PR [#4957](https://github.com/gnolang/gno/pull/4957) — initial migration
- PR [#5604](https://github.com/gnolang/gno/pull/5604) — follow-up
  refinements (`-remote-override` flag, discovery-mode warning,
  `-no-examples` import-graph diagnostic, `LookupFS` FS-only lookup,
  `KindUnknown` zero value, rootIdx Reload preservation, guessPath
  basename sanitization, staging progress logging,
  `-without-quarantined-examples` via `Config.ExcludeDirs`,
  transitive import closure for lazily-loaded packages
  (`Package.Imports` + deps-first expansion in `Reload`),
  discovery-mode warning routed through the column logger,
  conditional genesis `users/init` bootstrap tx with sized
  signature slots)
- `gnovm/pkg/packages/` — native loader
- `contribs/gnodev/pkg/packages/` — gnodev's loader package
- `contribs/gnodev/pkg/proxy/path_interceptor.go` — lazy proxy
