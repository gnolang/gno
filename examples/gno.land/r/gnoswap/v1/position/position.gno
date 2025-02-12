package position

import (
	"encoding/base64"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gnft"

	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	en "gno.land/r/gnoswap/v1/emission"
	pl "gno.land/r/gnoswap/v1/pool"
)

const (
	ZERO_LIQUIDITY_FOR_FEE_COLLECTION = "0"
)

var (
	positions = avl.NewTree() // tokenId[uint64] -> Position
	nextId    = uint64(1)
)

// MustGetPosition returns a position for a given tokenId
// panics if position doesn't exist
func MustGetPosition(tokenId uint64) Position {
	position, exist := GetPosition(tokenId)
	if !exist {
		panic(newErrorWithDetail(
			errPositionDoesNotExist,
			ufmt.Sprintf("position with tokenId(%d) doesn't exist", tokenId),
		))
	}
	return position
}

// mustUpdatePosition updates a position for a given tokenId
func mustUpdatePosition(tokenId uint64, position Position) {
	update := setPosition(tokenId, position)
	if !update {
		panic(newErrorWithDetail(
			errPositionDoesNotExist,
			ufmt.Sprintf("position with tokenId(%d) doesn't exist", tokenId),
		))
	}
}

// GetPosition returns a position for a given tokenId
// Returns false if position doesn't exist
func GetPosition(tokenId uint64) (Position, bool) {
	tokenIdStr := strconv.FormatUint(tokenId, 10)
	iPosition, exist := positions.Get(tokenIdStr)
	if !exist {
		return Position{}, false
	}

	return iPosition.(Position), true
}

// setPosition sets a position for a given tokenId
// Returns true if position is newly created, false if position already exists and just updated.
func setPosition(tokenId uint64, position Position) bool {
	tokenIdStr := strconv.FormatUint(tokenId, 10)
	return positions.Set(tokenIdStr, position)
}

// removePosition removes a position for a given tokenId
func removePosition(tokenId uint64) {
	tokenIdStr := strconv.FormatUint(tokenId, 10)
	positions.Remove(tokenIdStr)
}

func existPosition(tokenId uint64) bool {
	_, exist := GetPosition(tokenId)
	return exist
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

// nextId is the next tokenId to be minted
func getNextId() uint64 {
	return nextId
}

// incrementNextId increments the next tokenId to be minted
func incrementNextId() {
	nextId++
}

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
) (uint64, string, string, string) {
	assertOnlyNotHalted()
	prevCaller := getPrevRealm()
	assertOnlyUserOrStaker(prevCaller)
	if isUserCall() {
		// if user called, validate the prev address and input addresses(mintTo, caller)
		assertOnlyValidAddressWith(prevCaller.Addr(), mintTo)
		assertOnlyValidAddressWith(prevCaller.Addr(), caller)
	}
	checkDeadline(deadline)

	en.MintAndDistributeGns()

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

	mintParams := newMintParams(processedInput, mintInput)
	tokenId, liquidity, amount0, amount1 := mint(mintParams)
	if processedInput.tokenPair.token0IsNative && processedInput.tokenPair.wrappedAmount > amount0.Uint64() {
		// unwrap leftover wugnot
		err = unwrap(processedInput.tokenPair.wrappedAmount-amount0.Uint64(), caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}
	if processedInput.tokenPair.token1IsNative && processedInput.tokenPair.wrappedAmount > amount1.Uint64() {
		// unwrap leftover wugnot
		err = unwrap(processedInput.tokenPair.wrappedAmount-amount1.Uint64(), caller)
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
		"tickLower", formatInt(int64(processedInput.tickLower)),
		"tickUpper", formatInt(int64(processedInput.tickUpper)),
		"poolPath", processedInput.poolPath,
		"mintTo", mintTo.String(),
		"caller", caller.String(),
		"lpTokenId", formatUint(tokenId),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"sqrtPriceX96", poolSqrtPriceX96,
	)

	return tokenId, liquidity.ToString(), amount0.ToString(), amount1.ToString()
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
) (string, string, bool, bool, uint64) {
	err := validateTokenPath(token0, token1)
	if err != nil {
		panic(newErrorWithDetail(err, ufmt.Sprintf("token0(%s), token1(%s)", token0, token1)))
	}

	token0IsNative := false
	token1IsNative := false
	wrappedAmount := uint64(0)

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
//   - If the liquidity position (tokenId) already exists.
//   - If adding liquidity fails due to insufficient amounts or invalid tick ranges.
//
// Notes:
//   - This function relies on `addLiquidity` to perform the liquidity calculation and ensure proper slippage checks.
//   - The NFT minted is critical for tracking the user's liquidity in the pool.
//   - Position state management is handled by `setPosition`, ensuring the uniqueness of the tokenId.
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

	tokenId := getNextId()
	gnft.Mint(params.mintTo, tokenIdFrom(tokenId)) // owner, tokenId

	pool := pl.GetPoolFromPoolPath(poolKey)
	positionKey := computePositionKey(GetOrigPkgAddr(), params.tickLower, params.tickUpper)

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
		burned:                   false,
	}

	// The tokenId should not exist at the time of minting
	updated := setPosition(tokenId, position)
	if updated {
		panic(newErrorWithDetail(
			errPositionExist,
			ufmt.Sprintf("tokenId(%d) already exists", tokenId),
		))
	}
	incrementNextId()

	return tokenId, liquidity, amount0, amount1
}

