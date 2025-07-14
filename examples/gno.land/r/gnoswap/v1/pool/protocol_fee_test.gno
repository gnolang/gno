package pool

import (
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/foo"
)

func TestSetPoolCreationFee(t *testing.T) {
	tests := []struct {
		name        string
		action      func()
		verify      func() string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call by non-governance",
			action: func() {
				const newFee = 2_000_000_000
				testing.SetRealm(adminRealm)
				SetPoolCreationFee(cross, newFee)
			},
			verify: nil,
			// expected:    "caller(g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8c) has no permission",
			expected:    ufmt.Sprintf("caller(%s) has no permission", adminAddr.String()),
			shouldPanic: true,
		},
		{
			name: "Success call by governance",
			action: func() {
				const newFee = 2_000_000_000
				govRealm := std.NewUserRealm(govAddr)
				testing.SetRealm(govRealm)
				SetPoolCreationFee(cross, newFee)
			},
			verify: func() string {
				return strconv.FormatInt(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatInt(2_000_000_000, 10),
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action()
				if tc.verify != nil {
					got := tc.verify()
					uassert.Equal(t, got, tc.expected)
				}
			} else {
				uassert.AbortsWithMessage(t, tc.expected, func() {
					tc.action()
				})
			}
		})
	}
}

func TestSetPoolCreationFeeByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		action      func()
		verify      func() string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call by non-admin (gov contract)",
			action: func() {
				const newFee = 2_000_000_000
				govRealm := std.NewUserRealm(govAddr)
				testing.SetRealm(govRealm)
				SetPoolCreationFeeByAdmin(cross, newFee)
			},
			verify: nil,
			// expected:    "caller(g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd) has no permission",
			expected:    ufmt.Sprintf("caller(%s) has no permission", govAddr.String()),
			shouldPanic: true,
		},
		{
			name: "Panic call by non-admin (user)",
			action: func() {
				const newFee = 2_000_000_000
				testing.SetRealm(std.NewUserRealm(alice))
				SetPoolCreationFeeByAdmin(cross, newFee)
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call by admin",
			action: func() {
				const newFee = 2_000_000_000
				testing.SetRealm(adminRealm)
				SetPoolCreationFeeByAdmin(cross, newFee)
			},
			verify: func() string {
				return strconv.FormatInt(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatInt(2_000_000_000, 10),
			shouldPanic: false,
		},
		{
			name: "Success call by admin (rollback)",
			action: func() {
				const newFee = 1_000_000_000
				testing.SetRealm(adminRealm)
				SetPoolCreationFeeByAdmin(cross, newFee)
			},
			verify: func() string {
				return strconv.FormatInt(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatInt(1_000_000_000, 10),
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action()
				if tc.verify != nil {
					got := tc.verify()
					uassert.Equal(t, got, tc.expected)
				}
			} else {
				uassert.AbortsWithMessage(t, tc.expected, func() {
					tc.action()
				})
			}
		})
	}
}

func TestHandleWithdrawalFee(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) (string, string)
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic if caller is not position contract",
			action: func(t *testing.T) {
				testing.SetOriginCaller(adminAddr)
				CollectWithdrawalFee(cross, 0, "gno.land/r/onbloc/foo", "0", "gno.land/r/onbloc/foo", "0", "", adminAddr)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-001] caller has no permission || only position(g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5) can call this function, called from g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			shouldPanic: true,
		},
		{
			name: "Panic if pkgPath is not registered",
			action: func(t *testing.T) {
				testing.SetRealm(std.NewUserRealm(positionAddr))
				CollectWithdrawalFee(cross, 0, "pkgPath", "1000", "pkgPath", "1000", "poolPath", adminAddr)
			},
			verify:      nil,
			expected:    "[GNOSWAP-COMMON-004] token is not registered || token(pkgPath)",
			shouldPanic: true,
		},
		{
			name: "Success call by position contract",
			action: func(t *testing.T) {
				InitialisePoolTest(t)
				testing.SetRealm(std.NewUserRealm(alice))
				TokenApprove(t, wugnotPath, alice, poolAddr, 1000)
				TokenApprove(t, gnsPath, alice, poolAddr, 1000)

				testing.SetRealm(posRealm)
				Mint(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					positionAddr,
					int32(1020),
					int32(5040),
					"1000",
					alice,
				)
			},
			verify: func(t *testing.T) (string, string) {
				testing.SetRealm(std.NewUserRealm(alice))
				TokenApprove(t, wugnotPath, alice, poolAddr, 10000000000)
				TokenApprove(t, gnsPath, alice, poolAddr, 1000000000)
				testing.SetRealm(std.NewUserRealm(positionAddr))
				poolPath := GetPoolPath(wugnotPath, gnsPath, fee3000)
				return CollectWithdrawalFee(cross, 2, wugnotPath, "1000", gnsPath, "1000", poolPath, alice)
			},
			expected:    "990,990",
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				uassert.AbortsWithMessage(t, tc.expected, func() {
					tc.action(t)
				})
			} else {
				tc.action(t)
				if tc.verify != nil {
					gotAfterAmount0, gotAfterAmount1 := tc.verify(t)
					expected := strings.Split(tc.expected, ",")
					uassert.Equal(t, gotAfterAmount0, expected[0])
					uassert.Equal(t, gotAfterAmount1, expected[1])
				}
			}
		})
	}
}

