package position

import (
	"std"
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gnft"
)

var (
	mockRegistry = make(map[string]bool)
)

// MockRegister registers a token path in the mock registry.
func MockRegister(t *testing.T, tokenPath string) {
	t.Helper()
	mockRegistry[tokenPath] = true
}

// MockUnregister unregisters a token path in the mock registry.
func MockUnregister(t *testing.T, tokenPath string) {
	t.Helper()
	delete(mockRegistry, tokenPath)
}

// IsRegistered checks if a token path is registered in the mock registry.
func IsRegistered(t *testing.T, tokenPath string) error {
	t.Helper()
	if mockRegistry[tokenPath] {
		return nil
	}
	return errInvalidTokenPath
}

func newDummyPosition(tokenId uint64) Position {
	return Position{
		nonce:                    u256.NewUint(tokenId),
		operator:                 "user1",
		poolKey:                  "poolKey1",
		tickLower:                -500,
		tickUpper:                500,
		liquidity:                u256.NewUint(1000),
		feeGrowthInside0LastX128: u256.NewUint(10),
		feeGrowthInside1LastX128: u256.NewUint(20),
		tokensOwed0:              u256.NewUint(30),
		tokensOwed1:              u256.NewUint(40),
		burned:                   false,
	}
}

func TestMustGetPosition(t *testing.T) {
	tokenId := uint64(1)
	position := newDummyPosition(tokenId)
	setPosition(tokenId, position)

	t.Run("Success - Get existing position", func(t *testing.T) {
		result := MustGetPosition(tokenId)
		uassert.Equal(t, position.nonce.ToString(), result.nonce.ToString())
		uassert.Equal(t, position.operator, result.operator)
		uassert.Equal(t, position.poolKey, result.poolKey)
		uassert.Equal(t, position.tickLower, result.tickLower)
		uassert.Equal(t, position.tickUpper, result.tickUpper)
		uassert.Equal(t, position.liquidity.ToString(), result.liquidity.ToString())
		uassert.Equal(t, position.feeGrowthInside0LastX128.ToString(), result.feeGrowthInside0LastX128.ToString())
		uassert.Equal(t, position.feeGrowthInside1LastX128.ToString(), result.feeGrowthInside1LastX128.ToString())
		uassert.Equal(t, position.tokensOwed0.ToString(), result.tokensOwed0.ToString())
		uassert.Equal(t, position.tokensOwed1.ToString(), result.tokensOwed1.ToString())
		uassert.Equal(t, position.burned, result.burned)
	})

	t.Run("Fail - Non-existent position", func(t *testing.T) {
		uassert.PanicsWithMessage(t,
			"[GNOSWAP-POSITION-013] position does not exist || position with tokenId(999) doesn't exist",
			func() {
				MustGetPosition(999)
			})
	})
}

func TestGetPosition(t *testing.T) {
	tokenId := uint64(1)
	position := newDummyPosition(tokenId)
	t.Run("Success - Get existing position", func(t *testing.T) {
		result := MustGetPosition(tokenId)
		uassert.Equal(t, position.nonce.ToString(), result.nonce.ToString())
		uassert.Equal(t, position.operator, result.operator)
		uassert.Equal(t, position.poolKey, result.poolKey)
		uassert.Equal(t, position.tickLower, result.tickLower)
		uassert.Equal(t, position.tickUpper, result.tickUpper)
		uassert.Equal(t, position.liquidity.ToString(), result.liquidity.ToString())
		uassert.Equal(t, position.feeGrowthInside0LastX128.ToString(), result.feeGrowthInside0LastX128.ToString())
		uassert.Equal(t, position.feeGrowthInside1LastX128.ToString(), result.feeGrowthInside1LastX128.ToString())
		uassert.Equal(t, position.tokensOwed0.ToString(), result.tokensOwed0.ToString())
		uassert.Equal(t, position.tokensOwed1.ToString(), result.tokensOwed1.ToString())
		uassert.Equal(t, position.burned, result.burned)
		removePosition(tokenId)
	})
}

func TestSetPosition(t *testing.T) {
	tokenId := uint64(300)
	position := newDummyPosition(tokenId)

	t.Run("Success - Create new position", func(t *testing.T) {
		isNew := setPosition(tokenId, position)
		uassert.False(t, isNew)

		storedPosition, _ := GetPosition(tokenId)
		uassert.Equal(t, position.nonce.ToString(), storedPosition.nonce.ToString())
		uassert.Equal(t, position.operator, storedPosition.operator)
		uassert.Equal(t, position.poolKey, storedPosition.poolKey)
		uassert.Equal(t, position.tickLower, storedPosition.tickLower)
		uassert.Equal(t, position.tickUpper, storedPosition.tickUpper)
		uassert.Equal(t, position.liquidity.ToString(), storedPosition.liquidity.ToString())
		uassert.Equal(t, position.feeGrowthInside0LastX128.ToString(), storedPosition.feeGrowthInside0LastX128.ToString())
		uassert.Equal(t, position.feeGrowthInside1LastX128.ToString(), storedPosition.feeGrowthInside1LastX128.ToString())
		uassert.Equal(t, position.tokensOwed0.ToString(), storedPosition.tokensOwed0.ToString())
		uassert.Equal(t, position.tokensOwed1.ToString(), storedPosition.tokensOwed1.ToString())
		uassert.Equal(t, position.burned, storedPosition.burned)
	})

	t.Run("Success - Update existing position", func(t *testing.T) {
		position.tickUpper = 1000
		isNew := setPosition(tokenId, position)
		uassert.True(t, isNew)

		updatedPosition, _ := GetPosition(tokenId)
		uassert.Equal(t, int32(1000), updatedPosition.tickUpper)
	})
}

