package packages

import (
	"go/token"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type MockResolver struct {
	pkgs         map[string]*std.MemPackage
	resolveCalls map[string]int // Track resolve calls per path
	mtx          sync.Mutex
}

func NewMockResolver(pkgs ...*std.MemPackage) *MockResolver {
	mappkgs := make(map[string]*std.MemPackage, len(pkgs))
	for _, pkg := range pkgs {
		mappkgs[pkg.Path] = pkg
	}
	return &MockResolver{
		pkgs:         mappkgs,
		resolveCalls: make(map[string]int),
	}
}

func (m *MockResolver) ResolveCalls(fset *token.FileSet, path string) int {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	count := m.resolveCalls[path]
	return count
}

func (m *MockResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.resolveCalls[path]++ // Increment call count
	if mempkg, ok := m.pkgs[path]; ok {
		return &Package{MemPackage: *mempkg}, nil
	}
	return nil, ErrResolverPackageNotFound
}

func (m *MockResolver) Name() string {
	return "mock"
}
