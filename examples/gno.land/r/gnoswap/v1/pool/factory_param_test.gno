package pool

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
)

func TestNewPoolParams(t *testing.T) {
	t.Run(
		"success - normal token pair",
		func(t *testing.T) {
			params := newPoolParams(
				"token0",
				"token1",
				500,       // 0.05% fee
				"1000000", // example sqrt price
				GetFeeAmountTickSpacing(500),
			)

			if params.Token0Path() != "token0" {
				t.Errorf("Expected token0Path to be 'token0', got %s", params.Token0Path())
			}

			if params.TickSpacing() != 10 { // 500 fee should have 10 tick spacing
				t.Errorf("Expected tick spacing 10, got %d", params.TickSpacing())
			}

			if !params.isInOrder() {
				t.Errorf("Expected token0Path(token0) < token1Path(token1)")
			}
		},
	)

	t.Run(
		"Success - native token wrapped wugnot token",
		func(t *testing.T) {
			params := newPoolParams(
				consts.GNOT,
				"token1",
				500,
				"1000000",
				GetFeeAmountTickSpacing(500),
			)
			token0, _ := params.wrap()
			if token0 != consts.WRAPPED_WUGNOT {
				t.Errorf("Expected GNOT to be wrapped to WUGNOT")
			}
		},
	)

	t.Run(
		"Fail - same token path", func(t *testing.T) {
			params := newPoolParams(
				"token0",
				"token0",
				500,
				"1000000",
				GetFeeAmountTickSpacing(500),
			)
			if !params.isSameTokenPath() {
				t.Errorf("Expected token0Path(token0) == token1Path(token0)")
			}
		},
	)
}

func TestNewPoolParam(t *testing.T) {
	testCases := []struct {
		name            string
		token0Path      string
		token1Path      string
		fee             uint32
		sqrtPriceX96Str string
		sqrtPriceX96    *u256.Uint
		tickSpacing     int32
		expectedPanic   bool
		expectedMsg     string
	}{
		{
			name:            "success - normal token pair",
			token0Path:      gnsPath,
			token1Path:      wugnotPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "79228162514264337593543950337",
			sqrtPriceX96:    u256.MustFromDecimal("79228162514264337593543950337"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			params := newPoolParams(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPriceX96Str, tt.tickSpacing)
			uassert.Equal(t, params.token0Path, gnsPath, "Token0Path should match")
			uassert.Equal(t, params.token1Path, wugnotPath, "Token1Path should match")
			uassert.Equal(t, params.fee, FeeTier3000, "Fee should match")
			uassert.Equal(t, params.tickSpacing, params.TickSpacing(), "TickSpacing should match")
			uassert.Equal(t, params.SqrtPriceX96().ToString(), u256.MustFromDecimal(tt.sqrtPriceX96Str).ToString(), "SqrtPriceX96 should match")
		})
	}
}

func TestIsInOrder(t *testing.T) {
	const (
		gnsPath    = "gno.land/r/gnoswap/v1/gns"
		wugnotPath = "gno.land/r/gnoswap/v1/wugnot"
	)

	testCases := []struct {
		name            string
		token0Path      string
		token1Path      string
		fee             uint32
		sqrtPriceX96Str string
		sqrtPriceX96    *u256.Uint
		tickSpacing     int32
		expectedPanic   bool
		expectedMsg     string
	}{
		{
			name:            "success - normal token pair",
			token0Path:      gnsPath,
			token1Path:      wugnotPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "79228162514264337593543950337",
			sqrtPriceX96:    u256.MustFromDecimal("79228162514264337593543950337"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			params := newPoolParams(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPriceX96Str, tt.tickSpacing)
			poolPath := params.generatePoolPath()
			uassert.Equal(t, poolPath, "gno.land/r/gnoswap/v1/gns:gno.land/r/gnoswap/v1/wugnot:3000", "Pool path should match")
			uassert.Equal(t, params.isInOrder(), true, "Tokens order should match")
		})
	}
}

