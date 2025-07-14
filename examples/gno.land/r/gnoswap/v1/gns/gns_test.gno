package gns

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/halt"
)

const (
	// gnoVM test context default height
	// ref: https://github.com/gnolang/gno/blob/a85a53d5b38f0a21d66262a823a8b07f4f836b68/gnovm/pkg/test/test.go#L31-L32
	GNO_VM_DEFAULT_HEIGHT = int64(123)
)

var (
	emissionRealm = std.NewCodeRealm(consts.EMISSION_PATH)
	adminRealm    = std.NewUserRealm(adminAddr)
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")

	userRealm = std.NewUserRealm(testutils.TestAddress("user"))
)

func TestKnownAccounts(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(std.NewUserRealm(alice))
	func() {
		testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))
		uassert.Equal(t, 1, KnownAccounts())
	}()
}

func TestTotalSupply(t *testing.T) {
	uassert.Equal(t, INITIAL_MINT_AMOUNT, TotalSupply())
}

func TestBalanceOf(t *testing.T) {
	uassert.Equal(t, INITIAL_MINT_AMOUNT, BalanceOf(adminAddr))
}

func TestAssertTooManyEmission(t *testing.T) {
	tests := []struct {
		name        string
		amount      int64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "should panic if emission amount is too large",
			amount:      MAXIMUM_SUPPLY,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-GNS-002] too many emission reward || amount: 1000000000000000",
		},
		{
			name:   "should not panic if emission amount is not too large",
			amount: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					assertTooManyEmission(tt.amount)
				})
			} else {
				uassert.NotPanics(t, func() { assertTooManyEmission(tt.amount) })
			}
		})
	}
}

func TestIsLastBlockOfHalvingYear(t *testing.T) {
	tests := make([]struct {
		name   string
		height int64
		want   bool
	}, 0, 24)

	for i := HALVING_START_YEAR; i <= HALVING_END_YEAR; i++ {
		tests = append(tests, struct {
			name   string
			height int64
			want   bool
		}{
			name:   ufmt.Sprintf("last block of halving year %d", i),
			height: GetHalvingYearEndBlock(i),
			want:   true,
		})

		tests = append(tests, struct {
			name   string
			height int64
			want   bool
		}{
			name:   ufmt.Sprintf("not last block of halving year %d", i),
			height: GetHalvingYearEndBlock(i) - 1,
			want:   false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.want, isLastBlockOfHalvingYear(tt.height))
		})
	}
}

func TestHandleLeftEmissionAmount(t *testing.T) {
	tests := make([]struct {
		name   string
		year   int64
		amount int64
		want   int64
	}, 0, 24)

	for i := int64(1); i <= 12; i++ {
		tests = append(tests, struct {
			name   string
			year   int64
			amount int64
			want   int64
		}{
			name:   ufmt.Sprintf("handle left emission amount for year %d, non minted", i),
			year:   i,
			amount: 0,
			want:   GetHalvingYearMaxAmount(i),
		})

		tests = append(tests, struct {
			name   string
			year   int64
			amount int64
			want   int64
		}{
			name:   ufmt.Sprintf("handle left emission amount for year %d, minted", i),
			year:   i,
			amount: int64(123456),
			want:   GetHalvingYearMaxAmount(i) - int64(123456),
		})
	}
}

func TestSkipIfSameHeight(t *testing.T) {
	t.Run("should skip if height is same", func(t *testing.T) {
		uassert.True(t, skipIfSameHeight(1, 1))
	})

	t.Run("should not skip if height is different", func(t *testing.T) {
		uassert.False(t, skipIfSameHeight(1, 2))
	})
}

func TestGetterSetter(t *testing.T) {
	t.Run("last minted height", func(t *testing.T) {
		value := int64(123)
		setLastMintedHeight(value)
		uassert.Equal(t, value, GetLastMintedHeight())
	})

	t.Run("left emission amount", func(t *testing.T) {
		value := int64(0)
		setLeftEmissionAmount(value)
		uassert.Equal(t, value, GetLeftEmissionAmount())
	})
}

