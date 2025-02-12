package pool

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

type Slot0 struct {
	sqrtPriceX96 *u256.Uint // current price of the pool as a sqrt(token1/token0) Q96 value
	tick         int32      // current tick of the pool, i.e according to the last tick transition that was run
	feeProtocol  uint8      // protocol fee for both tokens of the pool
	unlocked     bool       // whether the pool is currently locked to reentrancy
}

func newSlot0(
	sqrtPriceX96 *u256.Uint,
	tick int32,
	feeProtocol uint8,
	unlocked bool,
) Slot0 {
	return Slot0{
		sqrtPriceX96: sqrtPriceX96,
		tick:         tick,
		feeProtocol:  feeProtocol,
		unlocked:     unlocked,
	}
}

type Balances struct {
	// current balance of the pool in token0/token1
	token0 *u256.Uint
	token1 *u256.Uint
}

func newBalances() Balances {
	return Balances{
		token0: u256.Zero(),
		token1: u256.Zero(),
	}
}

type ProtocolFees struct {
	// current protocol fees of the pool in token0/token1
	token0 *u256.Uint
	token1 *u256.Uint
}

func newProtocolFees() ProtocolFees {
	return ProtocolFees{
		token0: u256.Zero(),
		token1: u256.Zero(),
	}
}

// ModifyPositionParams repersents the parameters for modifying a liquidity position.
// This structure is used internally both `Mint` and `Burn` operation to manage
// the liquidity positions.
type ModifyPositionParams struct {
	// owner is the address that owns the position
	owner std.Address

	// tickLower and atickUpper define the price range
	// The actual price range is calculated as 1.0001^tick
	// This allows for precision in price range while using integer math.

	tickLower int32 // lower tick of the position
	tickUpper int32 // upper tick of the position

	// liquidityDelta represents the change in liquidity
	// Positive for minting, negative for burning
	liquidityDelta *i256.Int
}

// newModifyPositionParams creates a new `ModifyPositionParams` instance.
// This is used to preare parameters for the `modifyPosition` function,
// which handles both minting and burning of liquidity positions.
//
// Parameters:
//   - owner: address that will own (or owns) the position
//   - tickLower: lower tick bound of the position
//   - tickUpper: upper tick bound of the position
//   - liquidityDelta: amount of liquidity to add (positive) or remove (negative)
//
// The tick parameters represent prices as powers of 1.0001:
// - actual_price = 1.0001^tick
// - For example, tick = 100 means price = 1.0001^100
//
// Returns:
//   - ModifyPositionParams: a new instance of ModifyPositionParams
func newModifyPositionParams(
	owner std.Address,
	tickLower int32,
	tickUpper int32,
	liquidityDelta *i256.Int,
) ModifyPositionParams {
	return ModifyPositionParams{
		owner:          owner,
		tickLower:      tickLower,
		tickUpper:      tickUpper,
		liquidityDelta: liquidityDelta,
	}
}

// SwapCache holds data that remains constant throughout a swap.
type SwapCache struct {
	feeProtocol    uint8      // protocol fee for the input token
	liquidityStart *u256.Uint // liquidity at the beginning of the swap
}

func newSwapCache(
	feeProtocol uint8,
	liquidityStart *u256.Uint,
) SwapCache {
	return SwapCache{
		feeProtocol:    feeProtocol,
		liquidityStart: liquidityStart,
	}
}

// SwapState tracks the changing values during a swap.
// This type helps manage the state transiktions that occur as the swap progresses
// accross different price ranges.
type SwapState struct {
	amountSpecifiedRemaining *i256.Int  // amount remaining to be swapped in/out of the input/output token
	amountCalculated         *i256.Int  // amount already swapped out/in of the output/input token
	sqrtPriceX96             *u256.Uint // current sqrt(price)
	tick                     int32      // tick associated with the current sqrt(price)
	feeGrowthGlobalX128      *u256.Uint // global fee growth of the input token
	protocolFee              *u256.Uint // amount of input token paid as protocol fee
	liquidity                *u256.Uint // current liquidity in range
}

