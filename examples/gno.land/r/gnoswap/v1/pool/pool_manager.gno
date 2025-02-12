package pool

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

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

// createPoolParams holds the essential parameters for creating a new pool.
type createPoolParams struct {
	token0Path, token1Path string
	fee                    uint32
	sqrtPriceX96           *u256.Uint
	tickSpacing            int32
}

// newPoolParams defines the essential parameters for creating a new pool.
func newPoolParams(
	token0Path string,
	token1Path string,
	fee uint32,
	sqrtPriceX96 string,
) *createPoolParams {
	price := u256.MustFromDecimal(sqrtPriceX96)
	tickSpacing := GetFeeAmountTickSpacing(fee)
	return &createPoolParams{
		token0Path:   token0Path,
		token1Path:   token1Path,
		fee:          fee,
		sqrtPriceX96: price,
		tickSpacing:  tickSpacing,
	}
}

func (p *createPoolParams) updateWithWrapping() *createPoolParams {
	token0Path, token1Path := p.wrap()

	if !p.isInOrder() {
		token0Path, token1Path = token1Path, token0Path

		oldTick := common.TickMathGetTickAtSqrtRatio(p.sqrtPriceX96)
		if oldTick > consts.MAX_TICK || oldTick < consts.MIN_TICK {
			panic(addDetailToError(
				errInvalidTickRange,
				ufmt.Sprintf("expected tick(%d) to be within range", oldTick),
			))
		}
		newPrice := common.TickMathGetSqrtRatioAtTick(oldTick * int32(-1))
		if newPrice.IsZero() {
			panic(addDetailToError(
				errOverFlow,
				ufmt.Sprintf("expected newPrice(%s) to be non-zero", newPrice.ToString()),
			))
		}
		p.sqrtPriceX96 = new(u256.Uint).Set(newPrice)
	}

	return newPoolParams(token0Path, token1Path, p.fee, p.sqrtPriceX96.ToString())
}

func (p *createPoolParams) isSameTokenPath() bool {
	return p.token0Path == p.token1Path
}

// isInOrder checks if token paths are in lexicographical (or, alphabetical) order
func (p *createPoolParams) isInOrder() bool {
	if strings.Compare(p.token0Path, p.token1Path) < 0 {
		return true
	}

	return false
}

func (p *createPoolParams) wrap() (string, string) {
	if p.token0Path == consts.GNOT {
		p.token0Path = consts.WRAPPED_WUGNOT
	}
	if p.token1Path == consts.GNOT {
		p.token1Path = consts.WRAPPED_WUGNOT
	}

	return p.token0Path, p.token1Path
}

func (p *createPoolParams) Token0Path() string {
	return p.token0Path
}

func (p *createPoolParams) Token1Path() string {
	return p.token1Path
}

func (p *createPoolParams) Fee() uint32 {
	return p.fee
}

func (p *createPoolParams) TickSpacing() int32 {
	return p.tickSpacing
}

func (p *createPoolParams) SqrtPriceX96() *u256.Uint {
	return p.sqrtPriceX96
}

// CreatePool creates a new concentrated liquidity pool with the given parameters.
// It mints and distributes GNS tokens, validates the input parameters, and creates a new pool.
// If GNOT is used as one of the tokens, it is automatically wrapped to WUGNOT.
// The function ensures that token0Path is lexicographically smaller than token1Path.
// ref: https://docs.gnoswap.io/contracts/pool/pool_manager.gno#createpool
func CreatePool(
	token0Path string,
	token1Path string,
	fee uint32,
	sqrtPriceX96 string,
) {
	assertOnlyNotHalted()

	poolInfo := newPoolParams(token0Path, token1Path, fee, sqrtPriceX96)
	poolInfo = poolInfo.updateWithWrapping()
	if poolInfo.isSameTokenPath() {
		panic(addDetailToError(
			errDuplicateTokenInPool,
			ufmt.Sprintf(
				"expected token0Path(%s) != token1Path(%s)",
				poolInfo.token0Path, poolInfo.token1Path,
			),
		))
	}
	en.MintAndDistributeGns()

	// wrap first
	token0Path, token1Path = poolInfo.wrap()
	poolPath := GetPoolPath(token0Path, token1Path, fee)

	// reinitialize poolInfo with wrapped tokens
	poolInfo = newPoolParams(token0Path, token1Path, fee, sqrtPriceX96)

	// then check if token0Path == token1Path
	if poolInfo.isSameTokenPath() {
		panic(addDetailToError(
			errDuplicateTokenInPool,
			ufmt.Sprintf(
				"expected token0Path(%s) != token1Path(%s)",
				token0Path, token1Path,
			),
		))
	}

	if !poolInfo.isInOrder() {
		panic(addDetailToError(
			errTokenSortOrder,
			ufmt.Sprintf("expected token0Path(%s) < token1Path(%s)", token0Path, token1Path),
		))
	}

	prevAddr, prevRealm := getPrevAsString()

	// check whether the pool already exist
	pool, exist := pools.Get(poolPath)
	if exist {
		panic(addDetailToError(
			errPoolAlreadyExists,
			ufmt.Sprintf("expected poolPath(%s) not to exist", poolPath),
		))
	}

	if poolCreationFee > 0 {
		gns.TransferFrom(std.PrevRealm().Addr(), consts.PROTOCOL_FEE_ADDR, poolCreationFee)
		pf.AddToProtocolFee(consts.GNS_PATH, poolCreationFee)

		std.Emit(
			"PoolCreationFee",
			"prevAddr", prevAddr,
			"prevRealm", prevRealm,
			"poolPath", poolPath,
			"feeTokenPath", consts.GNS_PATH,
			"feeAmount", formatUint(poolCreationFee),
		)
	}

	pool = newPool(poolInfo)
	pools.Set(poolPath, pool)

	currentTick := common.TickMathGetTickAtSqrtRatio(u256.MustFromDecimal(sqrtPriceX96))

	std.Emit(
		"CreatePool",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"token0Path", token0Path,
		"token1Path", token1Path,
		"fee", formatUint(uint64(fee)),
		"sqrtPriceX96", sqrtPriceX96,
		"poolPath", poolPath,
		"tick", formatInt(int64(currentTick)),
		"tickSpacing", formatInt(int64(poolInfo.TickSpacing())),
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
		panic(addDetailToError(
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
	feeStr := strconv.FormatUint(uint64(fee), 10)
	iTickSpacing, exist := feeAmountTickSpacing.Get(feeStr)
	if !exist {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("expected feeAmountTickSpacing(%s) to exist", feeStr),
		))
	}

	return iTickSpacing.(int32)
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
	feeStr := strconv.FormatUint(uint64(fee), 10)
	feeAmountTickSpacing.Set(feeStr, tickSpacing)
}
