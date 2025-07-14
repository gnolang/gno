package protocol_fee

import (
	"std"
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/access"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/qux"
)

var (
	adminAddr, _  = access.GetAddress(access.ROLE_ADMIN)
	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	adminRealm    = std.NewUserRealm(adminAddr)
	adminUser     = adminAddr
	dummyRealm    = std.NewCodeRealm("gno.land/r/dummy")
	aliceAddr     = testutils.TestAddress("alice")
)

func TestDistributeProtocolFee(t *testing.T) {
	// Reset protocol fee state for clean test
	protocolFeeState = NewProtocolFeeState()

	// admin > protocol_fee
	testing.SetRealm(adminRealm)
	testing.SetOriginCaller(adminAddr)
	bar.Transfer(cross, protocolFeeAddr, 1000)
	qux.Transfer(cross, protocolFeeAddr, 1000)

	stakerRealm := std.NewCodeRealm(consts.STAKER_PATH)
	testing.SetRealm(stakerRealm)
	testing.SetOriginCaller(stakerAddr)
	AddToProtocolFee(cross, "gno.land/r/onbloc/bar", 1000)
	AddToProtocolFee(cross, "gno.land/r/onbloc/qux", 1000)

	uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(1000))
	uassert.Equal(t, bar.BalanceOf(devOpsAddr), int64(0))
	uassert.Equal(t, bar.BalanceOf(govStakerAddr), int64(0))

	uassert.Equal(t, qux.BalanceOf(protocolFeeAddr), int64(1000))
	uassert.Equal(t, qux.BalanceOf(devOpsAddr), int64(0))
	uassert.Equal(t, qux.BalanceOf(govStakerAddr), int64(0))

	// gov/staker > protocol_fee
	govStakerRealm := std.NewCodeRealm(consts.GOV_STAKER_PATH)
	testing.SetRealm(govStakerRealm)
	testing.SetOriginCaller(govStakerAddr)
	DistributeProtocolFee(cross)

	uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(0))
	uassert.Equal(t, bar.BalanceOf(devOpsAddr), int64(0))
	uassert.Equal(t, bar.BalanceOf(govStakerAddr), int64(1000))

	uassert.Equal(t, qux.BalanceOf(protocolFeeAddr), int64(0))
	uassert.Equal(t, qux.BalanceOf(devOpsAddr), int64(0))
	uassert.Equal(t, qux.BalanceOf(govStakerAddr), int64(1000))
}

func TestSetDevOpsPctByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		realm       std.Realm
		percentage  uint64
		shouldPanic bool
		panicMsg    string
		expectedPct uint64
	}{
		{
			name:        "no permission to set devops percentage",
			realm:       std.NewUserRealm(testutils.TestAddress("dummy")),
			percentage:  123,
			shouldPanic: true,
			panicMsg:    "caller(g1v36k6mteta047h6lta047h6lta047h6lz7gmv8) has no permission",
		},
		{
			name:        "invalid percentage value",
			realm:       adminRealm,
			percentage:  100001,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-002] invalid percentage || pct(100001) should not be bigger than 10000",
		},
		{
			name:        "successfully set devops percentage",
			realm:       adminRealm,
			percentage:  123,
			shouldPanic: false,
			expectedPct: 123,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset protocol fee state for clean test
			protocolFeeState = NewProtocolFeeState()

			testing.SetRealm(test.realm)

			if test.shouldPanic {
				uassert.AbortsWithMessage(t, test.panicMsg, func() {
					SetDevOpsPctByAdmin(cross, test.percentage)
				})
			} else {
				uassert.Equal(t, GetDevOpsPct(), uint64(0))
				SetDevOpsPctByAdmin(cross, test.percentage)
				uassert.Equal(t, GetDevOpsPct(), test.expectedPct)
			}
		})
	}
}

func TestTokenList(t *testing.T) {
	// Reset and setup test data
	protocolFeeState = NewProtocolFeeState()
	protocolFeeState.TokenListWithAmount = map[string]uint64{
		"gno.land/r/foo": 100,
		"gno.land/r/bar": 0,
		"gno.land/r/baz": 200,
	}

	uassert.Equal(t, len(TokenList(true)), 2)
	uassert.Equal(t, len(TokenList(false)), 3)
}

