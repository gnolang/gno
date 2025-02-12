package pool

import (
	"gno.land/p/demo/ufmt"
	"std"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

// Mint adds liquidity to a pool by minting a new position.
//
// This function mints a liquidity position within the specified tick range in a pool.
// It verifies caller permissions, validates inputs, and updates the pool's state. Additionally,
// it transfers the required amounts of token0 and token1 from the caller to the pool.
//
// Parameters:
//   - token0Path: string, the path or identifier for token 0.
//   - token1Path: string, the path or identifier for token 1.
//   - fee: uint32, the fee tier of the pool.
//   - recipient: std.Address, the address to receive the newly created position.
//   - tickLower: int32, the lower tick boundary of the position.
//   - tickUpper: int32, the upper tick boundary of the position.
//   - liquidityAmount: string, the amount of liquidity to add, provided as a decimal string.
//   - positionCaller: std.Address, the address of the entity calling the function (e.g., the position owner).
//
// Returns:
//   - string: The amount of token 0 transferred to the pool as a string.
//   - string: The amount of token 1 transferred to the pool as a string.
//
// Panic Conditions:
//   - The system is halted (`common.IsHalted()`).
//   - Caller lacks permission to mint a position when `common.GetLimitCaller()` is enforced.
//   - The provided `liquidityAmount` is zero.
//   - Any failure during token transfers or position modifications.
//
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#mint
func Mint(
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	tickLower int32,
	tickUpper int32,
	liquidityAmount string,
	positionCaller std.Address,
) (string, string) {
	assertOnlyNotHalted()
	if common.GetLimitCaller() {
		assertOnlyPositionContract()
	}

	liquidity := u256.MustFromDecimal(liquidityAmount)
	if liquidity.IsZero() {
		panic(errZeroLiquidity)
	}

	pool := GetPool(token0Path, token1Path, fee)
	liquidityDelta := safeConvertToInt128(liquidity)
	position := newModifyPositionParams(recipient, tickLower, tickUpper, liquidityDelta)
	_, amount0, amount1 := pool.modifyPosition(position)

	if amount0.Gt(u256.Zero()) {
		pool.safeTransferFrom(positionCaller, consts.POOL_ADDR, pool.token0Path, amount0, true)
	}

	if amount1.Gt(u256.Zero()) {
		pool.safeTransferFrom(positionCaller, consts.POOL_ADDR, pool.token1Path, amount1, false)
	}

	return amount0.ToString(), amount1.ToString()
}

// Burn removes liquidity from a position in the pool.
//
// This function allows the caller to burn (remove) a specified amount of liquidity from a position.
// It calculates the amounts of token0 and token1 released when liquidity is removed and updates
// the position's owed token amounts. The actual transfer of tokens back to the caller happens
// during a separate `Collect()` operation.
//
// Parameters:
//   - token0Path: string, the path or identifier for token 0.
//   - token1Path: string, the path or identifier for token 1.
//   - fee: uint32, the fee tier of the pool.
//   - tickLower: int32, the lower tick boundary of the position.
//   - tickUpper: int32, the upper tick boundary of the position.
//   - liquidityAmount: string, the amount of liquidity to remove, provided as a decimal string (uint128).
//
// Returns:
//   - string: The amount of token0 owed after removing liquidity, as a string.
//   - string: The amount of token1 owed after removing liquidity, as a string.
//
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#burn
func Burn(
	token0Path string,
	token1Path string,
	fee uint32,
	tickLower int32,
	tickUpper int32,
	liquidityAmount string, // uint128
) (string, string) { // uint256 x2
	assertOnlyNotHalted()
	if common.GetLimitCaller() {
		assertOnlyPositionContract()
	}
	pool := GetPool(token0Path, token1Path, fee)

	caller := getPrevAddr()
	liqAmount := u256.MustFromDecimal(liquidityAmount)
	liqAmountInt256 := safeConvertToInt128(liqAmount)
	liqDelta := i256.Zero().Neg(liqAmountInt256)
	posParams := newModifyPositionParams(caller, tickLower, tickUpper, liqDelta)
	position, amount0, amount1 := pool.modifyPosition(posParams)

	if amount0.Gt(u256.Zero()) || amount1.Gt(u256.Zero()) {
		amount0 = toUint128(amount0)
		amount1 = toUint128(amount1)
		position.tokensOwed0 = new(u256.Uint).Add(position.tokensOwed0, amount0)
		position.tokensOwed1 = new(u256.Uint).Add(position.tokensOwed1, amount1)
	}

	positionKey := getPositionKey(caller, tickLower, tickUpper)
	pool.setPosition(positionKey, position)

	// actual token transfer happens in Collect()
	return amount0.ToString(), amount1.ToString()
}

// Collect handles the collection of tokens (token0 and token1) from a liquidity position.
//
// This function allows the caller to collect a specified amount of tokens owed to a position
// in a liquidity pool. It calculates the collectible amount based on three constraints:
// the requested amount, the tokens owed to the position, and the pool's available balance.
// The collected tokens are transferred to the specified recipient.
//
// Parameters:
//   - token0Path: string, the path or identifier for token 0.
//   - token1Path: string, the path or identifier for token 1.
//   - fee: uint32, the fee tier of the pool.
//   - recipient: std.Address, the address to receive the collected tokens.
//   - tickLower: int32, the lower tick boundary of the position.
//   - tickUpper: int32, the upper tick boundary of the position.
//   - amount0Requested: string, the requested amount of token 0 to collect (decimal string).
//   - amount1Requested: string, the requested amount of token 1 to collect (decimal string).
//
// Returns:
//   - string: The actual amount of token 0 collected, as a string.
//   - string: The actual amount of token 1 collected, as a string.
//
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#collect
func Collect(
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	tickLower int32,
	tickUpper int32,
	amount0Requested string,
	amount1Requested string,
) (string, string) {
	assertOnlyNotHalted()
	if common.GetLimitCaller() {
		assertOnlyPositionContract()
	}

	pool := GetPool(token0Path, token1Path, fee)
	positionKey := getPositionKey(getPrevAddr(), tickLower, tickUpper)
	position := pool.mustGetPosition(positionKey)

	var amount0, amount1 *u256.Uint

	// Smallest of three: amount0Requested, position.tokensOwed0, pool.balances.token0
	amount0Req := u256.MustFromDecimal(amount0Requested)
	amount0 = collectToken(amount0Req, position.tokensOwed0, pool.BalanceToken0())
	amount1Req := u256.MustFromDecimal(amount1Requested)
	amount1 = collectToken(amount1Req, position.tokensOwed1, pool.BalanceToken1())

	if amount0.Gt(u256.Zero()) {
		position.tokensOwed0 = new(u256.Uint).Sub(position.tokensOwed0, amount0)
		pool.balances.token0 = new(u256.Uint).Sub(pool.balances.token0, amount0)
		token0 := common.GetTokenTeller(pool.token0Path)
		checkTransferError(token0.Transfer(recipient, amount0.Uint64()))
	}
	if amount1.Gt(u256.Zero()) {
		position.tokensOwed1 = new(u256.Uint).Sub(position.tokensOwed1, amount1)
		pool.balances.token1 = new(u256.Uint).Sub(pool.balances.token1, amount1)
		token1 := common.GetTokenTeller(pool.token1Path)
		checkTransferError(token1.Transfer(recipient, amount1.Uint64()))
	}

	pool.setPosition(positionKey, position)

	return amount0.ToString(), amount1.ToString()
}

// collectToken calculates the actual amount of tokens that can be collected.
//
// This function determines the smallest possible value among the requested amount (`amountReq`),
// the tokens owed (`tokensOwed`), and the pool's available balance (`poolBalance`). It ensures
// the collected amount does not exceed any of these constraints.
//
// Parameters:
//   - amountReq: *u256.Uint, the amount of tokens requested for collection.
//   - tokensOwed: *u256.Uint, the total amount of tokens owed to the position.
//   - poolBalance: *u256.Uint, the current balance of tokens available in the pool.
//
// Returns:
//   - amount: *u256.Uint, the actual amount that can be collected (minimum of the three inputs).
func collectToken(
	amountReq, tokensOwed, poolBalance *u256.Uint,
) (amount *u256.Uint) {
	// find smallest of three amounts
	amount = u256Min(amountReq, tokensOwed)
	amount = u256Min(amount, poolBalance)
	return amount.Clone()
}

// SetFeeProtocolByAdmin sets the fee protocol for all pools
// Also it will be applied to new created pools
func SetFeeProtocolByAdmin(
	feeProtocol0 uint8,
	feeProtocol1 uint8,
) {
	assertOnlyAdmin()
	setFeeProtocolInternal(feeProtocol0, feeProtocol1, "SetFeeProtocolByAdmin")
}

// SetFeeProtocol sets the fee protocol for all pools
// Only governance contract can execute this function via proposal
// Also it will be applied to new created pools
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#setfeeprotocol
func SetFeeProtocol(feeProtocol0, feeProtocol1 uint8) {
	assertOnlyGovernance()
	setFeeProtocolInternal(feeProtocol0, feeProtocol1, "SetFeeProtocol")
}

// setFeeProtocolInternal updates the protocol fee for all pools and emits an event.
//
// This function is an internal utility used to set the protocol fee for token0 and token1 in a compact
// format. The fee values are stored as a single `uint8` byte where:
//   - Lower 4 bits represent the fee for token0 (feeProtocol0).
//   - Upper 4 bits represent the fee for token1 (feeProtocol1).
//
// It also emits an event to log the changes, including the previous and new fee protocol values.
//
// Parameters:
//   - feeProtocol0: uint8, protocol fee for token0 (must be 0 or between 4 and 10 inclusive).
//   - feeProtocol1: uint8, protocol fee for token1 (must be 0 or between 4 and 10 inclusive).
//   - eventName: string, the name of the event to emit (e.g., "SetFeeProtocolByAdmin").
//
// Notes:
//   - This function is called by higher-level functions like `SetFeeProtocolByAdmin` or `SetFeeProtocol`.
//   - It does not validate caller permissions; validation must be performed by the calling function.
func setFeeProtocolInternal(feeProtocol0, feeProtocol1 uint8, eventName string) {
	oldFee := slot0FeeProtocol
	newFee := setFeeProtocol(feeProtocol0, feeProtocol1)

	feeProtocol0Old := oldFee % 16
	feeProtocol1Old := oldFee >> 4

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		eventName,
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFeeProtocol0", formatUint(uint64(feeProtocol0Old)),
		"prevFeeProtocol1", formatUint(uint64(feeProtocol1Old)),
		"feeProtocol0", formatUint(uint64(feeProtocol0)),
		"feeProtocol1", formatUint(uint64(feeProtocol1)),
		"newFee", formatUint(uint64(newFee)),
	)
}

