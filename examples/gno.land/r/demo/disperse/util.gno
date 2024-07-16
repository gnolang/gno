package disperse

import (
	"std"
	"strconv"
	"strings"
)

func parseAddresses(addresses string) ([]std.Address, error) {
	var ret []std.Address

	for _, str := range strings.Split(addresses, ",") {
		addr := std.Address(str)
		if !addr.IsValid() {
			return nil, errInvalidAddress
		}
		ret = append(ret, addr)
	}

	return ret, nil
}

func parseAmounts(amounts string) ([]int64, error) {
	var ret []int64

	for _, amt := range strings.Split(amounts, ",") {
		amount, _ := strconv.Atoi(amt)
		if amount < 0 {
			return nil, errNegativeCoinAmount
		}
		ret = append(ret, int64(amount))
	}

	return ret, nil
}