func newSwapState(
	amountSpecifiedRemaining *i256.Int,
	feeGrowthGlobalX128 *u256.Uint,
	liquidity *u256.Uint,
	slot0 Slot0,
) SwapState {
	return SwapState{
		amountSpecifiedRemaining: amountSpecifiedRemaining,
		amountCalculated:         i256.Zero(),
		sqrtPriceX96:             slot0.sqrtPriceX96,
		tick:                     slot0.tick,
		feeGrowthGlobalX128:      feeGrowthGlobalX128,
		protocolFee:              u256.Zero(),
		liquidity:                liquidity,
	}
}

func (s *SwapState) SetSqrtPriceX96(sqrtPriceX96 string) {
	s.sqrtPriceX96 = u256.MustFromDecimal(sqrtPriceX96)
}

func (s *SwapState) SetTick(tick int32) {
	s.tick = tick
}

func (s *SwapState) SetFeeGrowthGlobalX128(feeGrowthGlobalX128 *u256.Uint) {
	s.feeGrowthGlobalX128 = feeGrowthGlobalX128
}

func (s *SwapState) SetProtocolFee(fee *u256.Uint) {
	s.protocolFee = fee
}

// StepComputations holds intermediate values used during a single step of a swap.
// Each step represents movement from the current tick to the next initialized tick
// or the target price, whichever comes first.
type StepComputations struct {
	sqrtPriceStartX96 *u256.Uint // price at the beginning of the step
	tickNext          int32      // next tick to swap to from the current tick in the swap direction
	initialized       bool       // whether tickNext is initialized
	sqrtPriceNextX96  *u256.Uint // sqrt(price) for the next tick (token1/token0) Q96
	amountIn          *u256.Uint // how much being swapped in this step
	amountOut         *u256.Uint // how much is being swapped out in this step
	feeAmount         *u256.Uint // how much fee is being paid in this step
}

// init initializes the computation for a single swap step
func (step *StepComputations) initSwapStep(state SwapState, pool *Pool, zeroForOne bool) {
	step.sqrtPriceStartX96 = state.sqrtPriceX96
	step.tickNext, step.initialized = pool.tickBitmapNextInitializedTickWithInOneWord(
		state.tick,
		pool.tickSpacing,
		zeroForOne,
	)

	// prevent overshoot the min/max tick
	step.clampTickNext()

	// get the price for the next tick
	step.sqrtPriceNextX96 = common.TickMathGetSqrtRatioAtTick(step.tickNext)
}

// clampTickNext ensures that `tickNext` stays within the min, max tick boundaries
// as the tick bitmap is not aware of these bounds
func (step *StepComputations) clampTickNext() {
	if step.tickNext < consts.MIN_TICK {
		step.tickNext = consts.MIN_TICK
	} else if step.tickNext > consts.MAX_TICK {
		step.tickNext = consts.MAX_TICK
	}
}

type PositionInfo struct {
	liquidity *u256.Uint // amount of liquidity owned by this position

	// Fee growth per unit of liquidity as of the last update
	// Used to calculate uncollected fees for token0
	feeGrowthInside0LastX128 *u256.Uint

	// Fee growth per unit of liquidity as of the last update
	// Used to calculate uncollected fees for token1
	feeGrowthInside1LastX128 *u256.Uint

	// accumulated fees in token0 waiting to be collected
	tokensOwed0 *u256.Uint

	// accumulated fees in token1 waiting to be collected
	tokensOwed1 *u256.Uint
}

