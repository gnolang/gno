package position

import (
	"std"
	"time"

	"gno.land/p/demo/json"
	ufmt "gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gnft"
	pl "gno.land/r/gnoswap/v1/pool"
)

func ApiGetPositions() string {
	rpcPositions := []RpcPosition{}
	for positionId := uint64(1); positionId < nextId; positionId++ {
		rpcPosition := rpcMakePosition(positionId)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	return makeJsonResponse(&baseStat, &PositionsResponse{Positions: rpcPositions})
}

func ApiGetPosition(id uint64) string {
	_, exist := GetPosition(id)
	if !exist {
		return ""
	}

	rpcPosition := rpcMakePosition(id)
	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	return makeJsonResponse(&baseStat, &PositionsResponse{Positions: []RpcPosition{rpcPosition}})
}

func ApiGetPositionsByPoolPath(poolPath string) string {
	rpcPositions := []RpcPosition{}
	for id := uint64(1); id < nextId; id++ {
		position := MustGetPosition(id)
		if position.poolKey != poolPath {
			continue
		}
		rpcPosition := rpcMakePosition(id)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	return makeJsonResponse(&baseStat, &PositionsResponse{Positions: rpcPositions})
}

func ApiGetPositionsUnclaimedFee() string {
	rpcUnclaimedFee := []RpcUnclaimedFee{}
	for id := uint64(1); id < nextId; id++ {
		unclaimedFee0, unclaimedFee1 := unclaimedFee(id)
		rpcUnclaimedFee = append(rpcUnclaimedFee, RpcUnclaimedFee{
			LpPositionId: id,
			Fee0:         unclaimedFee0.ToString(),
			Fee1:         unclaimedFee1.ToString(),
		})
	}

	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	return makeJsonResponse(&baseStat, &UnclaimedFeesResponse{Fees: rpcUnclaimedFee})
}

func ApiGetPositionsByAddress(address std.Address) string {
	rpcPositions := []RpcPosition{}
	for lpPositionId := uint64(1); lpPositionId < nextId; lpPositionId++ {
		position := MustGetPosition(lpPositionId)
		owner, err := gnft.OwnerOf(positionIdFrom(lpPositionId))
		if err != nil {
			panic(ufmt.Sprintf("owner not found for positionId: %d", lpPositionId))
		}

		if !(position.operator == address || owner == address) {
			continue
		}

		rpcPosition := rpcMakePosition(lpPositionId)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	r := NewResponseApiGetPositions(baseStat, rpcPositions)
	stat := baseStat.JSON()

	// RESPONSE (ARRAY) NODE
	rsps := json.ArrayNode("", []*json.Node{})
	for _, position := range r.Response {
		owner, err := gnft.OwnerOf(positionIdFrom(position.LpPositionId))
		if err != nil {
			owner = consts.ZERO_ADDRESS
		}
		rsps.AppendArray(position.JSON(owner))
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": rsps,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPositionUnclaimedFeeByLpPositionId(lpPositionId uint64) string {
	unclaimedFee0, unclaimedFee1 := unclaimedFee(lpPositionId)
	fee := RpcUnclaimedFee{
		LpPositionId: lpPositionId,
		Fee0:         unclaimedFee0.ToString(),
		Fee1:         unclaimedFee1.ToString(),
	}

	baseStat := NewResponseQueryBase(std.ChainHeight(), time.Now().Unix())
	return makeJsonResponse(&baseStat, &UnclaimedFeesResponse{
		Fees: []RpcUnclaimedFee{fee},
	})
}

func rpcMakePosition(positionId uint64) RpcPosition {
	position := MustGetPosition(positionId)

	pool := pl.GetPoolFromPoolPath(position.poolKey)
	currentSqrtPriceX96 := pool.Slot0SqrtPriceX96()
	lowerTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickLower)
	upperTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickUpper)

	calculatedToken0Balance, calculatedToken1Balance := common.GetAmountsForLiquidity(
		currentSqrtPriceX96,
		lowerTickSqrtPriceX96,
		upperTickSqrtPriceX96,
		position.liquidity,
	)

	token0Balance, token1Balance := position.token0Balance, position.token1Balance

	unclaimedFee0 := i256.Zero()
	unclaimedFee1 := i256.Zero()
	burned := IsBurned(positionId)
	if !burned {
		unclaimedFee0, unclaimedFee1 = unclaimedFee(positionId)
	}

	owner, err := gnft.OwnerOf(positionIdFrom(positionId))
	if err != nil {
		owner = consts.ZERO_ADDRESS
	}

	return RpcPosition{
		LpPositionId:             positionId,
		Burned:                   burned,
		Owner:                    owner.String(),
		Operator:                 position.operator.String(),
		PoolKey:                  position.poolKey,
		TickLower:                position.tickLower,
		TickUpper:                position.tickUpper,
		Liquidity:                position.liquidity.ToString(),
		FeeGrowthInside0LastX128: position.feeGrowthInside0LastX128.ToString(),
		FeeGrowthInside1LastX128: position.feeGrowthInside1LastX128.ToString(),
		TokensOwed0:              position.tokensOwed0.ToString(),
		TokensOwed1:              position.tokensOwed1.ToString(),
		Token0Balance:            token0Balance.ToString(),
		Token1Balance:            token1Balance.ToString(),
		CalculatedToken0Balance:  calculatedToken0Balance,
		CalculatedToken1Balance:  calculatedToken1Balance,
		FeeUnclaimed0:            unclaimedFee0.ToString(),
		FeeUnclaimed1:            unclaimedFee1.ToString(),
	}
}

func UnclaimedFee(positionId uint64) (*i256.Int, *i256.Int) {
	return unclaimedFee(positionId)
}

func unclaimedFee(positionId uint64) (*i256.Int, *i256.Int) {
	// ref: https://blog.uniswap.org/uniswap-v3-math-primer-2#calculating-uncollected-fees

	position := MustGetPosition(positionId)

	liquidityU256 := position.liquidity
	liquidity := i256.FromUint256(liquidityU256)

	tickLower := position.tickLower
	tickUpper := position.tickUpper

	poolKey := position.poolKey
	pool := pl.GetPoolFromPoolPath(poolKey)

	currentTick := pool.Slot0Tick()

	feeGrowthGlobal0X128 := i256.FromUint256(pool.FeeGrowthGlobal0X128())
	feeGrowthGlobal1X128 := i256.FromUint256(pool.FeeGrowthGlobal1X128())

	tickUpperFeeGrowthOutside0X128 := i256.FromUint256(pool.GetTickFeeGrowthOutside0X128(tickUpper))
	tickUpperFeeGrowthOutside1X128 := i256.FromUint256(pool.GetTickFeeGrowthOutside1X128(tickUpper))

	tickLowerFeeGrowthOutside0X128 := i256.FromUint256(pool.GetTickFeeGrowthOutside0X128(tickLower))
	tickLowerFeeGrowthOutside1X128 := i256.FromUint256(pool.GetTickFeeGrowthOutside1X128(tickLower))

	feeGrowthInside0LastX128 := i256.FromUint256(position.feeGrowthInside0LastX128)
	feeGrowthInside1LastX128 := i256.FromUint256(position.feeGrowthInside1LastX128)

	var tickLowerFeeGrowthBelow0, tickLowerFeeGrowthBelow1, tickUpperFeeGrowthAbove0, tickUpperFeeGrowthAbove1 *i256.Int

	if currentTick >= tickUpper {
		tickUpperFeeGrowthAbove0 = subIn256(feeGrowthGlobal0X128, tickUpperFeeGrowthOutside0X128)
		tickUpperFeeGrowthAbove1 = subIn256(feeGrowthGlobal1X128, tickUpperFeeGrowthOutside1X128)
	} else {
		tickUpperFeeGrowthAbove0 = tickUpperFeeGrowthOutside0X128
		tickUpperFeeGrowthAbove1 = tickUpperFeeGrowthOutside1X128
	}

	if currentTick >= tickLower {
		tickLowerFeeGrowthBelow0 = tickLowerFeeGrowthOutside0X128
		tickLowerFeeGrowthBelow1 = tickLowerFeeGrowthOutside1X128
	} else {
		tickLowerFeeGrowthBelow0 = subIn256(feeGrowthGlobal0X128, tickLowerFeeGrowthOutside0X128)
		tickLowerFeeGrowthBelow1 = subIn256(feeGrowthGlobal1X128, tickLowerFeeGrowthOutside1X128)
	}

	feeGrowthInside0X128 := subIn256(feeGrowthGlobal0X128, tickLowerFeeGrowthBelow0)
	feeGrowthInside0X128 = subIn256(feeGrowthInside0X128, tickUpperFeeGrowthAbove0)

	feeGrowthInside1X128 := subIn256(feeGrowthGlobal1X128, tickLowerFeeGrowthBelow1)
	feeGrowthInside1X128 = subIn256(feeGrowthInside1X128, tickUpperFeeGrowthAbove1)

	value01 := subIn256(feeGrowthInside0X128, feeGrowthInside0LastX128)
	value02 := i256.Zero().Mul(liquidity, value01)
	unclaimedFee0 := i256.Zero().Div(value02, i256.MustFromDecimal(consts.Q128))

	value11 := subIn256(feeGrowthInside1X128, feeGrowthInside1LastX128)
	value12 := i256.Zero().Mul(liquidity, value11)
	unclaimedFee1 := i256.Zero().Div(value12, i256.MustFromDecimal(consts.Q128))

	return unclaimedFee0, unclaimedFee1
}

func subIn256(x, y *i256.Int) *i256.Int {
	value := i256.Zero()
	diff := value.Sub(x, y)

	if diff.IsNeg() {
		q256 := i256.MustFromDecimal(consts.MAX_UINT256)
		return diff.Add(diff, q256)
	}

	return diff
}

func IsBurned(positionId uint64) bool {
	position := MustGetPosition(positionId)
	return position.burned
}