func TestSetWithdrawalFee(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call to set withdrawal fee setFee by non-admin (user)",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				userRealm := std.NewUserRealm(alice)
				testing.SetRealm(userRealm)
				SetWithdrawalFee(cross, newFee)
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Panic call to set withdrawal fee by admin",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				adminRealm := std.NewUserRealm(adminAddr)
				testing.SetRealm(adminRealm)
				SetWithdrawalFee(cross, newFee)
			},
			verify:      nil,
			expected:    "caller(g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call to set withdrawal fee by governance",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				govRealm := std.NewUserRealm(govAddr)
				testing.SetRealm(govRealm)
				SetWithdrawalFee(cross, newFee)
			},
			verify: func(t *testing.T) string {
				return strconv.FormatUint(GetWithdrawalFee(), 10)
			},
			expected:    strconv.FormatUint(200, 10),
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action(t)
				if tc.verify != nil {
					gotWithdrawalFee := tc.verify(t)
					uassert.Equal(t, gotWithdrawalFee, tc.expected)
				}
			} else {
				uassert.AbortsWithMessage(t, tc.expected, func() {
					tc.action(t)
				})
			}
		})
	}
}

func TestSetWithdrawalFeeByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call to set withdrawal fee by non-admin (user)",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				userRealm := std.NewUserRealm(alice)
				testing.SetRealm(userRealm)
				SetWithdrawalFeeByAdmin(cross, newFee)
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Panic call to set withdrawal fee by non-admin (gov contract)",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				govRealm := std.NewUserRealm(govAddr)
				testing.SetRealm(govRealm)
				SetWithdrawalFeeByAdmin(cross, newFee)
			},
			verify:      nil,
			expected:    "caller(g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call to set withdrawal fee by admin",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				adminRealm := std.NewUserRealm(adminAddr)
				testing.SetRealm(adminRealm)
				SetWithdrawalFeeByAdmin(cross, newFee)
			},
			verify: func(t *testing.T) string {
				return strconv.FormatUint(GetWithdrawalFee(), 10)
			},
			expected:    strconv.FormatUint(100, 10),
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action(t)
				if tc.verify != nil {
					gotWithdrawalFee := tc.verify(t)
					uassert.Equal(t, gotWithdrawalFee, tc.expected)
				}
			} else {
				uassert.AbortsWithMessage(t, tc.expected, func() {
					tc.action(t)
				})
			}
		})
	}
}