// IncreaseLiquidity increases liquidity of the existing position
// Returns tokenId, liquidity, amount0, amount1, poolPath
// ref: https://docs.gnoswap.io/contracts/position/position.gno#increaseliquidity
func IncreaseLiquidity(
	tokenId uint64,
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

	en.MintAndDistributeGns()

	position := MustGetPosition(tokenId)
	token0, token1, _ := splitOf(position.poolKey)
	err := validateTokenPath(token0, token1)
	if err != nil {
		panic(newErrorWithDetail(err, ufmt.Sprintf("token0(%s), token1(%s)", token0, token1)))
	}
	caller := getPrevAddr()

	wrappedAmount := uint64(0)
	if isWrappedToken(token0) {
		wrappedAmount = safeWrapNativeToken(amount0DesiredStr, caller)
	} else if isWrappedToken(token1) {
		wrappedAmount = safeWrapNativeToken(amount1DesiredStr, caller)
	}

	amount0Desired, amount1Desired, amount0Min, amount1Min := parseAmounts(amount0DesiredStr, amount1DesiredStr, amount0MinStr, amount1MinStr)
	increaseLiquidityParams := IncreaseLiquidityParams{
		tokenId:        tokenId,
		amount0Desired: amount0Desired,
		amount1Desired: amount1Desired,
		amount0Min:     amount0Min,
		amount1Min:     amount1Min,
		deadline:       deadline,
	}

	_, liquidity, amount0, amount1, poolPath := increaseLiquidity(increaseLiquidityParams)

	if isWrappedToken(token0) && wrappedAmount > amount0.Uint64() {
		// unwrap leftover wugnot
		err = unwrap(wrappedAmount-amount0.Uint64(), caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}
	if isWrappedToken(token1) && wrappedAmount > amount1.Uint64() {
		// unwrap leftover wugnot
		err = unwrap(wrappedAmount-amount1.Uint64(), caller)
		if err != nil {
			panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
		}
	}

	poolSqrtPriceX96 := pl.PoolGetSlot0SqrtPriceX96(poolPath)
	token0Balance := pl.PoolGetBalanceToken0(poolPath)
	token1Balance := pl.PoolGetBalanceToken1(poolPath)

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"IncreaseLiquidity",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"poolPath", poolPath,
		"caller", caller.String(),
		"lpTokenId", formatUint(tokenId),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"sqrtPriceX96", poolSqrtPriceX96,
		"positionLiquidity", position.liquidity.ToString(),
		"token0Balance", token0Balance,
		"token1Balance", token1Balance,
	)

	return tokenId, liquidity.ToString(), amount0.ToString(), amount1.ToString(), poolPath
}

