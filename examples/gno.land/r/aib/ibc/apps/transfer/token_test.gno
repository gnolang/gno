package transfer_test

import (
	"chain"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	"gno.land/r/aib/ibc/apps/transfer"
)

func TestTokenToCoin(t *testing.T) {
	const (
		tokenDenom  = "atom/pool"
		tokenAmount = "100"
	)
	testCases := []struct {
		name     string
		token    transfer.Token
		expCoin  chain.Coin
		expError string
	}{
		{
			"success: convert token to coin",
			transfer.Token{
				Denom: transfer.Denom{
					Base:  tokenDenom,
					Trace: []transfer.Hop{},
				},
				Amount: tokenAmount,
			},
			chain.NewCoin(tokenDenom, 100),
			"",
		},
		{
			"failure: invalid amount string",
			transfer.Token{
				Denom: transfer.Denom{
					Base:  tokenDenom,
					Trace: []transfer.Hop{},
				},
				Amount: "value",
			},
			chain.Coin{},
			"unable to parse transfer amount",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			coin, err := tc.token.ToCoin()

			if tc.expError == "" {
				urequire.NoError(t, err, tc.name)
				return
			}
			urequire.ErrorContains(t, err, tc.expError)
			uassert.Equal(t, tc.expCoin.Amount, coin.Amount)
			uassert.Equal(t, tc.expCoin.Denom, coin.Denom)
		})
	}
}
