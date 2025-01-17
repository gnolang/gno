package packages

import (
	"errors"
	"fmt"
	"go/token"
	"strings"
)

var (
	ErrResolverPackageNotFound = errors.New("package not found")
	ErrResolverPackageSkip     = errors.New("package has been skip")
)

type Resolver interface {
	Name() string
	Resolve(fset *token.FileSet, path string) (*Package, error)
}

type NoopResolver struct{}

func (NoopResolver) Name() string { return "" }
func (NoopResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	return nil, ErrResolverPackageNotFound
}

// Chain Resolver

type ChainedResolver []Resolver

func ChainResolvers(rs ...Resolver) Resolver {
	switch len(rs) {
	case 0:
		return &NoopResolver{}
	case 1:
		return rs[0]
	default:
		return ChainedResolver(rs)
	}
}

func (cr ChainedResolver) Name() string {
	var name strings.Builder

	for i, r := range cr {
		rname := r.Name()
		if rname == "" {
			continue
		}

		if i > 0 {
			name.WriteRune('/')
		}
		name.WriteString(rname)
	}

	return name.String()
}

func (cr ChainedResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	for _, resolver := range cr {
		pkg, err := resolver.Resolve(fset, path)
		if err == nil {
			return pkg, nil
		} else if errors.Is(err, ErrResolverPackageNotFound) {
			continue
		}

		return nil, fmt.Errorf("resolver %q error: %w", resolver.Name(), err)
	}

	return nil, ErrResolverPackageNotFound
}