// FeeGrowthInside represents fee growth inside ticks
type FeeGrowthInside struct {
	feeGrowthInside0LastX128 *u256.Uint
	feeGrowthInside1LastX128 *u256.Uint
}

// PositionFeeUpdate represents fee update calculation result
type PositionFeeUpdate struct {
	tokensOwed0              *u256.Uint
	tokensOwed1              *u256.Uint
	feeGrowthInside0LastX128 *u256.Uint
	feeGrowthInside1LastX128 *u256.Uint
}

func calculatePositionFeeUpdate(
	position Position,
	currentFeeGrowth FeeGrowthInside,
) PositionFeeUpdate {
	tokensOwed0 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside0LastX128,
		position.feeGrowthInside0LastX128,
		position.liquidity,
	)

	tokensOwed1 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside1LastX128,
		position.feeGrowthInside1LastX128,
		position.liquidity,
	)

	return PositionFeeUpdate{
		tokensOwed0:              new(u256.Uint).Add(position.tokensOwed0, tokensOwed0),
		tokensOwed1:              new(u256.Uint).Add(position.tokensOwed1, tokensOwed1),
		feeGrowthInside0LastX128: new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside0LastX128),
		feeGrowthInside1LastX128: new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside1LastX128),
	}
}

// updatePosition updates the position with new liquidity and fee data
func updatePosition(
	position Position,
	feeUpdate PositionFeeUpdate,
	newLiquidity *u256.Uint,
) Position {
	position.tokensOwed0 = feeUpdate.tokensOwed0
	position.tokensOwed1 = feeUpdate.tokensOwed1
	position.feeGrowthInside0LastX128 = feeUpdate.feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = feeUpdate.feeGrowthInside1LastX128
	position.liquidity = new(u256.Uint).Add(position.liquidity, newLiquidity)
	position.burned = false

	return position
}

func increaseLiquidity(params IncreaseLiquidityParams) (uint64, *u256.Uint, *u256.Uint, *u256.Uint, string) {
	// verify tokenId exists
	assertTokenExists(params.tokenId)
	caller := getPrevAddr()
	assertOnlyOwnerOfToken(params.tokenId, caller)

	position := MustGetPosition(params.tokenId)
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
	positionKey := computePositionKey(GetOrigPkgAddr(), position.tickLower, position.tickUpper)
	feeGrowthInside0LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

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
	position.burned = false

	updated := setPosition(params.tokenId, position)
	if !updated {
		panic(newErrorWithDetail(
			errPositionDoesNotExist,
			ufmt.Sprintf("can not increase liquidity for non-existent position(%d)", params.tokenId),
		))
	}

	return params.tokenId, liquidity, amount0, amount1, position.poolKey
}

// DecreaseLiquidity decreases liquidity of the existing position
// It also handles the conversion between GNOT and WUGNOT transparently for the user.
// Returns tokenId, liquidity, fee0, fee1, amount0, amount1, poolPath
// ref: https://docs.gnoswap.io/contracts/position/position.gno#decreaseliquidity
func DecreaseLiquidity(
	tokenId uint64,
	liquidityStr string,
	amount0MinStr string,
	amount1MinStr string,
	deadline int64,
	unwrapResult bool,
) (uint64, string, string, string, string, string, string) {
	assertOnlyNotHalted()
	isAuthorizedForToken(tokenId)
	checkDeadline(deadline)
	assertValidLiquidityAmount(liquidityStr)

	en.MintAndDistributeGns()

	amount0Min := u256.MustFromDecimal(amount0MinStr)
	amount1Min := u256.MustFromDecimal(amount1MinStr)
	decreaseLiquidityParams := DecreaseLiquidityParams{
		tokenId:      tokenId,
		liquidity:    liquidityStr,
		amount0Min:   amount0Min,
		amount1Min:   amount1Min,
		deadline:     deadline,
		unwrapResult: unwrapResult,
	}

	tokenId, liquidity, fee0, fee1, amount0, amount1, poolPath := decreaseLiquidity(decreaseLiquidityParams)

	poolSqrtPriceX96 := pl.PoolGetSlot0SqrtPriceX96(poolPath)
	token0Balance := pl.PoolGetBalanceToken0(poolPath)
	token1Balance := pl.PoolGetBalanceToken1(poolPath)

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"DecreaseLiquidity",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpTokenId", formatUint(tokenId),
		"poolPath", poolPath,
		"decreasedLiquidity", liquidity.ToString(),
		"feeAmount0", fee0.ToString(),
		"feeAmount1", fee1.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"unwrapResult", formatBool(unwrapResult),
		"sqrtPriceX96", poolSqrtPriceX96,
		"positionLiquidity", PositionGetPositionLiquidityStr(tokenId),
		"token0Balance", token0Balance,
		"token1Balance", token1Balance,
	)

	return tokenId, liquidity.ToString(), fee0.ToString(), fee1.ToString(), amount0.ToString(), amount1.ToString(), poolPath
}