func TestAddToProtocolFee(t *testing.T) {
	// Reset protocol fee state for clean test
	protocolFeeState = NewProtocolFeeState()

	tests := []struct {
		name      string
		tokenPath string
		amount    uint64
		want      uint64
	}{
		{
			name:      "add foo to protocol fee",
			tokenPath: "gno.land/r/foo",
			amount:    100,
			want:      100,
		},
		{
			name:      "add baz to protocol fee",
			tokenPath: "gno.land/r/baz",
			amount:    50,
			want:      50,
		},
		{
			name:      "add more baz to protocol fee",
			tokenPath: "gno.land/r/baz",
			amount:    10,
			want:      60,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testing.SetRealm(std.NewCodeRealm(consts.POOL_PATH))
			AddToProtocolFee(cross, test.tokenPath, test.amount)
			uassert.Equal(t, protocolFeeState.TokenListWithAmount[test.tokenPath], test.want)
		})
	}
}

func TestClearTokenListWithAmount(t *testing.T) {
	tests := []struct {
		name        string
		prevRealm   std.Realm
		want        map[string]uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "no permission to clear",
			prevRealm:   dummyRealm,
			shouldPanic: true,
			panicMsg:    "caller(g1lvx5ssxvuz5tttx6uza3myv8xy6w36a46fv7sy) has no permission",
		},
		{
			name:      "clear protocol fee",
			prevRealm: std.NewCodeRealm(consts.GOV_STAKER_PATH),
			want:      map[string]uint64{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset and setup test data for each test case
			protocolFeeState = NewProtocolFeeState()
			protocolFeeState.TokenListWithAmount = map[string]uint64{
				"gno.land/r/foo": 100,
				"gno.land/r/baz": 200,
			}

			testing.SetRealm(test.prevRealm)

			if test.shouldPanic {
				uassert.AbortsWithMessage(t, test.panicMsg, func() {
					ClearTokenListWithAmount(cross)
				})
			} else {
				ClearTokenListWithAmount(cross)
				uassert.Equal(t, len(protocolFeeState.TokenListWithAmount), len(test.want))
			}
		})
	}
}

func TestAssertOnlyPoolRouterStaker(t *testing.T) {
	tests := []struct {
		name        string
		prevRealm   std.Realm
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:      "caller is pool contract",
			prevRealm: std.NewCodeRealm(consts.POOL_PATH),
		},
		{
			name:      "caller is router contract",
			prevRealm: std.NewCodeRealm(consts.ROUTER_PATH),
		},
		{
			name:      "caller is staker contract",
			prevRealm: std.NewCodeRealm(consts.STAKER_PATH),
		},
		{
			name:        "caller is stranger",
			prevRealm:   dummyRealm,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-001] caller has no permission",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testing.SetRealm(test.prevRealm)

			// When a function is defined to accept realm as a parameter,
			// there's no need to wrap it with an anonymous function,
			// but since the assertOnlyPoolRouterStaker function doesn't accept a realm parameter,
			// it appears that a separate execution scope needs to be set up to distinguish realms,
			// and this anonymous function wrapping pattern must be maintained for the tests to pass.
			if test.shouldPanic {
				uassert.AbortsWithMessage(t, test.panicMsg, func() {
					func(cur realm) {
						testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))
						assertOnlyPoolRouterStaker()
					}(cross)
				})
			} else {
				uassert.NotPanics(t, func() {
					func(cur realm) {
						testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))
						assertOnlyPoolRouterStaker()
					}(cross)
				})
			}
		})
	}
}

