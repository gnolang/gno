package packages

import (
	"errors"
	"fmt"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	vmpackages "github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ErrPackageNotFound is returned by Resolve when no index/FS/RPC lookup
// yielded the requested package path.
var ErrPackageNotFound = errors.New("package not found")

// Loader resolves gnodev's package set using gnovm's native loader for
// bulk operations and a local per-path lookup (filesystem + PackageFetcher)
// for the proxy's lazy-resolve path.
type Loader struct {
	cfg            Config
	modCache       string // gnomod.ModCachePath(), resolved once at construction
	modCachePrefix string // modCache + separator, for boundary-safe prefix checks
	wsPattern      string // gnovm.Load pattern for cfg.Workspace, resolved once at construction

	mu      sync.RWMutex
	fetcher pkgdownload.PackageFetcher
	index   map[string]*Package
	tracked map[string]struct{}          // paths added via Resolve, used by Reload
	rootIdx map[string]map[string]string // root → (importPath → dir); populated by Resolve on first lookup against that root
}

func New(cfg Config) *Loader {
	if cfg.GnoRoot == "" {
		cfg.GnoRoot = gnoenv.RootDir()
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	fetcher := cfg.Fetcher
	if fetcher == nil {
		fetcher = rpcpkgfetcher.New(cfg.RemoteOverrides)
	}
	modCache := filepath.Clean(gnomod.ModCachePath())
	return &Loader{
		cfg:            cfg,
		modCache:       modCache,
		modCachePrefix: modCache + string(filepath.Separator),
		wsPattern:      workspacePattern(cfg.Workspace),
		fetcher:        fetcher,
		index:          make(map[string]*Package),
		tracked:        make(map[string]struct{}),
		rootIdx:        make(map[string]map[string]string),
	}
}

// Resolve returns a previously-seen Package if known, else tries FS and RPC
// lookups in order. Hits are memoized in the index and added to tracked.
//
// Locking: fast path is RLock-only. The FS walk runs under the write lock
// (so the per-root index cache is built once). The RPC fetch runs WITHOUT
// the lock held, so a slow rpcLookup for one path does not block Resolve
// for unrelated paths. Two concurrent Resolve calls for the same missing
// path may both hit RPC; the second insert is a no-op via re-check.
func (l *Loader) Resolve(path string) (*Package, error) {
	l.mu.RLock()
	if p, ok := l.index[path]; ok {
		l.mu.RUnlock()
		return p, nil
	}
	l.mu.RUnlock()

	// FS lookup under write lock: ensureRootIndexLocked mutates l.rootIdx.
	l.mu.Lock()
	if p, ok := l.index[path]; ok {
		l.mu.Unlock()
		return p, nil
	}
	if pkg := l.fsLookupLocked(path); pkg != nil {
		l.index[pkg.ImportPath] = pkg
		l.tracked[pkg.ImportPath] = struct{}{}
		l.mu.Unlock()
		return pkg, nil
	}
	l.mu.Unlock()

	// RPC fetch with no lock held: the fetcher field is set once in New and
	// never mutated, and a slow network call must not block unrelated paths.
	pkg := l.rpcLookup(path)
	if pkg == nil {
		return nil, fmt.Errorf("%w: %s", ErrPackageNotFound, path)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	// Re-check: another goroutine may have inserted while we were fetching.
	if existing, ok := l.index[pkg.ImportPath]; ok {
		return existing, nil
	}
	l.index[pkg.ImportPath] = pkg
	l.tracked[pkg.ImportPath] = struct{}{}
	return pkg, nil
}

// LookupFS reports whether path is reachable via the loader's filesystem
// roots (extra roots + GNOROOT/examples when enabled; the workspace is NOT
// consulted — it is covered by the eager load). Walks any root not yet
// cached. Does NOT consult the rpc fetcher and does NOT mutate l.index or
// l.tracked, so it is safe for diagnostic / pre-flight use.
func (l *Loader) LookupFS(path string) bool {
	// The root list is derived from cfg, which is immutable after New.
	// Compute it once to avoid duplicate allocations across the lock dance.
	roots := l.lookupRoots()

	l.mu.RLock()
	for _, root := range roots {
		if rootIdx, ok := l.rootIdx[root]; ok {
			if _, hit := rootIdx[path]; hit {
				l.mu.RUnlock()
				return true
			}
		}
	}
	l.mu.RUnlock()

	// Cold path: ensure each root is walked. Take the write lock once so
	// concurrent callers serialize on the FS walk rather than duplicating it.
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, root := range roots {
		rootIdx := l.ensureRootIndexLocked(root)
		if _, hit := rootIdx[path]; hit {
			return true
		}
	}
	return false
}

// rpcLookup fetches a package via cfg.Fetcher. cfg.Fetcher is set once in
// New and never mutated, so no lock is required.
func (l *Loader) rpcLookup(path string) *Package {
	files, err := l.fetcher.FetchPackage(path)
	if err != nil {
		l.cfg.Logger.Debug("rpc fetch miss", "path", path, "err", err)
		return nil
	}
	mp := &std.MemPackage{
		Path:  path,
		Name:  extractPackageName(files),
		Files: files,
	}
	p := packageFromMemPackage(mp)
	p.Kind = KindRemote
	return p
}

// extractPackageName returns the package name from the first parseable
// non-test .gno file. Returns "" if none is found.
func extractPackageName(files []*std.MemFile) string {
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".gno") {
			continue
		}
		if strings.HasSuffix(f.Name, "_test.gno") || strings.HasSuffix(f.Name, "_filetest.gno") {
			continue
		}
		name, err := gnolang.PackageNameFromFileBody(f.Name, f.Body)
		if err != nil {
			continue
		}
		return string(name)
	}
	return ""
}

// fsLookupLocked assumes the caller holds l.mu (write).
// Uses a per-root cached import-path→dir map so we walk each root at most once.
func (l *Loader) fsLookupLocked(path string) *Package {
	for _, root := range l.lookupRoots() {
		rootIdx := l.ensureRootIndexLocked(root)
		if dir, ok := rootIdx[path]; ok {
			return &Package{
				ImportPath: path,
				Dir:        dir,
				Kind:       l.kindForDir(dir),
			}
		}
	}
	return nil
}

func (l *Loader) lookupRoots() []string {
	roots := make([]string, 0, len(l.cfg.ExtraRoots)+1)
	roots = append(roots, l.cfg.ExtraRoots...)
	if l.cfg.Examples && l.cfg.GnoRoot != "" {
		roots = append(roots, filepath.Join(l.cfg.GnoRoot, "examples"))
	}
	return roots
}

// ensureRootIndexLocked walks root once and caches the result.
// Missing/unreadable roots cache as an empty map to avoid repeated walk attempts.
func (l *Loader) ensureRootIndexLocked(root string) map[string]string {
	if idx, ok := l.rootIdx[root]; ok {
		return idx
	}
	idx := scanRoot(root, l.cfg.ExcludeDirs, l.cfg.Logger)
	l.rootIdx[root] = idx
	return idx
}

// scanRoot walks a root looking for gnomod.toml files and returns a
// module-path → dir map. Skips common noise dirs (dotfiles, node_modules,
// _build) to avoid descending into VCS/build trees, plus any directory
// whose absolute path matches an entry in excludeDirs. Errors from the
// walker or from ParseDir are logged at debug and do not abort the scan.
func scanRoot(root string, excludeDirs []string, logger *slog.Logger) map[string]string {
	excluded := make(map[string]struct{}, len(excludeDirs))
	for _, d := range excludeDirs {
		if d == "" {
			continue
		}
		excluded[filepath.Clean(d)] = struct{}{}
	}
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p == root {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "_build" {
				return fs.SkipDir
			}
			if _, skip := excluded[p]; skip {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() != "gnomod.toml" {
			return nil
		}
		dir := filepath.Dir(p)
		gm, err := gnomod.ParseDir(dir)
		if err != nil {
			// ParseDir stats the file itself; don't re-log the "missing" cases
			// even though we just matched a name — err still possible via i/o.
			if !errors.Is(err, gnomod.ErrNoModFile) && !errors.Is(err, os.ErrNotExist) {
				logger.Debug("skipping unparseable gnomod.toml", "dir", dir, "err", err)
			}
			return nil
		}
		if gm.Module == "" {
			return nil
		}
		out[gm.Module] = dir
		return nil
	})
	if err != nil {
		logger.Warn("root scan failed", "root", root, "err", err)
	}
	if len(out) == 0 {
		logger.Debug("root index empty", "root", root)
	}
	return out
}

// AddLocalPackage registers dir as the source of importPath for a dir that
// has no gnomod.toml — the `gnodev ./scratch-realm` flow, where the module
// path is generated from the directory name. The package is tracked so it
// reaches every reload, and ToMemPackage synthesizes the missing
// gnomod.toml at deploy time.
func (l *Loader) AddLocalPackage(importPath, dir string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.index[importPath] = &Package{
		ImportPath:    importPath,
		Dir:           dir,
		Kind:          KindFS,
		MissingGnoMod: true,
	}
	l.tracked[importPath] = struct{}{}
}

// Track registers paths to re-resolve on every Reload / LoadAll, exactly
// like paths previously seen by Resolve. Paths are not validated here: an
// unresolvable tracked path is warn-logged at reload time. Used for the
// -paths flag and -txs-file dependencies, which must reach genesis even
// though no query or transaction passes through the proxy for them.
func (l *Loader) Track(paths ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, p := range paths {
		if p == "" {
			continue
		}
		l.tracked[p] = struct{}{}
	}
}

// Reload re-runs the eager load for the workspace and every -extra-root;
// see reloadRoots.
func (l *Loader) Reload() ([]*Package, error) {
	return l.reloadRoots(l.cfg.ExtraRoots)
}

// reloadRoots eagerly loads the workspace plus the given roots, then merges
// in each tracked path's transitive closure. Tracked paths discovered via
// the RPC fetcher live outside any FS root, so they are re-resolved
// individually and merged with the eager result.
//
// The index is never evicted: FS-backed entries are content-free handles
// (ToMemPackage re-reads disk on every call), and remote packages are
// session-immutable — re-fetching them on every watcher tick would waste an
// RPC round-trip per file save. rootIdx is likewise preserved: directories
// are stable mid-session; new dirs (or deleted extra-roots) need a gnodev
// restart.
func (l *Loader) reloadRoots(roots []string) ([]*Package, error) {
	l.mu.RLock()
	trackedPaths := make([]string, 0, len(l.tracked))
	for p := range l.tracked {
		trackedPaths = append(trackedPaths, p)
	}
	l.mu.RUnlock()

	out, err := l.loadEager(roots)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(out))
	for _, p := range out {
		seen[p.ImportPath] = struct{}{}
	}

	for _, p := range trackedPaths {
		closure, err := l.resolveClosure(p, seen)
		if err != nil {
			l.cfg.Logger.Warn("reload tracked path failed", "path", p, "err", err)
			continue
		}
		out = append(out, closure...)
	}

	return out, nil
}

