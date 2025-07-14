package pool

import (
	"gno.land/p/demo/ufmt"
)

func createPool(
	token0Path string,
	token1Path string,
	fee uint32,
	sqrtPriceX96 string,
	tickSpacing int32,
) *poolCreateConfig {
	poolInfo := newPoolParams(token0Path, token1Path, fee, sqrtPriceX96, tickSpacing)
	// check if token0Path == token1Path
	if poolInfo.isSameTokenPath() {
		panic(newErrorWithDetail(
			errDuplicateTokenInPool,
			ufmt.Sprintf(
				"expected token0Path(%s) != token1Path(%s)",
				poolInfo.token0Path, poolInfo.token1Path,
			),
		))
	}
	if !poolInfo.isSupportedFee(fee) {
		panic(newErrorWithDetail(
			errUnsupportedFeeTier,
			ufmt.Sprintf("expected fee(%d) to be one of %d, %d, %d, %d", fee, FeeTier100, FeeTier500, FeeTier3000, FeeTier10000),
		))
	}

	poolInfo = poolInfo.updateWithWrapping()

	assertOnlyRegistered(poolInfo.token0Path)
	assertOnlyRegistered(poolInfo.token1Path)

	if !poolInfo.isInOrder() {
		panic(newErrorWithDetail(
			errTokenSortOrder,
			ufmt.Sprintf("expected token0Path(%s) < token1Path(%s)", token0Path, token1Path),
		))
	}
	return poolInfo
}
