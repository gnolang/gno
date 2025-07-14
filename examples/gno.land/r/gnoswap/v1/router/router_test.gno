package router

import (
	"std"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"
)

func TestFinalizeSwap(t *testing.T) {
	gnot := consts.GNOT

	newUint256 := func(val string) *u256.Uint {
		return u256.MustFromDecimal(val)
	}

	tests := []struct {
		name                    string
		inputToken              string
		outputToken             string
		resultAmountIn          *u256.Uint
		resultAmountOut         *u256.Uint
		swapType                SwapType
		tokenAmountLimit        *u256.Uint
		userBeforeWugnotBalance int64
		userWrappedWugnot       int64
		amountSpecified         *u256.Uint
		expectError             bool
		errorMessage            string
	}{
		{
			name:                    "Pass: ExactIn",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("100"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactIn,
			tokenAmountLimit:        newUint256("85"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             false,
		},
		{
			name:                    "Pass: ExactOut",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("110"),
			resultAmountOut:         newUint256("100"),
			swapType:                ExactOut,
			tokenAmountLimit:        newUint256("120"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             false,
		},
		{
			name:                    "ExactOut: Slippage error",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("100"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactOut,
			tokenAmountLimit:        newUint256("100"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             true,
			errorMessage:            "[GNOSWAP-ROUTER-002] slippage check failed || Received more than requested in [EXACT_OUT] requested=100, actual=90",
		},
		{
			name:                    "GNOT: Slippage error",
			inputToken:              gnot,
			outputToken:             barPath,
			resultAmountIn:          newUint256("300"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactIn,
			tokenAmountLimit:        newUint256("85"),
			userBeforeWugnotBalance: 1000000,
			userWrappedWugnot:       20000,
			expectError:             true,
			errorMessage:            "too much wugnot spent",
		},
		{
			name:                    "ExactOut: Real world slippage error",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("843455035"),
			resultAmountOut:         newUint256("600901351"),
			swapType:                ExactOut,
			tokenAmountLimit:        newUint256("843455036"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("600901352"),
			expectError:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)
			createBasicPool(t)

			if tt.expectError {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Error expected but not occurred")
						return
					}
					errorStr, ok := r.(string)
					if !ok {
						t.Errorf("Unexpected error type: %v", r)
						return
					}
					if tt.errorMessage != "" && !strings.Contains(errorStr, tt.errorMessage) {
						t.Errorf("Expected error message not included. got: %v, want: %v", errorStr, tt.errorMessage)
					}
				}()
			}

			amountIn, amountOut := finalizeSwap(
				tt.inputToken,
				tt.outputToken,
				tt.resultAmountIn,
				tt.resultAmountOut,
				tt.swapType,
				tt.tokenAmountLimit,
				tt.userBeforeWugnotBalance,
				tt.userWrappedWugnot,
				tt.amountSpecified,
			)

			if !tt.expectError {
				uassert.NotEqual(t, amountIn, "")
				uassert.NotEqual(t, amountOut, "")

				outVal := i256.MustFromDecimal(amountOut)
				if !outVal.IsNeg() {
					t.Error("amountOut is not negative")
				}
			}
		})
	}
}

func TestCompareExactInAndDrySwapWithNoLiquidityChanged(t *testing.T) {
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"
	const routerPath = consts.ROUTER_PATH
	const maxTimeout int64 = 9999999999
	const wugnotGnsPoolPath = "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:500"

	routerContract := routerAddr
	poolContract := poolAddr

	alice := testutils.TestAddress("alice")
	positionOwner := admin

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		feeProtocol0     uint8
		feeProtocol1     uint8
		inputToken       string
		outputToken      string
		amountIn         string
		routeArr         string
		quoteArr         string
		amountOutMin     string
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "success - gnot -> gns",
			tokenPath0:       wugnotTokenPath,
			tokenPath1:       gnsTokenPath,
			feeTier:          pl.FeeTier500,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100000000",
			amount1Requested: "100000000",
			feeProtocol0:     10,
			feeProtocol1:     10,
			inputToken:       wugnotTokenPath,
			outputToken:      gnsTokenPath,
			amountIn:         "100000000",
			routeArr:         wugnotGnsPoolPath,
			quoteArr:         "100",
			amountOutMin:     "85000000",
			expectedAmount0:  "100000000",
			expectedAmount1:  "-98817367",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)

			// Pool Creation
			testing.SetRealm(std.NewUserRealm(positionOwner))
			pl.SetPoolCreationFeeByAdmin(cross, 0)
			if !pl.DoesPoolPathExist(pl.GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				pl.CreatePool(cross, tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := pl.GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				pl.SetFeeProtocolByAdmin(cross, tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}

			testing.SetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(10000000000)}}
			testing.IssueCoins(positionOwner, newCoins)
			testing.SetOriginSend(newCoins)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, newCoins)
			wugnot.Deposit(cross)
			if tt.recipient != positionOwner {
				wugnot.Transfer(cross, tt.recipient, 10000000000)
			}
			gns.Transfer(cross, tt.recipient, 10000000000)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)

			func(cur realm) {
				teller := common.GetTokenTeller(wugnotTokenPath)
				teller.Approve(poolContract, maxApprove)
			}(cross)

			// Position Creation
			pn.Mint(
				cross,
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.tickLower,
				tt.tickUpper,
				tt.amount0Requested,
				tt.amount1Requested,
				"0",
				"0",
				maxTimeout,
				tt.recipient,
				tt.recipient,
				"",
			)
			testing.SkipHeights(1)

			// Swap
			testing.SetOriginCaller(tt.recipient)
			wugnot.Transfer(cross, routerContract, 20000000)
			gns.Transfer(cross, routerContract, 20000000)
			wugnot.Approve(cross, routerContract, maxApprove)
			gns.Approve(cross, routerContract, maxApprove)
			beforeWugnotBalance := wugnot.BalanceOf(tt.recipient)

			testing.SetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(cross, poolContract, maxApprove)

			_, drySwapAmountOut, _ := DrySwapRoute(
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				"EXACT_IN",
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
			)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			amountIn, amountOut := ExactInSwapRoute(
				cross,
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
			testing.SkipHeights(1)
			afterWugnotBalance := wugnot.BalanceOf(tt.recipient)

			func(cur realm) {
				absDrySwapAmountOut := strings.TrimPrefix(drySwapAmountOut, "-")
				withoutFee := handleSwapFee(tt.outputToken, u256.MustFromDecimal(absDrySwapAmountOut))
				absAmountOut := strings.TrimPrefix(amountOut, "-")
				uassert.Equal(t, withoutFee.ToString(), absAmountOut)
				uassert.Equal(t, tt.expectedAmount0, amountIn, "amountIn is not equal to expectedAmount0")
				uassert.Equal(t, tt.expectedAmount1, amountOut, "amountOut is not equal to expectedAmount1")
				uassert.Equal(t, tt.expectedAmount0, strconv.FormatInt(int64(beforeWugnotBalance-afterWugnotBalance), 10), "amountIn is not equal to expectedAmount0")
			}(cross)
		})
	}
}