func TestRemovePosition(t *testing.T) {
	tokenId := uint64(2)
	position := newDummyPosition(tokenId)

	t.Run("Success - Remove existing position", func(t *testing.T) {
		setPosition(tokenId, position)
		removePosition(tokenId)
		_, found := GetPosition(tokenId)
		uassert.False(t, found)
	})

	t.Run("Success - Remove non-existent position", func(t *testing.T) {
		removePosition(tokenId)
		_, found := GetPosition(tokenId)
		uassert.False(t, found)
	})
}

func TestExistPosition(t *testing.T) {
	tokenId := uint64(2)
	position := newDummyPosition(tokenId)

	t.Run("False - Position does not exist", func(t *testing.T) {
		uassert.False(t, existPosition(tokenId))
	})

	t.Run("True - Position exists", func(t *testing.T) {
		setPosition(tokenId, position)
		uassert.True(t, existPosition(tokenId))
		removePosition(tokenId)
	})
}

func TestComputePositionKey(t *testing.T) {
	tests := []struct {
		name      string
		owner     std.Address
		tickLower int32
		tickUpper int32
		expected  string
	}{
		{
			name:      "Basic Position Key",
			owner:     alice,
			tickLower: -100,
			tickUpper: 200,
			expected:  "ZzF2OWt4amNtOXRhMDQ3aDZsdGEwNDdoNmx0YTA0N2g2bHpkNDBnaF9fLTEwMF9fMjAw", // Base64 of "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh__-100__200"
		},
		{
			name:      "Zero Ticks",
			owner:     alice,
			tickLower: 0,
			tickUpper: 0,
			expected:  "ZzF2OWt4amNtOXRhMDQ3aDZsdGEwNDdoNmx0YTA0N2g2bHpkNDBnaF9fMF9fMA==", // Base64 of "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh__0__0"
		},
		{
			name:      "Negative Lower Tick",
			owner:     alice,
			tickLower: -50,
			tickUpper: 150,
			expected:  "ZzF2OWt4amNtOXRhMDQ3aDZsdGEwNDdoNmx0YTA0N2g2bHpkNDBnaF9fLTUwX18xNTA=", // Base64 of "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh__-50__150"
		},
		{
			name:      "Same Tick Bounds",
			owner:     alice,
			tickLower: 300,
			tickUpper: 300,
			expected:  "ZzF2OWt4amNtOXRhMDQ3aDZsdGEwNDdoNmx0YTA0N2g2bHpkNDBnaF9fMzAwX18zMDA=", // Base64 of "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh__300__300"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computePositionKey(tt.owner, tt.tickLower, tt.tickUpper)

			if result != tt.expected {
				t.Errorf("expected %s but got %s", tt.expected, result)
			}
		})
	}
}

func TestNextIdFunctions(t *testing.T) {
	nextId = uint64(1)

	t.Run("Initial nextId should return 1", func(t *testing.T) {
		uassert.Equal(t, uint64(1), getNextId(), "expected nextId to start at 1")
	})

	t.Run("After mint nextId should return 2", func(t *testing.T) {
		incrementNextId()
		uassert.Equal(t, getNextId(), uint64(2), "expected nextId to be 2 after mint")
	})

	t.Run("Increment nextId once", func(t *testing.T) {
		incrementNextId()
		uassert.Equal(t, uint64(3), getNextId(), "expected nextId to increment to 3")
	})

	t.Run("Increment nextId multiple times", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			incrementNextId()
		}
		uassert.Equal(t, uint64(5), getNextId(), "expected nextId to increment to 5 after 3 more increments")
	})

	t.Run("Ensure no overflow on normal increments", func(t *testing.T) {
		for i := uint64(5); i < 100; i++ {
			incrementNextId()
		}
		uassert.Equal(t, uint64(100), getNextId(), "expected nextId to reach 100 after continuous increments")
	})
}