func TestDistributeWithDifferentDevOpsPct(t *testing.T) {
	tests := []struct {
		name                    string
		devOpsPct               uint64
		tokenAmount             uint64
		expectedDevOpsAmount    uint64
		expectedGovStakerAmount uint64
	}{
		{
			name:                    "0% to DevOps",
			devOpsPct:               0,
			tokenAmount:             1000,
			expectedDevOpsAmount:    0,
			expectedGovStakerAmount: 1000,
		},
		{
			name:                    "25% to DevOps",
			devOpsPct:               2500,
			tokenAmount:             1000,
			expectedDevOpsAmount:    250,
			expectedGovStakerAmount: 750,
		},
		{
			name:                    "50% to DevOps",
			devOpsPct:               5000,
			tokenAmount:             1000,
			expectedDevOpsAmount:    500,
			expectedGovStakerAmount: 500,
		},
		{
			name:                    "100% to DevOps",
			devOpsPct:               10000,
			tokenAmount:             1000,
			expectedDevOpsAmount:    1000,
			expectedGovStakerAmount: 0,
		},
		{
			name:                    "33.33% to DevOps",
			devOpsPct:               3333,
			tokenAmount:             1000,
			expectedDevOpsAmount:    333,
			expectedGovStakerAmount: 667,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocolFeeState = NewProtocolFeeState()

			if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
				bar.Transfer(cross, adminAddr, balance)
			}
			if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
				testing.SetRealm(std.NewUserRealm(devOpsAddr))
				bar.Transfer(cross, adminAddr, balance)
				testing.SetRealm(adminRealm)
			}
			if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
				testing.SetRealm(std.NewUserRealm(govStakerAddr))
				bar.Transfer(cross, adminAddr, balance)
				testing.SetRealm(adminRealm)
			}

			testing.SetRealm(adminRealm)
			SetDevOpsPctByAdmin(cross, tt.devOpsPct)
			uassert.Equal(t, GetDevOpsPct(), tt.devOpsPct)

			testing.SetOriginCaller(adminAddr)
			bar.Transfer(cross, protocolFeeAddr, int64(tt.tokenAmount))

			// record the amount in protocol fee
			testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
			AddToProtocolFee(cross, "gno.land/r/onbloc/bar", tt.tokenAmount)

			// initial state
			uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(tt.tokenAmount))
			uassert.Equal(t, bar.BalanceOf(devOpsAddr), int64(0))
			uassert.Equal(t, bar.BalanceOf(govStakerAddr), int64(0))

			// distribute protocol fee
			testing.SetRealm(std.NewCodeRealm(consts.GOV_STAKER_PATH))
			testing.SetOriginCaller(govStakerAddr)
			result := DistributeProtocolFee(cross)

			// final state
			uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(0))
			uassert.Equal(t, bar.BalanceOf(devOpsAddr), int64(tt.expectedDevOpsAmount))
			uassert.Equal(t, bar.BalanceOf(govStakerAddr), int64(tt.expectedGovStakerAmount))

			if tt.expectedGovStakerAmount > 0 {
				uassert.Equal(t, result["gno.land/r/onbloc/bar"], tt.expectedGovStakerAmount)
			}

			uassert.Equal(t, len(protocolFeeState.TokenListWithAmount), 0)
		})
	}
}

