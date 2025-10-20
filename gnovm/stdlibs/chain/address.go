package chain

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_packageAddress(pkgPath string) string {
	return string(gno.DerivePkgBech32Addr(pkgPath))
}

func X_deriveStorageDepositAddr(pkgPath string) string {
	return string(gno.DeriveStorageDepositBech32Addr(pkgPath))
}