func TestCompareExactInAndDrySwapWithWhenLiquidityAdded(t *testing.T) {
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"
	const routerPath = consts.ROUTER_PATH
	const maxTimeout int64 = 9999999999
	const wugnotGnsPoolPath = "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:500"

	routerContract := routerAddr
	poolContract := poolAddr

	alice := testutils.TestAddress("alice")
	positionOwner := admin

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		feeProtocol0     uint8
		feeProtocol1     uint8
		inputToken       string
		outputToken      string
		amountIn         string
		routeArr         string
		quoteArr         string
		amountOutMin     string
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "success - gnot -> gns",
			tokenPath0:       wugnotTokenPath,
			tokenPath1:       gnsTokenPath,
			feeTier:          pl.FeeTier500,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100000000",
			amount1Requested: "100000000",
			feeProtocol0:     10,
			feeProtocol1:     10,
			inputToken:       wugnotTokenPath,
			outputToken:      gnsTokenPath,
			amountIn:         "50000000",
			routeArr:         wugnotGnsPoolPath,
			quoteArr:         "100",
			amountOutMin:     "20000000",
			expectedAmount0:  "50000000",
			expectedAmount1:  "-49776285",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)

			// Pool Creation
			testing.SetRealm(std.NewUserRealm(positionOwner))
			pl.SetPoolCreationFeeByAdmin(cross, 0)
			if !pl.DoesPoolPathExist(pl.GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				pl.CreatePool(cross, tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := pl.GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				pl.SetFeeProtocolByAdmin(cross, tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}

			testing.SetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(20000000000)}}
			testing.IssueCoins(positionOwner, newCoins)
			testing.SetOriginSend(newCoins)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, std.Coins{{"ugnot", int64(10000000000)}})
			wugnot.Deposit(cross)
			if tt.recipient != positionOwner {
				wugnot.Transfer(cross, tt.recipient, 10000000000)
			}
			banker.SendCoins(positionOwner, tt.recipient, std.Coins{{"ugnot", int64(10000000000)}})
			gns.Transfer(cross, tt.recipient, 10000000000)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)

			func(cur realm) {
				teller := common.GetTokenTeller(wugnotTokenPath)
				teller.Approve(poolContract, maxApprove)
			}(cross)

			// Position Creation
			positionId, _, _, _ := pn.Mint(
				cross,
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.tickLower,
				tt.tickUpper,
				tt.amount0Requested,
				tt.amount1Requested,
				"0",
				"0",
				maxTimeout,
				tt.recipient,
				tt.recipient,
				"",
			)
			testing.SkipHeights(1)

			// Add Liquidity
			testing.SetOriginCaller(tt.recipient)
			addAmount, err := strconv.ParseInt(tt.amount0Requested, 10, 64)
			if err != nil {
				t.Errorf("failed to parse amount0Requested: %v", err)
			}
			newCoinsForAdd := std.Coins{{"ugnot", addAmount}}
			testing.SetOriginSend(newCoinsForAdd)
			banker = std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(tt.recipient, positionAddr, newCoinsForAdd)
			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)
			pn.IncreaseLiquidity(
				cross,
				positionId,
				tt.amount0Requested,
				tt.amount1Requested,
				"0",
				"0",
				maxTimeout,
			)

			// Swap
			testing.SetOriginCaller(tt.recipient)
			wugnot.Transfer(cross, routerContract, 20000000)
			gns.Transfer(cross, routerContract, 20000000)
			wugnot.Approve(cross, routerContract, maxApprove)
			gns.Approve(cross, routerContract, maxApprove)
			testing.SetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(cross, poolContract, maxApprove)

			DrySwapRoute(
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				"EXACT_IN",
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
			)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			_, amountOut := ExactInSwapRoute(
				cross,
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
			testing.SkipHeights(1)

			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						if errMsg, ok := r.(string); ok {
							uassert.Equal(t, "not authorized", errMsg)
						}
					} else {
						t.Errorf("expected panic but got none")
					}
				}
			}()

			uassert.Equal(t, tt.expectedAmount1, amountOut)
		})
	}
}

