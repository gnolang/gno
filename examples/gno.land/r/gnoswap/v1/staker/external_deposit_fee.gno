package staker

import (
	"std"

	"gno.land/r/gnoswap/v1/common"
)

var depositGnsAmount = uint64(1_000_000_000) // 1_000 GNS

// GetDepositGnsAmount returns the current deposit amount in GNS.
//
// Returns:
// - uint64: The deposit amount in GNS.
func GetDepositGnsAmount() uint64 {
	return depositGnsAmount
}

// SetDepositGnsAmountByAdmin allows an admin to set the deposit amount in GNS.
//
// This function validates the caller as an admin using `common.AdminOnly`.
// If successful, it updates the deposit amount and emits an event with details
// of the change.
//
// Parameters:
// - amount (uint64): The new deposit amount in GNS.
//
// Panics:
// - If the caller is not an admin.
func SetDepositGnsAmountByAdmin(amount uint64) {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}

	prevDepositGnsAmount := getDepositGnsAmount()
	setDepositGnsAmount(amount)

	prevAddr, prevPkgPath := getPrev()

	std.Emit(
		"SetDepositGnsAmountByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevAmount", formatUint(prevDepositGnsAmount),
		"newAmount", formatUint(amount),
	)
}

// SetDepositGnsAmount modifies the deposit gns amount
// Only governance contract can execute this function via proposal
// ref: https://docs.gnoswap.io/contracts/staker/external_deposit_fee.gno
func SetDepositGnsAmount(amount uint64) {
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err.Error())
	}

	prevDepositGnsAmount := getDepositGnsAmount()
	setDepositGnsAmount(amount)

	prevAddr, prevPkgPath := getPrev()

	std.Emit(
		"SetDepositGnsAmount",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevAmount", formatUint(prevDepositGnsAmount),
		"newAmount", formatUint(amount),
	)
}

func setDepositGnsAmount(amount uint64) {
	depositGnsAmount = amount
}

func getDepositGnsAmount() uint64 {
	return depositGnsAmount
}
