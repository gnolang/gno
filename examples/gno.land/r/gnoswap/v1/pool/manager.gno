package pool

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	en "gno.land/r/gnoswap/v1/emission"
	pf "gno.land/r/gnoswap/v1/protocol_fee"

	"gno.land/r/gnoswap/v1/gns"
)

var (
	feeAmountTickSpacing = avl.NewTree() // feeBps(uint32) -> tickSpacing(int32)
	pools                = avl.NewTree() // poolPath -> *Pool

	slot0FeeProtocol uint8 = 0
)

func init() {
	setFeeAmountTickSpacing(100, 1)     // 0.01%
	setFeeAmountTickSpacing(500, 10)    // 0.05%
	setFeeAmountTickSpacing(3000, 60)   // 0.3%
	setFeeAmountTickSpacing(10000, 200) // 1%
}

// CreatePool creates a new concentrated liquidity pool with the given parameters.
// It mints and distributes GNS tokens, validates the input parameters, and creates a new pool.
// If GNOT is used as one of the tokens, it is automatically wrapped to WUGNOT.
// The function ensures that token0Path is lexicographically smaller than token1Path.
// ref: https://docs.gnoswap.io/contracts/pool/pool_manager.gno#createpool
func CreatePool(
	cur realm,
	token0Path string,
	token1Path string,
	fee uint32,
	sqrtPriceX96 string,
) {
	assertOnlyNotHalted()
	poolInfo := createPool(token0Path, token1Path, fee, sqrtPriceX96, GetFeeAmountTickSpacing(fee))
	en.MintAndDistributeGns(cross)

	poolPath := poolInfo.generatePoolPath()
	// check whether the pool already exist
	if DoesPoolPathExist(poolPath) {
		panic(newErrorWithDetail(
			errPoolAlreadyExists,
			ufmt.Sprintf("expected poolPath(%s) not to exist", poolPath),
		))
	}

	prevAddr, prevRealm := getPrevAsString()

	if poolCreationFee > 0 {
		gns.TransferFrom(cross, std.PreviousRealm().Address(), protocolFeeAddr, poolCreationFee)
		pf.AddToProtocolFee(cross, consts.GNS_PATH, uint64(poolCreationFee))

		std.Emit(
			"PoolCreationFee",
			"prevAddr", prevAddr,
			"prevRealm", prevRealm,
			"poolPath", poolPath,
			"feeTokenPath", consts.GNS_PATH,
			"feeAmount", formatInt(poolCreationFee),
		)
	}

	pool := newPool(poolInfo)
	pools.Set(poolPath, pool)

	currentTick := common.TickMathGetTickAtSqrtRatio(poolInfo.SqrtPriceX96())

	std.Emit(
		"CreatePool",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"token0Path", token0Path,
		"token1Path", token1Path,
		"fee", formatUint(fee),
		"sqrtPriceX96", sqrtPriceX96,
		"poolPath", poolPath,
		"tick", formatInt(currentTick),
		"tickSpacing", formatInt(poolInfo.TickSpacing()),
	)
}

// DoesPoolPathExist checks if a pool exists for the given poolPath.
// The poolPath is a unique identifier for a pool, combining token paths and fee.
func DoesPoolPathExist(poolPath string) bool {
	return pools.Has(poolPath)
}

// GetPool retrieves a pool instance based on the provided token paths and fee tier.
//
// This function determines the pool path by combining the paths of token0 and token1 along with the fee tier,
// and then retrieves the corresponding pool instance using that path.
//
// Parameters:
// - token0Path (string): The unique path for token0.
// - token1Path (string): The unique path for token1.
// - fee (uint32): The fee tier for the pool, expressed in basis points (e.g., 3000 for 0.3%).
//
// Returns:
// - *Pool: A pointer to the Pool instance corresponding to the provided tokens and fee tier.
//
// Notes:
// - The order of token paths (token0Path and token1Path) matters and should match the pool's configuration.
// - Ensure that the tokens and fee tier provided are valid and registered in the system.
//
// Example:
// pool := GetPool("gno.land/r/demo/wugnot", "gno.land/r/gnoswap/v1/gns", 3000)
func GetPool(token0Path, token1Path string, fee uint32) *Pool {
	poolPath := GetPoolPath(token0Path, token1Path, fee)
	return GetPoolFromPoolPath(poolPath)
}

// GetPoolFromPoolPath retrieves a pool instance based on the provided pool path.
//
// This function checks if a pool exists for the given poolPath in the `pools` mapping.
// If the pool exists, it returns the pool instance. Otherwise, it panics with a descriptive error.
//
// Parameters:
// - poolPath (string): The unique identifier or path for the pool.
//
// Returns:
// - *Pool: A pointer to the Pool instance corresponding to the given poolPath.
//
// Panics:
//   - If the `poolPath` does not exist in the `pools` mapping, it panics with an error message
//     indicating that the expected poolPath was not found.
//
// Notes:
// - Ensure that the `poolPath` provided is valid and corresponds to an existing pool in the `pools` mapping.
//
// Example:
// pool := GetPoolFromPoolPath("path/to/pool")
func GetPoolFromPoolPath(poolPath string) *Pool {
	iPool, exist := pools.Get(poolPath)
	if !exist {
		panic(newErrorWithDetail(
			errDataNotFound,
			ufmt.Sprintf("expected poolPath(%s) to exist", poolPath),
		))
	}
	return iPool.(*Pool)
}

