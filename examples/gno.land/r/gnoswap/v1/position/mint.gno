package position

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gnft"
	pl "gno.land/r/gnoswap/v1/pool"
)

// mint creates a new liquidity position by adding liquidity to a pool and minting an NFT representing the position.
//
// This function handles the entire lifecycle of creating a new liquidity position, including adding liquidity
// to a pool, minting an NFT to represent the position, and storing the position's state.
//
// Parameters:
//   - params (MintParams): A struct containing all necessary parameters to mint a new liquidity position, including:
//   - token0, token1: The addresses of the token pair.
//   - fee: The fee tier of the pool.
//   - tickLower, tickUpper: The price range (ticks) for the liquidity position.
//   - amount0Desired, amount1Desired: Desired amounts of token0 and token1 to provide.
//   - amount0Min, amount1Min: Minimum acceptable amounts to prevent slippage.
//   - caller: The address initiating the mint. The required token amounts (token0 and token1) will be withdrawn
//     from the caller's balance and deposited into the pool.
//   - mintTo: The address to receive the newly minted NFT.
//
// Returns:
//   - uint64: The token ID of the minted liquidity position NFT.
//   - *u256.Uint: The amount of liquidity added to the pool.
//   - *u256.Uint: The actual amount of token0 used in the liquidity addition.
//   - *u256.Uint: The actual amount of token1 used in the liquidity addition.
//
// Panics:
//   - If the liquidity position (position ID) already exists.
//   - If adding liquidity fails due to insufficient amounts or invalid tick ranges.
//
// Notes:
//   - This function relies on `addLiquidity` to perform the liquidity calculation and ensure proper slippage checks.
//   - The NFT minted is critical for tracking the user's liquidity in the pool.
//   - Position state management is handled by `setPosition`, ensuring the uniqueness of the position ID.
func mint(params MintParams) (uint64, *u256.Uint, *u256.Uint, *u256.Uint) {
	poolKey := pl.GetPoolPath(params.token0, params.token1, params.fee)
	liquidity, amount0, amount1 := addLiquidity(
		AddLiquidityParams{
			poolKey:        poolKey,
			tickLower:      params.tickLower,
			tickUpper:      params.tickUpper,
			amount0Desired: params.amount0Desired,
			amount1Desired: params.amount1Desired,
			amount0Min:     params.amount0Min,
			amount1Min:     params.amount1Min,
			caller:         params.caller,
		},
	)
	// Ensure liquidity is not zero before minting NFT
	if liquidity.IsZero() {
		panic(newErrorWithDetail(
			errZeroLiquidity,
			"Liquidity is zero, cannot mint position.",
		))
	}

	id := GetNextId()
	if ExistPosition(id) {
		panic(newErrorWithDetail(
			errPositionExist,
			ufmt.Sprintf("positionId(%d)", id),
		))
	}

	gnft.Mint(cross, params.mintTo, positionIdFrom(id)) // owner, position ID

	pool := pl.GetPoolFromPoolPath(poolKey)
	positionKey := computePositionKey(params.caller, params.tickLower, params.tickUpper)

	position := Position{
		nonce:                    u256.Zero(),
		operator:                 consts.ZERO_ADDRESS,
		poolKey:                  poolKey,
		tickLower:                params.tickLower,
		tickUpper:                params.tickUpper,
		liquidity:                liquidity,
		feeGrowthInside0LastX128: new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey)),
		feeGrowthInside1LastX128: new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey)),
		tokensOwed0:              u256.Zero(),
		tokensOwed1:              u256.Zero(),
		token0Balance:            amount0,
		token1Balance:            amount1,
		burned:                   false,
	}

	// The position ID should not exist at the time of minting
	updated := setPosition(id, position)
	if updated {
		panic(newErrorWithDetail(
			errPositionExist,
			ufmt.Sprintf("position ID(%d) already exists", id),
		))
	}
	incrementNextId()

	return id, liquidity, amount0, amount1
}

