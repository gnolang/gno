package uint256

import (
	"testing"
)

func TestFuncs(t *testing.T) {
	x := NewUint(0)
	y := NewUint(1)

	// println(x.Dec())

	var m256 bigint = 115792089237316195423570985008687907853269984665640564039457584007913129639935

	z := FromBigint(m256)

	println(z.Dec())
}