// GetPoolPath generates a unique pool path string based on the token paths and fee tier.
//
// This function ensures that the token paths are registered and sorted in alphabetical order
// before combining them with the fee tier to create a unique identifier for the pool.
//
// Parameters:
// - token0Path (string): The unique identifier or path for token0.
// - token1Path (string): The unique identifier or path for token1.
// - fee (uint32): The fee tier for the pool, expressed in basis points (e.g., 3000 for 0.3%).
//
// Returns:
// - string: A unique pool path string in the format "token0Path:token1Path:fee".
//
// Notes:
//   - The function validates that both `token0Path` and `token1Path` are registered in the system
//     using `common.MustRegistered`.
//   - The token paths are sorted alphabetically to ensure consistent pool path generation, regardless
//     of the input order.
//   - This sorting guarantees that the pool path remains deterministic for the same pair of tokens and fee.
//
// Example:
// poolPath := GetPoolPath("path/to/token0", "path/to/token1", 3000)
// // Output: "path/to/token0:path/to/token1:3000"
func GetPoolPath(token0Path, token1Path string, fee uint32) string {
	assertOnlyRegistered(token0Path)
	assertOnlyRegistered(token1Path)

	// all the token paths in the pool are sorted in alphabetical order.
	if strings.Compare(token1Path, token0Path) < 0 {
		token0Path, token1Path = token1Path, token0Path
	}
	return ufmt.Sprintf("%s:%s:%d", token0Path, token1Path, fee)
}

// GetFeeAmountTickSpacing retrieves the tick spacing associated with a given fee amount.
// The tick spacing determines the minimum distance between ticks in the pool.
//
// Parameters:
//   - fee (uint32): The fee tier in basis points (e.g., 3000 for 0.3%)
//
// Returns:
//   - int32: The tick spacing value for the given fee tier
//
// Panics:
//   - If the fee amount is not registered in feeAmountTickSpacing
func GetFeeAmountTickSpacing(fee uint32) int32 {
	feeStr := formatUint(fee)
	iTickSpacing, exist := feeAmountTickSpacing.Get(feeStr)
	if !exist {
		panic(newErrorWithDetail(
			errUnsupportedFeeTier,
			ufmt.Sprintf("expected fee(%d) to be one of %d, %d, %d, %d", fee, FeeTier100, FeeTier500, FeeTier3000, FeeTier10000),
		))
	}

	return iTickSpacing.(int32)
}

// SetFeeProtocolByAdmin sets the fee protocol for all pools
// Also it will be applied to new created pools
func SetFeeProtocolByAdmin(
	cur realm,
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
func SetFeeProtocol(cur realm, feeProtocol0, feeProtocol1 uint8) {
	assertOnlyGovernance()
	setFeeProtocolInternal(feeProtocol0, feeProtocol1, "SetFeeProtocol")
}

// setFeeAmountTickSpacing associates a tick spacing value with a fee amount.
// This is typically called during initialization to set up supported fee tiers.
//
// Parameters:
//   - fee (uint32): The fee tier in basis points (e.g., 3000 for 0.3%)
//   - tickSpacing (int32): The minimum tick spacing for this fee tier
//
// Note: Smaller tick spacing allows for more granular price points but increases
// computational overhead. Higher fee tiers typically use larger tick spacing.
func setFeeAmountTickSpacing(fee uint32, tickSpacing int32) {
	feeStr := formatUint(fee)
	feeAmountTickSpacing.Set(feeStr, tickSpacing)
}

// mustGetPool retrieves a pool instance by its path and ensures it exists.
//
// This function attempts to fetch the pool using GetPoolFromPoolPath().
// If the pool does not exist, it panics with a detailed error message.
//
// Parameters:
//   - poolPath: The unique identifier for the pool.
//
// Returns:
//   - *Pool: A valid Pool instance.
//
// Panics:
//   - If the pool does not exist, an error is thrown with the message
//     "expected poolPath(<poolPath>) to exist".
func mustGetPool(poolPath string) *Pool {
	pool := GetPoolFromPoolPath(poolPath)
	if pool == nil {
		panic(newErrorWithDetail(
			errDataNotFound,
			ufmt.Sprintf("expected poolPath(%s) to exist", poolPath),
		))
	}
	return pool
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
		"prevFeeProtocol0", formatUint(feeProtocol0Old),
		"prevFeeProtocol1", formatUint(feeProtocol1Old),
		"feeProtocol0", formatUint(feeProtocol0),
		"feeProtocol1", formatUint(feeProtocol1),
		"newFee", formatUint(newFee),
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
		panic(newErrorWithDetail(
			err,
			ufmt.Sprintf("expected (feeProtocol0(%d) == 0 || (feeProtocol0(%d) >= 4 && feeProtocol0(%d) <= 10)) && (feeProtocol1(%d) == 0 || (feeProtocol1(%d) >= 4 && feeProtocol1(%d) <= 10))", feeProtocol0, feeProtocol0, feeProtocol0, feeProtocol1, feeProtocol1, feeProtocol1),
		))
	}
	// combine both protocol fee into a single byte:
	// - feePrtocol0 occupies the lower 4 bits
	// - feeProtocol1 is shifted the lower 4 positions to occupy the upper 4 bits
	newFee := feeProtocol0 + (feeProtocol1 << 4) // ( << 4 ) = ( * 16 )

	// Update slot0 for each pool
	pools.Iterate("", "", func(poolPath string, iPool any) bool {
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
