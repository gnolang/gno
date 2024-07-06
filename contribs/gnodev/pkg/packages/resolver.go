package packages

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/gnovm"
)

var ErrResolverPackageNotFound = errors.New("package not found")

type PackageKind int

const (
	PackageKindOther = iota
	PackageKindFS
)

type Package struct {
	gnovm.MemPackage
	Kind     PackageKind
	Location string
}

type Resolver interface {
	Resolve(path string) (*Package, error)
}

type ChainedResolver []Resolver

func (cr ChainedResolver) Resolve(path string) (*Package, error) {
	for _, resolver := range cr {
		pkg, err := resolver.Resolve(path)
		if err == nil {
			return pkg, nil
		} else if errors.Is(err, ErrResolverPackageNotFound) {
			continue
		}

		return nil, fmt.Errorf("unable to resolve %q: %w", path, err)
	}

	return nil, ErrResolverPackageNotFound
}
