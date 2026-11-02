package std

import (
	native256 "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// The gnolang package is imported under an alias other than the one the
// generated file uses.
func Echo256(a native256.Uint256, b native256.Int256, c [2]native256.Uint256, d *native256.Int256) (native256.Uint256, native256.Int256) {
	return a, b
}
