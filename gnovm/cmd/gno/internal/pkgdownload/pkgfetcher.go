package pkgdownload

import "github.com/gnolang/gno/gnovm"

type PackageFetcher interface {
	FetchPackage(pkgPath string) ([]*gnovm.MemFile, error)
}
