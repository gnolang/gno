package main

import (
	"fmt"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// getAmountFromEntry
func getAmountFromEntry(entry string) (int64, error) {
	matches := amountRegex.FindStringSubmatch(entry)

	// Check if there is a match
	if len(matches) != 2 {
		return 0, fmt.Errorf(
			"invalid amount, %s",
			entry,
		)
	}

	amount, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount, %s", matches[1])
	}

	return amount, nil
}

// getBalanceFromEntry extracts the account balance information
// from a single line in the form of: <address>=<amount>ugnot
func getBalanceFromEntry(entry string) (*accountBalance, error) {
	matches := balanceRegex.FindStringSubmatch(entry)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid balance encountered, %s", entry)
	}

	// Validate the address
	address, err := crypto.AddressFromString(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid address, %w", err)
	}

	// Validate the amount
	amount, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount, %w", err)
	}

	return &accountBalance{
		address: address,
		amount:  amount,
	}, nil
}
