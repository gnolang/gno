package disperse

import (
	"std"
	"strconv"
	"strings"
	"unicode"
)

func parseAddresses(addresses string) ([]std.Address, error) {
	var ret []std.Address

	for _, str := range strings.Split(addresses, ",") {
		addr := std.Address(str)
		if !addr.IsValid() {
			return nil, ErrInvalidAddress
		}
		
		ret = append(ret, addr)
	}

	return ret, nil
}

func splitString(input string) (string, string) {
	var pos int
	for i, char := range input {
		if !unicode.IsDigit(char) {
			pos = i
			break
		}
	}
	return input[:pos], input[pos:]
}

func parseTokens(tokens string) ([]uint64, []string, error) {
	var amounts []uint64
	var symbols []string

	for _, token := range strings.Split(tokens, ",") {
		amountStr, symbol := splitString(token)
		amount, _ := strconv.Atoi(amountStr)
		if amount < 0 {
			return nil, nil, ErrNegativeCoinAmount
		}
		amounts = append(amounts, uint64(amount))
		symbols = append(symbols, symbol)
	}

	return amounts, symbols, nil
}

func parseAmounts(amounts string) ([]int64, error) {
	var ret []int64

	for _, amt := range strings.Split(amounts, ",") {
		amount, _ := strconv.Atoi(amt)
		if amount < 0 {
			return nil, ErrNegativeCoinAmount
		}
		ret = append(ret, int64(amount))
	}

	return ret, nil
}