// setFeeProtocol updates the protocol fee configuration for all managed pools.
//
// This function combines the protocol fee values for token0 and token1 into a single `uint8` value,
// where:
//   - Lower 4 bits store feeProtocol0 (for token0).
//   - Upper 4 bits store feeProtocol1 (for token1).
//
// The updated fee protocol is applied uniformly to all pools managed by the system.
//
// Parameters:
//   - feeProtocol0: protocol fee for token0 (must be 0 or between 4 and 10 inclusive).
//   - feeProtocol1: protocol fee for token1 (must be 0 or between 4 and 10 inclusive).
//
// Returns:
//   - newFee (uint8): the combined fee protocol value.
//
// Example:
// If feeProtocol0 = 4 and feeProtocol1 = 5:
//
//	newFee = 4 + (5 << 4)
//	// Results in: 0x54 (84 in decimal)
//	// Binary: 0101 0100
//	//         ^^^^ ^^^^
//	//       fee1=5  fee0=4
//
// Notes:
//   - This function ensures that all pools under management are updated to use the same fee protocol.
//   - Caller restrictions (e.g., admin or governance) are not enforced in this function.
//   - Ensure the system is not halted before updating fees.
func setFeeProtocol(feeProtocol0, feeProtocol1 uint8) uint8 {
	assertOnlyNotHalted()
	if err := validateFeeProtocol(feeProtocol0, feeProtocol1); err != nil {
		panic(addDetailToError(
			err,
			ufmt.Sprintf("expected (feeProtocol0(%d) == 0 || (feeProtocol0(%d) >= 4 && feeProtocol0(%d) <= 10)) && (feeProtocol1(%d) == 0 || (feeProtocol1(%d) >= 4 && feeProtocol1(%d) <= 10))", feeProtocol0, feeProtocol0, feeProtocol0, feeProtocol1, feeProtocol1, feeProtocol1),
		))
	}
	// combine both protocol fee into a single byte:
	// - feePrtocol0 occupies the lower 4 bits
	// - feeProtocol1 is shifted the lower 4 positions to occupy the upper 4 bits
	newFee := feeProtocol0 + (feeProtocol1 << 4) // ( << 4 ) = ( * 16 )

	// Update slot0 for each pool
	pools.Iterate("", "", func(poolPath string, iPool interface{}) bool {
		pool := iPool.(*Pool)
		pool.slot0.feeProtocol = newFee

		return false
	})

	// update slot0
	slot0FeeProtocol = newFee
	return newFee
}

