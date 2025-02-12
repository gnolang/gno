package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

func TestInitTwoPools(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(consts.ADMIN))
	SetPoolCreationFeeByAdmin(0)

	// foo:bar
	CreatePool("gno.land/r/onbloc/bar", "gno.land/r/onbloc/foo", uint32(500), common.TickMathGetSqrtRatioAtTick(-10000).ToString())

	// bar:baz
	CreatePool(barPath, bazPath, fee500, "130621891405341611593710811006") // tick 10000

	uassert.Equal(t, pools.Size(), 2)
}

func TestApiGetPools(t *testing.T) {
	getPools := ApiGetPools()

	root, err := json.Unmarshal([]byte(getPools))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}

	uassert.Equal(t, 2, response.Size()) // should be same as the number of pools
}

func TestApiGetPool(t *testing.T) {
	t.Run("existing pool", func(t *testing.T) {
		getPool := ApiGetPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
		uassert.Equal(t, getPool, `{"stat":{"height":123,"timestamp":1234567890},"response":{"poolPath":"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500","token0Path":"gno.land/r/onbloc/bar","token1Path":"gno.land/r/onbloc/foo","token0Balance":"0","token1Balance":"0","fee":500,"tickSpacing":10,"maxLiquidityPerTick":"1917569901783203986719870431555990","sqrtPriceX96":"48055510970269007215549348797","tick":-10000,"feeProtocol":0,"unlocked":true,"feeGrowthGlobal0X128":"0","feeGrowthGlobal1X128":"0","token0ProtocolFee":"0","token1ProtocolFee":"0","liquidity":"0","ticks":{},"tickBitmaps":{},"positions":[]}}`)
	})

	t.Run("non-existing pool", func(t *testing.T) {
		getPool := ApiGetPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000")
		uassert.Equal(t, getPool, "")
	})
}

func TestMakeStatNode(t *testing.T) {
	t.Run("default block height and timestamp", func(t *testing.T) {
		statNode := makeStatNode()

		statHeight, _ := statNode.GetKey("height")
		uassert.Equal(t, statHeight.String(), "123")

		statTimestamp, _ := statNode.GetKey("timestamp")
		uassert.Equal(t, statTimestamp.String(), "1234567890")
	})

	t.Run("increase block height and timestamp", func(t *testing.T) {
		std.TestSkipHeights(1)

		statNode := makeStatNode()

		statHeight, _ := statNode.GetKey("height")
		uassert.Equal(t, statHeight.String(), "124")
	})
}

func TestRpcMakePool(t *testing.T) {
	t.Run("existing pool", func(t *testing.T) {
		rpcPool := rpcMakePool("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
		uassert.Equal(t, rpcPool.PoolPath, "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
		uassert.Equal(t, rpcPool.Token0Path, "gno.land/r/onbloc/bar")
		uassert.Equal(t, rpcPool.Token1Path, "gno.land/r/onbloc/foo")
		uassert.Equal(t, rpcPool.Fee, uint32(500))
		uassert.Equal(t, rpcPool.TickSpacing, int32(10))
		uassert.Equal(t, rpcPool.MaxLiquidityPerTick, "1917569901783203986719870431555990")
		uassert.Equal(t, rpcPool.Slot0SqrtPriceX96, "48055510970269007215549348797")
		uassert.Equal(t, rpcPool.Slot0Tick, int32(-10000))
		uassert.Equal(t, rpcPool.Slot0FeeProtocol, uint8(0))
		uassert.Equal(t, rpcPool.Slot0Unlocked, true)
		uassert.Equal(t, rpcPool.FeeGrowthGlobal0X128, "0")
		uassert.Equal(t, rpcPool.FeeGrowthGlobal1X128, "0")
		uassert.Equal(t, rpcPool.Token0ProtocolFee, "0")
		uassert.Equal(t, rpcPool.Token1ProtocolFee, "0")
		uassert.Equal(t, rpcPool.Liquidity, "0")

		if len(rpcPool.Ticks) != 0 {
			t.Errorf("expected 0 ticks, got %d", len(rpcPool.Ticks))
		}

		if len(rpcPool.TickBitmaps) != 0 {
			t.Errorf("expected 0 tickBitmaps, got %d", len(rpcPool.TickBitmaps))
		}
	})

	t.Run("non-existing pool", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic, got %v", r)
			}
		}()

		rpcMakePool("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000")
	})
}

func TestMakeTicksJson(t *testing.T) {
	t.Run("existing tick", func(t *testing.T) {
		ticks := RpcTicks{
			10000: {
				LiquidityGross:              "100000000",
				LiquidityNet:                "100000000",
				FeeGrowthOutside0X128:       "100000000",
				FeeGrowthOutside1X128:       "100000000",
				TickCumulativeOutside:       100000000,
				SecondsPerLiquidityOutsideX: "100000000",
				SecondsOutside:              100000000,
				Initialized:                 true,
			},
		}

		ticksJson := makeTicksJson(ticks)
		uassert.Equal(t, ticksJson["10000"].String(), `{"liquidityGross":"100000000","liquidityNet":"100000000","feeGrowthOutside0X128":"100000000","feeGrowthOutside1X128":"100000000","tickCumulativeOutside":100000000,"secondsPerLiquidityOutsideX":"100000000","secondsOutside":100000000,"initialized":true}`)
	})

	t.Run("non-existing tick", func(t *testing.T) {
		ticksJson := makeTicksJson(RpcTicks{})
		uassert.Equal(t, len(ticksJson), 0)
	})
}

func TestMakeRpcTickBitmapsJson(t *testing.T) {
	t.Run("existing tickBitmap", func(t *testing.T) {
		tickBitmaps := RpcTickBitmaps{
			10000: "100000000",
			10001: "100000001",
		}

		tickBitmapsJson := makeRpcTickBitmapsJson(tickBitmaps)
		uassert.Equal(t, tickBitmapsJson["10000"].String(), `"100000000"`)
		uassert.Equal(t, tickBitmapsJson["10001"].String(), `"100000001"`)
	})

	t.Run("non-existing tickBitmap", func(t *testing.T) {
		tickBitmapsJson := makeRpcTickBitmapsJson(RpcTickBitmaps{})
		uassert.Equal(t, len(tickBitmapsJson), 0)
	})
}

func TestMakeRpcPositionsArray(t *testing.T) {
	t.Run("existing positions", func(t *testing.T) {
		positions := []RpcPosition{
			{Owner: "gno.land/r/onbloc/bar", TickLower: 10000, TickUpper: 10001, Liquidity: "100000000", Token0Owed: "100000000", Token1Owed: "100000000"},
		}

		positionsJson := makeRpcPositionsArray(positions)
		uassert.Equal(t, positionsJson[0].String(), `{"owner":"gno.land/r/onbloc/bar","tickLower":10000,"tickUpper":10001,"liquidity":"100000000","token0Owed":"100000000","token1Owed":"100000000"}`)
	})

	t.Run("non-existing positions", func(t *testing.T) {
		positionsJson := makeRpcPositionsArray([]RpcPosition{})
		uassert.Equal(t, len(positionsJson), 0)
	})
}