func TestBalanceAmountMismatch_Fix(t *testing.T) {
	tests := []struct {
		name                      string
		setupFunc                 func()
		tokenPath                 string
		recordedAmount            uint64 // Amount recorded in TokenListWithAmount
		actualBalance             uint64 // Actual token balance in protocol_fee
		externalTransferAmount    uint64 // Amount transferred externally
		devOpsPct                 uint64 // DevOps percentage (basis points)
		expectedDevOpsReceived    uint64 // Expected amount DevOps should receive
		expectedGovStakerReceived uint64 // Expected amount GovStaker should receive
		expectPanic               bool
		panicMsg                  string
	}{
		{
			name: "exact match - balance equals recorded amount",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Admin transfers tokens to protocol_fee
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 1000)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            1000,
			actualBalance:             1000,
			externalTransferAmount:    0,
			devOpsPct:                 2000, // 20%
			expectedDevOpsReceived:    200,
			expectedGovStakerReceived: 800,
			expectPanic:               false,
		},
		{
			name: "external transfer - balance exceeds recorded amount",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// Reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Admin transfers initial amount
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 1000)

				// External user sends additional tokens
				bar.Transfer(cross, protocolFeeAddr, 500)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            1000,
			actualBalance:             1500,
			externalTransferAmount:    500,
			devOpsPct:                 2000, // 20%
			expectedDevOpsReceived:    200,  // 20% of 1000 (recorded amount only)
			expectedGovStakerReceived: 800,  // 80% of 1000 (recorded amount only)
			expectPanic:               false,
		},
		{
			name: "recorded exceeds balance - should panic",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// Reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Admin transfers tokens
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 500)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            1000, // More than actual balance
			actualBalance:             500,
			externalTransferAmount:    0,
			devOpsPct:                 2000,
			expectedDevOpsReceived:    0,
			expectedGovStakerReceived: 0,
			expectPanic:               true,
			panicMsg:                  "[GNOSWAP-PROTOCOL_FEE-003] invalid amount || amount: 1000 should be less than or equal to balance: 500",
		},
		{
			name: "multiple external transfers - all distributed",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// Reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Admin transfers initial amount
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 100)

				// Multiple external transfers
				bar.Transfer(cross, protocolFeeAddr, 200)
				bar.Transfer(cross, protocolFeeAddr, 300)
				bar.Transfer(cross, protocolFeeAddr, 400)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            100,  // Only the recorded amount
			actualBalance:             1000, // Total including external transfers
			externalTransferAmount:    900,
			devOpsPct:                 1000, // 10%
			expectedDevOpsReceived:    10,   // 10% of 100 (recorded amount only)
			expectedGovStakerReceived: 90,   // 90% of 100 (recorded amount only)
			expectPanic:               false,
		},
		{
			name: "zero recorded amount with external balance",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// Reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Only external transfer, no recorded amount
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 1000)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            0,    // Nothing recorded
			actualBalance:             1000, // But balance exists
			externalTransferAmount:    1000,
			devOpsPct:                 5000, // 50%
			expectedDevOpsReceived:    0,    // Nothing distributed as recorded is 0
			expectedGovStakerReceived: 0,    // Nothing distributed as recorded is 0
			expectPanic:               false,
		},
		{
			name: "100% to DevOps with external transfer",
			setupFunc: func() {
				protocolFeeState = NewProtocolFeeState()

				// Reset balances
				if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
					bar.Transfer(cross, adminAddr, balance)
				}
				if balance := bar.BalanceOf(devOpsAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(devOpsAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}
				if balance := bar.BalanceOf(govStakerAddr); balance > 0 {
					testing.SetRealm(std.NewUserRealm(govStakerAddr))
					bar.Transfer(cross, adminAddr, balance)
					testing.SetRealm(adminRealm)
				}

				// Admin transfers
				testing.SetRealm(adminRealm)
				testing.SetOriginCaller(adminAddr)
				bar.Transfer(cross, protocolFeeAddr, 800)

				// External transfer
				bar.Transfer(cross, protocolFeeAddr, 200)
			},
			tokenPath:                 "gno.land/r/onbloc/bar",
			recordedAmount:            800,
			actualBalance:             1000,
			externalTransferAmount:    200,
			devOpsPct:                 10000, // 100%
			expectedDevOpsReceived:    800,   // All recorded amount goes to DevOps
			expectedGovStakerReceived: 0,
			expectPanic:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			testing.SetRealm(adminRealm)
			SetDevOpsPctByAdmin(cross, tt.devOpsPct)

			// record the amount in protocol fee
			if tt.recordedAmount > 0 {
				testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
				AddToProtocolFee(cross, tt.tokenPath, tt.recordedAmount)
			}

			// check initial state
			uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(tt.actualBalance))
			uassert.Equal(t, protocolFeeState.TokenListWithAmount[tt.tokenPath], tt.recordedAmount)

			// distribute protocol fee
			testing.SetRealm(std.NewCodeRealm(consts.GOV_STAKER_PATH))
			testing.SetOriginCaller(govStakerAddr)

			if tt.expectPanic {
				uassert.AbortsWithMessage(t, tt.panicMsg, func() {
					DistributeProtocolFee(cross)
				})
			} else {
				// record initial balances before distribution
				prevDevOpsBalance := bar.BalanceOf(devOpsAddr)
				prevGovStakerBalance := bar.BalanceOf(govStakerAddr)

				result := DistributeProtocolFee(cross)

				// check the actual amounts received (delta from previous balance)
				actualDevOpsReceived := bar.BalanceOf(devOpsAddr) - prevDevOpsBalance
				actualGovStakerReceived := bar.BalanceOf(govStakerAddr) - prevGovStakerBalance

				// verify balances after distribution
				if tt.recordedAmount == 0 {
					// If recorded amount is 0, nothing should be distributed even if balance exists
					uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), int64(tt.actualBalance))
				} else {
					// Only recorded amount should be distributed, external transfers remain
					expectedRemaining := int64(tt.actualBalance) - int64(tt.recordedAmount)
					uassert.Equal(t, bar.BalanceOf(protocolFeeAddr), expectedRemaining)
				}

				uassert.Equal(t, actualDevOpsReceived, int64(tt.expectedDevOpsReceived))
				uassert.Equal(t, actualGovStakerReceived, int64(tt.expectedGovStakerReceived))

				// verify returned result
				if tt.expectedGovStakerReceived > 0 {
					uassert.Equal(t, result[tt.tokenPath], tt.expectedGovStakerReceived)
				}

				// verify TokenListWithAmount is cleared
				uassert.Equal(t, len(protocolFeeState.TokenListWithAmount), 0)
			}
		})
	}
}