func TestIsValidTokenPath(t *testing.T) {
	tests := []struct {
		name      string
		tokenPath string
		register  bool
		expected  bool
	}{
		{
			name:      "Valid Token Path",
			tokenPath: gnsPath,
			register:  true,
			expected:  true,
		},
		{
			name:      "Invalid Token Path",
			tokenPath: "invalid/path",
			register:  false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTokenPath(tt.tokenPath)
			uassert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTokenPath(t *testing.T) {
	tests := []struct {
		name      string
		token0    string
		token1    string
		register0 bool
		register1 bool
		expected  error
	}{
		{
			name:      "Valid Token Path",
			token0:    gnsPath,
			token1:    barPath,
			register0: true,
			register1: true,
			expected:  nil,
		},
		{
			name:      "Same Token Path",
			token0:    "tokenA",
			token1:    "tokenA",
			register0: true,
			register1: true,
			expected:  errInvalidTokenPath,
		},
		{
			name:      "Conflicting Tokens (GNOT ↔ WUGNOT)",
			token0:    consts.GNOT,
			token1:    consts.WRAPPED_WUGNOT,
			register0: true,
			register1: true,
			expected:  errInvalidTokenPath,
		},
		{
			name:      "Invalid Token Path",
			token0:    "tokenX",
			token1:    "tokenY",
			register0: false,
			register1: true,
			expected:  errInvalidTokenPath,
		},
		{
			name:      "Both Invalid Tokens",
			token0:    "invalidA",
			token1:    "invalidB",
			register0: false,
			register1: false,
			expected:  errInvalidTokenPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.register0 {
				MockRegister(t, tt.token0)
			} else {
				MockUnregister(t, tt.token0)
			}

			if tt.register1 {
				MockRegister(t, tt.token1)
			} else {
				MockUnregister(t, tt.token1)
			}

			if tt.expected != nil {
				err := validateTokenPath(tt.token0, tt.token1)
				uassert.Equal(t, tt.expected.Error(), err.Error())
			} else {
				uassert.NotPanics(t, func() {
					err := validateTokenPath(tt.token0, tt.token1)
					if err != nil {
						t.Errorf("expected no error but got %s", err.Error())
					}
				})
			}
		})
	}
}

func TestProcessTokens(t *testing.T) {
	tests := []struct {
		name           string
		token0         string
		token1         string
		amount0Desired string
		amount1Desired string
		caller         std.Address
		expected0      string
		expected1      string
		isNative0      bool
		isNative1      bool
		expectedWrap   uint64
		expectPanic    bool
		expectMsg      string
	}{
		{
			name:           "Both tokens valid and not native",
			token0:         gnsPath,
			token1:         "tokenB",
			amount0Desired: "100",
			amount1Desired: "200",
			caller:         alice,
			expected0:      "tokenA",
			expected1:      "tokenB",
			isNative0:      false,
			isNative1:      false,
			expectedWrap:   0,
			expectPanic:    true,
			expectMsg:      "[GNOSWAP-POSITION-016] invalid token address || token0(gno.land/r/gnoswap/v1/gns), token1(tokenB)",
		},
		{
			name:           "token0 is native",
			token0:         consts.GNOT,
			token1:         gnsPath,
			amount0Desired: "1300",
			amount1Desired: "200",
			caller:         alice,
			expected0:      consts.WRAPPED_WUGNOT,
			expected1:      "gno.land/r/gnoswap/v1/gns",
			isNative0:      true,
			isNative1:      false,
			expectedWrap:   1300,
			expectPanic:    false,
			expectMsg:      "[GNOSWAP-POSITION-016] invalid token address || token0(gnot), token1(gno.land/r/gnoswap/v1/gns)",
		},
		{
			name:           "token1 is native",
			token0:         gnsPath,
			token1:         consts.GNOT,
			amount0Desired: "150",
			amount1Desired: "1250",
			caller:         testutils.TestAddress("user3"),
			expected0:      "gno.land/r/gnoswap/v1/gns",
			expected1:      consts.WRAPPED_WUGNOT,
			isNative0:      false,
			isNative1:      true,
			expectedWrap:   1250,
			expectPanic:    false,
		},
		{
			name:           "Both tokens are native",
			token0:         consts.GNOT,
			token1:         consts.GNOT,
			amount0Desired: "1100",
			amount1Desired: "1200",
			caller:         testutils.TestAddress("user4"),
			expected0:      consts.WRAPPED_WUGNOT,
			expected1:      consts.WRAPPED_WUGNOT,
			isNative0:      true,
			isNative1:      true,
			expectedWrap:   2300,
			expectPanic:    true,
			expectMsg:      "[GNOSWAP-POSITION-016] invalid token address || token0(gnot), token1(gnot)",
		},
		{
			name:           "Invalid token path",
			token0:         "invalidToken",
			token1:         "tokenB",
			amount0Desired: "150",
			amount1Desired: "200",
			caller:         testutils.TestAddress("user5"),
			expectPanic:    true,
			expectMsg:      "[GNOSWAP-POSITION-016] invalid token address || token0(invalidToken), token1(tokenB)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MockRegister(t, tt.token0)
			MockRegister(t, tt.token1)

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()

			if tt.token0 == consts.GNOT {
				amount, _ := strconv.ParseUint(tt.amount0Desired, 10, 64)
				ugnotFaucet(t, consts.POSITION_ADDR, amount)
				std.TestSetRealm(std.NewUserRealm(consts.POSITION_ADDR))
				transferUGNOT(consts.POSITION_ADDR, consts.POSITION_ADDR, amount)
			}
			if tt.token1 == consts.GNOT {
				amount, _ := strconv.ParseUint(tt.amount1Desired, 10, 64)
				ugnotFaucet(t, consts.POSITION_ADDR, amount)
				std.TestSetRealm(std.NewUserRealm(consts.POSITION_ADDR))
				transferUGNOT(consts.POSITION_ADDR, consts.POSITION_ADDR, amount)
			}

			if !tt.expectPanic {
				token0, token1, native0, native1, wrapped := processTokens(
					tt.token0,
					tt.token1,
					tt.amount0Desired,
					tt.amount1Desired,
					tt.caller,
				)

				uassert.Equal(t, tt.expected0, token0)
				uassert.Equal(t, tt.expected1, token1)
				uassert.Equal(t, tt.isNative0, native0)
				uassert.Equal(t, tt.isNative1, native1)
				uassert.Equal(t, tt.expectedWrap, wrapped)
			} else {
				uassert.PanicsWithMessage(t, tt.expectMsg, func() {
					processTokens(tt.token0, tt.token1, tt.amount0Desired, tt.amount1Desired, tt.caller)
				})
			}
		})
	}
}

func TestProcessMintInput(t *testing.T) {
	tests := []struct {
		name            string
		input           MintInput
		expectedToken0  string
		expectedToken1  string
		expectedAmount0 string
		expectedAmount1 string
		expectedTickL   int32
		expectedTickU   int32
		expectError     bool
	}{
		{
			name: "Standard Mint - Token0 < Token1",
			input: MintInput{
				token0:         gnsPath,
				token1:         barPath,
				amount0Desired: "1000",
				amount1Desired: "2000",
				amount0Min:     "800",
				amount1Min:     "1800",
				tickLower:      -10000,
				tickUpper:      10000,
				caller:         alice,
			},
			expectedToken0:  gnsPath,
			expectedToken1:  barPath,
			expectedAmount0: "1000",
			expectedAmount1: "2000",
			expectedTickL:   -10000,
			expectedTickU:   10000,
			expectError:     false,
		},
		{
			name: "Token Swap - Token1 < Token0",
			input: MintInput{
				token0:         barPath,
				token1:         gnsPath,
				amount0Desired: "2000",
				amount1Desired: "1000",
				amount0Min:     "1800",
				amount1Min:     "800",
				tickLower:      -20000,
				tickUpper:      20000,
				caller:         alice,
			},
			expectedToken0:  gnsPath,
			expectedToken1:  barPath,
			expectedAmount0: "1000",
			expectedAmount1: "2000",
			expectedTickL:   -20000,
			expectedTickU:   20000,
			expectError:     false,
		},
		{
			name: "Error Case - Invalid Amounts",
			input: MintInput{
				token0:         gnsPath,
				token1:         barPath,
				amount0Desired: "invalid",
				amount1Desired: "2000",
				amount0Min:     "800",
				amount1Min:     "1800",
				tickLower:      -5000,
				tickUpper:      5000,
				caller:         alice,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectError {
						t.Errorf("unexpected panic: %v", r)
					} else {
						uassert.Equal(t, "[GNOSWAP-POSITION-005] invalid input data || input string : invalid", r)
					}
				}
			}()
			processed, err := processMintInput(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				uassert.Equal(t, tt.expectedToken0, processed.tokenPair.token0)
				uassert.Equal(t, tt.expectedToken1, processed.tokenPair.token1)
				uassert.Equal(t, tt.expectedAmount0, processed.amount0Desired.ToString())
				uassert.Equal(t, tt.expectedAmount1, processed.amount1Desired.ToString())
				uassert.Equal(t, tt.expectedTickL, processed.tickLower)
				uassert.Equal(t, tt.expectedTickU, processed.tickUpper)
			}
		})
	}
}

