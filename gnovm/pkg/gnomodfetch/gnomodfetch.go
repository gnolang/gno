package gnomodfetch

import "fmt"

func Fetch(pkgPath string, dst string) error {
	if FetchFn != nil {
		return FetchFn(pkgPath, dst)
	}

	return fmt.Errorf("not implemented")
}

// FetchFn allows to override default Fetch behavior, this is mostly useful for testing
var FetchFn func(pkgPath string, dst string) error