// validateFeeProtocol validates the fee protocol values for token0 and token1.
//
// This function checks whether the provided fee protocol values (`feeProtocol0` and `feeProtocol1`)
// are valid using the `isValidFeeProtocolValue` function. If either value is invalid, it returns
// an error indicating that the protocol fee percentage is invalid.
//
// Parameters:
//   - feeProtocol0: uint8, the fee protocol value for token0.
//   - feeProtocol1: uint8, the fee protocol value for token1.
//
// Returns:
//   - error: Returns `errInvalidProtocolFeePct` if either `feeProtocol0` or `feeProtocol1` is invalid.
//     Returns `nil` if both values are valid.
func validateFeeProtocol(feeProtocol0, feeProtocol1 uint8) error {
	if !isValidFeeProtocolValue(feeProtocol0) || !isValidFeeProtocolValue(feeProtocol1) {
		return errInvalidProtocolFeePct
	}
	return nil
}

// isValidFeeProtocolValue checks if a fee protocol value is within acceptable range.
// valid values are either 0 or between 4 and 10 inclusive.
func isValidFeeProtocolValue(value uint8) bool {
	return value == 0 || (value >= 4 && value <= 10)
}

// CollectProtocolByAdmin collects protocol fees for the given pool that accumulated while it was being used for swap
// Returns collected amount0, amount1 in string
func CollectProtocolByAdmin(
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	amount0Requested string,
	amount1Requested string,
) (string, string) {
	assertOnlyAdmin()
	amount0, amount1 := collectProtocol(
		token0Path,
		token1Path,
		fee,
		recipient,
		amount0Requested,
		amount1Requested,
	)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"CollectProtocolByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"token0Path", token0Path,
		"token1Path", token1Path,
		"fee", formatUint(fee),
		"recipient", recipient.String(),
		"collectedAmount0", amount0,
		"collectedAmount1", amount1,
	)

	return amount0, amount1
}

