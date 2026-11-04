package disperse

import (
	"strconv"
	"strings"
	"unicode"
)

func parseAddresses(addresses string) ([]address, error) {
	var ret []address

	for _, str := range strings.Split(addresses, ",") {
		addr := address(str)
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

// parseTokens parses "<amount><symbol>,..." where amount is a decimal uint256.
func parseTokens(tokens string) ([]uint256, []string, error) {
	var amounts []uint256
	var symbols []string

	for _, token := range strings.Split(tokens, ",") {
		amountStr, symbol := splitString(token)
		amount, err := strconv.ParseUint256(amountStr, 10)
		if err != nil {
			return nil, nil, ErrInvalidAmount
		}

		amounts = append(amounts, amount)
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