// valueOrZero initializes nil fields in PositionInfo to zero.
//
// This function ensures that all numeric fields in the PositionInfo struct are not nil.
// If a field is nil, it is replaced with a zero value, maintaining consistency and preventing
// potential null pointer issues during calculations.
//
// Fields affected:
//   - liquidity: The liquidity amount associated with the position.
//   - feeGrowthInside0LastX128: Fee growth for token 0 inside the tick range, last recorded value.
//   - feeGrowthInside1LastX128: Fee growth for token 1 inside the tick range, last recorded value.
//   - tokensOwed0: The amount of token 0 owed to the position owner.
//   - tokensOwed1: The amount of token 1 owed to the position owner.
//
// Behavior:
//   - If a field is nil, it is set to its equivalent zero value.
//   - If a field already has a value, it remains unchanged.
//
// Example:
//
//	position := &PositionInfo{}
//	position.valueOrZero()
//	fmt.Println(position.liquidity) // Output: 0
//
// Notes:
//   - This function is useful for ensuring numeric fields are properly initialized
//     before performing operations or calculations.
//   - Prevents runtime errors caused by nil values.
func (p *PositionInfo) valueOrZero() {
	p.liquidity = p.liquidity.NilToZero()
	p.feeGrowthInside0LastX128 = p.feeGrowthInside0LastX128.NilToZero()
	p.feeGrowthInside1LastX128 = p.feeGrowthInside1LastX128.NilToZero()
	p.tokensOwed0 = p.tokensOwed0.NilToZero()
	p.tokensOwed1 = p.tokensOwed1.NilToZero()
}

// TickInfo stores information about a specific tick in the pool.
// TIcks represent discrete price points that can be used as boundaries for positions.
type TickInfo struct {
	liquidityGross *u256.Uint // total position liquidity that references this tick
	liquidityNet   *i256.Int  // amount of net liquidity added (subtracted) when tick is crossed from left to right (right to left)

	// fee growth per unit of liquidity on the _other_ side of this tick (relative to the current tick)
	// only has relative meaning, not absolute — the value depends on when the tick is initialized
	feeGrowthOutside0X128 *u256.Uint
	feeGrowthOutside1X128 *u256.Uint

	tickCumulativeOutside int64 // cumulative tick value on the other side of the tick

	// the seconds per unit of liquidity on the _other_ side of this tick (relative to the current tick)
	// only has relative meaning, not absolute — the value depends on when the tick is initialized
	secondsPerLiquidityOutsideX128 *u256.Uint

	// the seconds spent on the other side of the tick (relative to the current tick)
	// only has relative meaning, not absolute — the value depends on when the tick is initialized
	secondsOutside uint32

	initialized bool // whether the tick is initialized
}

// valueOrZero ensures that all fields of TickInfo are valid by setting nil fields to zero,
// while retaining existing values if they are not nil.
// This function updates the TickInfo struct to replace any nil values in its fields
// with their respective zero values, ensuring data consistency.
//
// Behavior:
// - If a field is nil, it is replaced with its zero value.
// - If a field already has a valid value, the value remains unchanged.
//
// Fields:
// - liquidityGross: Gross liquidity for the tick, set to zero if nil, otherwise retains its value.
// - liquidityNet: Net liquidity for the tick, set to zero if nil, otherwise retains its value.
// - feeGrowthOutside0X128: Accumulated fee growth for token0 outside the tick, set to zero if nil, otherwise retains its value.
// - feeGrowthOutside1X128: Accumulated fee growth for token1 outside the tick, set to zero if nil, otherwise retains its value.
// - secondsPerLiquidityOutsideX128: Time per liquidity outside the tick, set to zero if nil, otherwise retains its value.
//
// Use Case:
// This function ensures all numeric fields in TickInfo are non-nil and have valid values,
// preventing potential runtime errors caused by nil values during operations like arithmetic or comparisons.
func (t *TickInfo) valueOrZero() {
	t.liquidityGross = t.liquidityGross.NilToZero()
	t.liquidityNet = t.liquidityNet.NilToZero()
	t.feeGrowthOutside0X128 = t.feeGrowthOutside0X128.NilToZero()
	t.feeGrowthOutside1X128 = t.feeGrowthOutside1X128.NilToZero()
	t.secondsPerLiquidityOutsideX128 = t.secondsPerLiquidityOutsideX128.NilToZero()
}