func TestGrc20Methods(t *testing.T) {
	tests := []struct {
		name          string
		fn            func()
		shouldPanic   bool
		exceptionKind string // abort / panic
		panicMsg      string
	}{
		{
			name: "TotalSupply",
			fn: func() {
				uassert.Equal(t, INITIAL_MINT_AMOUNT, TotalSupply())
			},
		},
		{
			name: "BalanceOf(admin)",
			fn: func() {
				uassert.Equal(t, INITIAL_MINT_AMOUNT, BalanceOf(adminAddr))
			},
		},
		{
			name: "BalanceOf(alice)",
			fn: func() {
				uassert.Equal(t, int64(0), BalanceOf(alice))
			},
		},
		{
			name: "Allowance(admin, alice)",
			fn: func() {
				uassert.Equal(t, int64(0), Allowance(adminAddr, alice))
			},
		},
		{
			name: "MintGns success",
			fn: func() {
				testing.SetRealm(emissionRealm)
				MintGns(cross, adminAddr)
			},
		},
		{
			name: "MintGns without permission should panic",
			fn: func() {
				testing.SkipHeights(1)
				MintGns(cross, adminAddr)
			},
			shouldPanic:   true,
			exceptionKind: "abort",
			panicMsg:      `caller(g1jgqwaa2le3yr63d533fj785qkjspumzv22ys5m) has no permission`,
		},
		{
			name: "Burn without permission should panic",
			fn: func() {
				Burn(cross, adminAddr, int64(1))
			},
			shouldPanic:   true,
			exceptionKind: "abort",
			panicMsg:      `ownable: caller is not owner`,
		},
		{
			name: "Transfer success",
			fn: func() {
				testing.SetOriginCaller(adminAddr)
				testing.SetRealm(adminRealm)
				Transfer(cross, alice, int64(1))
			},
		},
		{
			name: "Transfer without enough balance should panic",
			fn: func() {
				testing.SetOriginCaller(alice)
				testing.SetRealm(std.NewUserRealm(alice))
				Transfer(cross, bob, int64(1))
			},
			shouldPanic:   true,
			exceptionKind: "abort",
			panicMsg:      `insufficient balance`,
		},
		{
			name: "Transfer to self should panic",
			fn: func() {
				testing.SetOriginCaller(adminAddr)
				testing.SetRealm(adminRealm)
				Transfer(cross, adminAddr, int64(1))
			},
			shouldPanic:   true,
			exceptionKind: "abort",
			panicMsg:      `cannot send transfer to self`,
		},
		{
			name: "TransferFrom success",
			fn: func() {
				// approve first
				testing.SetRealm(adminRealm)
				Approve(cross, alice, int64(1))

				// alice transfer admin's balance to bob
				testing.SetOriginCaller(alice)
				testing.SetRealm(std.NewUserRealm(alice))
				TransferFrom(cross, adminAddr, bob, int64(1))
			},
		},
		{
			name: "TransferFrom without enough allowance should panic",
			fn: func() {
				testing.SetRealm(adminRealm)
				Approve(cross, alice, int64(1))

				testing.SetOriginCaller(alice)
				testing.SetRealm(std.NewUserRealm(alice))
				TransferFrom(cross, adminAddr, bob, int64(2))
			},
			shouldPanic:   true,
			exceptionKind: "abort",
			panicMsg:      `insufficient allowance`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGnsTokenObject(t)

			if tt.shouldPanic {
				switch tt.exceptionKind {
				case "abort":
					uassert.AbortsWithMessage(t, tt.panicMsg, tt.fn)
				case "panic":
					uassert.PanicsWithMessage(t, tt.panicMsg, tt.fn)
				default:
					t.Fatalf("unknown exception kind: %s", tt.exceptionKind)
				}
			} else {
				uassert.NotPanics(t, func() { tt.fn() })
			}
		})
	}
}

