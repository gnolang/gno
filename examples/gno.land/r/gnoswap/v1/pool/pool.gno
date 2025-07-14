package pool

import (
	"std"

	"gno.land/r/gnoswap/v1/common"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/access"
)

var (
	positionAddr, _    = access.GetAddress(access.ROLE_POSITION)
	poolAddr, _        = access.GetAddress(access.ROLE_POOL)
	protocolFeeAddr, _ = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	routerAddr, _      = access.GetAddress(access.ROLE_ROUTER)
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
//   - The system is halted (`halt.IsHalted()`).
//   - Caller lacks permission to mint a position when `common.GetLimitCaller()` is enforced.
//   - The provided `liquidityAmount` is zero.
//   - Any failure during token transfers or position modifications.
//
// ref: https://docs.gnoswap.io/contracts/pool/pool.gno#mint
func Mint(cur realm,
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
	assertOnlyValidAddress(positionCaller)

	liquidity := u256.MustFromDecimal(liquidityAmount)
	if liquidity.IsZero() {
		panic(errZeroLiquidity)
	}

	pool := GetPool(token0Path, token1Path, fee)
	liquidityDelta := safeConvertToInt128(liquidity)
	positionParam := newModifyPositionParams(positionCaller, tickLower, tickUpper, liquidityDelta)
	_, amount0, amount1 := pool.modifyPosition(positionParam)

	if amount0.Gt(u256.Zero()) {
		pool.safeTransferFrom(positionCaller, poolAddr, pool.token0Path, amount0, true)
	}

	if amount1.Gt(u256.Zero()) {
		pool.safeTransferFrom(positionCaller, poolAddr, pool.token1Path, amount1, false)
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
	cur realm,
	token0Path string,
	token1Path string,
	fee uint32,
	tickLower int32,
	tickUpper int32,
	liquidityAmount string, // uint128
	positionCaller std.Address,
) (string, string) { // uint256 x2
	assertOnlyNotHalted()
	if common.GetLimitCaller() {
		assertOnlyPositionContract()
	}

	pool := GetPool(token0Path, token1Path, fee)

	liqAmount := u256.MustFromDecimal(liquidityAmount)
	liqAmountInt256 := safeConvertToInt128(liqAmount)
	liqDelta := i256.Zero().Neg(liqAmountInt256)
	posParams := newModifyPositionParams(positionCaller, tickLower, tickUpper, liqDelta)
	position, amount0, amount1 := pool.modifyPosition(posParams)

	if amount0.Gt(u256.Zero()) || amount1.Gt(u256.Zero()) {
		amount0 = toUint128(amount0)
		amount1 = toUint128(amount1)
		position.tokensOwed0 = new(u256.Uint).Add(position.tokensOwed0, amount0)
		position.tokensOwed1 = new(u256.Uint).Add(position.tokensOwed1, amount1)
	}

	positionKey := getPositionKey(positionCaller, tickLower, tickUpper)
	pool.setPosition(positionKey, position)

	// mustGetPosition() is called to ensure the position exists
	pool.mustGetPosition(positionKey)

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
func Collect(cur realm,
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
	// Use recipient address instead of getPrevAddr() for position key generation
	// because positions are created with the recipient's address in Mint function,
	// and we need to access the same position that was originally created.
	// getPrevAddr() would return the position contract address, but the actual
	// position is stored under the recipient's address key.
	positionKey := getPositionKey(recipient, tickLower, tickUpper)
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
		checkTransferError(common.Transfer(cross, pool.token0Path, recipient, amount0.Int64()))
	}
	if amount1.Gt(u256.Zero()) {
		position.tokensOwed1 = new(u256.Uint).Sub(position.tokensOwed1, amount1)
		pool.balances.token1 = new(u256.Uint).Sub(pool.balances.token1, amount1)
		checkTransferError(common.Transfer(cross, pool.token1Path, recipient, amount1.Int64()))
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

// CollectProtocolByAdmin collects protocol fees for the given pool that accumulated while it was being used for swap
// Returns collected amount0, amount1 in string
func CollectProtocolByAdmin(cur realm,
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
func CollectProtocol(cur realm,
	token0Path string,
	token1Path string,
	fee uint32,
	recipient std.Address,
	amount0Requested string, // uint128
	amount1Requested string, // uint128
) (string, string) {
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
	if amount0Req.IsZero() && amount1Req.IsZero() {
		return u256.Zero().ToString(), u256.Zero().ToString()
	}

	amount0 := u256Min(amount0Req, pool.ProtocolFeesToken0())
	amount1 := u256Min(amount1Req, pool.ProtocolFeesToken1())

	amount0, amount1 = pool.saveProtocolFees(amount0.Clone(), amount1.Clone())
	uAmount0 := amount0.Int64()
	uAmount1 := amount1.Int64()

	checkTransferError(common.Transfer(cross, pool.token0Path, recipient, uAmount0))
	newBalanceToken0, err := updatePoolBalance(pool.BalanceToken0(), pool.BalanceToken1(), amount0, true)
	if err != nil {
		panic(err)
	}
	pool.balances.token0 = newBalanceToken0

	checkTransferError(common.Transfer(cross, pool.token1Path, recipient, uAmount1))
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
	p.protocolFees.token0 = new(u256.Uint).Sub(p.ProtocolFeesToken0(), amount0)
	p.protocolFees.token1 = new(u256.Uint).Sub(p.ProtocolFeesToken1(), amount1)

	return amount0, amount1
}
