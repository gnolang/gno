package packages

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// MockLoader is a simple loader for testing that uses in-memory packages.
type MockLoader struct {
	packages map[string]*std.MemPackage
}

// NewMockLoader creates a loader from a list of in-memory packages.
func NewMockLoader(pkgs ...*std.MemPackage) *MockLoader {
	m := &MockLoader{
		packages: make(map[string]*std.MemPackage, len(pkgs)),
	}
	for _, pkg := range pkgs {
		m.packages[pkg.Path] = pkg
	}
	return m
}

func (l *MockLoader) Name() string {
	return "mock"
}

func (l *MockLoader) Load(paths ...string) ([]*Package, error) {
	result := make([]*Package, 0, len(paths))
	for _, path := range paths {
		pkg, err := l.Resolve(path)
		if err != nil {
			continue // Skip packages that don't exist
		}
		result = append(result, pkg)
	}
	return result, nil
}

func (l *MockLoader) Resolve(path string) (*Package, error) {
	mempkg, ok := l.packages[path]
	if !ok {
		return nil, ErrResolverPackageNotFound
	}

	return NewPackageFromMemPackage(mempkg), nil
}
