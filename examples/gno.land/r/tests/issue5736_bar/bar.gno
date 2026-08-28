// Package issue5736_bar is the caller-side fixture for
// https://github.com/gnolang/gno/issues/5736. See the sibling
// gno.land/r/tests/issue5736_common package for context.
package issue5736_bar

import (
	u256 "gno.land/p/onbloc/uint256"
	"gno.land/r/tests/issue5736_common"
)

func Call(cur realm) string {
	out := issue5736_common.DoLsh(u256.NewUint(123), 1)
	return out.Dec()
}