func TestMintInternal(t *testing.T) {
	MakeMintPositionWithoutFee(t)
	TokenFaucet(t, fooPath, alice)
	TokenFaucet(t, barPath, alice)
	TokenApprove(t, fooPath, alice, pool, maxApprove)
	TokenApprove(t, barPath, alice, pool, maxApprove)

	tests := []struct {
		name              string
		params            MintParams
		expectedTokenId   uint64
		expectedLiquidity string
		expectedAmount0   string
		expectedAmount1   string
		expectPanic       bool
		expectedError     string
		tokenId           uint64
	}{
		{
			name: "Successful Mint",
			params: MintParams{
				token0:         barPath,
				token1:         fooPath,
				fee:            fee500,
				tickLower:      -100,
				tickUpper:      100,
				amount0Desired: u256.MustFromDecimal("10000"),
				amount1Desired: u256.MustFromDecimal("10000"),
				amount0Min:     u256.MustFromDecimal("10"),
				amount1Min:     u256.MustFromDecimal("10"),
				caller:         alice,
				mintTo:         alice,
			},
			expectedTokenId:   101,
			expectedLiquidity: "2005104",
			expectedAmount0:   "10000",
			expectedAmount1:   "10000",
			expectPanic:       false,
		},
		{
			name: "Position Exists",
			params: MintParams{
				token0:         barPath,
				token1:         fooPath,
				fee:            fee500,
				tickLower:      -100,
				tickUpper:      100,
				amount0Desired: u256.MustFromDecimal("10000"),
				amount1Desired: u256.MustFromDecimal("10000"),
				amount0Min:     u256.MustFromDecimal("10"),
				amount1Min:     u256.MustFromDecimal("10"),
				caller:         alice,
				mintTo:         alice,
			},
			expectPanic:   true,
			expectedError: "token id already exists",
			tokenId:       1,
		},
		{
			name: "Zero Liquidity Mint",
			params: MintParams{
				token0:         barPath,
				token1:         fooPath,
				fee:            fee500,
				tickLower:      -200,
				tickUpper:      200,
				amount0Desired: u256.Zero(),
				amount1Desired: u256.Zero(),
				amount0Min:     u256.NewUint(5),
				amount1Min:     u256.NewUint(5),
				caller:         alice,
				mintTo:         alice,
			},
			expectPanic:   true,
			expectedError: "[GNOSWAP-POOL-010] zero liquidity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tokenId != 0 {
				nextId = tt.tokenId
			}

			if !tt.expectPanic {
				std.TestSetRealm(std.NewUserRealm(tt.params.mintTo))
				tokenId, liquidity, amount0, amount1 := mint(tt.params)
				uassert.Equal(t, tt.expectedTokenId, tokenId)
				uassert.Equal(t, tt.expectedLiquidity, liquidity.ToString())
				uassert.Equal(t, tt.expectedAmount0, amount0.ToString())
				uassert.Equal(t, tt.expectedAmount1, amount1.ToString())
			} else {
				uassert.PanicsWithMessage(t, tt.expectedError, func() {
					mint(tt.params)
				})
			}
		})
	}
}