func TestCompareExactInAndDrySwapWithWhenLiquidityPartiallyRemoved(t *testing.T) {
	t.Skip("permission problem")
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"
	const routerPath = consts.ROUTER_PATH
	const maxTimeout int64 = 9999999999
	const wugnotGnsPoolPath = "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:10000"

	routerContract := routerAddr
	poolContract := poolAddr
	positionContract := positionAddr

	alice := testutils.TestAddress("alice")
	positionOwner := admin

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		feeProtocol0     uint8
		feeProtocol1     uint8
		inputToken       string
		outputToken      string
		amountIn         string
		routeArr         string
		quoteArr         string
		amountOutMin     string
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "success - gnot -> gns",
			tokenPath0:       wugnotTokenPath,
			tokenPath1:       gnsTokenPath,
			feeTier:          pl.FeeTier10000,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100000000",
			amount1Requested: "100000000",
			feeProtocol0:     10,
			feeProtocol1:     10,
			inputToken:       wugnotTokenPath,
			outputToken:      gnsTokenPath,
			amountIn:         "10000000",
			routeArr:         wugnotGnsPoolPath,
			quoteArr:         "100",
			amountOutMin:     "850000",
			expectedAmount0:  "10000000",
			expectedAmount1:  "-9872994",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)

			// Pool Creation
			testing.SetRealm(std.NewUserRealm(positionOwner))
			pl.SetPoolCreationFeeByAdmin(cross, 0)
			if !pl.DoesPoolPathExist(pl.GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				pl.CreatePool(cross, tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := pl.GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				pl.SetFeeProtocolByAdmin(cross, tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}

			testing.SetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(20000000000)}}
			testing.IssueCoins(positionOwner, newCoins)
			testing.SetOriginSend(newCoins)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, std.Coins{{"ugnot", int64(10000000000)}})
			wugnot.Deposit(cross)
			if tt.recipient != positionOwner {
				wugnot.Transfer(cross, tt.recipient, 10000000000)
			}
			banker.SendCoins(positionOwner, tt.recipient, std.Coins{{"ugnot", int64(10000000000)}})
			gns.Transfer(cross, tt.recipient, 10000000000)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)

			func(cur realm) {
				teller := common.GetTokenTeller(wugnotTokenPath)
				teller.Approve(poolContract, maxApprove)
			}(cross)

			// Position Creation
			positionId, liquidity, _, _ := pn.Mint(
				cross,
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.tickLower,
				tt.tickUpper,
				tt.amount0Requested,
				tt.amount1Requested,
				"0",
				"0",
				maxTimeout,
				tt.recipient,
				tt.recipient,
				"",
			)
			testing.SkipHeights(1)

			// Decrease Liquidity
			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)
			wugnot.Approve(cross, positionContract, maxApprove)
			gns.Approve(cross, positionContract, maxApprove)
			liquidityUint64, _ := strconv.ParseUint(liquidity, 10, 64)
			partialLiquidity := liquidityUint64 / 5
			reqLiquidity := strconv.FormatUint(partialLiquidity, 10)

			testing.SetOriginCaller(tt.recipient)
			pn.DecreaseLiquidity(
				cross,
				positionId,
				reqLiquidity,
				"0",
				"0",
				maxTimeout,
				true,
			)

			// Swap
			testing.SetOriginCaller(tt.recipient)
			wugnot.Transfer(cross, routerContract, 20000000)
			gns.Transfer(cross, routerContract, 20000000)
			wugnot.Approve(cross, routerContract, maxApprove)
			gns.Approve(cross, routerContract, maxApprove)
			beforeWugnotBalance := wugnot.BalanceOf(tt.recipient)

			testing.SetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(cross, poolContract, maxApprove)

			DrySwapRoute(
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				"EXACT_IN",
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
			)

			amountIn, amountOut := ExactInSwapRoute(
				cross,
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
			testing.SkipHeights(1)
			afterWugnotBalance := wugnot.BalanceOf(tt.recipient)

			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						if errMsg, ok := r.(string); ok {
							uassert.Equal(t, "not authorized", errMsg)
						}
					} else {
						t.Errorf("expected panic but got none")
					}
				}
			}()

			uassert.Equal(t, tt.expectedAmount0, amountIn)
			uassert.Equal(t, tt.expectedAmount1, amountOut)
			uassert.Equal(t, tt.expectedAmount0, strconv.FormatInt(int64(beforeWugnotBalance-afterWugnotBalance), 10))
		})
	}
}

