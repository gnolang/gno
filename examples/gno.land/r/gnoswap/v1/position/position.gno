package position

import (
	"encoding/base64"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/access"
	pl "gno.land/r/gnoswap/v1/pool"
	"gno.land/r/gnoswap/v1/referral"
)

var (
	positionAddr, _ = access.GetAddress(access.ROLE_POSITION)
	stakerAddr, _   = access.GetAddress(access.ROLE_STAKER)
)

const (
	ZERO_LIQUIDITY_FOR_FEE_COLLECTION = "0"
)

// Mint creates a new liquidity position by depositing token pairs into the pool and minting a new LP token.
//
// Parameters:
//   - token0: The address of token0.
//   - token1: The address of token1.
//   - fee: The fee tier of the pool, in basis points.
//   - tickLower: The lower tick boundary of the position.
//   - tickUpper: The upper tick boundary of the position.
//   - amount0Desired: Desired amount of token0 to add as liquidity, as a string.
//   - amount1Desired: Desired amount of token1 to add as liquidity, as a string.
//   - amount0Min: Minimum acceptable amount of token0 to add as liquidity, as a string.
//   - amount1Min: Minimum acceptable amount of token1 to add as liquidity, as a string.
//   - deadline: Expiration timestamp for the transaction.
//   - mintTo: Address to receive the minted LP token.
//   - caller: The address of the entity (contract or user) providing liquidity; assets will be withdrawn from this address.
//
// Returns:
//   - uint64: The ID of the newly minted liquidity position.
//   - string: The amount of liquidity provided to the position.
//   - string: The amount of token0 used in the mint.
//   - string: The amount of token1 used in the mint.
//
// Behavior:
//  1. **Validation**:
//     - Ensures the contract is not halted.
//     - Validates that the caller is either a user or a staker contract.
//     - If the caller is a user, validates the `mintTo` and `caller` addresses to ensure they match.
//     - Checks the transaction's deadline to prevent expired transactions.
//  2. **Pre-Mint Setup**:
//     - Calls `MintAndDistributeGns` to handle GNS emissions.
//     - Processes the input parameters for minting (`processMintInput`) to standardize and validate the inputs.
//  3. **Mint Execution**:
//     - Executes the mint operation using the processed parameters.
//     - Withdraws the required token amounts (`token0` and `token1`) from the `caller` address.
//     - Mints a new LP token, and the resulting LP token is sent to the `mintTo` address.
//  4. **Post-Mint Cleanup**:
//     - If native tokens were used (e.g., `ugnot`), unwraps any leftover wrapped tokens (`wugnot`) and refunds them to the `caller` address.
//  5. **Event Emission**:
//     - Emits a "Mint" event containing detailed information about the mint operation.
//
// Panics:
//   - If the contract is halted.
//   - If the caller is not authorized.
//   - If the transaction deadline has passed.
//   - If input validation fails.
//   - If errors occur during the minting process or leftover token unwrapping.
//
// ref: https://docs.gnoswap.io/contracts/position/position.gno#mint
func Mint(
	cur realm,
	token0 string,
	token1 string,
	fee uint32,
	tickLower int32,
	tickUpper int32,
	amount0Desired string,
	amount1Desired string,
	amount0Min string,
	amount1Min string,
	deadline int64,
	mintTo std.Address,
	caller std.Address,
	referrer string,
) (uint64, string, string, string) {
	assertOnlyNotHalted()
	prevCaller := getPrevRealm()
	assertOnlyUserOrStaker(prevCaller)
	if isUserCall() {
		// if user called, validate the prev address and input addresses(mintTo, caller)
		assertOnlyValidAddressWith(prevCaller.Address(), mintTo)
		assertOnlyValidAddressWith(prevCaller.Address(), caller)
	}
	checkDeadline(deadline)

	referral.TryRegister(cross, caller, referrer)

	mintAndDistributeGnsCallback()

	mintInput := MintInput{
		token0:         token0,
		token1:         token1,
		fee:            fee,
		tickLower:      tickLower,
		tickUpper:      tickUpper,
		amount0Desired: amount0Desired,
		amount1Desired: amount1Desired,
		amount0Min:     amount0Min,
		amount1Min:     amount1Min,
		deadline:       deadline,
		mintTo:         mintTo,
		caller:         caller,
	}
	processedInput, err := processMintInput(mintInput)
	if err != nil {
		panic(newErrorWithDetail(errInvalidInput, err.Error()))
	}

	// create MintParams
	mintParams := newMintParams(processedInput, mintInput)

	// mint liquidity
	id, liquidity, amount0, amount1 := mint(mintParams)

	// refund leftover wrapped tokens
	if processedInput.tokenPair.token0IsNative && processedInput.tokenPair.wrappedAmount > amount0.Int64() {
		err = unwrap(processedInput.tokenPair.wrappedAmount-amount0.Int64(), caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}
	if processedInput.tokenPair.token1IsNative && processedInput.tokenPair.wrappedAmount > amount1.Int64() {
		err = unwrap(processedInput.tokenPair.wrappedAmount-amount1.Int64(), caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}

	poolSqrtPriceX96 := pl.PoolGetSlot0SqrtPriceX96(processedInput.poolPath)
	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"Mint",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"tickLower", formatInt(processedInput.tickLower),
		"tickUpper", formatInt(processedInput.tickUpper),
		"poolPath", processedInput.poolPath,
		"mintTo", mintTo.String(),
		"caller", caller.String(),
		"lpPositionId", formatUint(id),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"sqrtPriceX96", poolSqrtPriceX96,
		"token0Balance", pl.PoolGetBalanceToken0(processedInput.poolPath),
		"token1Balance", pl.PoolGetBalanceToken1(processedInput.poolPath),
	)

	return id, liquidity.ToString(), amount0.ToString(), amount1.ToString()
}

// IncreaseLiquidity increases liquidity of the existing position
// Returns position ID, liquidity, amount0, amount1, poolPath
// ref: https://docs.gnoswap.io/contracts/position/position.gno#increaseliquidity
func IncreaseLiquidity(
	cur realm,
	positionId uint64,
	amount0DesiredStr string,
	amount1DesiredStr string,
	amount0MinStr string,
	amount1MinStr string,
	deadline int64,
) (uint64, string, string, string, string) {
	assertOnlyNotHalted()
	assertValidNumberString(amount0DesiredStr)
	assertValidNumberString(amount1DesiredStr)
	assertValidNumberString(amount0MinStr)
	assertValidNumberString(amount1MinStr)
	checkDeadline(deadline)

	mintAndDistributeGnsCallback()

	position := MustGetPosition(positionId)
	token0, token1, _ := splitOf(position.poolKey)
	err := validateTokenPath(token0, token1)
	if err != nil {
		panic(newErrorWithDetail(err, ufmt.Sprintf("token0(%s), token1(%s)", token0, token1)))
	}

	caller := getPrevAddr()

	wrappedAmount := int64(0)
	if isWrappedToken(token0) {
		wrappedAmount = safeWrapNativeToken(amount0DesiredStr, caller)
	} else if isWrappedToken(token1) {
		wrappedAmount = safeWrapNativeToken(amount1DesiredStr, caller)
	}

	amount0Desired, amount1Desired, amount0Min, amount1Min := parseAmounts(amount0DesiredStr, amount1DesiredStr, amount0MinStr, amount1MinStr)
	increaseLiquidityParams := IncreaseLiquidityParams{
		positionId:     positionId,
		amount0Desired: amount0Desired,
		amount1Desired: amount1Desired,
		amount0Min:     amount0Min,
		amount1Min:     amount1Min,
		deadline:       deadline,
		caller:         caller,
	}

	_, liquidity, amount0, amount1, poolPath := increaseLiquidity(increaseLiquidityParams)

	amount0Int64 := amount0.Int64()
	amount1Int64 := amount1.Int64()

	if isWrappedToken(token0) && wrappedAmount > amount0Int64 {
		// unwrap leftover wugnot
		err = unwrap(wrappedAmount-amount0Int64, caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}
	if isWrappedToken(token1) && wrappedAmount > amount1Int64 {
		// unwrap leftover wugnot
		err = unwrap(wrappedAmount-amount1Int64, caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"IncreaseLiquidity",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"poolPath", poolPath,
		"caller", caller.String(),
		"lpPositionId", formatUint(positionId),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"sqrtPriceX96", pl.PoolGetSlot0SqrtPriceX96(poolPath),
		"positionLiquidity", PositionGetPositionLiquidityStr(positionId),
		"token0Balance", pl.PoolGetBalanceToken0(poolPath),
		"token1Balance", pl.PoolGetBalanceToken1(poolPath),
	)

	return positionId, liquidity.ToString(), amount0.ToString(), amount1.ToString(), poolPath
}

// DecreaseLiquidity decreases liquidity of the existing position
// It also handles the conversion between GNOT and WUGNOT transparently for the user.
// Returns position ID, liquidity, fee0, fee1, amount0, amount1, poolPath
// ref: https://docs.gnoswap.io/contracts/position/position.gno#decreaseliquidity
func DecreaseLiquidity(
	cur realm,
	positionId uint64,
	liquidityStr string,
	amount0MinStr string,
	amount1MinStr string,
	deadline int64,
	unwrapResult bool,
) (uint64, string, string, string, string, string, string) {
	caller := getPrevAddr()

	assertOnlyNotHalted()
	assertAuthorizedForToken(caller, positionId)
	checkDeadline(deadline)
	assertValidLiquidityAmount(liquidityStr)
	mintAndDistributeGnsCallback()

	amount0Min := u256.MustFromDecimal(amount0MinStr)
	amount1Min := u256.MustFromDecimal(amount1MinStr)
	decreaseLiquidityParams := DecreaseLiquidityParams{
		positionId:   positionId,
		liquidity:    liquidityStr,
		amount0Min:   amount0Min,
		amount1Min:   amount1Min,
		deadline:     deadline,
		unwrapResult: unwrapResult,
		caller:       caller,
	}

	positionId, liquidity, fee0, fee1, amount0, amount1, poolPath := decreaseLiquidity(decreaseLiquidityParams)

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"DecreaseLiquidity",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpPositionId", formatUint(positionId),
		"poolPath", poolPath,
		"decreasedLiquidity", liquidity.ToString(),
		"feeAmount0", fee0.ToString(),
		"feeAmount1", fee1.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"unwrapResult", formatBool(unwrapResult),
		"sqrtPriceX96", pl.PoolGetSlot0SqrtPriceX96(poolPath),
		"positionLiquidity", PositionGetPositionLiquidityStr(positionId),
		"token0Balance", pl.PoolGetBalanceToken0(poolPath),
		"token1Balance", pl.PoolGetBalanceToken1(poolPath),
	)

	return positionId, liquidity.ToString(), fee0.ToString(), fee1.ToString(), amount0.ToString(), amount1.ToString(), poolPath
}

// CollectFee collects swap fee from the position
// Returns position ID, afterFee0, afterFee1, poolPath, origFee0, origFee1
// ref: https://docs.gnoswap.io/contracts/position/position.gno#collectfee
func CollectFee(
	cur realm,
	positionId uint64,
	unwrapResult bool,
) (uint64, string, string, string, string, string) {
	caller := getPrevAddr()

	return collectFee(cur, positionId, unwrapResult, caller)
}

func collectFee(
	cur realm,
	positionId uint64,
	unwrapResult bool,
	caller std.Address,
) (uint64, string, string, string, string, string) {
	assertOnlyNotHalted()
	assertTokenExists(positionId)
	assertAuthorizedForToken(caller, positionId)
	mintAndDistributeGnsCallback()

	// verify position
	position := MustGetPosition(positionId)
	token0, token1, fee := splitOf(position.poolKey)

	pl.Burn(
		cross,
		token0,
		token1,
		fee,
		position.tickLower,
		position.tickUpper,
		ZERO_LIQUIDITY_FOR_FEE_COLLECTION, // burn '0' liquidity to collect fee
		caller,
	)

	currentFeeGrowth, err := getCurrentFeeGrowth(position, caller)
	if err != nil {
		panic(newErrorWithDetail(err, "failed to get current fee growth"))
	}

	tokensOwed0, tokensOwed1 := calculateFees(position, currentFeeGrowth)

	position.feeGrowthInside0LastX128 = new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside0LastX128)
	position.feeGrowthInside1LastX128 = new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside1LastX128)

	// check user wugnot amount
	// need this value to unwrap fee
	userWugnot := wugnot.BalanceOf(caller)

	// collect fee
	amount0, amount1 := pl.Collect(
		cross,
		token0, token1, fee,
		caller,
		position.tickLower, position.tickUpper,
		tokensOwed0.ToString(), tokensOwed1.ToString(),
	)

	// sometimes there will be a few less uBase amount than expected due to rounding down in core, but we just subtract the full amount expected
	// instead of the actual amount so we can burn the token
	position.tokensOwed0 = new(u256.Uint).Sub(tokensOwed0, u256.MustFromDecimal(amount0))
	position.tokensOwed1 = new(u256.Uint).Sub(tokensOwed1, u256.MustFromDecimal(amount1))
	mustUpdatePosition(positionId, position)

	// handle withdrawal fee
	withoutFee0, withoutFee1 := pl.CollectWithdrawalFee(
		cross,
		positionId,
		token0,
		amount0,
		token1,
		amount1,
		position.poolKey,
		caller,
	)

	// UNWRAP
	pToken0, pToken1, _ := splitOf(position.poolKey)
	handleUnwrap(pToken0, pToken1, unwrapResult, userWugnot, caller)

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"CollectSwapFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpPositionId", formatUint(positionId),
		"feeAmount0", withoutFee0,
		"feeAmount1", withoutFee1,
		"poolPath", position.poolKey,
		"unwrapResult", formatBool(unwrapResult),
	)

	return positionId, withoutFee0, withoutFee1, position.poolKey, amount0, amount1
}