func TestMint(t *testing.T) {
	nextId = gnft.TotalSupply() + 1
	tests := []struct {
		name            string
		token0          string
		token1          string
		fee             uint32
		tickLower       int32
		tickUpper       int32
		amount0         string
		amount1         string
		minAmount0      string
		minAmount1      string
		deadline        int64
		mintTo          std.Address
		caller          std.Address
		expectPanic     bool
		expectedError   string
		expectedAmount0 uint64
		expectedAmount1 uint64
	}{
		{
			name:            "Success - Valid mint request",
			token0:          barPath,
			token1:          fooPath,
			fee:             fee500,
			tickLower:       -500,
			tickUpper:       500,
			amount0:         "1000000",
			amount1:         "2000000",
			minAmount0:      "950000",
			minAmount1:      "900000",
			deadline:        time.Now().Add(10 * time.Minute).Unix(),
			mintTo:          alice,
			caller:          alice,
			expectPanic:     false,
			expectedAmount0: 1000000,
			expectedAmount1: 1000000,
		},
		{
			name:            "Fail - Deadline exceeded",
			token0:          barPath,
			token1:          fooPath,
			fee:             fee500,
			tickLower:       -500,
			tickUpper:       500,
			amount0:         "1000000",
			amount1:         "2000000",
			minAmount0:      "950000",
			minAmount1:      "1900000",
			deadline:        time.Now().Add(-10 * time.Minute).Unix(),
			mintTo:          alice,
			caller:          alice,
			expectPanic:     true,
			expectedError:   "[GNOSWAP-POSITION-007] transaction expired || transaction too old, now(1234567890) > deadline(1234567290)",
			expectedAmount0: 950000,
			expectedAmount1: 1900000,
		},
		{
			name:            "Fail - Invalid tick range",
			token0:          barPath,
			token1:          fooPath,
			fee:             fee500,
			tickLower:       600,
			tickUpper:       500,
			amount0:         "1000000",
			amount1:         "2000000",
			minAmount0:      "950000",
			minAmount1:      "1900000",
			deadline:        time.Now().Add(10 * time.Minute).Unix(),
			mintTo:          alice,
			caller:          alice,
			expectPanic:     true,
			expectedError:   "[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(600), tickUpper(500)",
			expectedAmount0: 950000,
			expectedAmount1: 1900000,
		},
		{
			name:            "Fail - Caller not authorized",
			token0:          barPath,
			token1:          fooPath,
			fee:             fee500,
			tickLower:       -500,
			tickUpper:       500,
			amount0:         "1000000",
			amount1:         "2000000",
			minAmount0:      "950000",
			minAmount1:      "1900000",
			deadline:        time.Now().Add(10 * time.Minute).Unix(),
			mintTo:          admin,
			caller:          alice,
			expectPanic:     true,
			expectedError:   "[GNOSWAP-POSITION-012] invalid address || (g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh, g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d)",
			expectedAmount0: 950000,
			expectedAmount1: 1900000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetRealm(std.NewUserRealm(tc.caller))
			if tc.expectPanic {
				uassert.PanicsWithMessage(t,
					tc.expectedError,
					func() {
						Mint(
							tc.token0,
							tc.token1,
							tc.fee,
							tc.tickLower,
							tc.tickUpper,
							tc.amount0,
							tc.amount1,
							tc.minAmount0,
							tc.minAmount1,
							tc.deadline,
							tc.mintTo,
							tc.caller,
						)
					})
			} else {
				tokenId, liquidity, amount0, amount1 := Mint(
					tc.token0,
					tc.token1,
					tc.fee,
					tc.tickLower,
					tc.tickUpper,
					tc.amount0,
					tc.amount1,
					tc.minAmount0,
					tc.minAmount1,
					tc.deadline,
					tc.mintTo,
					tc.caller,
				)

				uassert.Equal(t, uint64(5), tokenId)
				uassert.NotEmpty(t, liquidity)
				t0, _ := strconv.ParseUint(amount0, 10, 64)
				t1, _ := strconv.ParseUint(amount1, 10, 64)
				uassert.Equal(t, tc.expectedAmount0, t0)
				uassert.Equal(t, tc.expectedAmount1, t1)
			}
		})
	}
}

