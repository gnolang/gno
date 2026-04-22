package packages

import (
	"errors"
	"fmt"
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
	cfg      Config
	modCache string // gnomod.ModCachePath(), resolved once at construction

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
	fetcher := cfg.Fetcher
	if fetcher == nil {
		fetcher = rpcpkgfetcher.New(cfg.RemoteOverrides)
	}
	return &Loader{
		cfg:      cfg,
		modCache: filepath.Clean(gnomod.ModCachePath()),
		fetcher:  fetcher,
		index:    make(map[string]*Package),
		tracked:  make(map[string]struct{}),
		rootIdx:  make(map[string]map[string]string),
	}
}

// Resolve returns a previously-seen Package if known, else tries FS and RPC
// lookups in order. Hits are memoized in the index and added to tracked.
//
// Locking: fast path is RLock-only; cold path takes the write lock for the
// duration of the FS walk and RPC fetch so concurrent Resolve calls for the
// same path serialize rather than duplicate work.
func (l *Loader) Resolve(path string) (*Package, error) {
	l.mu.RLock()
	if p, ok := l.index[path]; ok {
		l.mu.RUnlock()
		return p, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Re-check: another goroutine may have inserted it while we waited.
	if p, ok := l.index[path]; ok {
		return p, nil
	}
	if pkg := l.fsLookupLocked(path); pkg != nil {
		l.index[pkg.ImportPath] = pkg
		l.tracked[pkg.ImportPath] = struct{}{}
		return pkg, nil
	}
	if pkg := l.rpcLookup(path); pkg != nil {
		l.index[pkg.ImportPath] = pkg
		l.tracked[pkg.ImportPath] = struct{}{}
		return pkg, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrPackageNotFound, path)
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
	idx := scanRoot(root, l.cfg.Logger)
	l.rootIdx[root] = idx
	return idx
}

// scanRoot walks a root looking for gnomod.toml files and returns a
// module-path → dir map. Skips common noise dirs (dotfiles, node_modules,
// _build) to avoid descending into VCS/build trees. Errors from the walker
// or from ParseDir are logged at debug and do not abort the scan.
func scanRoot(root string, logger *slog.Logger) map[string]string {
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

// Reload re-runs gnovm.Load for the workspace and re-Resolves each tracked
// path. Tracked paths discovered via ExtraRoots or the fetcher live outside
// the workspace, so they are re-resolved individually and merged with the
// workspace result.
func (l *Loader) Reload() ([]*Package, error) {
	l.mu.RLock()
	wsPattern := ""
	if l.cfg.Workspace != "" {
		wsPattern = l.cfg.Workspace + "/..."
	}
	trackedPaths := make([]string, 0, len(l.tracked))
	for p := range l.tracked {
		trackedPaths = append(trackedPaths, p)
	}
	l.mu.RUnlock()

	var out []*Package
	seen := map[string]struct{}{}
	appendUnique := func(pkgs ...*Package) {
		for _, p := range pkgs {
			if _, dup := seen[p.ImportPath]; dup {
				continue
			}
			seen[p.ImportPath] = struct{}{}
			out = append(out, p)
		}
	}

	if wsPattern != "" {
		pkgs, err := l.loadWithPatterns(wsPattern)
		if err != nil {
			return nil, err
		}
		appendUnique(pkgs...)
	}

	// Drop tracked paths from the index so Resolve re-derives them from
	// FS or fetcher (picking up any changes on disk).
	l.mu.Lock()
	for _, p := range trackedPaths {
		delete(l.index, p)
	}
	// Invalidate root indexes so scanRoot re-walks and picks up added dirs.
	l.rootIdx = make(map[string]map[string]string)
	l.mu.Unlock()

	for _, p := range trackedPaths {
		pkg, err := l.Resolve(p)
		if err != nil {
			l.cfg.Logger.Warn("reload tracked path failed", "path", p, "err", err)
			continue
		}
		appendUnique(pkg)
	}

	return out, nil
}

// LoadWorkspace eagerly loads packages in the configured workspace.
// Returns nil (no error) if no workspace is set.
func (l *Loader) LoadWorkspace() ([]*Package, error) {
	if l.cfg.Workspace == "" {
		return nil, nil
	}
	return l.loadWithPatterns(l.cfg.Workspace + "/...")
}

// LoadAll eagerly loads the workspace, every ExtraRoot, and GNOROOT/examples
// (when Examples=true). Used by the staging subcommand which wants to
// materialize every reachable package at startup.
func (l *Loader) LoadAll() ([]*Package, error) {
	var out []*Package
	seen := map[string]struct{}{}
	appendUnique := func(pkgs []*Package) {
		for _, p := range pkgs {
			if _, dup := seen[p.ImportPath]; dup {
				continue
			}
			seen[p.ImportPath] = struct{}{}
			out = append(out, p)
		}
	}

	if l.cfg.Workspace != "" {
		pkgs, err := l.loadWithPatterns(l.cfg.Workspace + "/...")
		if err != nil {
			return nil, err
		}
		appendUnique(pkgs)
	}

	extraRoots := append([]string(nil), l.cfg.ExtraRoots...)
	if l.cfg.Examples && l.cfg.GnoRoot != "" {
		extraRoots = append(extraRoots, filepath.Join(l.cfg.GnoRoot, "examples"))
	}
	for _, root := range extraRoots {
		pkgs, err := l.loadRootStandalone(root)
		if err != nil {
			l.cfg.Logger.Warn("load extra root failed", "root", root, "err", err)
			continue
		}
		appendUnique(pkgs)
	}

	return out, nil
}

// loadRootStandalone loads every gnomod.toml-rooted package found under root
// by resolving each import path individually. Avoids gnovm.Load's "pattern
// must be inside workspace" check for roots outside the current workspace.
func (l *Loader) loadRootStandalone(root string) ([]*Package, error) {
	l.mu.Lock()
	idx := l.ensureRootIndexLocked(root)
	paths := make([]string, 0, len(idx))
	for p := range idx {
		paths = append(paths, p)
	}
	l.mu.Unlock()

	out := make([]*Package, 0, len(paths))
	for _, p := range paths {
		pkg, err := l.Resolve(p)
		if err != nil {
			l.cfg.Logger.Warn("resolve failed", "path", p, "err", err)
			continue
		}
		out = append(out, pkg)
	}
	return out, nil
}

func (l *Loader) loadWithPatterns(patterns ...string) ([]*Package, error) {
	// l.fetcher and l.cfg are set in New and never mutated; no lock needed.
	conf := vmpackages.LoadConfig{
		Deps:                true,
		AllowEmpty:          true,
		GnoRoot:             l.cfg.GnoRoot,
		Out:                 &logWriter{logger: l.cfg.Logger},
		Fetcher:             l.fetcher,
		ExtraWorkspaceRoots: l.cfg.ExtraRoots,
	}
	pkgList, err := vmpackages.Load(conf, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	sorted, err := pkgList.Sort()
	if err != nil {
		return nil, fmt.Errorf("sort packages: %w", err)
	}
	sorted = sorted.GetNonIgnoredPkgs()

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
	return out, nil
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
	if strings.HasPrefix(filepath.Clean(dir), l.modCache) {
		return KindRemote
	}
	return KindFS
}
