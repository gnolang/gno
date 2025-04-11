package pkgdownload

import (
	"errors"

	"github.com/gnolang/gno/gnovm"
)

type PackageFetcher interface {
	FetchPackage(pkgPath string) ([]*gnovm.MemFile, error)
}

func NewNoopFetcher() PackageFetcher {
	return &noopFetcher{}
}

type noopFetcher struct{}

func (nf *noopFetcher) FetchPackage(pkgPath string) ([]*gnovm.MemFile, error) {
	return nil, errors.New("noop")
}
