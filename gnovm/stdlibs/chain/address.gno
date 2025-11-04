package chain

func packageAddress(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func PackageAddress(pkgPath string) address {
	addr := packageAddress(pkgPath)
	return address(addr)
}

func deriveStorageDepositAddr(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func DeriveStorageDepositAddr(pkgPath string) address {
	addr := deriveStorageDepositAddr(pkgPath)
	return address(addr)
}