func TestCompareExactInAndDrySwapWithWhenZeroForOneIsFalse(t *testing.T) {
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"
	const routerPath = consts.ROUTER_PATH
	const maxTimeout int64 = 9999999999
	const gnsWugnotPoolPath = "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:10000"

	routerContract := routerAddr
	poolContract := poolAddr

	alice := testutils.TestAddress("alice")
	positionOwner := admin

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		feeProtocol0     uint8
		feeProtocol1     uint8
		inputToken       string
		outputToken      string
		amountIn         string
		routeArr         string
		quoteArr         string
		amountOutMin     string
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "success - gns -> gnot",
			tokenPath0:       wugnotTokenPath,
			tokenPath1:       gnsTokenPath,
			feeTier:          pl.FeeTier10000,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100000000",
			amount1Requested: "100000000",
			feeProtocol0:     10,
			feeProtocol1:     10,
			inputToken:       gnsTokenPath,
			outputToken:      wugnotTokenPath,
			amountIn:         "10000000",
			routeArr:         gnsWugnotPoolPath,
			quoteArr:         "100",
			amountOutMin:     "850000",
			expectedAmount0:  "10000000",
			expectedAmount1:  "-9875422",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)

			// Pool Creation
			testing.SetRealm(std.NewUserRealm(positionOwner))
			pl.SetPoolCreationFeeByAdmin(cross, 0)
			if !pl.DoesPoolPathExist(pl.GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				pl.CreatePool(cross, tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := pl.GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				pl.SetFeeProtocolByAdmin(cross, tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}

			testing.SetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(20000000000)}}
			testing.IssueCoins(positionOwner, newCoins)
			testing.SetOriginSend(newCoins)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, std.Coins{{"ugnot", int64(10000000000)}})
			wugnot.Deposit(cross)
			if tt.recipient != positionOwner {
				wugnot.Transfer(cross, tt.recipient, 10000000000)
			}
			banker.SendCoins(positionOwner, tt.recipient, std.Coins{{"ugnot", int64(10000000000)}})
			gns.Transfer(cross, tt.recipient, 10000000000)

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(cross, poolContract, maxApprove)
			gns.Approve(cross, poolContract, maxApprove)

			func(cur realm) {
				teller := common.GetTokenTeller(wugnotTokenPath)
				teller.Approve(poolContract, maxApprove)
			}(cross)

			// Position Creation
			pn.Mint(
				cross,
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.tickLower,
				tt.tickUpper,
				tt.amount0Requested,
				tt.amount1Requested,
				"0",
				"0",
				maxTimeout,
				tt.recipient,
				tt.recipient,
				"",
			)
			testing.SkipHeights(1)

			// Swap
			testing.SetOriginCaller(tt.recipient)
			wugnot.Transfer(cross, routerContract, 20000000)
			gns.Transfer(cross, routerContract, 20000000)
			wugnot.Approve(cross, routerContract, maxApprove)
			gns.Approve(cross, routerContract, maxApprove)
			wugnot.BalanceOf(tt.recipient)

			testing.SetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(cross, poolContract, maxApprove)

			_, drySwapAmountOut, _ := DrySwapRoute(
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				"EXACT_IN",
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
			)
			handleSwapFee(tt.outputToken, u256.MustFromDecimal(drySwapAmountOut))

			testing.SetRealm(std.NewUserRealm(tt.recipient))
			amountIn, amountOut := ExactInSwapRoute(
				cross,
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
			testing.SkipHeights(1)

			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						if errMsg, ok := r.(string); ok {
							uassert.Equal(t, "not authorized", errMsg)
						}
					} else {
						t.Errorf("expected panic but got none")
					}
				}
			}()

			uassert.Equal(t, tt.expectedAmount0, amountIn)
			uassert.Equal(t, tt.expectedAmount1, amountOut)
		})
	}
}

