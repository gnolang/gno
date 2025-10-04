package transfer

import (
	"chain"
	"strconv"

	"gno.land/p/nt/ufmt"
)

// Token defines a struct which represents a token to be transferred.
type Token struct {
	// the token denomination
	Denom Denom
	// the token amount to be transferred
	Amount string
}

// ToCoin converts a Token to an chain.Coin.
//
// The function parses the Token.Amount field and returns a new sdk.Coin with
// the IBCDenom of the Token's Denom field and the parsed Amount.
// If the Amount cannot be parsed, an error is returned with a wrapped error
// message.
func (t Token) ToCoin() (chain.Coin, error) {
	transferAmount, err := strconv.ParseInt(t.Amount, 10, 64)
	if err != nil {
		return chain.Coin{}, ufmt.Errorf("unable to parse transfer amount (%s) into int64", t.Amount)
	}

	coin := chain.NewCoin(t.Denom.IBCDenom(), transferAmount)
	return coin, nil
}
