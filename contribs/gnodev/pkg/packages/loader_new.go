package packages

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	vmpackages "github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
)

// LoaderImpl resolves gnodev's package set using gnovm's native loader for
// bulk operations and a local per-path lookup (filesystem + PackageFetcher)
// for the proxy's lazy-resolve path.
// Renamed to Loader in Phase D after the legacy Loader interface is removed.
type LoaderImpl struct {
	cfg Config

	mu      sync.RWMutex
	fetcher pkgdownload.PackageFetcher
	index   map[string]*NewPackage
	tracked map[string]struct{}          // paths added via Resolve, used by Reload
	rootIdx map[string]map[string]string // root → (importPath → dir); populated by Resolve on first lookup against that root
}

func NewLoaderImpl(cfg Config) *LoaderImpl {
	if cfg.GnoRoot == "" {
		cfg.GnoRoot = gnoenv.RootDir()
	}
	fetcher := cfg.Fetcher
	if fetcher == nil {
		fetcher = rpcpkgfetcher.New(cfg.RemoteOverrides)
	}
	return &LoaderImpl{
		cfg:     cfg,
		fetcher: fetcher,
		index:   make(map[string]*NewPackage),
		tracked: make(map[string]struct{}),
		rootIdx: make(map[string]map[string]string),
	}
}

// LoadWorkspace eagerly loads packages in the configured workspace.
// Returns nil (no error) if no workspace is set.
func (l *LoaderImpl) LoadWorkspace() ([]*NewPackage, error) {
	if l.cfg.Workspace == "" {
		return nil, nil
	}
	return l.loadWithPatterns(l.cfg.Workspace + "/...")
}

func (l *LoaderImpl) loadWithPatterns(patterns ...string) ([]*NewPackage, error) {
	l.mu.RLock()
	fetcher := l.fetcher
	l.mu.RUnlock()

	conf := vmpackages.LoadConfig{
		Deps:       true,
		AllowEmpty: true,
		GnoRoot:    l.cfg.GnoRoot,
		Out:        &logWriter{logger: l.cfg.Logger},
		Fetcher:    fetcher,
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

	out := make([]*NewPackage, 0, len(sorted))
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, vp := range sorted {
		if len(vp.Errors) > 0 {
			for _, e := range vp.Errors {
				l.cfg.Logger.Warn("package error", "path", vp.ImportPath, "err", e.Error())
			}
			continue
		}
		p := &NewPackage{
			ImportPath: vp.ImportPath,
			Dir:        vp.Dir,
			Name:       vp.Name,
			Kind:       kindForDir(vp.Dir),
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
func kindForDir(dir string) Kind {
	modCache := gnomod.ModCachePath()
	if modCache == "" {
		return KindFS
	}
	if strings.HasPrefix(filepath.Clean(dir), filepath.Clean(modCache)) {
		return KindRemote
	}
	return KindFS
}