func TestApiGetActualBalance(t *testing.T) {
	protocolFeeState = NewProtocolFeeState()

	testing.SetRealm(adminRealm)
	if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
		bar.Burn(cross, protocolFeeAddr, balance)
	}
	if balance := qux.BalanceOf(protocolFeeAddr); balance > 0 {
		qux.Burn(cross, protocolFeeAddr, balance)
	}

	// Transfer tokens to protocol fee
	bar.Transfer(cross, protocolFeeAddr, 1500)
	qux.Transfer(cross, protocolFeeAddr, 2500)

	// Record only part of the amounts
	testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
	AddToProtocolFee(cross, "gno.land/r/onbloc/bar", 1000)
	AddToProtocolFee(cross, "gno.land/r/onbloc/qux", 2000)

	result := ApiGetActualBalance()

	node := json.Must(json.Unmarshal([]byte(result)))

	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))

	// Verify actual balances (not recorded amounts)
	barBalance, err := node.MustKey("gno.land/r/onbloc/bar").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, barBalance, "1500")

	quxBalance, err := node.MustKey("gno.land/r/onbloc/qux").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, quxBalance, "2500")

	protocolFeeState = NewProtocolFeeState()
}

func TestApiGetRecordedBalance(t *testing.T) {
	protocolFeeState = NewProtocolFeeState()

	testing.SetRealm(adminRealm)
	if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
		bar.Burn(cross, protocolFeeAddr, balance)
	}
	if balance := qux.BalanceOf(protocolFeeAddr); balance > 0 {
		qux.Burn(cross, protocolFeeAddr, balance)
	}

	// Transfer tokens to protocol fee
	bar.Transfer(cross, protocolFeeAddr, 1500)
	qux.Transfer(cross, protocolFeeAddr, 2500)

	// Record only part of the amounts
	testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
	AddToProtocolFee(cross, "gno.land/r/onbloc/bar", 1000)
	AddToProtocolFee(cross, "gno.land/r/onbloc/qux", 2000)

	// Call API function
	result := ApiGetRecordedBalance()

	// Parse JSON response
	node := json.Must(json.Unmarshal([]byte(result)))

	// Verify metadata fields
	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))

	// Verify recorded balances (not actual balances)
	barBalance, err := node.MustKey("gno.land/r/onbloc/bar").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, barBalance, "1000")

	quxBalance, err := node.MustKey("gno.land/r/onbloc/qux").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, quxBalance, "2000")
}

func TestApiGetUnrecordedBalance(t *testing.T) {
	// Reset protocol fee state
	protocolFeeState = NewProtocolFeeState()

	testing.SetRealm(adminRealm)
	if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
		bar.Burn(cross, protocolFeeAddr, balance)
	}
	if balance := qux.BalanceOf(protocolFeeAddr); balance > 0 {
		qux.Burn(cross, protocolFeeAddr, balance)
	}

	// Test case 1: Unrecorded amounts exist
	bar.Transfer(cross, protocolFeeAddr, 1500)
	qux.Transfer(cross, protocolFeeAddr, 2500)

	// Record only part of the amounts
	testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
	AddToProtocolFee(cross, "gno.land/r/onbloc/bar", 1000)
	AddToProtocolFee(cross, "gno.land/r/onbloc/qux", 2000)

	result := ApiGetUnrecordedBalance()

	node := json.Must(json.Unmarshal([]byte(result)))

	// Verify metadata fields
	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))

	// Verify unrecorded balances (actual - recorded)
	barBalance, err := node.MustKey("gno.land/r/onbloc/bar").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, barBalance, "500") // 1500 - 1000

	quxBalance, err := node.MustKey("gno.land/r/onbloc/qux").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, quxBalance, "500") // 2500 - 2000
}

func TestApiGetUnrecordedBalance_NoUnrecorded(t *testing.T) {
	protocolFeeState = NewProtocolFeeState()

	testing.SetRealm(adminRealm)
	testing.SetOriginCaller(adminAddr)
	if balance := bar.BalanceOf(protocolFeeAddr); balance > 0 {
		bar.Burn(cross, protocolFeeAddr, balance)
	}

	// Transfer exact amount that will be recorded
	bar.Transfer(cross, protocolFeeAddr, 1000)

	// Record the exact amount
	testing.SetRealm(std.NewCodeRealm(consts.STAKER_PATH))
	AddToProtocolFee(cross, "gno.land/r/onbloc/bar", 1000)

	// Call API function
	result := ApiGetUnrecordedBalance()

	// Parse JSON response
	node := json.Must(json.Unmarshal([]byte(result)))

	// Verify metadata fields
	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))

	// Should not have bar token as there's no unrecorded amount
	uassert.False(t, node.HasKey("gno.land/r/onbloc/bar"))
}