func TestCalculateAmountWithFee(t *testing.T) {
	tests := []struct {
		name         string
		amount       *u256.Uint
		fee          *u256.Uint
		expectedFee  *u256.Uint
		expectedRest *u256.Uint
	}{
		{
			name:         "normal case",
			amount:       u256.NewUint(1000),
			fee:          u256.NewUint(100), // 1%
			expectedFee:  u256.NewUint(10),
			expectedRest: u256.NewUint(990),
		},
		{
			name:         "zero amount",
			amount:       u256.NewUint(0),
			fee:          u256.NewUint(100),
			expectedFee:  u256.NewUint(0),
			expectedRest: u256.NewUint(0),
		},
		{
			name:         "zero fee",
			amount:       u256.NewUint(1000),
			fee:          u256.NewUint(0),
			expectedFee:  u256.NewUint(0),
			expectedRest: u256.NewUint(1000),
		},
		{
			name:         "max fee",
			amount:       u256.NewUint(1000),
			fee:          u256.NewUint(10000), // 100%
			expectedFee:  u256.NewUint(1000),
			expectedRest: u256.NewUint(0),
		},
		{
			name:         "large amount",
			amount:       u256.MustFromDecimal("1000000000000000000"),
			fee:          u256.NewUint(100),
			expectedFee:  u256.MustFromDecimal("10000000000000000"),
			expectedRest: u256.MustFromDecimal("990000000000000000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeAmount, restAmount := calculateAmountWithFee(tt.amount, tt.fee)
			uassert.Equal(t, feeAmount.ToString(), tt.expectedFee.ToString())
			uassert.Equal(t, restAmount.ToString(), tt.expectedRest.ToString())
		})
	}
}

func TestProtocolFee(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) interface{}
		expected    interface{}
		shouldPanic bool
	}{
		{
			name: "initial fee is 0",
			action: func(t *testing.T) {
				// No action needed, just verify initial state
			},
			verify: func(t *testing.T) interface{} {
				return slot0FeeProtocol
			},
			expected:    uint8(0),
			shouldPanic: false,
		},
		{
			name: "create first pool with zero creation fee",
			action: func(t *testing.T) {
				testing.SetRealm(adminRealm)
				SetPoolCreationFeeByAdmin(cross, 0)
				CreatePool(cross, barPath, fooPath, 3000, common.TickMathGetSqrtRatioAtTick(0).ToString())
			},
			verify: func(t *testing.T) interface{} {
				return slot0FeeProtocol
			},
			expected:    uint8(0),
			shouldPanic: false,
		},
		{
			name: "get protocol fee from created pool",
			action: func(t *testing.T) {
				// Pool already created in previous test
			},
			verify: func(t *testing.T) interface{} {
				pool := GetPool(barPath, fooPath, 3000)
				return pool.slot0.feeProtocol
			},
			expected:    uint8(0),
			shouldPanic: false,
		},
		{
			name: "panic when setting fee protocol with invalid values (3, 3)",
			action: func(t *testing.T) {
				setFeeProtocol(3, 3)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(3) == 0 || (feeProtocol0(3) >= 4 && feeProtocol0(3) <= 10)) && (feeProtocol1(3) == 0 || (feeProtocol1(3) >= 4 && feeProtocol1(3) <= 10))",
			shouldPanic: true,
		},
		{
			name: "panic when setting fee protocol with invalid values (6, 3)",
			action: func(t *testing.T) {
				setFeeProtocol(6, 3)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(6) == 0 || (feeProtocol0(6) >= 4 && feeProtocol0(6) <= 10)) && (feeProtocol1(3) == 0 || (feeProtocol1(3) >= 4 && feeProtocol1(3) <= 10))",
			shouldPanic: true,
		},
		{
			name: "panic when setting fee protocol with invalid values (3, 6)",
			action: func(t *testing.T) {
				setFeeProtocol(3, 6)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(3) == 0 || (feeProtocol0(3) >= 4 && feeProtocol0(3) <= 10)) && (feeProtocol1(6) == 0 || (feeProtocol1(6) >= 4 && feeProtocol1(6) <= 10))",
			shouldPanic: true,
		},
		{
			name: "panic when setting fee protocol with invalid values (11, 11)",
			action: func(t *testing.T) {
				setFeeProtocol(11, 11)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(11) == 0 || (feeProtocol0(11) >= 4 && feeProtocol0(11) <= 10)) && (feeProtocol1(11) == 0 || (feeProtocol1(11) >= 4 && feeProtocol1(11) <= 10))",
			shouldPanic: true,
		},
		{
			name: "panic when setting fee protocol with invalid values (6, 11)",
			action: func(t *testing.T) {
				setFeeProtocol(6, 11)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(6) == 0 || (feeProtocol0(6) >= 4 && feeProtocol0(6) <= 10)) && (feeProtocol1(11) == 0 || (feeProtocol1(11) >= 4 && feeProtocol1(11) <= 10))",
			shouldPanic: true,
		},
		{
			name: "panic when setting fee protocol with invalid values (11, 6)",
			action: func(t *testing.T) {
				setFeeProtocol(11, 6)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-016] invalid protocol fee percentage || expected (feeProtocol0(11) == 0 || (feeProtocol0(11) >= 4 && feeProtocol0(11) <= 10)) && (feeProtocol1(6) == 0 || (feeProtocol1(6) >= 4 && feeProtocol1(6) <= 10))",
			shouldPanic: true,
		},
		{
			name: "set fee protocol to valid values (10, 10)",
			action: func(t *testing.T) {
				setFeeProtocol(10, 10)
			},
			verify: func(t *testing.T) interface{} {
				return slot0FeeProtocol
			},
			expected:    uint8(170), // 10 + (10 << 4) = 170
			shouldPanic: false,
		},
		{
			name: "check if previous pool's protocol fee is updated",
			action: func(t *testing.T) {
				// Fee already set in previous test
			},
			verify: func(t *testing.T) interface{} {
				pool := GetPool(barPath, fooPath, 3000)
				return pool.slot0.feeProtocol
			},
			expected:    uint8(170),
			shouldPanic: false,
		},
		{
			name: "create second pool with updated protocol fee",
			action: func(t *testing.T) {
				CreatePool(cross, barPath, fooPath, 500, common.TickMathGetSqrtRatioAtTick(0).ToString())
			},
			verify: func(t *testing.T) interface{} {
				pool := GetPool(barPath, fooPath, 500)
				return pool.slot0.feeProtocol
			},
			expected:    uint8(170),
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action(t)
				if tc.verify != nil {
					got := tc.verify(t)
					uassert.Equal(t, got, tc.expected)
				}
			} else {
				uassert.PanicsWithMessage(t, tc.expected.(string), func() {
					tc.action(t)
				})
			}
		})
	}
}

