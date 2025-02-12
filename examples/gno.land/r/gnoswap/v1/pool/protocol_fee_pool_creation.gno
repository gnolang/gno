package pool

import (
	"std"

	"gno.land/r/gnoswap/v1/common"
)

// poolCreationFee is the fee that is charged when a user creates a pool.
// The fee is denominated in GNS tokens.
var (
	poolCreationFee = uint64(100_000_000) // 100_GNS
)

// GetPoolCreationFee returns the poolCreationFee
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_pool_creation.gno#getpoolcreationfee
func GetPoolCreationFee() uint64 {
	return poolCreationFee
}

// SetPoolCreationFee sets the poolCreationFee.
// Only governance contract can execute this function via proposal
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_pool_creation.gno#setpoolcreationfee
func SetPoolCreationFee(fee uint64) {
	assertOnlyNotHalted()
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err.Error())
	}
	prevPoolCreationFee := poolCreationFee
	setPoolCreationFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetPoolCreationFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatUint(prevPoolCreationFee),
		"newFee", formatUint(fee),
	)
}

// SetPoolCreationFeeByAdmin sets the poolCreationFee by Admin.
// Only admin can execute this function.
func SetPoolCreationFeeByAdmin(fee uint64) {
	assertOnlyNotHalted()

	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}
	prevPoolCreationFee := poolCreationFee
	setPoolCreationFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetPoolCreationFeeByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatUint(prevPoolCreationFee),
		"newFee", formatUint(fee),
	)
}

// setPoolCreationFee this function is internal function called by SetPoolCreationFee
// And SetPoolCreationFeeByAdmin
func setPoolCreationFee(fee uint64) {
	poolCreationFee = fee
}
