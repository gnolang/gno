package pool

import (
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"
)

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

type Slot0 struct {
	sqrtPriceX96 *u256.Uint // current price of the pool as a sqrt(token1/token0) Q96 value
	tick         int32      // current tick of the pool, i.e according to the last tick transition that was run
	feeProtocol  uint8      // protocol fee for both tokens of the pool
	unlocked     bool       // whether the pool is currently locked to reentrancy
}

func (s *Slot0) Tick() int32        { return s.tick }
func (s *Slot0) FeeProtocol() uint8 { return s.feeProtocol }

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

// type Pool describes a single Pool's state
// A pool is identificed with a unique key (token0, token1, fee), where token0 < token1
type Pool struct {
	// token0/token1 path of the pool
	token0Path           string
	token1Path           string
	fee                  uint32 // fee tier of the pool
	tickSpacing          int32  // spacing between ticks
	slot0                Slot0
	balances             Balances // balances of the pool
	protocolFees         ProtocolFees
	maxLiquidityPerTick  *u256.Uint // the maximum amount of liquidity that can be added per tick
	feeGrowthGlobal0X128 *u256.Uint // uint256
	feeGrowthGlobal1X128 *u256.Uint // uint256
	liquidity            *u256.Uint // total amount of liquidity in the pool
	ticks                *avl.Tree  // tick(int32) -> TickInfo
	tickBitmaps          *avl.Tree  // tick(wordPos)(int16) -> bitMap(tickWord ^ mask)(*u256.Uint)
	positions            *avl.Tree  // maps the key (caller, lower tick, upper tick) to a unique position
}

func (p *Pool) PoolPath() string                 { return GetPoolPath(p.token0Path, p.token1Path, p.fee) }
func (p *Pool) Token0Path() string               { return p.token0Path }
func (p *Pool) Token1Path() string               { return p.token1Path }
func (p *Pool) Fee() uint32                      { return p.fee }
func (p *Pool) BalanceToken0() *u256.Uint        { return p.balances.token0 }
func (p *Pool) BalanceToken1() *u256.Uint        { return p.balances.token1 }
func (p *Pool) TickSpacing() int32               { return p.tickSpacing }
func (p *Pool) MaxLiquidityPerTick() *u256.Uint  { return p.maxLiquidityPerTick }
func (p *Pool) Slot0() Slot0                     { return p.slot0 }
func (p *Pool) Slot0SqrtPriceX96() *u256.Uint    { return p.slot0.sqrtPriceX96 }
func (p *Pool) Slot0Tick() int32                 { return p.slot0.tick }
func (p *Pool) Slot0FeeProtocol() uint8          { return p.slot0.feeProtocol }
func (p *Pool) Slot0Unlocked() bool              { return p.slot0.unlocked }
func (p *Pool) FeeGrowthGlobal0X128() *u256.Uint { return p.feeGrowthGlobal0X128 }
func (p *Pool) FeeGrowthGlobal1X128() *u256.Uint { return p.feeGrowthGlobal1X128 }
func (p *Pool) ProtocolFeesToken0() *u256.Uint   { return p.protocolFees.token0 }
func (p *Pool) ProtocolFeesToken1() *u256.Uint   { return p.protocolFees.token1 }
func (p *Pool) Liquidity() *u256.Uint            { return p.liquidity }

func (p *Pool) Ticks() string {
	if p.ticks == nil {
		return "[]"
	}

	tickInfoStrings := []string{}

	p.ticks.Iterate("", "", func(tickKey string, tickValue any) bool {
		tick, _ := strconv.Atoi(tickKey)
		tickInfo := tickValue.(TickInfo)

		tickInfoStrings = append(tickInfoStrings, ufmt.Sprintf(
			`{"tick":%d,"feeGrowthOutside0X128":"%s","feeGrowthOutside1X128":"%s"}`,
			tick,
			tickInfo.feeGrowthOutside0X128.ToString(),
			tickInfo.feeGrowthOutside1X128.ToString(),
		))

		return false
	})

	return "[" + strings.Join(tickInfoStrings, ",") + "]"
}
