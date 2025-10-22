package transfer

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"crypto/bech32"
	"crypto/sha256"

	"gno.land/p/aeddi/panictoerr"
	"gno.land/p/nt/ufmt"
)

// sendTransfer sends the specified amount of tokens from the sender to the
// escrow account.
func sendTransfer(banker banker.Banker, sourceClient string, token Token) error {
	// TODO handle when token is an IBC denom
	coin, err := token.ToCoin()
	if err != nil {
		return ufmt.Errorf("token to coin error: %v", err)
	}

	// Send the amount to the escrow address
	var (
		from   = runtime.CurrentRealm().Address()
		to     = GetEscrowAddress(sourceClient)
		amount = chain.NewCoins(coin)
	)
	if err := panictoerr.PanicAbortToError(func() {
		banker.SendCoins(from, to, amount)
	}); err != nil {
		return ufmt.Errorf("send coins: %v", err)
	}

	// TODO track escrowed coins by denom in store?

	return nil
}

// GetEscrowAddress returns the escrow address for the specified client.
// The escrow address follows the format as outlined in ADR 028:
// https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md
func GetEscrowAddress(clientID string) address {
	// a slash is used to create domain separation between port and client
	// identifiers to prevent address collisions between escrow addresses created
	// for different clients.
	contents := ufmt.Sprintf("%s/%s", PortID, clientID)

	// ADR 028 AddressHash construction
	preImage := []byte(escrowAddressVersion)
	preImage = append(preImage, 0)
	preImage = append(preImage, contents...)
	hash := sha256.Sum256(preImage)
	addr, err := bech32.EncodeFromBase256("g", hash[:20])
	if err != nil {
		panic(err)
	}
	return address(addr)
}

func Render(_ string) string {
	return "hello transfer app"
}
