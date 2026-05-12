package chain

import "errors"

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

func pubKeyAddress(string) (addr string, errStr string)

// PubKeyAddress derives the bech32 address from a bech32-encoded public key.
// Returns an error if the input is malformed or amino-decodes to a nil
// concrete pubkey (e.g., empty-payload "gpub1mdgqmw").
//
// This is the inverse half of the on-chain valset entry format
// "<bech32-pubkey>:<power>": split the entry, then call this helper to
// recover the address from the pubkey component.
func PubKeyAddress(bech32PubKey string) (address, error) {
	addr, errStr := pubKeyAddress(bech32PubKey)
	if errStr != "" {
		return "", errors.New(errStr)
	}
	return address(addr), nil
}