// decreaseLiquidity reduces the liquidity of a given position and collects the corresponding tokens.
// If unwrapResult is true and the position involves WUGNOT, any leftover WUGNOT will be
// unwrapped to GNOT at the end of the operation.
// Returns tokenId, liquidity, fee0, fee1, amount0, amount1, poolPath
func decreaseLiquidity(params DecreaseLiquidityParams) (uint64, *u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint, string) {
	verifyTokenIdAndOwnership(params.tokenId)

	// BEFORE DECREASE LIQUIDITY, COLLECT FEE FIRST
	_, fee0Str, fee1Str, _, _, _ := CollectFee(params.tokenId, params.unwrapResult)
	fee0 := u256.MustFromDecimal(fee0Str)
	fee1 := u256.MustFromDecimal(fee1Str)

	position := MustGetPosition(params.tokenId)
	positionLiquidity := position.liquidity
	if positionLiquidity.IsZero() {
		panic(newErrorWithDetail(
			errZeroLiquidity,
			ufmt.Sprintf("position(tokenId:%d) has 0 liquidity", params.tokenId),
		))
	}

	liquidityToRemove := u256.MustFromDecimal(params.liquidity)
	if liquidityToRemove.Gt(positionLiquidity) {
		panic(newErrorWithDetail(
			errInvalidLiquidity,
			ufmt.Sprintf("Liquidity requested(%s) is greater than liquidity held(%s)", liquidityToRemove.ToString(), positionLiquidity.ToString()),
		))
	}

	caller := getPrevAddr()
	beforeWugnotBalance := wugnot.BalanceOf(caller) // before unwrap

	pToken0, pToken1, pFee := splitOf(position.poolKey)
	burn0, burn1 := pl.Burn(pToken0, pToken1, pFee, position.tickLower, position.tickUpper, liquidityToRemove.ToString())

	burnedAmount0 := u256.MustFromDecimal(burn0)
	burnedAmount1 := u256.MustFromDecimal(burn1)
	verifySlippageAmounts(burnedAmount0, burnedAmount1, params.amount0Min, params.amount1Min)

	positionKey := computePositionKey(GetOrigPkgAddr(), position.tickLower, position.tickUpper)
	pool := pl.GetPoolFromPoolPath(position.poolKey)
	feeGrowthInside0LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

	position.tokensOwed0 = updateTokensOwed(
		feeGrowthInside0LastX128,
		position.feeGrowthInside0LastX128,
		position.liquidity,
		burnedAmount0,
		position.tokensOwed0,
	)

	position.tokensOwed1 = updateTokensOwed(
		feeGrowthInside1LastX128,
		position.feeGrowthInside1LastX128,
		position.liquidity,
		burnedAmount1,
		position.tokensOwed1,
	)

	position.feeGrowthInside0LastX128 = feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = feeGrowthInside1LastX128
	position.liquidity = new(u256.Uint).Sub(positionLiquidity, liquidityToRemove)
	mustUpdatePosition(params.tokenId, position)

	collect0, collect1 := pl.Collect(
		pToken0,
		pToken1,
		pFee,
		caller,
		position.tickLower,
		position.tickUpper,
		burn0,
		burn1,
	)

	collectAmount0 := u256.MustFromDecimal(collect0)
	collectAmount1 := u256.MustFromDecimal(collect1)

	underflow := false
	position.tokensOwed0, underflow = new(u256.Uint).SubOverflow(position.tokensOwed0, collectAmount0)
	if underflow {
		panic(newErrorWithDetail(
			errUnderflow,
			"tokensOwed0 underflow",
		))
	}
	position.tokensOwed1, underflow = new(u256.Uint).SubOverflow(position.tokensOwed1, collectAmount1)
	if underflow {
		panic(newErrorWithDetail(
			errUnderflow,
			"tokensOwed1 underflow",
		))
	}
	mustUpdatePosition(params.tokenId, position)

	if position.isClear() {
		burnPosition(params.tokenId) // just update flag (we don't want to burn actual position)
	}

	// NO UNWRAP
	if !params.unwrapResult {
		return params.tokenId, liquidityToRemove, fee0, fee1, collectAmount0, collectAmount1, position.poolKey
	}

	handleUnwrap(pToken0, pToken1, params.unwrapResult, beforeWugnotBalance, caller)

	return params.tokenId, liquidityToRemove, fee0, fee1, collectAmount0, collectAmount1, position.poolKey
}