func TestIncreaseLiquidityInternal(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(admin))
	position := Position{
		poolKey:                  "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500",
		tickLower:                -10000,
		tickUpper:                10000,
		liquidity:                u256.NewUint(1000000),
		feeGrowthInside0LastX128: u256.Zero(),
		feeGrowthInside1LastX128: u256.Zero(),
		tokensOwed0:              u256.Zero(),
		tokensOwed1:              u256.Zero(),
		burned:                   false,
	}
	id := getNextId()
	setPosition(id, position)
	incrementNextId()

	MakeMintPositionWithoutFee(t)

	tests := []struct {
		name              string
		params            IncreaseLiquidityParams
		expectedLiquidity string
		expectedAmount0   string
		expectedAmount1   string
		expectPanic       bool
		expectedErrorMsg  string
	}{
		{
			name: "Fail - Position Does Not Exist",
			params: IncreaseLiquidityParams{
				tokenId:        999,
				amount0Desired: u256.NewUint(1000000),
				amount1Desired: u256.NewUint(2000000),
				amount0Min:     u256.NewUint(900000),
				amount1Min:     u256.NewUint(1800000),
				deadline:       time.Now().Add(10 * time.Minute).Unix(),
			},
			expectPanic:      true,
			expectedErrorMsg: "[GNOSWAP-POSITION-006] requested data not found || tokenId(999) doesn't exist",
		},
		{
			name: "Fail - Zero Liquidity",
			params: IncreaseLiquidityParams{
				tokenId:        7,
				amount0Desired: u256.Zero(),
				amount1Desired: u256.Zero(),
				amount0Min:     u256.NewUint(900000),
				amount1Min:     u256.NewUint(1800000),
				deadline:       time.Now().Add(10 * time.Minute).Unix(),
			},
			expectPanic:      true,
			expectedErrorMsg: "[GNOSWAP-POOL-010] zero liquidity",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				uassert.PanicsWithMessage(t,
					tc.expectedErrorMsg,
					func() {
						increaseLiquidity(tc.params)
					})
			} else {
				tokenId, liquidity, amount0, amount1, _ := increaseLiquidity(tc.params)
				uassert.Equal(t, tc.params.tokenId, tokenId)
				uassert.Equal(t, tc.expectedLiquidity, liquidity.ToString())
				uassert.Equal(t, tc.expectedAmount0, amount0.ToString())
				uassert.Equal(t, tc.expectedAmount1, amount1.ToString())
			}
		})
	}
}

func TestIncreaseLiquidity(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(admin))
	tests := []struct {
		name             string
		tokenId          uint64
		amount0Desired   string
		amount1Desired   string
		amount0Min       string
		amount1Min       string
		deadline         int64
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
		expectedErrorMsg string
	}{
		{
			name:            "Success - Valid Increase",
			tokenId:         7,
			amount0Desired:  "1000000",
			amount1Desired:  "2000000",
			amount0Min:      "950000",
			amount1Min:      "900000",
			deadline:        time.Now().Add(10 * time.Minute).Unix(),
			expectedAmount0: "1000000",
			expectedAmount1: "1000000",
			expectPanic:     false,
		},
		{
			name:             "Fail - Deadline Exceeded",
			tokenId:          7,
			amount0Desired:   "1000000",
			amount1Desired:   "2000000",
			amount0Min:       "950000",
			amount1Min:       "900000",
			deadline:         time.Now().Add(-10 * time.Minute).Unix(),
			expectPanic:      true,
			expectedErrorMsg: "[GNOSWAP-POSITION-007] transaction expired || transaction too old, now(1234567890) > deadline(1234567290)",
		},
		{
			name:             "Fail - Invalid Amount String",
			tokenId:          7,
			amount0Desired:   "invalid_amount",
			amount1Desired:   "2000000",
			amount0Min:       "950000",
			amount1Min:       "900000",
			deadline:         time.Now().Add(10 * time.Minute).Unix(),
			expectPanic:      true,
			expectedErrorMsg: "[GNOSWAP-POSITION-005] invalid input data || input string : invalid_amount",
		},
		{
			name:             "Fail - Position Does Not Exist",
			tokenId:          999,
			amount0Desired:   "1000000",
			amount1Desired:   "2000000",
			amount0Min:       "950000",
			amount1Min:       "900000",
			deadline:         time.Now().Add(10 * time.Minute).Unix(),
			expectPanic:      true,
			expectedErrorMsg: "[GNOSWAP-POSITION-013] position does not exist || position with tokenId(999) doesn't exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				uassert.PanicsWithMessage(t,
					tc.expectedErrorMsg,
					func() {
						IncreaseLiquidity(
							tc.tokenId,
							tc.amount0Desired,
							tc.amount1Desired,
							tc.amount0Min,
							tc.amount1Min,
							tc.deadline,
						)
					})
			} else {
				tokenId, _, amount0, amount1, _ := IncreaseLiquidity(
					tc.tokenId,
					tc.amount0Desired,
					tc.amount1Desired,
					tc.amount0Min,
					tc.amount1Min,
					tc.deadline,
				)
				uassert.Equal(t, tc.tokenId, tokenId)
				uassert.Equal(t, tc.expectedAmount0, amount0)
				uassert.Equal(t, tc.expectedAmount1, amount1)
			}
		})
	}
}