func SetPositionOperator(
	cur realm,
	id uint64,
	operator std.Address,
) {
	caller := getPrevRealm().PkgPath()
	if caller != consts.STAKER_PATH {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("caller(%s) is not staker", caller),
		))
	}

	position := MustGetPosition(id)
	position.operator = operator
	mustUpdatePosition(id, position)
}

// computePositionKey generates a unique base64-encoded key for a liquidity position.
//
// This function takes an owner's address and the lower and upper tick bounds of a position,
// and generates a unique key by concatenating the parameters into a string. The resulting
// string is base64 encoded to ensure it is compact and unique.
//
// Parameters:
//   - owner (std.Address): The address of the position owner.
//   - tickLower (int32): The lower tick boundary of the position.
//   - tickUpper (int32): The upper tick boundary of the position.
//
// Returns:
//   - string: A base64-encoded string representing the unique key for the position.
//
// Notes:
//   - This function is useful in scenarios where unique identifiers for liquidity positions
//     are required (e.g., decentralized exchange positions).
//   - The key format follows the pattern "ownerAddress__tickLower__tickUpper" to ensure uniqueness.
func computePositionKey(
	owner std.Address,
	tickLower int32,
	tickUpper int32,
) string {
	key := ufmt.Sprintf("%s__%d__%d", owner.String(), tickLower, tickUpper)
	encoded := base64.StdEncoding.EncodeToString([]byte(key))
	return encoded
}

func getCurrentFeeGrowth(position Position, owner std.Address) (FeeGrowthInside, error) {
	pool := pl.GetPoolFromPoolPath(position.poolKey)
	positionKey := computePositionKey(owner, position.tickLower, position.tickUpper)

	feeGrowthInside0 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

	feeGrowthInside := FeeGrowthInside{
		feeGrowthInside0LastX128: feeGrowthInside0,
		feeGrowthInside1LastX128: feeGrowthInside1,
	}

	return feeGrowthInside, nil
}