// CollectFee collects swap fee from the position
// Returns tokenId, afterFee0, afterFee1, poolPath, origFee0, origFee1
// ref: https://docs.gnoswap.io/contracts/position/position.gno#collectfee
func CollectFee(tokenId uint64, unwrapResult bool) (uint64, string, string, string, string, string) {
	assertOnlyNotHalted()
	assertTokenExists(tokenId)
	isAuthorizedForToken(tokenId)

	en.MintAndDistributeGns()

	// verify position
	position := MustGetPosition(tokenId)
	token0, token1, fee := splitOf(position.poolKey)

	pl.Burn(
		token0,
		token1,
		fee,
		position.tickLower,
		position.tickUpper,
		ZERO_LIQUIDITY_FOR_FEE_COLLECTION, // burn '0' liquidity to collect fee
	)

	currentFeeGrowth, err := getCurrentFeeGrowth(position, token0, token1, fee)
	if err != nil {
		panic(newErrorWithDetail(err, "failed to get current fee growth"))
	}

	tokensOwed0, tokensOwed1 := calculateFees(position, currentFeeGrowth)

	position.feeGrowthInside0LastX128 = new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside0LastX128)
	position.feeGrowthInside1LastX128 = new(u256.Uint).Set(currentFeeGrowth.feeGrowthInside1LastX128)

	// check user wugnot amount
	// need this value to unwrap fee
	caller := getPrevAddr()
	userWugnot := wugnot.BalanceOf(caller)

	// collect fee
	amount0, amount1 := pl.Collect(
		token0, token1, fee,
		caller,
		position.tickLower, position.tickUpper,
		tokensOwed0.ToString(), tokensOwed1.ToString(),
	)

	// sometimes there will be a few less uBase amount than expected due to rounding down in core, but we just subtract the full amount expected
	// instead of the actual amount so we can burn the token
	position.tokensOwed0 = new(u256.Uint).Sub(tokensOwed0, u256.MustFromDecimal(amount0))
	position.tokensOwed1 = new(u256.Uint).Sub(tokensOwed1, u256.MustFromDecimal(amount1))
	mustUpdatePosition(tokenId, position)

	// handle withdrawal fee
	withoutFee0, withoutFee1 := pl.HandleWithdrawalFee(
		tokenId,
		token0,
		amount0,
		token1,
		amount1,
		position.poolKey,
		std.PrevRealm().Addr(),
	)

	// UNWRAP
	pToken0, pToken1, _ := splitOf(position.poolKey)
	handleUnwrap(pToken0, pToken1, unwrapResult, userWugnot, caller)

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"CollectSwapFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpTokenId", formatUint(tokenId),
		"feeAmount0", withoutFee0,
		"feeAmount1", withoutFee1,
		"poolPath", position.poolKey,
		"unwrapResult", formatBool(unwrapResult),
	)

	return tokenId, withoutFee0, withoutFee1, position.poolKey, amount0, amount1
}

