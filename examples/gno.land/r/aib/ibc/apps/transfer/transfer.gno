package transfer

import (
	"chain"
	"chain/banker"
	"crypto/sha256"

	"gno.land/p/nt/ufmt"
)

func SendTransfer(sourcePort, sourceClient string, token Token, senderAddr address) error {
	// TODO check transfer keeper.SendTransfer for the usecases
	coin, err := token.ToCoin()
	if err != nil {
		return ufmt.Errorf("token to coin error: %v", err)
	}
	var (
		banker     = banker.NewBanker(banker.BankerTypeRealmSend)
		escrowAddr = getEscrowAddress(sourcePort, sourceClient)
		amt        = chain.NewCoins(coin)
	)
	banker.SendCoins(senderAddr, escrowAddr, amt)
	return nil
}

// getEscrowAddress returns the escrow address for the specified channel.
// The escrow address follows the format as outlined in ADR 028:
// https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md
func getEscrowAddress(portID, channelID string) address {
	// a slash is used to create domain separation between port and channel identifiers to
	// prevent address collisions between escrow addresses created for different channels
	contents := ufmt.Sprintf("%s/%s", portID, channelID)

	// ADR 028 AddressHash construction
	preImage := []byte(escrowAddressVersion)
	preImage = append(preImage, 0)
	preImage = append(preImage, contents...)
	hash := sha256.Sum256(preImage)
	return address(hash[:20])
}
