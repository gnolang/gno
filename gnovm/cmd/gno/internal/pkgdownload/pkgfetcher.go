package pkgdownload

type PackageFetcher interface {
	FetchPackage(pkgPath string) ([]PackageFile, error)
}

type PackageFile struct {
	Name string
	Body []byte
}