func TestCalculateAmountToMint(t *testing.T) {
	startHeight := std.ChainHeight()

	tests := []struct {
		name            string
		inputFromHeight int64
		inputToHeight   int64
		expectedAmount  int64
	}{
		{
			name:            "calculate amount to mint by 1 block for same year(1st year)",
			inputFromHeight: startHeight + 1,
			inputToHeight:   startHeight + 1,
			expectedAmount:  14269406,
		},
		{
			name:            "calculate amount to mint by 1 block for same year(2nd year)",
			inputFromHeight: GetHalvingYearStartBlock(2),
			inputToHeight:   GetHalvingYearStartBlock(2),
			expectedAmount:  14269406,
		},
		{
			name:            "calculate amount to mint by 1 block for same year(3rd year)",
			inputFromHeight: GetHalvingYearStartBlock(3),
			inputToHeight:   GetHalvingYearStartBlock(3),
			expectedAmount:  7134703, // halving year's amount / 2
		},
		{
			name:            "calculate amount to mint by 2 block for same year(1st year)",
			inputFromHeight: startHeight + 1,
			inputToHeight:   startHeight + 2,
			expectedAmount:  28538812, // 14269406 * 2
		},
		{
			name:            "calculate amount to mint by 1st year's entire block - 1 block",
			inputFromHeight: GetHalvingYearStartBlock(1),
			inputToHeight:   GetHalvingYearEndBlock(1) - 1,
			expectedAmount:  224_999_979_538_594, // 225000000000000 - 14269406 - 6192000
		},
		{
			name:            "calculate amount to mint by 1st year's entire block",
			inputFromHeight: GetHalvingYearStartBlock(1),
			inputToHeight:   GetHalvingYearEndBlock(1),
			expectedAmount:  225_000_000_000_000, // 225000000000000
		},
		{
			name:            "calculate amount to mint by 2nd year's entire block",
			inputFromHeight: GetHalvingYearStartBlock(2),
			inputToHeight:   GetHalvingYearEndBlock(2),
			expectedAmount:  225_000_000_000_000, // 225000000000000
		},
		{
			name:            "calculate amount to mint by 3rd year's entire block",
			inputFromHeight: GetHalvingYearStartBlock(3),
			inputToHeight:   GetHalvingYearEndBlock(3),
			expectedAmount:  112_500_000_000_000, // 112500000000000
		},
		{
			name:            "calculate amount to mint by 1st year's entire block and 2nd year's 1 block",
			inputFromHeight: GetHalvingYearStartBlock(1),
			inputToHeight:   GetHalvingYearStartBlock(2),
			expectedAmount:  225_000_014_269_406, // 225000000000000 + 14269406
		},
		{
			name:            "calculate amount to mint by entire block for 12 years",
			inputFromHeight: GetHalvingYearStartBlock(1),
			inputToHeight:   GetHalvingYearEndBlock(12),
			expectedAmount:  900_000_000_000_000, // 225000000000000
		},
		{
			name:            "calculate amount to mint is zero amount by after emission end height",
			inputFromHeight: GetHalvingYearEndBlock(12) + 1,
			inputToHeight:   GetHalvingYearEndBlock(12) + 10000,
			expectedAmount:  0,
		},
		{
			name:            "calculate amount to mint is zero amount by before emission start height",
			inputFromHeight: GetHalvingYearStartBlock(1) - 100,
			inputToHeight:   GetHalvingYearStartBlock(1) - 1,
			expectedAmount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGnsTest(t)

			amount := calculateAmountToMint(tt.inputFromHeight, tt.inputToHeight)

			uassert.Equal(t, tt.expectedAmount, amount)
		})
	}
}

func TestMintGns(t *testing.T) {
	tests := []struct {
		name                 string
		halted               bool
		inputRealm           std.Realm
		inputSkipHeight      int64
		expectedMintedAmount int64
		expectedHasAbort     bool
		expectedAbortMessage string
	}{
		{
			name:                 "mint gns is success by emission realm",
			halted:               false,
			inputRealm:           std.NewCodeRealm(consts.EMISSION_PATH),
			inputSkipHeight:      1,
			expectedMintedAmount: 14269406,
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
		{
			name:                 "mint gns is failed by halted",
			halted:               true,
			inputRealm:           std.NewCodeRealm(consts.EMISSION_PATH),
			inputSkipHeight:      1,
			expectedMintedAmount: 0,
			expectedHasAbort:     true,
			expectedAbortMessage: "halted",
		},
		{
			name:                 "mint gns is failed by caller is not emission realm",
			halted:               false,
			inputRealm:           std.NewUserRealm(testutils.TestAddress("user")),
			inputSkipHeight:      1,
			expectedMintedAmount: 0,
			expectedHasAbort:     true,
			expectedAbortMessage: "caller(g1w4ek2ujlta047h6lta047h6lta047h6lkvk347) has no permission",
		},
		{
			name:                 "mint gns is failed by not allowed to mint same block",
			halted:               false,
			inputRealm:           std.NewCodeRealm(consts.EMISSION_PATH),
			inputSkipHeight:      0,
			expectedMintedAmount: 0,
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGnsTest(t)

			testing.SetOriginCaller(adminAddr)
			halt.SetHaltByAdmin(cross, tt.halted) // set halt

			var mintedAmount int64

			mintGnsFn := func() {
				testing.SetRealm(tt.inputRealm)

				firstMintedAmount := MintGns(cross, adminAddr)
				testing.SkipHeights(tt.inputSkipHeight)
				secondMintedAmount := MintGns(cross, adminAddr)

				mintedAmount = firstMintedAmount + secondMintedAmount
			}

			if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMessage, mintGnsFn)
			} else {
				uassert.NotPanics(t, mintGnsFn)
			}

			uassert.Equal(t, tt.expectedMintedAmount, mintedAmount)
		})
	}
}
