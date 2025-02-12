package position

import (
	"std"

	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/gnft"
	pl "gno.land/r/gnoswap/v1/pool"
)

func PositionGetPosition(tokenId uint64) Position {
	position, _ := GetPosition(tokenId)
	return position
}

func PositionGetPositionNonce(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.nonce
}

func PositionGetPositionOperator(tokenId uint64) std.Address {
	position := MustGetPosition(tokenId)
	return position.operator
}

func PositionGetPositionPoolKey(tokenId uint64) string {
	position := MustGetPosition(tokenId)
	return position.poolKey
}

func PositionGetPositionTickLower(tokenId uint64) int32 {
	position := MustGetPosition(tokenId)
	return position.tickLower
}

func PositionGetPositionTickUpper(tokenId uint64) int32 {
	position := MustGetPosition(tokenId)
	return position.tickUpper
}

func PositionGetPositionLiquidity(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.liquidity
}

func PositionGetPositionFeeGrowthInside0LastX128(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.feeGrowthInside0LastX128
}

func PositionGetPositionFeeGrowthInside1LastX128(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.feeGrowthInside1LastX128
}

func PositionGetPositionTokensOwed0(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.tokensOwed0
}

func PositionGetPositionTokensOwed1(tokenId uint64) *u256.Uint {
	position := MustGetPosition(tokenId)
	return position.tokensOwed1
}

func PositionGetPositionIsBurned(tokenId uint64) bool {
	position := MustGetPosition(tokenId)
	return position.burned
}

func PositionIsInRange(tokenId uint64) bool {
	position := MustGetPosition(tokenId)
	poolPath := position.poolKey
	poolCurrentTick := pl.PoolGetSlot0Tick(poolPath)

	if position.tickLower <= poolCurrentTick && poolCurrentTick <= position.tickUpper {
		return true
	}
	return false
}

func PositionGetPositionOwner(tokenId uint64) std.Address {
	owner, err := gnft.OwnerOf(tokenIdFrom(tokenId))
	if err != nil {
		panic(newErrorWithDetail(
			errDataNotFound, err.Error()))
	}
	return owner
}

func PositionGetPositionNonceStr(tokenId uint64) string {
	return MustGetPosition(tokenId).nonce.ToString()
}

func PositionGetPositionOperatorStr(tokenId uint64) string {
	return MustGetPosition(tokenId).operator.String()
}

func PositionGetPositionLiquidityStr(tokenId uint64) string {
	return MustGetPosition(tokenId).liquidity.ToString()
}

func PositionGetPositionFeeGrowthInside0LastX128Str(tokenId uint64) string {
	return MustGetPosition(tokenId).feeGrowthInside0LastX128.ToString()
}

func PositionGetPositionFeeGrowthInside1LastX128Str(tokenId uint64) string {
	return MustGetPosition(tokenId).feeGrowthInside1LastX128.ToString()
}

func PositionGetPositionTokensOwed0Str(tokenId uint64) string {
	return MustGetPosition(tokenId).tokensOwed0.ToString()
}

func PositionGetPositionTokensOwed1Str(tokenId uint64) string {
	return MustGetPosition(tokenId).tokensOwed1.ToString()
}
