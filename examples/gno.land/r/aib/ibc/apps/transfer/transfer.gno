package transfer

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"crypto/bech32"

	"gno.land/p/nt/ufmt"
)

// escrowCoins sends the specified amount of tokens from the sender to the
// realm address.
func escrowCoin(banker banker.Banker, token Token) error {
	// TODO handle when token is an IBC denom
	coin, err := token.ToCoin()
	if err != nil {
		return ufmt.Errorf("token to coin error: %v", err)
	}

	// Send the amount from sender to realm address
	var (
		from   = runtime.OriginCaller()
		to     = runtime.CurrentRealm().Address()
		amount = chain.NewCoins(coin)
	)
	banker.SendCoins(from, to, amount)

	// TODO track escrowed coins by denom/client in store?

	return nil
}

// escrowCoins sends the specified amount of tokens from the realm to the
// original sender.
func unescrowCoin(sender string, token Token) error {
	// TODO handle when token is an IBC denom
	coin, err := token.ToCoin()
	if err != nil {
		return ufmt.Errorf("token to coin error: %v", err)
	}

	// Send back the amount from escrow address back to sender
	var (
		from   = runtime.CurrentRealm().Address()
		amount = chain.NewCoins(coin)
		banker = banker.NewBanker(banker.BankerTypeRealmSend)
	)
	_, to, err := bech32.Decode(sender)
	if err != nil {
		return ufmt.Errorf("sender address %s conversion error: %v", sender, err)
	}
	banker.SendCoins(from, address(to), amount)

	// TODO track escrowed coins by denom/client in store?

	return nil
}

func Render(_ string) string {
	return "hello transfer app"
}
