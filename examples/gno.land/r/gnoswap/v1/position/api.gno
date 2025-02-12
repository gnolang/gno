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

type RpcPosition struct {
	LpTokenId                uint64 `json:"lpTokenId"`
	Burned                   bool   `json:"burned"`
	Owner                    string `json:"owner"`
	Operator                 string `json:"operator"`
	PoolKey                  string `json:"poolKey"`
	TickLower                int32  `json:"tickLower"`
	TickUpper                int32  `json:"tickUpper"`
	Liquidity                string `json:"liquidity"`
	FeeGrowthInside0LastX128 string `json:"feeGrowthInside0LastX128"`
	FeeGrowthInside1LastX128 string `json:"feeGrowthInside1LastX128"`
	TokensOwed0              string `json:"token0Owed"`
	TokensOwed1              string `json:"token1Owed"`

	Token0Balance string `json:"token0Balance"`
	Token1Balance string `json:"token1Balance"`
	FeeUnclaimed0 string `json:"fee0Unclaimed"`
	FeeUnclaimed1 string `json:"fee1Unclaimed"`
}

type RpcUnclaimedFee struct {
	LpTokenId uint64 `json:"lpTokenId"`
	Fee0      string `json:"fee0"`
	Fee1      string `json:"fee1"`
}

type ResponseQueryBase struct {
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

type ResponseApiGetPositions struct {
	Stat     ResponseQueryBase `json:"stat"`
	Response []RpcPosition     `json:"response"`
}

func ApiGetPositions() string {
	rpcPositions := []RpcPosition{}

	for tokenId := uint64(1); tokenId < nextId; tokenId++ {
		rpcPosition := rpcMakePosition(tokenId)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	r := ResponseApiGetPositions{
		Stat: ResponseQueryBase{
			Height:    std.GetHeight(),
			Timestamp: time.Now().Unix(),
		},
		Response: rpcPositions,
	}

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, position := range r.Response {
		owner, err := gnft.OwnerOf(tokenIdFrom(position.LpTokenId))
		if err != nil {
			panic(ufmt.Sprintf("owner not found for tokenId: %d", position.LpTokenId))
		}

		positionNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId":                json.NumberNode("lpTokenId", float64(position.LpTokenId)),
			"burned":                   json.BoolNode("burned", position.Burned),
			"owner":                    json.StringNode("owner", owner.String()),
			"operator":                 json.StringNode("operator", position.Operator),
			"poolKey":                  json.StringNode("poolKey", position.PoolKey),
			"tickLower":                json.NumberNode("tickLower", float64(position.TickLower)),
			"tickUpper":                json.NumberNode("tickUpper", float64(position.TickUpper)),
			"liquidity":                json.StringNode("liquidity", position.Liquidity),
			"feeGrowthInside0LastX128": json.StringNode("feeGrowthInside0LastX128", position.FeeGrowthInside0LastX128),
			"feeGrowthInside1LastX128": json.StringNode("feeGrowthInside1LastX128", position.FeeGrowthInside1LastX128),
			"token0Owed":               json.StringNode("token0Owed", position.TokensOwed0),
			"token1Owed":               json.StringNode("token1Owed", position.TokensOwed1),
			"token0Balance":            json.StringNode("token0Balance", position.Token0Balance),
			"token1Balance":            json.StringNode("token1Balance", position.Token1Balance),
			"fee0Unclaimed":            json.StringNode("fee0Unclaimed", position.FeeUnclaimed0),
			"fee1Unclaimed":            json.StringNode("fee1Unclaimed", position.FeeUnclaimed1),
		})
		responses.AppendArray(positionNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPosition(lpTokenId uint64) string {
	rpcPositions := []RpcPosition{}

	_, exist := GetPosition(lpTokenId)
	if !exist {
		return ""
	}

	rpcPosition := rpcMakePosition(lpTokenId)
	rpcPositions = append(rpcPositions, rpcPosition)

	r := ResponseApiGetPositions{
		Stat: ResponseQueryBase{
			Height:    std.GetHeight(),
			Timestamp: time.Now().Unix(),
		},
		Response: rpcPositions,
	}

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, position := range r.Response {
		owner, err := gnft.OwnerOf(tokenIdFrom(position.LpTokenId))
		if err != nil {
			owner = consts.ZERO_ADDRESS
		}
		positionNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId":                json.NumberNode("lpTokenId", float64(position.LpTokenId)),
			"burned":                   json.BoolNode("burned", position.Burned),
			"owner":                    json.StringNode("owner", owner.String()),
			"operator":                 json.StringNode("operator", position.Operator),
			"poolKey":                  json.StringNode("poolKey", position.PoolKey),
			"tickLower":                json.NumberNode("tickLower", float64(position.TickLower)),
			"tickUpper":                json.NumberNode("tickUpper", float64(position.TickUpper)),
			"liquidity":                json.StringNode("liquidity", position.Liquidity),
			"feeGrowthInside0LastX128": json.StringNode("feeGrowthInside0LastX128", position.FeeGrowthInside0LastX128),
			"feeGrowthInside1LastX128": json.StringNode("feeGrowthInside1LastX128", position.FeeGrowthInside1LastX128),
			"token0Owed":               json.StringNode("token0Owed", position.TokensOwed0),
			"token1Owed":               json.StringNode("token1Owed", position.TokensOwed1),
			"token0Balance":            json.StringNode("token0Balance", position.Token0Balance),
			"token1Balance":            json.StringNode("token1Balance", position.Token1Balance),
			"fee0Unclaimed":            json.StringNode("fee0Unclaimed", position.FeeUnclaimed0),
			"fee1Unclaimed":            json.StringNode("fee1Unclaimed", position.FeeUnclaimed1),
		})
		responses.AppendArray(positionNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPositionsByPoolPath(poolPath string) string {
	rpcPositions := []RpcPosition{}

	for lpTokenId := uint64(1); lpTokenId < nextId; lpTokenId++ {
		position := MustGetPosition(lpTokenId)
		if position.poolKey != poolPath {
			continue
		}

		rpcPosition := rpcMakePosition(lpTokenId)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	r := ResponseApiGetPositions{
		Stat: ResponseQueryBase{
			Height:    std.GetHeight(),
			Timestamp: time.Now().Unix(),
		},
		Response: rpcPositions,
	}

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, position := range r.Response {
		owner, err := gnft.OwnerOf(tokenIdFrom(position.LpTokenId))
		if err != nil {
			owner = consts.ZERO_ADDRESS
		}
		positionNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId":                json.NumberNode("lpTokenId", float64(position.LpTokenId)),
			"burned":                   json.BoolNode("burned", position.Burned),
			"owner":                    json.StringNode("owner", owner.String()),
			"operator":                 json.StringNode("operator", position.Operator),
			"poolKey":                  json.StringNode("poolKey", position.PoolKey),
			"tickLower":                json.NumberNode("tickLower", float64(position.TickLower)),
			"tickUpper":                json.NumberNode("tickUpper", float64(position.TickUpper)),
			"liquidity":                json.StringNode("liquidity", position.Liquidity),
			"feeGrowthInside0LastX128": json.StringNode("feeGrowthInside0LastX128", position.FeeGrowthInside0LastX128),
			"feeGrowthInside1LastX128": json.StringNode("feeGrowthInside1LastX128", position.FeeGrowthInside1LastX128),
			"token0Owed":               json.StringNode("token0Owed", position.TokensOwed0),
			"token1Owed":               json.StringNode("token1Owed", position.TokensOwed1),
			"token0Balance":            json.StringNode("token0Balance", position.Token0Balance),
			"token1Balance":            json.StringNode("token1Balance", position.Token1Balance),
			"fee0Unclaimed":            json.StringNode("fee0Unclaimed", position.FeeUnclaimed0),
			"fee1Unclaimed":            json.StringNode("fee1Unclaimed", position.FeeUnclaimed1),
		})
		responses.AppendArray(positionNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPositionsByAddress(address std.Address) string {
	rpcPositions := []RpcPosition{}
	for lpTokenId := uint64(1); lpTokenId < nextId; lpTokenId++ {
		position := MustGetPosition(lpTokenId)
		owner, err := gnft.OwnerOf(tokenIdFrom(lpTokenId))
		if err != nil {
			panic(ufmt.Sprintf("owner not found for tokenId: %d", lpTokenId))
		}

		if !(position.operator == address || owner == address) {
			continue
		}

		rpcPosition := rpcMakePosition(lpTokenId)
		rpcPositions = append(rpcPositions, rpcPosition)
	}

	r := ResponseApiGetPositions{
		Stat: ResponseQueryBase{
			Height:    std.GetHeight(),
			Timestamp: time.Now().Unix(),
		},
		Response: rpcPositions,
	}

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, position := range r.Response {
		owner, err := gnft.OwnerOf(tokenIdFrom(position.LpTokenId))
		if err != nil {
			owner = consts.ZERO_ADDRESS
		}
		positionNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId":                json.NumberNode("lpTokenId", float64(position.LpTokenId)),
			"burned":                   json.BoolNode("burned", position.Burned),
			"owner":                    json.StringNode("owner", owner.String()),
			"operator":                 json.StringNode("operator", position.Operator),
			"poolKey":                  json.StringNode("poolKey", position.PoolKey),
			"tickLower":                json.NumberNode("tickLower", float64(position.TickLower)),
			"tickUpper":                json.NumberNode("tickUpper", float64(position.TickUpper)),
			"liquidity":                json.StringNode("liquidity", position.Liquidity),
			"feeGrowthInside0LastX128": json.StringNode("feeGrowthInside0LastX128", position.FeeGrowthInside0LastX128),
			"feeGrowthInside1LastX128": json.StringNode("feeGrowthInside1LastX128", position.FeeGrowthInside1LastX128),
			"token0Owed":               json.StringNode("token0Owed", position.TokensOwed0),
			"token1Owed":               json.StringNode("token1Owed", position.TokensOwed1),
			"token0Balance":            json.StringNode("token0Balance", position.Token0Balance),
			"token1Balance":            json.StringNode("token1Balance", position.Token1Balance),
			"fee0Unclaimed":            json.StringNode("fee0Unclaimed", position.FeeUnclaimed0),
			"fee1Unclaimed":            json.StringNode("fee1Unclaimed", position.FeeUnclaimed1),
		})
		responses.AppendArray(positionNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPositionsUnclaimedFee() string {
	rpcUnclaimedFee := []RpcUnclaimedFee{}
	for lpTokenId := uint64(1); lpTokenId < nextId; lpTokenId++ {
		unclaimedFee0, unclaimedFee1 := unclaimedFee(lpTokenId)
		rpcUnclaimedFee = append(rpcUnclaimedFee, RpcUnclaimedFee{
			LpTokenId: lpTokenId,
			Fee0:      unclaimedFee0.ToString(),
			Fee1:      unclaimedFee1.ToString(),
		})
	}

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, unclaimedFee := range rpcUnclaimedFee {
		unclaimedFeeNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId": json.NumberNode("lpTokenId", float64(unclaimedFee.LpTokenId)),
			"fee0":      json.StringNode("fee0", unclaimedFee.Fee0),
			"fee1":      json.StringNode("fee1", unclaimedFee.Fee1),
		})
		responses.AppendArray(unclaimedFeeNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetPositionUnclaimedFeeByLpTokenId(lpTokenId uint64) string {
	rpcUnclaimedFee := []RpcUnclaimedFee{}

	unclaimedFee0, unclaimedFee1 := unclaimedFee(lpTokenId)
	rpcUnclaimedFee = append(rpcUnclaimedFee, RpcUnclaimedFee{
		LpTokenId: lpTokenId,
		Fee0:      unclaimedFee0.ToString(),
		Fee1:      unclaimedFee1.ToString(),
	})

	// STAT NODE
	stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, unclaimedFee := range rpcUnclaimedFee {
		unclaimedFeeNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId": json.NumberNode("lpTokenId", float64(unclaimedFee.LpTokenId)),
			"fee0":      json.StringNode("fee0", unclaimedFee.Fee0),
			"fee1":      json.StringNode("fee1", unclaimedFee.Fee1),
		})
		responses.AppendArray(unclaimedFeeNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func rpcMakePosition(lpTokenId uint64) RpcPosition {
	position := MustGetPosition(lpTokenId)

	burned := isBurned(lpTokenId)

	pool := pl.GetPoolFromPoolPath(position.poolKey)
	currentX96 := pool.Slot0SqrtPriceX96()
	lowerX96 := common.TickMathGetSqrtRatioAtTick(position.tickLower)
	upperX96 := common.TickMathGetSqrtRatioAtTick(position.tickUpper)

	token0Balance, token1Balance := common.GetAmountsForLiquidity(
		currentX96,
		lowerX96,
		upperX96,
		position.liquidity,
	)

	unclaimedFee0 := i256.Zero()
	unclaimedFee1 := i256.Zero()
	if !burned {
		unclaimedFee0, unclaimedFee1 = unclaimedFee(lpTokenId)
	}

	owner, err := gnft.OwnerOf(tokenIdFrom(lpTokenId))
	if err != nil {
		owner = consts.ZERO_ADDRESS
	}

	return RpcPosition{
		LpTokenId:                lpTokenId,
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
		Token0Balance:            token0Balance,
		Token1Balance:            token1Balance,
		FeeUnclaimed0:            unclaimedFee0.ToString(),
		FeeUnclaimed1:            unclaimedFee1.ToString(),
	}
}

func unclaimedFee(tokenId uint64) (*i256.Int, *i256.Int) {
	// ref: https://blog.uniswap.org/uniswap-v3-math-primer-2#calculating-uncollected-fees

	position := MustGetPosition(tokenId)

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

func isBurned(tokenId uint64) bool {
	position := MustGetPosition(tokenId)
	return position.burned
}
