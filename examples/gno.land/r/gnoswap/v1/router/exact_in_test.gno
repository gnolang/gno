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

	std.TestSkipHeights(100)
	user1Realm := std.NewUserRealm(user1Addr)
	std.TestSetRealm(user1Realm)

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
				bar.Approve(routerAddr, maxApprove)
				baz.Approve(routerAddr, maxApprove)
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
			std.TestSetRealm(user1Realm)
			bar.Approve(routerAddr, maxApprove)
			baz.Approve(routerAddr, maxApprove)
			bar.Approve(poolAddr, maxApprove)
			baz.Approve(poolAddr, maxApprove)
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
			std.TestSetRealm(std.NewUserRealm(positionOwner))
			pl.SetPoolCreationFeeByAdmin(0)
			if !pl.DoesPoolPathExist(pl.GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				pl.CreatePool(tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := pl.GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				pl.SetFeeProtocolByAdmin(tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}

			std.TestSetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(10000000000)}}
			std.TestIssueCoins(positionOwner, newCoins)
			std.TestSetOriginSend(newCoins, nil)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, newCoins)
			wugnot.Deposit()
			if tt.recipient != positionOwner {
				wugnot.Transfer(tt.recipient, 10000000000)
			}
			gns.Transfer(tt.recipient, 10000000000)

			std.TestSetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(poolContract, maxApprove)
			gns.Approve(poolContract, maxApprove)

			teller := common.GetTokenTeller(wugnotTokenPath)
			teller.Approve(poolContract, maxApprove)

			// Position Creation
			pn.Mint(
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
			std.TestSkipHeights(1)

			// Swap
			std.TestSetOriginCaller(tt.recipient)
			wugnot.Transfer(routerContract, 20000000)
			gns.Transfer(routerContract, 20000000)
			wugnot.Approve(routerContract, maxApprove)
			gns.Approve(routerContract, maxApprove)
			beforeGnotBalance := uint64(banker.GetCoins(tt.recipient).AmountOf("ugnot"))

			std.TestSetRealm(std.NewCodeRealm(routerPath))
			wugnot.Approve(poolContract, maxApprove)

			amountIn, amountOut := ExactInSwapRoute(
				tt.inputToken,
				tt.outputToken,
				tt.amountIn,
				tt.routeArr,
				tt.quoteArr,
				tt.amountOutMin,
				time.Now().Add(time.Hour).Unix(),
				"", // referrer
			)
			std.TestSkipHeights(1)
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
