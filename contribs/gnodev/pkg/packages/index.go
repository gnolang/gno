package packages

import (
	"sync"
)

// PathIndex maintains a mapping between import paths and filesystem directories.
// This is needed for lazy loading where package paths are returned but we need
// to resolve them to filesystem directories.
type PathIndex struct {
	mu     sync.RWMutex
	byPath map[string]*Package // ImportPath -> Package
	byDir  map[string]*Package // Dir -> Package
}

func NewPathIndex() *PathIndex {
	return &PathIndex{
		byPath: make(map[string]*Package),
		byDir:  make(map[string]*Package),
	}
}

func (idx *PathIndex) Add(pkg *Package) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byPath[pkg.ImportPath] = pkg
	if pkg.Dir != "" {
		idx.byDir[pkg.Dir] = pkg
	}
}

func (idx *PathIndex) GetByPath(importPath string) (*Package, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	pkg, ok := idx.byPath[importPath]
	return pkg, ok
}

func (idx *PathIndex) GetByDir(dir string) (*Package, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	pkg, ok := idx.byDir[dir]
	return pkg, ok
}

func (idx *PathIndex) List() []*Package {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	pkgs := make([]*Package, 0, len(idx.byPath))
	for _, pkg := range idx.byPath {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func (idx *PathIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byPath = make(map[string]*Package)
	idx.byDir = make(map[string]*Package)
}

func (idx *PathIndex) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.byPath)
}
