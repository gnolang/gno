package router

import (
	"std"
	"testing"
	"time"

	"gno.land/p/gnoswap/consts"

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
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
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
			bar.Approve(consts.ROUTER_ADDR, maxApprove)
			baz.Approve(consts.ROUTER_ADDR, maxApprove)
			bar.Approve(consts.POOL_ADDR, maxApprove)
			baz.Approve(consts.POOL_ADDR, maxApprove)
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