// resolveClosure resolves path and its transitive package imports,
// dependency-first, so genesis can deploy them in order. Paths already in
// seen are skipped; resolved paths are added to it. An unresolvable or
// unreadable import is logged and skipped rather than failing the closure:
// the chain reports the precise type-check error at deploy time.
func (l *Loader) resolveClosure(path string, seen map[string]struct{}) ([]*Package, error) {
	if _, ok := seen[path]; ok {
		return nil, nil
	}
	pkg, err := l.Resolve(path)
	if err != nil {
		return nil, err
	}
	// Mark before walking imports so an import cycle terminates.
	seen[pkg.ImportPath] = struct{}{}

	imports, err := pkg.Imports()
	if err != nil {
		l.cfg.Logger.Warn("unable to read package imports", "path", pkg.ImportPath, "err", err)
		return []*Package{pkg}, nil
	}

	out := make([]*Package, 0, len(imports)+1)
	for _, imp := range imports {
		deps, err := l.resolveClosure(imp, seen)
		if err != nil {
			l.cfg.Logger.Warn("unresolvable import", "path", imp, "importer", pkg.ImportPath, "err", err)
			continue
		}
		out = append(out, deps...)
	}
	return append(out, pkg), nil
}

// LoadWorkspace eagerly loads packages in the configured workspace.
// Returns nil (no error) if no workspace is set.
func (l *Loader) LoadWorkspace() ([]*Package, error) {
	if l.cfg.Workspace == "" {
		return nil, nil
	}
	return l.loadWithPatterns(l.wsPattern)
}

