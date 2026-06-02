package packages

import (
	"log/slog"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
)

// Config configures the Loader.
type Config struct {
	// Workspace is the workspace root (dir containing gnowork.toml or gnomod.toml).
	// Empty if no workspace was detected.
	Workspace string

	// Examples, when true, includes $GNOROOT/examples in the lazy-loadable set.
	Examples bool

	// ExtraRoots are additional workspace roots supplied by the user.
	// Each must be an existing directory; invalid entries are skipped with a warning.
	ExtraRoots []string

	// ExcludeDirs is a set of directory paths to skip during scanRoot's
	// FS walk. Each entry, after filepath.Clean, is compared exactly
	// against directory paths emitted by the walker. The walker emits
	// paths in the form of the walked root (absolute when the root is
	// absolute), so entries should match that form. Empty entries are
	// ignored; entries that don't match any walked directory are no-ops.
	ExcludeDirs []string

	// GnoRoot is the installed gno root; defaults to gnoenv.RootDir().
	GnoRoot string

	// RemoteOverrides maps a chain domain (e.g. "gno.land") to an RPC URL for rpcpkgfetcher.
	// Ignored when Fetcher is non-nil.
	RemoteOverrides map[string]string

	// Fetcher overrides the default rpcpkgfetcher. Primarily for tests that
	// use InMemoryFetcher. Leave nil in production.
	Fetcher pkgdownload.PackageFetcher

	// Logger is the slog logger used for all loader output.
	// Defaults to slog.Default() if nil.
	Logger *slog.Logger
}