// type Pool describes a single Pool's state
// A pool is identificed with a unique key (token0, token1, fee), where token0 < token1
type Pool struct {
	// token0/token1 path of the pool
	token0Path string
	token1Path string

	balances Balances // balances of the pool

	fee uint32 // fee tier of the pool

	tickSpacing int32 // spacing between ticks

	maxLiquidityPerTick *u256.Uint // the maximum amount of liquidity that can be added per tick

	slot0 Slot0

	feeGrowthGlobal0X128 *u256.Uint // uint256
	feeGrowthGlobal1X128 *u256.Uint // uint256

	protocolFees ProtocolFees

	liquidity *u256.Uint // total amount of liquidity in the pool

	ticks *avl.Tree // tick(int32) -> TickInfo

	tickBitmaps *avl.Tree // tick(wordPos)(int16) -> bitMap(tickWord ^ mask)(*u256.Uint)

	positions *avl.Tree // maps the key (caller, lower tick, upper tick) to a unique position

	tickCrossHook func(poolPath string, tickId int32, zeroForOne bool)
}

var tickCrossHook func(poolPath string, tickId int32, zeroForOne bool)

func SetTickCrossHook(hook func(poolPath string, tickId int32, zeroForOne bool)) {
	if tickCrossHook != nil {
		panic("tickCrossHook already set")
	}
	tickCrossHook = hook
}

func newPool(poolInfo *createPoolParams) *Pool {
	maxLiquidityPerTick := calculateMaxLiquidityPerTick(poolInfo.tickSpacing)
	tick := common.TickMathGetTickAtSqrtRatio(poolInfo.SqrtPriceX96())
	slot0 := newSlot0(poolInfo.SqrtPriceX96(), tick, slot0FeeProtocol, true)

	return &Pool{
		token0Path:           poolInfo.Token0Path(),
		token1Path:           poolInfo.Token1Path(),
		balances:             newBalances(),
		fee:                  poolInfo.Fee(),
		tickSpacing:          poolInfo.TickSpacing(),
		maxLiquidityPerTick:  maxLiquidityPerTick,
		slot0:                slot0,
		feeGrowthGlobal0X128: u256.Zero(),
		feeGrowthGlobal1X128: u256.Zero(),
		protocolFees:         newProtocolFees(),
		liquidity:            u256.Zero(),
		ticks:                avl.NewTree(),
		tickBitmaps:          avl.NewTree(),
		positions:            avl.NewTree(),
	}
}

func (p *Pool) PoolPath() string {
	return GetPoolPath(p.token0Path, p.token1Path, p.fee)
}

func (p *Pool) Token0Path() string {
	return p.token0Path
}

func (p *Pool) Token1Path() string {
	return p.token1Path
}

func (p *Pool) Fee() uint32 {
	return p.fee
}

func (p *Pool) BalanceToken0() *u256.Uint {
	return p.balances.token0
}

func (p *Pool) BalanceToken1() *u256.Uint {
	return p.balances.token1
}

func (p *Pool) TickSpacing() int32 {
	return p.tickSpacing
}

func (p *Pool) MaxLiquidityPerTick() *u256.Uint {
	return p.maxLiquidityPerTick
}

func (p *Pool) Slot0() Slot0 {
	return p.slot0
}

func (p *Pool) Slot0SqrtPriceX96() *u256.Uint {
	return p.slot0.sqrtPriceX96
}

func (p *Pool) Slot0Tick() int32 {
	return p.slot0.tick
}

func (p *Pool) Slot0FeeProtocol() uint8 {
	return p.slot0.feeProtocol
}

func (p *Pool) Slot0Unlocked() bool {
	return p.slot0.unlocked
}

func (p *Pool) FeeGrowthGlobal0X128() *u256.Uint {
	return p.feeGrowthGlobal0X128
}

func (p *Pool) FeeGrowthGlobal1X128() *u256.Uint {
	return p.feeGrowthGlobal1X128
}

func (p *Pool) ProtocolFeesToken0() *u256.Uint {
	return p.protocolFees.token0
}

func (p *Pool) ProtocolFeesToken1() *u256.Uint {
	return p.protocolFees.token1
}

func (p *Pool) Liquidity() *u256.Uint {
	return p.liquidity
}

func mustGetPool(poolPath string) *Pool {
	pool := GetPoolFromPoolPath(poolPath)
	if pool == nil {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("expected poolPath(%s) to exist", poolPath),
		))
	}
	return pool
}