// CollectProtocol collects protocol fees for the given pool that accumulated while it was being used for swap
// Only governance contract can execute this function via proposal
// Returns collected amount0, amount1 in string
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#collectprotocol
func CollectProtocol(
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	amount0Requested string, // uint128
	amount1Requested string, // uint128
) (string, string) { // uint128 x2
	assertOnlyGovernance()
	amount0, amount1 := collectProtocol(
		token0Path,
		token1Path,
		fee,
		recipient,
		amount0Requested,
		amount1Requested,
	)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"CollectProtocol",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"token0Path", token0Path,
		"token1Path", token1Path,
		"fee", formatUint(fee),
		"recipient", recipient.String(),
		"internal_amount0", amount0,
		"internal_amount1", amount1,
	)

	return amount0, amount1
}

// collectProtocol collects protocol fees for token0 and token1 from the specified pool.
//
// This function allows the collection of accumulated protocol fees for token0 and token1. It ensures
// the requested amounts do not exceed the available protocol fees in the pool and transfers the
// collected amounts to the specified recipient.
//
// Parameters:
//   - token0Path: string, the path or identifier for token0.
//   - token1Path: string, the path or identifier for token1.
//   - fee: uint32, the fee tier of the pool.
//   - recipient: std.Address, the address to receive the collected protocol fees.
//   - amount0Requested: string, the requested amount of token0 to collect (decimal string).
//   - amount1Requested: string, the requested amount of token1 to collect (decimal string).
//
// Returns:
//   - string: The actual amount of token0 collected, as a string.
//   - string: The actual amount of token1 collected, as a string.
func collectProtocol(
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	amount0Requested string,
	amount1Requested string,
) (string, string) {
	assertOnlyRegistered(token0Path)
	assertOnlyRegistered(token1Path)
	assertOnlyNotHalted()

	pool := GetPool(token0Path, token1Path, fee)

	amount0Req := u256.MustFromDecimal(amount0Requested)
	amount1Req := u256.MustFromDecimal(amount1Requested)

	amount0 := u256Min(amount0Req, pool.ProtocolFeesToken0())
	amount1 := u256Min(amount1Req, pool.ProtocolFeesToken1())

	amount0, amount1 = pool.saveProtocolFees(amount0.Clone(), amount1.Clone())
	uAmount0 := amount0.Uint64()
	uAmount1 := amount1.Uint64()

	token0Teller := common.GetTokenTeller(pool.token0Path)
	checkTransferError(token0Teller.Transfer(recipient, uAmount0))
	newBalanceToken0, err := updatePoolBalance(pool.BalanceToken0(), pool.BalanceToken1(), amount0, true)
	if err != nil {
		panic(err)
	}
	pool.balances.token0 = newBalanceToken0

	token1Teller := common.GetTokenTeller(pool.token1Path)
	checkTransferError(token1Teller.Transfer(recipient, uAmount1))
	newBalanceToken1, err := updatePoolBalance(pool.BalanceToken0(), pool.BalanceToken1(), amount1, false)
	if err != nil {
		panic(err)
	}
	pool.balances.token1 = newBalanceToken1

	return amount0.ToString(), amount1.ToString()
}

// saveProtocolFees updates the protocol fee balances after collection.
//
// Parameters:
//   - amount0: amount of token0 fees to collect
//   - amount1: amount of token1 fees to collect
//
// Returns the adjusted amounts that will actually be collected for both tokens.
func (p *Pool) saveProtocolFees(amount0, amount1 *u256.Uint) (*u256.Uint, *u256.Uint) {
	cond01 := amount0.Gt(u256.Zero())
	cond02 := amount0.Eq(p.ProtocolFeesToken0())
	if cond01 && cond02 {
		amount0 = new(u256.Uint).Sub(amount0, u256.One())
	}

	cond11 := amount1.Gt(u256.Zero())
	cond12 := amount1.Eq(p.ProtocolFeesToken1())
	if cond11 && cond12 {
		amount1 = new(u256.Uint).Sub(amount1, u256.One())
	}

	p.protocolFees.token0 = new(u256.Uint).Sub(p.ProtocolFeesToken0(), amount0)
	p.protocolFees.token1 = new(u256.Uint).Sub(p.ProtocolFeesToken1(), amount1)

	// return rest fee
	return amount0, amount1
}