// workspacePattern returns the gnovm.Load pattern for a workspace root. A
// gnowork.toml root is a multi-package workspace and loads recursively; a
// gnomod.toml-only root (the `cd myrealm && gnodev` case) is gnovm
// single-package mode, which rejects recursive patterns, so the bare
// directory is passed instead. Resolved once at construction: the marker
// file is part of the session-stable directory layout.
func workspacePattern(workspace string) string {
	if workspace == "" {
		return ""
	}
	if hasFile(workspace, "gnowork.toml") {
		return filepath.Join(workspace, "...")
	}
	return workspace
}

// LoadAll eagerly loads the workspace, every ExtraRoot, GNOROOT/examples
// (when Examples=true), and every tracked path. Used by the staging
// subcommand which wants to materialize every reachable package at startup.
// The returned slice is topologically sorted: dependencies precede
// dependents across all roots so genesis deploy can apply packages in order.
func (l *Loader) LoadAll() ([]*Package, error) {
	return l.reloadRoots(l.lookupRoots())
}

// loadEager runs gnovm.Load against the workspace pattern (implicit:
// l.cfg.Workspace) and walks each root in roots (explicit: callers choose
// what to walk) via loadExtraRootVm, merging the results into a single
// topologically-sorted package list. Used by LoadAll (roots =
// lookupRoots(), includes $GNOROOT/examples) and Reload (roots =
// cfg.ExtraRoots, examples stay lazy via the proxy). Per-step progress is
// logged at Debug; users see it with -v.
func (l *Loader) loadEager(roots []string) ([]*Package, error) {
	var unified vmpackages.PkgList
	seen := map[string]struct{}{}
	appendUnique := func(pl vmpackages.PkgList) {
		for _, p := range pl {
			if _, dup := seen[p.ImportPath]; dup {
				continue
			}
			seen[p.ImportPath] = struct{}{}
			unified = append(unified, p)
		}
	}

	if l.cfg.Workspace != "" {
		l.cfg.Logger.Debug("loading workspace", "workspace", l.cfg.Workspace)
		ws, err := l.loadWithPatternsVm(l.wsPattern)
		if err != nil {
			return nil, err
		}
		l.cfg.Logger.Debug("loaded workspace", "packages", len(ws))
		appendUnique(ws)
	}

	for i, root := range roots {
		l.cfg.Logger.Debug("loading root", "root", root, "n", i+1, "of", len(roots))
		rp := l.loadExtraRootVm(root)
		l.cfg.Logger.Debug("loaded root", "root", root, "packages", len(rp))
		appendUnique(rp)
	}

	// Cross-root, remote, and stdlib deps are not in `unified`; PkgList.Sort
	// errors on missing deps, so trim them out of each pkg's source imports.
	// Safe because workspace deps are already pulled in by vmpackages.Load,
	// and at deploy time every dep we still reference is in `unified`.
	unified = stripStdlibs(unified)
	dropMissingDepImports(unified)

	sorted, err := unified.Sort()
	if err != nil {
		return nil, fmt.Errorf("sort packages: %w", err)
	}
	sorted = sorted.GetNonIgnoredPkgs()
	return l.vmPkgListToPackages(sorted), nil
}

