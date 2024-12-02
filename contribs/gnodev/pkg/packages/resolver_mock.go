package packages

import (
	"go/token"

	"github.com/gnolang/gno/gnovm"
)

type MockResolver struct {
	pkgs map[string]gnovm.MemPackage
}

func NewMockResolver(pkgs ...gnovm.MemPackage) *MockResolver {
	mappkgs := make(map[string]gnovm.MemPackage, len(pkgs))
	for _, pkg := range pkgs {
		mappkgs[pkg.Path] = pkg
	}

	return &MockResolver{
		pkgs: mappkgs,
	}
}

func (m *MockResolver) Name() string {
	return "mock"
}

func (m *MockResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	if mempkg, ok := m.pkgs[path]; ok {
		return &Package{
			MemPackage: mempkg,
			Kind:       PackageKindOther,
			Location:   "",
		}, nil
	}

	return nil, ErrResolverPackageNotFound
}
