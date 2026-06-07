// Package issue5736_common is a regression fixture for
// https://github.com/gnolang/gno/issues/5736.
//
// It mimics the gnoswap pattern that surfaced the bug: a non-crossing
// helper in an /r/ realm that performs in-place uint256 arithmetic on
// a value handed in from another /r/ realm. Before the fix to
// {Array,Struct}Value.Copy, this triggered:
//
//	panic: cannot directly modify readonly tainted object
//	  gno.land/p/onbloc/uint256/bitwise.gno (z.arr[0] = z.arr[0] << n)
package issue5736_common

import (
	u256 "gno.land/p/onbloc/uint256"
)

func DoLsh(x *u256.Uint, n uint) *u256.Uint {
	return u256.Zero().Lsh(x, n)
}
