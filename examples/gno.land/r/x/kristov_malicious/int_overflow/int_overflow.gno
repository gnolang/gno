// Demonstrate an integer overflow using built-in int type vs arbitrary-precision arithmetic lib

package overflow

import (
	"math/big"
	"strconv"
)

func Render(_ string) string {
	var num int = 9223372036854775807 // Maximum value for int64
	num = num + 1

	bigInt := big.NewInt(9223372036854775807)
	one := big.NewInt(1)
	bigSum := bigInt.Add(bigInt, one)

	return strconv.Itoa(num) + " vs " + bigSum.String()
}
