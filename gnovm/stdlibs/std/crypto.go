package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_derivePkgAddr(pkgPath string) string {
	return string(gno.DerivePkgBech32Addr(pkgPath))
}