// calculateFees calculates the fees for the current position.
func calculateFees(position Position, currentFeeGrowth FeeGrowthInside) (*u256.Uint, *u256.Uint) {
	fee0 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside0LastX128,
		position.feeGrowthInside0LastX128,
		position.liquidity,
	)

	fee1 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside1LastX128,
		position.feeGrowthInside1LastX128,
		position.liquidity,
	)

	tokensOwed0 := new(u256.Uint).Add(new(u256.Uint).Set(position.tokensOwed0), fee0)
	tokensOwed1 := new(u256.Uint).Add(new(u256.Uint).Set(position.tokensOwed1), fee1)

	return tokensOwed0, tokensOwed1
}

func getCurrentFeeGrowth(postion Position, token0, token1 string, fee uint32) (FeeGrowthInside, error) {
	pool := pl.GetPoolFromPoolPath(postion.poolKey)
	positionKey := computePositionKey(GetOrigPkgAddr(), postion.tickLower, postion.tickUpper)

	feeGrowthInside0 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

	feeGrowthInside := FeeGrowthInside{
		feeGrowthInside0LastX128: feeGrowthInside0,
		feeGrowthInside1LastX128: feeGrowthInside1,
	}

	return feeGrowthInside, nil
}

// Reposition adjusts the price range and liquidity of an existing position.
// Reposition is executed via multiple messages as follows:
//
// 1. Decrease the liquidity of the existing position.
// 2. Swap tokens as needed to reach the desired token ratio.
// 3. Add new liquidity with the desired price range and amount.
//
// Returns tokenId, liquidity, tickLower, tickUpper, amount0, amount1
// ref: https://docs.gnoswap.io/contracts/position/position.gno#reposition
func Reposition(
	tokenId uint64,
	tickLower int32,
	tickUpper int32,
	amount0DesiredStr string,
	amount1DesiredStr string,
	amount0MinStr string,
	amount1MinStr string,
) (uint64, string, int32, int32, string, string) {
	assertOnlyNotHalted()
	verifyTokenIdAndOwnership(tokenId)

	en.MintAndDistributeGns()

	// position should be burned to reposition
	position := MustGetPosition(tokenId)
	oldTickLower := position.tickLower
	oldTickUpper := position.tickUpper

	if !(position.isClear()) {
		panic(newErrorWithDetail(
			errNotClear,
			ufmt.Sprintf("position(%d) isn't clear(liquidity:%s, tokensOwed0:%s, tokensOwed1:%s)", tokenId, position.liquidity.ToString(), position.tokensOwed0.ToString(), position.tokensOwed1.ToString()),
		))
	}

	caller := getPrevAddr()
	token0, token1, fee := splitOf(position.poolKey)
	token0, token1, _, _, _ = processTokens(token0, token1, amount0DesiredStr, amount1DesiredStr, caller)

	liquidity, amount0, amount1 := addLiquidity(
		AddLiquidityParams{
			poolKey:        position.poolKey,
			tickLower:      tickLower,
			tickUpper:      tickUpper,
			amount0Desired: u256.MustFromDecimal(amount0DesiredStr),
			amount1Desired: u256.MustFromDecimal(amount1DesiredStr),
			amount0Min:     u256.MustFromDecimal(amount0MinStr),
			amount1Min:     u256.MustFromDecimal(amount1MinStr),
			caller:         caller,
		},
	)

	// update position tickLower, tickUpper to new value
	// because getCurrentFeeGrowth() uses tickLower, tickUpper
	position.tickLower = tickLower
	position.tickUpper = tickUpper

	currentFeeGrowth, err := getCurrentFeeGrowth(position, token0, token1, fee)
	if err != nil {
		panic(newErrorWithDetail(err, "failed to get current fee growth"))
	}
	position.feeGrowthInside0LastX128 = currentFeeGrowth.feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = currentFeeGrowth.feeGrowthInside1LastX128

	position.liquidity = liquidity
	// OBS: do not reset feeGrowthInside1LastX128 and feeGrowthInside1LastX128 to zero
	// if so, ( decrease 100% -> reposition )
	// > at this point, that position will have unclaimedFee which isn't intended
	position.tokensOwed0 = u256.Zero()
	position.tokensOwed1 = u256.Zero()
	position.burned = false
	mustUpdatePosition(tokenId, position)

	poolSqrtPriceX96 := pl.PoolGetSlot0SqrtPriceX96(position.poolKey)
	token0Balance := pl.PoolGetBalanceToken0(position.poolKey)
	token1Balance := pl.PoolGetBalanceToken1(position.poolKey)
	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"Reposition",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpTokenId", formatUint(tokenId),
		"tickLower", formatInt(int64(tickLower)),
		"tickUpper", formatInt(int64(tickUpper)),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"prevTickLower", formatInt(int64(oldTickLower)),
		"prevTickUpper", formatInt(int64(oldTickUpper)),
		"poolPath", position.poolKey,
		"sqrtPriceX96", poolSqrtPriceX96,
		"positionLiquidity", PositionGetPositionLiquidityStr(tokenId),
		"token0Balance", token0Balance,
		"token1Balance", token1Balance,
	)

	return tokenId, liquidity.ToString(), tickLower, tickUpper, amount0.ToString(), amount1.ToString()
}

