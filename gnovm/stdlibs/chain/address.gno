package chain

import (
	"errors"
	"strings"
)

func packageAddress(string) string

// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func PackageAddress(pkgPath string) address {
	addr := packageAddress(pkgPath)
	return address(addr)
}

// DerivePkgSubAddr derives the address of a sub-realm identity minted
// by cur.Sub(subpath) on the host realm at pkgpath. Pure computation;
// callable from anywhere — lets other realms and off-chain tooling
// compute a sub-address without consulting the host. Enforces the same
// subpath rules as Sub(); the host-colon-free assert also subsumes the
// run-path derivation assert (the run-path shape is colon-free).
func DerivePkgSubAddr(pkgpath, subpath string) address {
	assertValidSubpath(pkgpath, subpath)
	return PackageAddress(pkgpath + ":" + subpath)
}

// SplitPkgSubPath splits a possibly-synthesized pkgpath into host and
// subpath. ok reports whether p is a synthesized sub-realm pkgpath
// (":" is reserved in real package paths and appears only via
// realm.Sub synthesis).
func SplitPkgSubPath(p string) (host, subpath string, ok bool) {
	return strings.Cut(p, ":")
}

// assertValidSubpath mirrors the Sub() native's subpath validation so
// unmintable addresses cannot be derived by accident.
func assertValidSubpath(pkgpath, subpath string) {
	switch {
	case subpath == "":
		panic("Sub: subpath cannot be empty")
	case len(subpath) > 256:
		panic("Sub: subpath too long (max 256 bytes)")
	case strings.Contains(subpath, ":"):
		panic("Sub: subpath cannot contain ':'")
	case strings.Contains(subpath, "\x00"):
		panic("Sub: subpath cannot contain NUL byte")
	case strings.Contains(pkgpath, ":"):
		panic("Sub: receiver pkgpath is already synthesized or invalid")
	}
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
