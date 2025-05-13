package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_derivePkgAddr(pkgPath string) string {
	addr := string(gno.DerivePkgBech32Addr(pkgPath))
	return addr
}
