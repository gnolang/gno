package pkgdownload

import "github.com/gnolang/gno/tm2/pkg/std"

type PackageFetcher interface {
	FetchPackage(pkgPath string) ([]*std.MemFile, error)
}
