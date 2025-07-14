package router

import (
	"std"
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/qux"
)

func TestExactInSwapRouteOperation_Validate(t *testing.T) {
	tests := []struct {
		name         string
		inputToken   string
		outputToken  string
		amountIn     string
		amountOutMin string
		routeArr     string
		quoteArr     string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Pass: single pool path",
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "100",
			amountOutMin: "90",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			wantErr:      false,
		},
		{
			name:         "Fail: amountIn is 0",
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "0",
			amountOutMin: "100",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			wantErr:      true,
			errMsg:       "invalid amountIn(0), must be positive",
		},
		{
			name:         "Fail: amountIn is negative",
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "-100",
			amountOutMin: "10",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			wantErr:      true,
			errMsg:       "invalid amountIn(-100), must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseParams := BaseSwapParams{
				InputToken:  tt.inputToken,
				OutputToken: tt.outputToken,
				RouteArr:    tt.routeArr,
				QuoteArr:    tt.quoteArr,
			}

			pp := NewExactInParams(
				baseParams,
				tt.amountIn,
				tt.amountOutMin,
			)

			op := NewExactInSwapOperation(pp)
			err := op.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q but got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExactInSwapRoute(t *testing.T) {
	routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)
	poolAddr, _ := access.GetAddress(access.ROLE_POOL)

	CreatePoolWithoutFee(t)
	MakeMintPositionWithoutFee(t)

	testing.SkipHeights(100)
	user1Realm := std.NewUserRealm(user1Addr)
	testing.SetRealm(user1Realm)

	tests := []struct {
		name         string
		setup        func()
		inputToken   string
		outputToken  string
		amountIn     string
		routeArr     string
		quoteArr     string
		amountOutMin string
		wantErr      bool
	}{
		{
			name: "BAR -> BAZ",
			setup: func() {
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "100",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			amountOutMin: "85",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testing.SetRealm(user1Realm)
			bar.Approve(cross, routerAddr, maxApprove)
			baz.Approve(cross, routerAddr, maxApprove)
			bar.Approve(cross, poolAddr, maxApprove)
			baz.Approve(cross, poolAddr, maxApprove)
			if tt.setup != nil {
				tt.setup()
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("ExactInSwapRoute() panic = %v", r)
					}
				}
			}()

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

			if !tt.wantErr {
				if amountIn == "" || amountOut == "" {
					t.Errorf("ExactInSwapRoute() returned empty values")
				}
			}
		})
	}
}

func TestExactInSwapRouteWithReferral(t *testing.T) {
	t.Skip("Currently we don't have a way to test emitted events")
	routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)
	poolAddr, _ := access.GetAddress(access.ROLE_POOL)

	CreatePoolWithoutFee(t)
	MakeMintPositionWithoutFee(t)

	testing.SkipHeights(100)
	user1Realm := std.NewUserRealm(user1Addr)
	testing.SetRealm(user1Realm)

	self := getPrevAddr().String()
	validReferral := testutils.TestAddress("validReferral")
	println("validReferral", validReferral)

	tests := []struct {
		name         string
		setup        func()
		inputToken   string
		outputToken  string
		amountIn     string
		routeArr     string
		quoteArr     string
		amountOutMin string
		referrer     string
		wantErr      bool
	}{
		{
			name: "self referral",
			setup: func() {
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "100",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			amountOutMin: "85",
			referrer:     self,
			wantErr:      false,
		},
		{
			name: "change referral",
			setup: func() {
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			inputToken:   barPath,
			outputToken:  bazPath,
			amountIn:     "100",
			routeArr:     singlePoolPath,
			quoteArr:     "100",
			amountOutMin: "85",
			referrer:     validReferral.String(),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testing.SetRealm(user1Realm)
			bar.Approve(cross, routerAddr, maxApprove)
			baz.Approve(cross, routerAddr, maxApprove)
			bar.Approve(cross, poolAddr, maxApprove)
			baz.Approve(cross, poolAddr, maxApprove)
			if tt.setup != nil {
				tt.setup()
			}

			ExactInSwapRoute(
				cross,
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				tt.referrer,
			)
		})
	}
}

func TestExactInZeroForOneFalse(t *testing.T) {
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"
	const positionContract = consts.POSITION_ADDR
	const routerContract = consts.ROUTER_ADDR
	const routerPath = consts.ROUTER_PATH
	const poolContract = consts.POOL_ADDR
	const maxTimeout int64 = 9999999999
	const wugnotGnsPoolPath = "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:500"

	alice := testutils.TestAddress("alice")
	positionOwner := consts.ADMIN

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
			feeTier:          pl.FeeTier500,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100",
			amount1Requested: "100",
			feeProtocol0:     10,
			feeProtocol1:     10,
			inputToken:       gnsTokenPath,
			outputToken:      "gnot",
			amountIn:         "100",
			routeArr:         wugnotGnsPoolPath,
			quoteArr:         "100",
			amountOutMin:     "85",
			expectedAmount0:  "100",
			expectedAmount1:  "-98",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				"10000000",
				"10000000",
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
			beforeGnotBalance := uint64(banker.GetCoins(tt.recipient).AmountOf("ugnot"))

			testing.SetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(cross, poolContract, maxApprove)

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
			afterGnotBalance := uint64(banker.GetCoins(tt.recipient).AmountOf("ugnot"))

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
			uassert.Equal(t, tt.expectedAmount1, strconv.FormatInt(int64(beforeGnotBalance-afterGnotBalance), 10))
		})
	}
}

/*
This test demonstrates a issue(Gnoswap-16) in SwapRoute function. The key steps are:

· Set up the test environment with necessary approvals
· Set the original send amount to 12345 ugnot
· Attempt to call SwapRoute with:

	· GNOT as input token
	· qux as output token
	· An amountSpecified of "3" (which doesn't match the sent amount)

· Check if the function panics as expected
*/
func TestSwapRouteWugnotquxExactInDifferentAmountCoinShouldPanic(t *testing.T) {
	testing.SetRealm(adminRealm)

	wugnot.Approve(cross, routerAddr, 1000000)
	qux.Approve(cross, routerAddr, 1000000)

	testing.SetOriginCaller(adminAddr)
	newCoins := std.Coins{{"ugnot", 12345}}
	testing.IssueCoins(adminAddr, newCoins)
	testing.SetOriginSend(newCoins)
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(adminAddr, routerAddr, newCoins)

	uassert.AbortsWithMessage(
		t,
		`[GNOSWAP-POOL-008] requested data not found || expected poolPath(gno.land/r/demo/wugnot:gno.land/r/onbloc/qux:3000) to exist`,
		func() {
			ExactOutSwapRoute(
				cross,
				consts.GNOT, // inputToken
				quxPath,     // outputToken
				"3",         // amountSpecified
				"gno.land/r/demo/wugnot:gno.land/r/onbloc/qux:3000", // strRouteArr
				"100",   // quoteArr
				"12345", // tokenAmountLimit
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
		},
	)
}