func TestUpdateWithWrapping(t *testing.T) {
	const (
		gnsPath    = "gno.land/r/gnoswap/v1/gns"
		wugnotPath = "gno.land/r/gnoswap/v1/wugnot"
	)

	tick := int32(-11242)
	sqrtPriceX96FromTick := common.TickMathGetSqrtRatioAtTick(tick)

	tests := []struct {
		name            string
		token0Path      string
		token1Path      string
		fee             uint32
		sqrtPriceX96Str string
		sqrtPriceX96    *u256.Uint
		tickSpacing     int32
		expectedPanic   bool
		expectedMsg     string
	}{
		{
			name:            "success - update with wrapping",
			token0Path:      wugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "79228162514264337593543950337",
			sqrtPriceX96:    u256.MustFromDecimal("79228162514264337593543950337"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
		{
			name:            "success - near MIN_SQRT_RATIO",
			token0Path:      wugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "4295128740", // MIN_SQRT_RATIO + 1
			sqrtPriceX96:    u256.MustFromDecimal("4295128740"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
		{
			name:            "fail - below MIN_SQRT_RATIO",
			token0Path:      wugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "4295128738", // MIN_SQRT_RATIO - 1
			sqrtPriceX96:    u256.MustFromDecimal("4295128739"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   true,
			expectedMsg:     "out of range for numeric value",
		},
		{
			name:            "fail - above MAX_SQRT_RATIO",
			token0Path:      wugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "1461446703485210103287273052203988822378723970342", // MAX_SQRT_RATIO + 1
			sqrtPriceX96:    u256.MustFromDecimal("1461446703485210103287273052203988822378723970342"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   true,
			expectedMsg:     "out of range for numeric value",
		},
		{
			name:            "success - update with wrapping",
			token0Path:      wugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: sqrtPriceX96FromTick.ToString(),
			sqrtPriceX96:    sqrtPriceX96FromTick,
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but got none")
					} else if !strings.Contains(r.(string), tt.expectedMsg) {
						t.Errorf("Expected panic message to contain %s, got %s", tt.expectedMsg, r)
					}
				}()
			}

			params := newPoolParams(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPriceX96Str, tt.tickSpacing)
			updatedParams := params.updateWithWrapping()

			if !tt.expectedPanic {
				expected0 := params.token0Path
				expected1 := params.token1Path
				if !params.isInOrder() {
					expected0, expected1 = expected1, expected0
				}

				uassert.Equal(t, expected0, updatedParams.Token0Path(), "Token0Path should be swapped if needed")
				uassert.Equal(t, expected1, updatedParams.Token1Path(), "Token1Path should be swapped if needed")

				// Verify that the inverted price is within valid range
				if !params.isInOrder() {
					minSqrtRatio := u256.MustFromDecimal(consts.MIN_SQRT_RATIO)
					maxSqrtRatio := u256.MustFromDecimal(consts.MAX_SQRT_RATIO)
					newSqrtPriceX96 := updatedParams.SqrtPriceX96()

					if newSqrtPriceX96.Lt(minSqrtRatio) || newSqrtPriceX96.Gt(maxSqrtRatio) {
						t.Errorf("Inverted price %s is out of valid range [%s, %s]",
							newSqrtPriceX96.ToString(),
							minSqrtRatio.ToString(),
							maxSqrtRatio.ToString())
					}
				}
			}
		})
	}
}

func TestWrap(t *testing.T) {
	testCases := []struct {
		name            string
		token0Path      string
		token1Path      string
		fee             uint32
		sqrtPriceX96Str string
		sqrtPriceX96    *u256.Uint
		tickSpacing     int32
		expectedPanic   bool
		expectedMsg     string
	}{
		{
			name:            "success - native token wrapped wugnot token",
			token0Path:      ugnotPath,
			token1Path:      gnsPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "79228162514264337593543950337",
			sqrtPriceX96:    u256.MustFromDecimal("79228162514264337593543950337"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			params := newPoolParams(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPriceX96Str, tt.tickSpacing)
			wrapped0, wrapped1 := params.wrap()

			expected0 := params.token0Path
			expected1 := params.token1Path

			if params.token0Path == consts.GNOT {
				expected0 = consts.WRAPPED_WUGNOT
			}
			if params.token1Path == consts.GNOT {
				expected1 = consts.WRAPPED_WUGNOT
			}

			uassert.Equal(t, expected0, wrapped0, "Token0 should be wrapped if GNOT")
			uassert.Equal(t, expected1, wrapped1, "Token1 should be wrapped if GNOT")
		})
	}
}

func TestGeneratePoolPath(t *testing.T) {
	testCases := []struct {
		name            string
		token0Path      string
		token1Path      string
		fee             uint32
		sqrtPriceX96Str string
		sqrtPriceX96    *u256.Uint
		tickSpacing     int32
		expectedPanic   bool
		expectedMsg     string
	}{
		{
			name:            "success - pool path generated correctly",
			token0Path:      gnsPath,
			token1Path:      wugnotPath,
			fee:             FeeTier3000,
			sqrtPriceX96Str: "79228162514264337593543950337",
			sqrtPriceX96:    u256.MustFromDecimal("79228162514264337593543950337"),
			tickSpacing:     GetFeeAmountTickSpacing(FeeTier3000),
			expectedPanic:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			params := newPoolParams(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPriceX96Str, tt.tickSpacing)
			expected := ufmt.Sprintf("%s:%s:%d", params.token0Path, params.token1Path, params.fee)
			if !params.isInOrder() {
				expected = ufmt.Sprintf("%s:%s:%d", params.token1Path, params.token0Path, params.fee)
			}
			actual := params.generatePoolPath()

			uassert.Equal(t, expected, actual, "Pool path should match expected")
		})
	}
}
