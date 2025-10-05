package transfer

import (
	"chain"
	"chain/banker"
	"crypto/sha256"

	"gno.land/p/nt/ufmt"
)

// sendTransfer sends the specified amount of tokens from the sender to the
// escrow account. It ensures that the packet amount matches the origin send
// (using origin send tx parameter is the only way for a contract to transfer
// funds from signer).
func sendTransfer(sourcePort, sourceClient string, token Token, senderAddr address) error {
	// TODO handle when token is an IBC denom
	coin, err := token.ToCoin()
	if err != nil {
		return ufmt.Errorf("token to coin error: %v", err)
	}
	amt := chain.NewCoins(coin)
	sent := banker.OriginSend()
	if sent.String() != amt.String() {
		return ufmt.Errorf("packet amount and origin send do not match, %s != %s", amt, sent)
	}
	escrowAddr := getEscrowAddress(sourcePort, sourceClient)

	banker.NewBanker(banker.BankerTypeOriginSend).
		SendCoins(senderAddr, escrowAddr, sent)

	// TODO track escrowed coins by denom in store?

	return nil
}

// getEscrowAddress returns the escrow address for the specified client.
// The escrow address follows the format as outlined in ADR 028:
// https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md
func getEscrowAddress(portID, clientID string) address {
	// a slash is used to create domain separation between port and client
	// identifiers to prevent address collisions between escrow addresses created
	// for different clients.
	contents := ufmt.Sprintf("%s/%s", portID, clientID)

	// ADR 028 AddressHash construction
	preImage := []byte(escrowAddressVersion)
	preImage = append(preImage, 0)
	preImage = append(preImage, contents...)
	hash := sha256.Sum256(preImage)
	return address(hash[:20])
}
