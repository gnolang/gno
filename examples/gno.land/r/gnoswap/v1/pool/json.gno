package pool

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"

	u256 "gno.land/p/gnoswap/uint256"
)

// JsonResponse is an interface that all JSON response types must implement.
type JsonResponse interface {
	JSON() *json.Node
}

type RpcPool struct {
	PoolPath string `json:"poolPath"`

	Token0Path string `json:"token0Path"`
	Token1Path string `json:"token1Path"`

	Token0Balance string `json:"token0Balance"`
	Token1Balance string `json:"token1Balance"`

	Fee uint32 `json:"fee"`

	TickSpacing int32 `json:"tickSpacing"`

	MaxLiquidityPerTick string `json:"maxLiquidityPerTick"`

	Slot0SqrtPriceX96 string `json:"sqrtPriceX96"`
	Slot0Tick         int32  `json:"tick"`
	Slot0FeeProtocol  uint8  `json:"feeProtocol"`
	Slot0Unlocked     bool   `json:"unlocked"`

	FeeGrowthGlobal0X128 string `json:"feeGrowthGlobal0X128"`
	FeeGrowthGlobal1X128 string `json:"feeGrowthGlobal1X128"`

	Token0ProtocolFee string `json:"token0ProtocolFee"`
	Token1ProtocolFee string `json:"token1ProtocolFee"`

	Liquidity string `json:"liquidity"`

	Ticks RpcTicks `json:"ticks"`

	TickBitmaps RpcTickBitmaps `json:"tickBitmaps"`

	Positions RpcPositions `json:"positions"`
}

func newRpcPool(poolPath string) RpcPool {
	rpcPool := RpcPool{}
	pool := GetPoolFromPoolPath(poolPath)

	rpcPool.PoolPath = poolPath

	rpcPool.Token0Path = pool.token0Path
	rpcPool.Token1Path = pool.token1Path

	rpcPool.Token0Balance = pool.balances.token0.ToString()
	rpcPool.Token1Balance = pool.balances.token1.ToString()

	rpcPool.Fee = pool.fee

	rpcPool.TickSpacing = pool.tickSpacing

	rpcPool.MaxLiquidityPerTick = pool.maxLiquidityPerTick.ToString()

	rpcPool.Slot0SqrtPriceX96 = pool.slot0.sqrtPriceX96.ToString()
	rpcPool.Slot0Tick = pool.slot0.tick
	rpcPool.Slot0FeeProtocol = pool.slot0.feeProtocol
	rpcPool.Slot0Unlocked = pool.slot0.unlocked

	rpcPool.FeeGrowthGlobal0X128 = pool.feeGrowthGlobal0X128.ToString()
	rpcPool.FeeGrowthGlobal1X128 = pool.feeGrowthGlobal1X128.ToString()

	rpcPool.Token0ProtocolFee = pool.protocolFees.token0.ToString()
	rpcPool.Token1ProtocolFee = pool.protocolFees.token1.ToString()

	rpcPool.Liquidity = pool.liquidity.ToString()

	rpcPool.Ticks = RpcTicks{}
	pool.ticks.Iterate("", "", func(tickStr string, iTickInfo interface{}) bool {
		tick, _ := strconv.Atoi(tickStr)
		tickInfo := iTickInfo.(TickInfo)

		rpcPool.Ticks[int32(tick)] = RpcTickInfo{
			LiquidityGross:              tickInfo.liquidityGross.ToString(),
			LiquidityNet:                tickInfo.liquidityNet.ToString(),
			FeeGrowthOutside0X128:       tickInfo.feeGrowthOutside0X128.ToString(),
			FeeGrowthOutside1X128:       tickInfo.feeGrowthOutside1X128.ToString(),
			TickCumulativeOutside:       tickInfo.tickCumulativeOutside,
			SecondsPerLiquidityOutsideX: tickInfo.secondsPerLiquidityOutsideX128.ToString(),
			SecondsOutside:              tickInfo.secondsOutside,
			Initialized:                 tickInfo.initialized,
		}

		return false
	})

	rpcPool.TickBitmaps = RpcTickBitmaps{}
	pool.tickBitmaps.Iterate("", "", func(tickStr string, iTickBitmap interface{}) bool {
		tick, _ := strconv.Atoi(tickStr)
		pool.setTickBitmap(int16(tick), iTickBitmap.(*u256.Uint))

		return false
	})

	rpcPositions := []RpcPosition{}
	pool.positions.Iterate("", "", func(posKey string, iPositionInfo interface{}) bool {
		owner, tickLower, tickUpper := posKeyDivide(posKey)
		posInfo := iPositionInfo.(PositionInfo)

		rpcPositions = append(rpcPositions, RpcPosition{
			Owner:      owner,
			TickLower:  tickLower,
			TickUpper:  tickUpper,
			Liquidity:  posInfo.liquidity.ToString(),
			Token0Owed: posInfo.tokensOwed0.ToString(),
			Token1Owed: posInfo.tokensOwed1.ToString(),
		})

		return false
	})

	rpcPool.Positions = rpcPositions

	return rpcPool
}

func (r RpcPool) JSON() *json.Node {
	return makePoolNode(r)
}

