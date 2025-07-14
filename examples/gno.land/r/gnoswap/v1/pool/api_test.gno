package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"

	"gno.land/r/gnoswap/v1/common"
)

func TestInitTwoPools_GetPools(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	SetPoolCreationFeeByAdmin(cross, 0)
	// foo:bar
	CreatePool(cross, "gno.land/r/onbloc/bar", "gno.land/r/onbloc/foo", FeeTier500, common.TickMathGetSqrtRatioAtTick(-10000).ToString())

	// bar:baz
	CreatePool(cross, barPath, bazPath, fee500, "130621891405341611593710811006") // tick 10000

	uassert.Equal(t, pools.Size(), 2)

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

	t.Run("existing pool", func(t *testing.T) {
		getPool := ApiGetPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
		uassert.Equal(t, getPool, `{"stat":{"height":123,"timestamp":1234567890},"response":{"poolPath":"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500","token0Path":"gno.land/r/onbloc/bar","token1Path":"gno.land/r/onbloc/foo","token0Balance":"0","token1Balance":"0","fee":500,"tickSpacing":10,"maxLiquidityPerTick":"1917569901783203986719870431555990","sqrtPriceX96":"48055510970269007215549348797","tick":-10000,"feeProtocol":0,"unlocked":true,"feeGrowthGlobal0X128":"0","feeGrowthGlobal1X128":"0","token0ProtocolFee":"0","token1ProtocolFee":"0","liquidity":"0","ticks":{},"tickBitmaps":{},"positions":[]}}`)
	})

	t.Run("non-existing pool", func(t *testing.T) {
		getPool := ApiGetPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000")
		uassert.Equal(t, getPool, "")
	})

	t.Run("existing pool", func(t *testing.T) {
		rpcPool := newRpcPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
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

		newRpcPool("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000")
	})
}

func TestMakeStatNode(t *testing.T) {
	t.Run("default block height and timestamp", func(t *testing.T) {
		statNode := newStatNode().JSON()

		statHeight, _ := statNode.GetKey("height")
		uassert.Equal(t, statHeight.String(), "123")

		statTimestamp, _ := statNode.GetKey("timestamp")
		uassert.Equal(t, statTimestamp.String(), "1234567890")
	})

	t.Run("increase block height and timestamp", func(t *testing.T) {
		testing.SkipHeights(1)

		statNode := newStatNode().JSON()

		statHeight, _ := statNode.GetKey("height")
		uassert.Equal(t, statHeight.String(), "124")
	})
}