// processMintInput processes and validates user input for minting liquidity.
//
// This function standardizes and verifies minting parameters by ensuring token order, parsing desired amounts,
// and handling native token wrapping. It returns a structured `ProcessedMintInput` that can be used for further minting operations.
//
// Parameters:
//   - input (MintInput): Raw user input containing token addresses, tick bounds, and liquidity amounts.
//
// Returns:
//   - ProcessedMintInput: A structured and validated version of the minting input.
//   - error: Returns an error if input parsing, token processing, or number validation fails.
//
// Behavior:
//
//  1. **Number Validation**:
//     - Validates `amount0Desired`, `amount1Desired`, `amount0Min`, and `amount1Min` to ensure they are valid numeric strings.
//     - If validation fails, the function panics immediately with an error indicating invalid input.
//
//  2. **Token Processing**:
//     - Calls `processTokens` to validate and convert tokens (`token0` and `token1`) into their final forms.
//     - Handles wrapping of native tokens (e.g., UGNOT to WUGNOT) if necessary.
//     - Stores token metadata in a `TokenPair` struct, including wrapped amounts and native token status.
//
//  3. **Amount Parsing**:
//     - Converts `amount0Desired`, `amount1Desired`, `amount0Min`, and `amount1Min` from string to `u256.Uint` using `parseAmounts`.
//     - Ensures accurate representation of liquidity amounts for further processing.
//
//  4. **Token Order Enforcement**:
//     - If `token1` is lexicographically smaller than `token0`, the function swaps their order to enforce consistent pool identification.
//     - Along with token swaps, the tick bounds (`tickLower`, `tickUpper`) are inverted to preserve correct price boundaries.
//     - This step guarantees pool uniqueness by ensuring `token0 < token1`.
//
//  5. **Pool Path Calculation**:
//     - Computes the pool path (`poolPath`) using the finalized token addresses and fee tier.
//     - The pool path uniquely identifies the pool in which liquidity will be minted.
//
//  6. **Return**:
//     - Returns a populated `ProcessedMintInput` struct containing the finalized minting parameters.
//
// Panics:
//   - If any of the provided amount strings are invalid (non-numeric or empty).
//   - If `processTokens` encounters errors during token validation or wrapping.
//
// Notes:
//   - This function enforces token order and validates amounts to ensure the integrity of liquidity minting operations.
func processMintInput(input MintInput) (ProcessedMintInput, error) {
	assertValidNumberString(input.amount0Desired)
	assertValidNumberString(input.amount1Desired)
	assertValidNumberString(input.amount0Min)
	assertValidNumberString(input.amount1Min)
	var result ProcessedMintInput

	// process tokens
	token0, token1, token0IsNative, token1IsNative, wrappedAmount := processTokens(input.token0, input.token1, input.amount0Desired, input.amount1Desired, input.caller)
	pair := TokenPair{
		token0:         token0,
		token1:         token1,
		token0IsNative: token0IsNative,
		token1IsNative: token1IsNative,
		wrappedAmount:  wrappedAmount,
	}

	// parse amounts
	amount0Desired, amount1Desired, amount0Min, amount1Min := parseAmounts(input.amount0Desired, input.amount1Desired, input.amount0Min, input.amount1Min)

	tickLower, tickUpper := input.tickLower, input.tickUpper

	// swap if token1 < token0
	if token1 < token0 {
		pair.token0, pair.token1 = pair.token1, pair.token0
		amount0Desired, amount1Desired = amount1Desired, amount0Desired
		amount0Min, amount1Min = amount1Min, amount0Min
		tickLower, tickUpper = -tickUpper, -tickLower
		pair.token0IsNative, pair.token1IsNative = pair.token1IsNative, pair.token0IsNative
	}

	poolPath := computePoolPath(pair.token0, pair.token1, input.fee)

	result = ProcessedMintInput{
		tokenPair:      pair,
		amount0Desired: new(u256.Uint).Set(amount0Desired),
		amount1Desired: new(u256.Uint).Set(amount1Desired),
		amount0Min:     new(u256.Uint).Set(amount0Min),
		amount1Min:     new(u256.Uint).Set(amount1Min),
		tickLower:      tickLower,
		tickUpper:      tickUpper,
		poolPath:       poolPath,
	}

	return result, nil
}