func calculateTokensOwed(
	feeGrowthInsideLastX128 *u256.Uint,
	positionFeeGrowthInsideLastX128 *u256.Uint,
	positionLiquidity *u256.Uint,
) *u256.Uint {
	diff := new(u256.Uint).Sub(feeGrowthInsideLastX128, positionFeeGrowthInsideLastX128)
	return u256.MulDiv(diff, positionLiquidity, u256.MustFromDecimal(consts.Q128))
}

func updateTokensOwed(
	feeGrowthInsideLastX128 *u256.Uint,
	positionFeeGrowthInsideLastX128 *u256.Uint,
	positionLiquidity *u256.Uint,
	burnedAmount *u256.Uint,
	tokensOwed *u256.Uint,
) *u256.Uint {
	additionalTokensOwed := calculateTokensOwed(feeGrowthInsideLastX128, positionFeeGrowthInsideLastX128, positionLiquidity)
	add := new(u256.Uint).Add(burnedAmount, additionalTokensOwed)
	return new(u256.Uint).Add(tokensOwed, add)
}

func burnPosition(tokenId uint64) {
	position := MustGetPosition(tokenId)
	if !(position.isClear()) {
		panic(newErrorWithDetail(
			errNotClear,
			ufmt.Sprintf("position(%d) isn't clear(liquidity:%s, tokensOwed0:%s, tokensOwed1:%s)", tokenId, position.liquidity.ToString(), position.tokensOwed0.ToString(), position.tokensOwed1.ToString()),
		))
	}

	position.burned = true
	mustUpdatePosition(tokenId, position)
}

func calculateLiquidityToRemove(positionLiquidity *u256.Uint, liquidityRatio uint64) *u256.Uint {
	liquidityToRemove := new(u256.Uint).Mul(positionLiquidity, u256.NewUint(liquidityRatio))
	liquidityToRemove = new(u256.Uint).Div(liquidityToRemove, u256.NewUint(100))
	if positionLiquidity.Lt(liquidityToRemove) || liquidityRatio == 100 {
		return positionLiquidity
	}
	return liquidityToRemove
}

func SetPositionOperator(tokenId uint64, operator std.Address) {
	caller := getPrevRealm().PkgPath()
	if caller != consts.STAKER_PATH {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("caller(%s) is not staker", caller),
		))
	}

	position := MustGetPosition(tokenId)
	position.operator = operator
	mustUpdatePosition(tokenId, position)
}
