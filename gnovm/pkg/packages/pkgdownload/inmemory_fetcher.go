package pkgdownload

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// InMemoryFetcher is a PackageFetcher backed by a map of pre-registered
// MemPackages. Primarily intended for tests where network/filesystem
// access is undesirable.
//
// InMemoryFetcher is read-only after construction. Register all packages
// when calling NewInMemoryFetcher; concurrent FetchPackage calls are safe.
// There is intentionally no public API to mutate the map after creation.
type InMemoryFetcher struct {
	pkgs map[string][]*std.MemFile
}

var _ PackageFetcher = (*InMemoryFetcher)(nil)

// NewInMemoryFetcher registers the given MemPackages by their Path.
// If two MemPackages share the same Path, the later one wins.
func NewInMemoryFetcher(pkgs ...*std.MemPackage) *InMemoryFetcher {
	m := make(map[string][]*std.MemFile, len(pkgs))
	for _, p := range pkgs {
		m[p.Path] = p.Files
	}
	return &InMemoryFetcher{pkgs: m}
}

// FetchPackage implements [PackageFetcher].
func (f *InMemoryFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	files, ok := f.pkgs[pkgPath]
	if !ok {
		return nil, fmt.Errorf("in-memory fetcher: package %q not found", pkgPath)
	}
	return files, nil
}