func TestDecreaseLiquidity(t *testing.T) {
	tests := []struct {
		name              string
		beforeIncrease    bool
		increaseAmount0   string
		increaseAmount1   string
		liquidityToRemove string
		amount0Min        string
		amount1Min        string
		deadlineOffset    time.Duration
		unwrapResult      bool
		expectPanic       bool
		expectedErrorMsg  string
		expectedLiquidity string
		expectedFee0      string
		expectedFee1      string
	}{
		{
			name:              "Success - Decrease 50% Liquidity",
			liquidityToRemove: "400000",
			amount0Min:        "8000",
			amount1Min:        "15000",
			deadlineOffset:    time.Hour,
			unwrapResult:      false,
			expectPanic:       false,
			expectedLiquidity: "400000",
			expectedFee0:      "0",
			expectedFee1:      "0",
		},
		{
			name:              "Success - Decrease 100% Liquidity",
			liquidityToRemove: "600000",
			amount0Min:        "10000",
			amount1Min:        "20000",
			deadlineOffset:    time.Hour,
			unwrapResult:      false,
			expectPanic:       false,
			expectedLiquidity: "600000",
			expectedFee0:      "0",
			expectedFee1:      "0",
			expectedErrorMsg:  "",
		},
		{
			name:              "Fail - Zero Liquidity",
			liquidityToRemove: "0",
			amount0Min:        "10000",
			amount1Min:        "20000",
			deadlineOffset:    time.Hour,
			unwrapResult:      false,
			expectPanic:       true,
			expectedErrorMsg:  "[GNOSWAP-POSITION-010] zero liquidity || liquidity amount must be greater than 0, got 0",
		},
		{
			name:              "Fail - Underflow",
			liquidityToRemove: "1541592",
			amount0Min:        "200000",
			amount1Min:        "400000",
			deadlineOffset:    time.Hour,
			unwrapResult:      false,
			expectPanic:       true,
			expectedErrorMsg:  "[GNOSWAP-POOL-010] zero liquidity || both liquidityDelta and current position's liquidity are zero",
		},
		{
			name:              "Fail - Deadline Exceeded",
			liquidityToRemove: "50",
			amount0Min:        "10000",
			amount1Min:        "20000",
			deadlineOffset:    -time.Hour,
			unwrapResult:      false,
			expectPanic:       true,
			expectedErrorMsg:  "[GNOSWAP-POSITION-007] transaction expired || transaction too old, now(1234567890) > deadline(1234564290)",
		},
		{
			name:              "Success - Increase and Decrease",
			beforeIncrease:    true,
			increaseAmount0:   "500000",
			increaseAmount1:   "1000000",
			liquidityToRemove: "1270796",
			amount0Min:        "12000",
			amount1Min:        "25000",
			deadlineOffset:    time.Hour,
			unwrapResult:      false,
			expectPanic:       false,
			expectedLiquidity: "1270796",
			expectedFee0:      "0",
			expectedFee1:      "0",
		},
	}

	CreatePoolWithoutFee(t)
	std.TestSetRealm(std.NewUserRealm(admin))
	TokenApprove(t, barPath, admin, pool, consts.UINT64_MAX)
	TokenApprove(t, fooPath, admin, pool, consts.UINT64_MAX)
	tokenId, _, _, _ := Mint(
		barPath,
		fooPath,
		fee500,
		-10000,
		10000,
		"1000000",
		"1000000",
		"0",
		"0",
		time.Now().Add(time.Hour).Unix(),
		admin,
		admin,
	)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					} else {
						uassert.Equal(t, tc.expectedErrorMsg, r)
					}
				}
			}()

			deadline := time.Now().Add(tc.deadlineOffset).Unix()
			if tc.beforeIncrease {
				_, _, _, _, _ = IncreaseLiquidity(
					tokenId,
					tc.increaseAmount0,
					tc.increaseAmount1,
					"0",
					"0",
					deadline,
				)
			}

			if tc.expectPanic {
				DecreaseLiquidity(
					tokenId,
					tc.liquidityToRemove,
					tc.amount0Min,
					tc.amount1Min,
					deadline,
					tc.unwrapResult,
				)
			} else {
				_, liquidity, fee0, fee1, _, _, _ := DecreaseLiquidity(
					tokenId,
					tc.liquidityToRemove,
					tc.amount0Min,
					tc.amount1Min,
					deadline,
					tc.unwrapResult,
				)
				uassert.Equal(t, tc.expectedLiquidity, liquidity)
				uassert.Equal(t, tc.expectedFee0, fee0)
				uassert.Equal(t, tc.expectedFee1, fee1)
			}
		})
	}
}

func TestCollectFees(t *testing.T) {
	tests := []struct {
		name               string
		mintAmount0        string
		mintAmount1        string
		increaseAmount0    string
		increaseAmount1    string
		unwrapResult       bool
		liquidityToRemove  string
		expectedFee0       string
		expectedFee1       string
		expectedAmount0    string
		expectedAmount1    string
		expectPanic        bool
		expectedPanicError string
	}{
		{
			name:              "Success - Collect fee after multiple mints",
			mintAmount0:       "1000000",
			mintAmount1:       "2000000",
			increaseAmount0:   "500000",
			increaseAmount1:   "1000000",
			liquidityToRemove: "50",
			unwrapResult:      false,
			expectedFee0:      "0",
			expectedFee1:      "0",
			expectedAmount0:   "0",
			expectedAmount1:   "0",
			expectPanic:       false,
		},
		{
			name:              "Success - Collect fee after full liquidity removal",
			mintAmount0:       "1000000",
			mintAmount1:       "2000000",
			liquidityToRemove: "100",
			unwrapResult:      true,
			expectedFee0:      "0",
			expectedFee1:      "0",
			expectedAmount0:   "0",
			expectedAmount1:   "0",
			expectPanic:       false,
		},
		{
			name:               "Fail - No liquidity to collect",
			mintAmount0:        "100000",
			mintAmount1:        "200000",
			liquidityToRemove:  "0",
			unwrapResult:       false,
			expectPanic:        true,
			expectedPanicError: "[GNOSWAP-POSITION-010] zero liquidity || liquidity amount must be greater than 0, got 0",
		},
	}

	CreatePoolWithoutFee(t)
	std.TestSetRealm(std.NewUserRealm(admin))
	TokenApprove(t, barPath, admin, pool, consts.UINT64_MAX)
	TokenApprove(t, fooPath, admin, pool, consts.UINT64_MAX)

	// Mint liquidity
	tokenId, _, _, _ := Mint(
		barPath,
		fooPath,
		fee500,
		-10000,
		10000,
		"1000000",
		"1000000",
		"0",
		"0",
		time.Now().Add(time.Hour).Unix(),
		admin,
		admin,
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					} else {
						uassert.Equal(t, tt.expectedPanicError, r)
					}
				}
			}()

			// Increase liquidity if specified
			if tt.increaseAmount0 != "" && tt.increaseAmount1 != "" {
				_, _, _, _, _ = IncreaseLiquidity(
					tokenId,
					tt.increaseAmount0,
					tt.increaseAmount1,
					"0",
					"0",
					time.Now().Add(time.Hour).Unix(),
				)
			}

			// Decrease liquidity
			_, _, _, _, _, _, _ = DecreaseLiquidity(
				tokenId,
				tt.liquidityToRemove,
				"0",
				"0",
				time.Now().Add(time.Hour).Unix(),
				false,
			)

			if tt.expectPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicError, func() {
					CollectFee(tokenId, tt.unwrapResult)
				})
			} else {
				_, fee0, fee1, _, amount0, amount1 := CollectFee(tokenId, tt.unwrapResult)
				uassert.Equal(t, tt.expectedFee0, fee0)
				uassert.Equal(t, tt.expectedFee1, fee1)
				uassert.Equal(t, tt.expectedAmount0, amount0)
				uassert.Equal(t, tt.expectedAmount1, amount1)
			}
		})
	}
}

