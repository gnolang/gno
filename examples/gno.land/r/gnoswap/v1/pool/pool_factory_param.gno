package pool

import (
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
)

const (
	FeeTier100   uint32 = 100
	FeeTier500   uint32 = 500
	FeeTier3000  uint32 = 3000
	FeeTier10000 uint32 = 10000
)

// createPoolParams holds the essential parameters for creating a new pool.
type createPoolParams struct {
	token0Path   string
	token1Path   string
	fee          uint32
	sqrtPriceX96 *u256.Uint
	tickSpacing  int32
}

// newPoolParams defines the essential parameters for creating a new pool.
func newPoolParams(
	token0Path string,
	token1Path string,
	fee uint32,
	sqrtPriceX96 string,
	tickSpacing int32,
) *createPoolParams {
	price := u256.MustFromDecimal(sqrtPriceX96)
	return &createPoolParams{
		token0Path:   token0Path,
		token1Path:   token1Path,
		fee:          fee,
		sqrtPriceX96: price,
		tickSpacing:  tickSpacing,
	}
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

func (p *createPoolParams) updateWithWrapping() *createPoolParams {
	token0Path, token1Path := p.wrap()

	if !p.isInOrder() {
		token0Path, token1Path = token1Path, token0Path

		oldTick := common.TickMathGetTickAtSqrtRatio(p.sqrtPriceX96)
		if oldTick > consts.MAX_TICK || oldTick <= consts.MIN_TICK {
			panic(newErrorWithDetail(
				errInvalidTickMinMaxRange,
				ufmt.Sprintf("The tick value(%d) is out of range due to reverse", oldTick),
			))
		}
		newPrice := common.TickMathGetSqrtRatioAtTick(oldTick * int32(-1))
		if newPrice.IsZero() {
			panic(newErrorWithDetail(
				errOverFlow,
				ufmt.Sprintf("expected newPrice(%s) to be non-zero", newPrice.ToString()),
			))
		}
		p.sqrtPriceX96 = new(u256.Uint).Set(newPrice)
	}
	return newPoolParams(token0Path, token1Path, p.fee, p.sqrtPriceX96.ToString(), GetFeeAmountTickSpacing(p.fee))
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

func (p *createPoolParams) generatePoolPath() string {
	if !p.isInOrder() {
		p.token0Path, p.token1Path = p.token1Path, p.token0Path
	}

	return ufmt.Sprintf("%s:%s:%d", p.token0Path, p.token1Path, p.fee)
}

func (p *createPoolParams) isSupportedFee(feeTier uint32) bool {
	switch feeTier {
	case FeeTier100, FeeTier500, FeeTier3000, FeeTier10000:
		return true
	}
	return false
}
