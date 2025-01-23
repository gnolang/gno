package packages

import (
	"go/token"

	"github.com/gnolang/gno/gnovm"
)

type MockResolver struct {
	pkgs         map[string]*gnovm.MemPackage
	resolveCalls map[string]int // Track resolve calls per path
}

func NewMockResolver(pkgs ...*gnovm.MemPackage) *MockResolver {
	mappkgs := make(map[string]*gnovm.MemPackage, len(pkgs))
	for _, pkg := range pkgs {
		mappkgs[pkg.Path] = pkg
	}
	return &MockResolver{
		pkgs:         mappkgs,
		resolveCalls: make(map[string]int),
	}
}

func (m *MockResolver) ResolveCalls(fset *token.FileSet, path string) int {
	count, _ := m.resolveCalls[path]
	return count
}

func (m *MockResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	m.resolveCalls[path]++ // Increment call count
	if mempkg, ok := m.pkgs[path]; ok {
		return &Package{MemPackage: *mempkg}, nil
	}
	return nil, ErrResolverPackageNotFound
}

func (m *MockResolver) Name() string {
	return "mock"
}