// processTokens processes two token paths, validates them, and handles the wrapping of native tokens into wrapped tokens if applicable.
//
// Parameters:
//   - token0: The first token path to process.
//   - token1: The second token path to process.
//   - caller: The address of the user initiating the token processing.
//
// Returns:
//   - string: Processed token0 path (potentially modified if it was a native token).
//   - string: Processed token1 path (potentially modified if it was a native token).
//   - bool: Indicates whether token0 was a native token (`true` if native, `false` otherwise).
//   - bool: Indicates whether token1 was a native token (`true` if native, `false` otherwise).
//   - uint64: The amount of the native token that was wrapped into the wrapped token.
//
// Behavior:
//  1. Validates the token paths using `validateTokenPath`.
//     - Panics with a detailed error if validation fails.
//  2. Checks if `token0` or `token1` is a native token using `isNative`.
//     - If a token is native, it is replaced with the wrapped token path (`WRAPPED_WUGNOT`).
//     - The native token is then wrapped into the wrapped token using `safeWrapNativeToken`.
//  3. Returns the processed token paths, flags indicating if the tokens were native, and the wrapped amount.
//
// Panics:
//   - If `validateTokenPath` fails validation.
//   - If wrapping the native token using `safeWrapNativeToken` encounters an issue.
func processTokens(
	token0 string,
	token1 string,
	amount0Desired string,
	amount1Desired string,
	caller std.Address,
) (string, string, bool, bool, int64) {
	err := validateTokenPath(token0, token1)
	if err != nil {
		panic(newErrorWithDetail(err, ufmt.Sprintf("token0(%s), token1(%s)", token0, token1)))
	}

	token0IsNative := false
	token1IsNative := false
	wrappedAmount := int64(0)

	if isNative(token0) {
		token0 = consts.WRAPPED_WUGNOT
		token0IsNative = true

		wrappedAmount = safeWrapNativeToken(amount0Desired, caller)
	} else if isNative(token1) {
		token1 = consts.WRAPPED_WUGNOT
		token1IsNative = true

		wrappedAmount = safeWrapNativeToken(amount1Desired, caller)
	}

	return token0, token1, token0IsNative, token1IsNative, wrappedAmount
}

func increaseLiquidity(params IncreaseLiquidityParams) (uint64, *u256.Uint, *u256.Uint, *u256.Uint, string) {
	// verify position ID exists
	caller := params.caller
	assertTokenExists(params.positionId)
	assertOnlyOwnerOfToken(params.positionId, caller)

	position := MustGetPosition(params.positionId)
	liquidity, amount0, amount1 := addLiquidity(
		AddLiquidityParams{
			poolKey:        position.poolKey,
			tickLower:      position.tickLower,
			tickUpper:      position.tickUpper,
			amount0Desired: params.amount0Desired,
			amount1Desired: params.amount1Desired,
			amount0Min:     params.amount0Min,
			amount1Min:     params.amount1Min,
			caller:         caller,
		},
	)

	pool := pl.GetPoolFromPoolPath(position.poolKey)
	positionKey := computePositionKey(caller, position.tickLower, position.tickUpper)
	feeGrowthInside0LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

	currentSqrtPriceX96 := pool.Slot0SqrtPriceX96()
	lowerTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickLower)
	upperTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickUpper)
	calculatedToken0BalanceStr, calculatedToken1BalanceStr := common.GetAmountsForLiquidity(
		currentSqrtPriceX96,
		lowerTickSqrtPriceX96,
		upperTickSqrtPriceX96,
		position.liquidity,
	)
	calculatedToken0Balance := u256.MustFromDecimal(calculatedToken0BalanceStr)
	calculatedToken1Balance := u256.MustFromDecimal(calculatedToken1BalanceStr)

	{
		diff := new(u256.Uint).Sub(feeGrowthInside0LastX128, position.feeGrowthInside0LastX128)
		mulDiv := u256.MulDiv(diff, new(u256.Uint).Set(position.liquidity), u256.MustFromDecimal(consts.Q128))

		position.tokensOwed0 = new(u256.Uint).Add(position.tokensOwed0, mulDiv)
	}

	{
		diff := new(u256.Uint).Sub(feeGrowthInside1LastX128, position.feeGrowthInside1LastX128)
		mulDiv := u256.MulDiv(diff, new(u256.Uint).Set(position.liquidity), u256.MustFromDecimal(consts.Q128))

		position.tokensOwed1 = new(u256.Uint).Add(position.tokensOwed1, mulDiv)
	}

	position.feeGrowthInside0LastX128 = feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = feeGrowthInside1LastX128
	position.liquidity = new(u256.Uint).Add(new(u256.Uint).Set(position.liquidity), liquidity)
	position.token0Balance = new(u256.Uint).Add(new(u256.Uint).Set(calculatedToken0Balance), amount0)
	position.token1Balance = new(u256.Uint).Add(new(u256.Uint).Set(calculatedToken1Balance), amount1)
	position.burned = false

	updated := setPosition(params.positionId, position)
	if !updated {
		panic(newErrorWithDetail(
			errPositionDoesNotExist,
			ufmt.Sprintf("can not increase liquidity for non-existent position(%d)", params.positionId),
		))
	}

	return params.positionId, liquidity, amount0, amount1, position.poolKey
}