func TestValidateQuoteSum(t *testing.T) {
	parser := NewRouteParser()

	tests := []struct {
		name      string
		quotes    []string
		wantError bool
		errMsg    string
	}{
		{
			name:      "Valid quotes summing to 100",
			quotes:    []string{"30", "40", "30"},
			wantError: false,
		},
		{
			name:      "Single quote of 100",
			quotes:    []string{"100"},
			wantError: false,
		},
		{
			name:      "Sum not equal to 100",
			quotes:    []string{"30", "40", "20"},
			wantError: true,
			errMsg:    "quote sum(90) must be 100",
		},
		{
			name:      "Negative quote",
			quotes:    []string{"-10", "110"},
			wantError: true,
			errMsg:    "quote(-10) at index(0) must be positive value",
		},
		{
			name:      "Zero quote",
			quotes:    []string{"0", "100"},
			wantError: true,
			errMsg:    "quote(0) at index(0) must be positive value",
		},
		{
			name:      "Quote exceeding 100",
			quotes:    []string{"101", "-1"},
			wantError: true,
			errMsg:    "quote(101) at index(0) must be less than or equal to 100",
		},
		{
			name:      "Invalid number format",
			quotes:    []string{"abc", "100"},
			wantError: true,
			errMsg:    "invalid quote(abc) at index(0)",
		},
		{
			name:      "Empty quotes",
			quotes:    []string{},
			wantError: true,
			errMsg:    "quote sum(0) must be 100",
		},
		{
			name:      "Overflow check",
			quotes:    []string{"50", "60"},
			wantError: true,
			errMsg:    "quote sum exceeds 100 at index(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateQuoteSum(tt.quotes)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