// TestProtocolFeesDisabledWhenNotConfigured verifies that protocol fees are not collected
// when protocol fee configuration is not enabled. This test ensures that:
//  1. A pool can be created and liquidity can be minted without protocol fees
//  2. Swaps can be performed in both directions (token0->token1 and token1->token0)
//  3. No protocol fees are accumulated when protocol fees are not configured
//  4. Position liquidity remains protected from unexpected fee deductions
func TestProtocolFeesDisabledWhenNotConfigured(t *testing.T) {
	InitPoolTest(t)

	// Create pool with initial price at tick -10000
	testing.SetRealm(adminRealm)
	gns.Approve(cross, poolAddr, poolCreationFee)
	CreatePool(cross, barPath, fooPath, fee500, common.TickMathGetSqrtRatioAtTick(-10000).ToString())

	// Approve tokens for minting liquidity
	foo.Approve(cross, poolAddr, maxApprove)
	bar.Approve(cross, poolAddr, maxApprove)

	// Mint liquidity position across full tick range
	testing.SetRealm(posRealm)
	Mint(
		cross,
		barPath,
		fooPath,
		fee500,
		positionAddr,
		-887160,
		887160,
		"100000000",
		adminAddr,
	)

	// Perform swap from token0 to token1 (bar -> foo)
	testing.SetRealm(rouRealm)
	Swap(
		cross,
		barPath,
		fooPath,
		fee500,
		adminAddr,
		true,
		"10000",
		consts.MIN_PRICE,
		adminAddr,
	)

	// Perform swap from token1 to token0 (foo -> bar)
	Swap(
		cross,
		barPath,
		fooPath,
		fee500,
		adminAddr,
		false,
		"10000",
		consts.MAX_PRICE,
		adminAddr,
	)

	// Verify that no protocol fees were collected since protocol fees are not enabled
	thisPool := GetPool(barPath, fooPath, fee500)
	uassert.Equal(t, thisPool.protocolFees.token0.ToString(), "0")
	uassert.Equal(t, thisPool.protocolFees.token1.ToString(), "0")

	resetObject(t)
}
