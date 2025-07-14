package pool

import (
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

var Q192 = new(u256.Uint).Lsh(u256.One(), 192)

const (
	FeeTier100   uint32 = 100
	FeeTier500   uint32 = 500
	FeeTier3000  uint32 = 3000
	FeeTier10000 uint32 = 10000
)

// poolCreateConfig holds the essential parameters for creating a new pool.
type poolCreateConfig struct {
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
) *poolCreateConfig {
	price := u256.MustFromDecimal(sqrtPriceX96)
	return &poolCreateConfig{
		token0Path:   token0Path,
		token1Path:   token1Path,
		fee:          fee,
		sqrtPriceX96: price,
		tickSpacing:  tickSpacing,
	}
}

func (p *poolCreateConfig) SqrtPriceX96() *u256.Uint { return p.sqrtPriceX96 }
func (p *poolCreateConfig) TickSpacing() int32       { return p.tickSpacing }
func (p *poolCreateConfig) Token0Path() string       { return p.token0Path }
func (p *poolCreateConfig) Token1Path() string       { return p.token1Path }
func (p *poolCreateConfig) Fee() uint32              { return p.fee }

func (p *poolCreateConfig) updateWithWrapping() *poolCreateConfig {
	token0Path, token1Path := p.wrap()

	if !p.isInOrder() {
		token0Path, token1Path = token1Path, token0Path

		// Check if initial price is within valid range
		minSqrtRatio := u256.MustFromDecimal(consts.MIN_SQRT_RATIO)
		maxSqrtRatio := u256.MustFromDecimal(consts.MAX_SQRT_RATIO)

		if p.sqrtPriceX96.Lt(minSqrtRatio) || p.sqrtPriceX96.Gt(maxSqrtRatio) {
			panic(newErrorWithDetail(
				errOutOfRange,
				ufmt.Sprintf("sqrtPriceX96(%s) is out of range", p.sqrtPriceX96.ToString()),
			))
		}

		// newPrice = 2^192 / oldPrice
		newPrice := u256.Zero().Div(Q192, p.sqrtPriceX96)

		// Check if calculated price is within valid range
		if newPrice.Lt(minSqrtRatio) || newPrice.Gt(maxSqrtRatio) {
			panic(newErrorWithDetail(
				errOutOfRange,
				ufmt.Sprintf("sqrtPriceX96(%s) is out of range", newPrice.ToString()),
			))
		}

		p.sqrtPriceX96 = newPrice
	}
	return newPoolParams(token0Path, token1Path, p.fee, p.sqrtPriceX96.ToString(), GetFeeAmountTickSpacing(p.fee))
}

func (p *poolCreateConfig) isSameTokenPath() bool {
	return p.token0Path == p.token1Path
}

// isInOrder checks if token paths are in lexicographical (or, alphabetical) order
func (p *poolCreateConfig) isInOrder() bool {
	if strings.Compare(p.token0Path, p.token1Path) < 0 {
		return true
	}
	return false
}

func (p *poolCreateConfig) wrap() (string, string) {
	if p.token0Path == consts.GNOT {
		p.token0Path = consts.WRAPPED_WUGNOT
	}
	if p.token1Path == consts.GNOT {
		p.token1Path = consts.WRAPPED_WUGNOT
	}
	return p.token0Path, p.token1Path
}

func (p *poolCreateConfig) generatePoolPath() string {
	if !p.isInOrder() {
		p.token0Path, p.token1Path = p.token1Path, p.token0Path
	}

	return ufmt.Sprintf("%s:%s:%d", p.token0Path, p.token1Path, p.fee)
}

func (p *poolCreateConfig) isSupportedFee(feeTier uint32) bool {
	switch feeTier {
	case FeeTier100, FeeTier500, FeeTier3000, FeeTier10000:
		return true
	}
	return false
}