func (l *Loader) loadWithPatterns(patterns ...string) ([]*Package, error) {
	pkgList, err := l.loadWithPatternsVm(patterns...)
	if err != nil {
		return nil, err
	}
	sorted, err := pkgList.Sort()
	if err != nil {
		return nil, fmt.Errorf("sort packages: %w", err)
	}
	sorted = sorted.GetNonIgnoredPkgs()
	return l.vmPkgListToPackages(sorted), nil
}

// loadWithPatternsVm runs vmpackages.Load with Deps:true and returns the raw
// (unsorted) PkgList after stripping stdlibs. Used both by loadWithPatterns
// (which sorts immediately) and by LoadAll (which merges with extra roots
// before a unified sort).
func (l *Loader) loadWithPatternsVm(patterns ...string) (vmpackages.PkgList, error) {
	// l.fetcher and l.cfg are set in New and never mutated; no lock needed.
	conf := vmpackages.LoadConfig{
		Deps:       true,
		AllowEmpty: true,
		GnoRoot:    l.cfg.GnoRoot,
		Out:        &logWriter{logger: l.cfg.Logger},
		Fetcher:    l.fetcher,
		// Dependencies resolve from the same FS roots the lazy path serves;
		// only paths reachable from none of them go through the fetcher.
		// gnovm's dep discovery has no exclude-dir support, so ExcludeDirs
		// does not apply to transitive dependencies.
		ExtraWorkspaceRoots: l.lookupRoots(),
	}
	pkgList, err := vmpackages.Load(conf, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	// Drop stdlib packages and stdlib imports. gnovm.Load returns stdlibs and
	// skips native-stdlib deps during traversal (they're handled by the VM,
	// not deployed as on-chain packages). Without this filter, pkgList.Sort
	// fails on native-stdlib imports like "chain" that are never in the list.
	return stripStdlibs(pkgList), nil
}

// loadExtraRootVm walks one root and returns the packages found there as an
// unsorted PkgList. Each package's imports are parsed so the caller can run
// a unified PkgList.Sort against this list combined with other roots.
// Per-package failures (unreadable mempackage, parse error) are warning-logged
// and skipped; the function never errors.
//
// Note: on first access to a given root, ensureRootIndexLocked walks the
// entire root under the write lock, briefly blocking concurrent Resolve /
// LookupFS calls. Acceptable at startup; the per-package ReadMemPackage and
// Imports work below runs without the lock.
func (l *Loader) loadExtraRootVm(root string) vmpackages.PkgList {
	type entry struct{ path, dir string }
	l.mu.Lock()
	idx := l.ensureRootIndexLocked(root)
	entries := make([]entry, 0, len(idx))
	for p, d := range idx {
		entries = append(entries, entry{p, d})
	}
	l.mu.Unlock()

	fset := token.NewFileSet()
	out := make(vmpackages.PkgList, 0, len(entries))
	var skipped int
	for _, e := range entries {
		// Re-parse gnomod.toml to pick up the Ignore flag; scanRoot only
		// captured the module path. Without this, GetNonIgnoredPkgs lets
		// ignored realms through and they fail at genesis deploy time.
		mod, err := gnomod.ParseDir(e.dir)
		if err != nil {
			l.cfg.Logger.Warn("parse gnomod.toml failed", "path", e.path, "err", err)
			skipped++
			continue
		}
		mp, err := gnolang.ReadMemPackage(e.dir, e.path, gnolang.MPUserAll)
		if err != nil {
			l.cfg.Logger.Warn("read mempackage failed", "path", e.path, "err", err)
			skipped++
			continue
		}
		imps, err := vmpackages.Imports(mp, fset)
		if err != nil {
			l.cfg.Logger.Warn("parse imports failed", "path", e.path, "err", err)
			skipped++
			continue
		}
		out = append(out, &vmpackages.Package{
			Dir:          e.dir,
			ImportPath:   e.path,
			Name:         mp.Name,
			Ignore:       mod.Ignore,
			Imports:      imps.ToStrings(),
			ImportsSpecs: imps,
		})
	}
	if skipped > 0 {
		l.cfg.Logger.Warn("extra-root packages skipped due to load errors", "root", root, "skipped", skipped)
	}
	return out
}

// vmPkgListToPackages converts a sorted vmpackages list into gnodev's Package
// form, registering each entry in the loader index. Packages with load errors
// are warning-logged and skipped.
func (l *Loader) vmPkgListToPackages(sorted vmpackages.SortedPkgList) []*Package {
	out := make([]*Package, 0, len(sorted))
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, vp := range sorted {
		if len(vp.Errors) > 0 {
			for _, e := range vp.Errors {
				l.cfg.Logger.Warn("package error", "path", vp.ImportPath, "err", e.Error())
			}
			continue
		}
		p := &Package{
			ImportPath: vp.ImportPath,
			Dir:        vp.Dir,
			Name:       vp.Name,
			Kind:       l.kindForDir(vp.Dir),
		}
		l.index[p.ImportPath] = p
		out = append(out, p)
	}
	return out
}

// logWriter adapts an slog.Logger to io.Writer for gnovm's Out.
type logWriter struct{ logger *slog.Logger }

func (w *logWriter) Write(p []byte) (int, error) {
	if msg := strings.TrimSpace(string(p)); msg != "" {
		w.logger.Info(msg)
	}
	return len(p), nil
}

// kindForDir classifies a package directory. Packages resolved from the
// modcache are treated as Remote (they won't be watched and aren't part of
// the user's editable workspace). Everything else is FS.
func (l *Loader) kindForDir(dir string) Kind {
	if l.modCache == "" {
		return KindFS
	}
	dir = filepath.Clean(dir)
	if dir == l.modCache || strings.HasPrefix(dir, l.modCachePrefix) {
		return KindRemote
	}
	return KindFS
}

// filterSourceImports MUTATES p: source imports failing keep are dropped
// from BOTH import views. PkgList.Sort errors on imports missing from the
// list and GetNonIgnoredPkgs walks ImportsSpecs, so Imports and ImportsSpecs
// must stay consistent — every import filter goes through here.
func filterSourceImports(p *vmpackages.Package, keep func(path string) bool) {
	if imps := p.Imports[vmpackages.FileKindPackageSource]; len(imps) > 0 {
		kept := imps[:0]
		for _, imp := range imps {
			if keep(imp) {
				kept = append(kept, imp)
			}
		}
		p.Imports[vmpackages.FileKindPackageSource] = kept
	}
	if specs := p.ImportsSpecs[vmpackages.FileKindPackageSource]; len(specs) > 0 {
		kept := specs[:0]
		for _, sp := range specs {
			if keep(sp.PkgPath) {
				kept = append(kept, sp)
			}
		}
		p.ImportsSpecs[vmpackages.FileKindPackageSource] = kept
	}
}

// dropMissingDepImports MUTATES pl: each pkg's source imports lose entries
// whose paths aren't in pl. Used before PkgList.Sort so that cross-root,
// remote, or otherwise-absent deps don't block the toposort.
func dropMissingDepImports(pl vmpackages.PkgList) {
	present := make(map[string]struct{}, len(pl))
	for _, p := range pl {
		present[p.ImportPath] = struct{}{}
	}
	for _, p := range pl {
		filterSourceImports(p, func(imp string) bool {
			_, ok := present[imp]
			return ok
		})
	}
}

// stripStdlibs returns a pkgList with stdlib packages removed and stdlib
// imports filtered out of each remaining package's import views. This
// mirrors the convention used by gno.land/pkg/gnoland/genesis.go (via
// ReadPkgListFromDir): stdlibs are handled natively by the VM, not deployed
// as on-chain packages.
func stripStdlibs(pkgs vmpackages.PkgList) vmpackages.PkgList {
	out := pkgs[:0]
	for _, p := range pkgs {
		if gnolang.IsStdlib(p.ImportPath) {
			continue
		}
		filterSourceImports(p, func(imp string) bool {
			return !gnolang.IsStdlib(imp)
		})
		out = append(out, p)
	}
	return out
}