func makePoolNode(pool RpcPool) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"poolPath":             json.StringNode("poolPath", pool.PoolPath),
		"token0Path":           json.StringNode("token0Path", pool.Token0Path),
		"token1Path":           json.StringNode("token1Path", pool.Token1Path),
		"token0Balance":        json.StringNode("token0Balance", pool.Token0Balance),
		"token1Balance":        json.StringNode("token1Balance", pool.Token1Balance),
		"fee":                  json.NumberNode("fee", float64(pool.Fee)),
		"tickSpacing":          json.NumberNode("tickSpacing", float64(pool.TickSpacing)),
		"maxLiquidityPerTick":  json.StringNode("maxLiquidityPerTick", pool.MaxLiquidityPerTick),
		"sqrtPriceX96":         json.StringNode("sqrtPriceX96", pool.Slot0SqrtPriceX96),
		"tick":                 json.NumberNode("tick", float64(pool.Slot0Tick)),
		"feeProtocol":          json.NumberNode("feeProtocol", float64(pool.Slot0FeeProtocol)),
		"unlocked":             json.BoolNode("unlocked", pool.Slot0Unlocked),
		"feeGrowthGlobal0X128": json.StringNode("feeGrowthGlobal0X128", pool.FeeGrowthGlobal0X128),
		"feeGrowthGlobal1X128": json.StringNode("feeGrowthGlobal1X128", pool.FeeGrowthGlobal1X128),
		"token0ProtocolFee":    json.StringNode("token0ProtocolFee", pool.Token0ProtocolFee),
		"token1ProtocolFee":    json.StringNode("token1ProtocolFee", pool.Token1ProtocolFee),
		"liquidity":            json.StringNode("liquidity", pool.Liquidity),
		"ticks":                pool.Ticks.JSON(),
		"tickBitmaps":          pool.TickBitmaps.JSON(),
		"positions":            pool.Positions.JSON(),
	})
}

type RpcTickInfo struct {
	LiquidityGross string `json:"liquidityGross"`
	LiquidityNet   string `json:"liquidityNet"`

	FeeGrowthOutside0X128 string `json:"feeGrowthOutside0X128"`
	FeeGrowthOutside1X128 string `json:"feeGrowthOutside1X128"`

	TickCumulativeOutside int64 `json:"tickCumulativeOutside"`

	SecondsPerLiquidityOutsideX string `json:"secondsPerLiquidityOutsideX"`
	SecondsOutside              uint32 `json:"secondsOutside"`

	Initialized bool `json:"initialized"`
}

func (r RpcTickInfo) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"liquidityGross":              json.StringNode("liquidityGross", r.LiquidityGross),
		"liquidityNet":                json.StringNode("liquidityNet", r.LiquidityNet),
		"feeGrowthOutside0X128":       json.StringNode("feeGrowthOutside0X128", r.FeeGrowthOutside0X128),
		"feeGrowthOutside1X128":       json.StringNode("feeGrowthOutside1X128", r.FeeGrowthOutside1X128),
		"tickCumulativeOutside":       json.NumberNode("tickCumulativeOutside", float64(r.TickCumulativeOutside)),
		"secondsPerLiquidityOutsideX": json.StringNode("secondsPerLiquidityOutsideX", r.SecondsPerLiquidityOutsideX),
		"secondsOutside":              json.NumberNode("secondsOutside", float64(r.SecondsOutside)),
		"initialized":                 json.BoolNode("initialized", r.Initialized),
	})
}

type RpcTickBitmaps map[int16]string // tick(wordPos) => bitmap(tickWord ^ mask)

func (r RpcTickBitmaps) JSON() *json.Node {
	tickBitmapsJson := map[string]*json.Node{}
	for tick, tickBitmap := range r {
		tickBitmapsJson[strconv.Itoa(int(tick))] = json.StringNode("", tickBitmap)
	}
	return json.ObjectNode("", tickBitmapsJson)
}

type RpcTicks map[int32]RpcTickInfo // tick => RpcTickInfo

func (r RpcTicks) JSON() *json.Node {
	ticksJson := map[string]*json.Node{}
	for tick, tickInfo := range r {
		ticksJson[strconv.Itoa(int(tick))] = tickInfo.JSON()
	}
	return json.ObjectNode("", ticksJson)
}

type RpcPositions []RpcPosition

func (r RpcPositions) JSON() *json.Node {
	positionsJson := make([]*json.Node, len(r))
	for i, pos := range r {
		positionsJson[i] = pos.JSON()
	}
	return json.ArrayNode("", positionsJson)
}

type RpcPosition struct {
	Owner string `json:"owner"`

	TickLower int32 `json:"tickLower"`
	TickUpper int32 `json:"tickUpper"`

	Liquidity string `json:"liquidity"`

	Token0Owed string `json:"token0Owed"`
	Token1Owed string `json:"token1Owed"`
}

func (r RpcPosition) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"owner":      json.StringNode("owner", r.Owner),
		"tickLower":  json.NumberNode("tickLower", float64(r.TickLower)),
		"tickUpper":  json.NumberNode("tickUpper", float64(r.TickUpper)),
		"liquidity":  json.StringNode("liquidity", r.Liquidity),
		"token0Owed": json.StringNode("token0Owed", r.Token0Owed),
		"token1Owed": json.StringNode("token1Owed", r.Token1Owed),
	})
}

type statNode struct {
	height    int64
	timestamp int64
}

func newStatNode() statNode {
	return statNode{
		height:    std.ChainHeight(),
		timestamp: time.Now().Unix(),
	}
}

func (s statNode) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(s.height)),
		"timestamp": json.NumberNode("timestamp", float64(s.timestamp)),
	})
}