// validateTokenPath validates the relationship and format of token paths.
// Ensures that token paths are not identical, not conflicting (e.g., GNOT and WUGNOT),
// and each token path is in a valid format.
//
// Parameters:
//   - token0: The first token path to validate.
//   - token1: The second token path to validate.
//
// Returns:
//   - error: Returns `errInvalidTokenPath` or nil
//
// Example:
//
//	validateTokenPath("tokenA", "tokenB") -> nil
//	validateTokenPath("tokenA", "tokenA") -> errInvalidTokenPath
//	validateTokenPath(GNOT, WUGNOT) -> errInvalidTokenPath
func validateTokenPath(token0, token1 string) error {
	if token0 == token1 {
		return errInvalidTokenPath
	}
	if (token0 == consts.GNOT && token1 == consts.WRAPPED_WUGNOT) ||
		(token0 == consts.WRAPPED_WUGNOT && token1 == consts.GNOT) {
		return errInvalidTokenPath
	}
	if (!isNative(token0) && !isValidTokenPath(token0)) ||
		(!isNative(token1) && !isValidTokenPath(token1)) {
		return errInvalidTokenPath
	}
	return nil
}

// isValidTokenPath checks if the provided token path is registered.
//
// This function verifies if the specified token path exists in the system registry
// by invoking the `IsRegistered` method. A path is considered valid if no error
// is returned during the registration check.
//
// Parameters:
//   - tokenPath: The string representing the token path to validate.
//
// Returns:
//   - bool: Returns `true` if the token path is registered; otherwise, `false`.
func isValidTokenPath(tokenPath string) bool {
	return common.IsRegistered(tokenPath) == nil
}

// parseAmounts converts strings to u256.Uint values for amount0Desired, amount1Desired, amount0Min, and amount1Min.
func parseAmounts(amount0Desired, amount1Desired, amount0Min, amount1Min string) (*u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint) {
	return u256.MustFromDecimal(amount0Desired), u256.MustFromDecimal(amount1Desired), u256.MustFromDecimal(amount0Min), u256.MustFromDecimal(amount1Min)
}

// computePoolPath returns the pool path based on the token pair and fee tier.
//
// This function constructs a unique pool path identifier by utilizing the two token addresses
// and the pool fee tier. It helps identify the specific liquidity pool on the platform.
//
// Parameters:
// - token0: The address of the first token (string).
// - token1: The address of the second token (string).
// - fee: The fee for the liquidity pool (uint32).
//
// Returns:
// - string: A unique path string representing the liquidity pool.
func computePoolPath(token0, token1 string, fee uint32) string {
	return pl.GetPoolPath(token0, token1, fee)
}