func TestReposition(t *testing.T) {
	tests := []struct {
		name               string
		mintAmount0        string
		mintAmount1        string
		increaseAmount0    string
		increaseAmount1    string
		liquidityToRemove  string
		tickLower          int32
		tickUpper          int32
		expectPanic        bool
		expectedPanicError string
	}{
		{
			name:               "Fail - Reposition after full liquidity removal",
			mintAmount0:        "1000000",
			mintAmount1:        "2000000",
			increaseAmount0:    "500000",
			increaseAmount1:    "1000000",
			liquidityToRemove:  "100",
			tickLower:          -5000,
			tickUpper:          5000,
			expectPanic:        true,
			expectedPanicError: "[GNOSWAP-POSITION-009] position is not clear || position(10) isn't clear(liquidity:3812288, tokensOwed0:0, tokensOwed1:0)",
		},
		{
			name:               "Fail - Reposition with partial liquidity removal",
			mintAmount0:        "1000000",
			mintAmount1:        "2000000",
			increaseAmount0:    "500000",
			increaseAmount1:    "1000000",
			liquidityToRemove:  "50",
			tickLower:          -3000,
			tickUpper:          3000,
			expectPanic:        true,
			expectedPanicError: "[GNOSWAP-POSITION-009] position is not clear || position(10) isn't clear(liquidity:5083034, tokensOwed0:0, tokensOwed1:0)",
		},
		{
			name:               "Fail - Reposition on non-existent tokenId",
			mintAmount0:        "1000000",
			mintAmount1:        "2000000",
			liquidityToRemove:  "10",
			tickLower:          -4000,
			tickUpper:          4000,
			expectPanic:        true,
			expectedPanicError: "[GNOSWAP-POSITION-009] position is not clear || position(10) isn't clear(liquidity:5083024, tokensOwed0:0, tokensOwed1:0)",
		},
	}

	CreatePoolWithoutFee(t)
	std.TestSetRealm(std.NewUserRealm(admin))
	TokenApprove(t, barPath, admin, pool, consts.UINT64_MAX)
	TokenApprove(t, fooPath, admin, pool, consts.UINT64_MAX)

	// Step 1: Mint a position
	tokenId, _, _, _ := Mint(
		barPath,
		fooPath,
		fee500,
		-10000,
		10000,
		"1000000",
		"1000000",
		"0",
		"0",
		time.Now().Add(time.Hour).Unix(),
		admin,
		admin,
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 2: Increase liquidity
			if tt.increaseAmount0 != "" && tt.increaseAmount1 != "" {
				_, _, _, _, _ = IncreaseLiquidity(
					tokenId,
					tt.increaseAmount0,
					tt.increaseAmount1,
					"0",
					"0",
					time.Now().Add(time.Hour).Unix(),
				)
			}

			// Step 3: Decrease liquidity to clear the position
			_, _, _, _, _, _, _ = DecreaseLiquidity(
				tokenId,
				tt.liquidityToRemove,
				"0",
				"0",
				time.Now().Add(time.Hour).Unix(),
				false,
			)

			// Step 4: Attempt Reposition
			if tt.expectPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicError, func() {
					Reposition(
						tokenId,
						tt.tickLower,
						tt.tickUpper,
						tt.mintAmount0,
						tt.mintAmount1,
						"0",
						"0",
					)
				})
			} else {
				//tokenId, liquidity.ToString(), tickLower, tickUpper, amount0.ToString(), amount1.ToString()
				tid, _, tickL, tickH, _, _ := Reposition(
					tokenId,
					tt.tickLower,
					tt.tickUpper,
					tt.mintAmount0,
					tt.mintAmount1,
					"0",
					"0",
				)
				uassert.Equal(t, tid, tokenId)
				uassert.Equal(t, tickL, tt.tickLower)
				uassert.Equal(t, tickH, tt.tickUpper)
			}
		})
	}
}
