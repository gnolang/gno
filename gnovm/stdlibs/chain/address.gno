package chain

// packageAddress returns the blockchain address associated with the given realm's
// pkgPath. If pkgPath is the path of a MsgRun ephemeral package, the deployer's
// address is returned instead. It will panic if pkgPath is empty.
//
// The given package path is considered to be a MsgRun path only if:
//
//   - It matches the following regular expression:
//
//     ^(?P<PKGPATH>(?:(?P<DOMAIN>(?:(?P<SLD>(?:(?:(?:[a-z0-9-])+\.)+))(?P<TLD>(?:(?:[a-z]){2,63}))))(?P<URLPATH>(?:/(?P<LETTER>(?:[a-z]))/(?P<USER>(?:(?:(?:_)?[a-z](?:[a-z0-9_])*)))(?:/(?P<REPO>(?:(?:(?:_)?[a-z](?:[a-z0-9_])*)(?:/(?:(?:_)?[a-z](?:[a-z0-9_])*))*)))?))))$
//
//   - The value of the LETTER capturing group from the above regular expression is "e".
//
//   - The value of the REPO capturing group from the above regular expression is "run".
//
//   - The value of the USER capturing group from the above regular expression matches the following regular expression:
//
//     ^(?P<ADDRESS>(?:g1(?:[a-z0-9])+))$
//
// If these rules are matched, the ADDRESS from the USER capturing group is returned.
//
// In all other cases, the address is obtained with the following process:
//
//   - A SHA-256 hash sum is calculated of the string "pkgPath:" concatenated with the given pkgPath.
//   - The resulting hash is trimmed to its first 20 bytes (160 bits).
//   - The resulting 20 bytes are encoded using bech32 encoding.
func packageAddress(pkgPath string) string

// PackageAddress returns the blockchain address associated with the given realm's
// pkgPath. If pkgPath is the path of a MsgRun ephemeral package, the deployer's
// address is returned instead. It will panic if pkgPath is empty.
func PackageAddress(pkgPath string) address {
	addr := packageAddress(pkgPath)
	return address(addr)
}

// deriveStorageDepositAddr returns the blockchain address where tokens are locked
// up for storing data on the realm with the given pkgPath. If pkgPath is the
// path of a MsgRun ephemeral package, the deployer's address is returned
// instead. It will panic if pkgPath is empty.
//
// The logic to determine if a package path comes from MsgRun is thoroughly
// described in [packageAddress].
//
// In all other cases, the address is obtained with the following process:
//
//   - A SHA-256 hash sum is calculated of the string "pkgPath:", concatenated
//     with the given pkgPath, then additionally concatenated with the string ".storageDeposit".
//   - The resulting hash is trimmed to its first 20 bytes (160 bits).
//   - The resulting 20 bytes are encoded using bech32 encoding.
func deriveStorageDepositAddr(pkgPath string) string

// DeriveStorageDepositAddr returns the blockchain address where tokens are locked
// up for storing data on the realm with the given pkgPath. If pkgPath is the
// path of a MsgRun ephemeral package, the deployer's address is returned
// instead. It will panic if pkgPath is empty.
func DeriveStorageDepositAddr(pkgPath string) address {
	addr := deriveStorageDepositAddr(pkgPath)
	return address(addr)
}
