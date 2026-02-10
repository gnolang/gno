package fomo3d

import (
	"strconv"

	"gno.land/p/demo/grc/grc721"
)

var fomo3dNFT = grc721.NewNFTWithMetadata("Fomo3D Winner", "FOMO")

// Public getters

func Name() string {
	return fomo3dNFT.Name()
}

func Symbol() string {
	return fomo3dNFT.Symbol()
}

func BalanceOf(owner address) (uint64, error) {
	return fomo3dNFT.BalanceOf(owner)
}

func OwnerOf(tokenID grc721.TokenID) (address, error) {
	return fomo3dNFT.OwnerOf(tokenID)
}

func TokenMetadata(tokenID grc721.TokenID) (grc721.Metadata, error) {
	return fomo3dNFT.TokenMetadata(tokenID)
}

// Transfer and approval methods

func TransferFrom(from, to address, tokenID grc721.TokenID) error {
	return fomo3dNFT.TransferFrom(from, to, tokenID)
}

func SafeTransferFrom(from, to address, tokenID grc721.TokenID) error {
	return fomo3dNFT.SafeTransferFrom(from, to, tokenID)
}

func Approve(approved address, tokenID grc721.TokenID) error {
	return fomo3dNFT.Approve(approved, tokenID)
}

func GetApproved(tokenID grc721.TokenID) (address, error) {
	return fomo3dNFT.GetApproved(tokenID)
}

func SetApprovalForAll(operator address, approved bool) error {
	return fomo3dNFT.SetApprovalForAll(operator, approved)
}

func IsApprovedForAll(owner, operator address) bool {
	return fomo3dNFT.IsApprovedForAll(owner, operator)
}

// Mints a new NFT for the round winner
func mintRoundWinnerNFT(winner address, roundNumber int64) error {
	if winner == "" {
		return ErrZeroAddress
	}

	roundStr := strconv.FormatInt(roundNumber, 10)
	tokenID := grc721.TokenID(roundStr)

	// Create metadata
	metadata := grc721.Metadata{
		Name:            "Fomo3D Winner - Round #" + roundStr,
		Description:     "Winner of Fomo3D round #" + roundStr,
		Image:           "https://ipfs.io/ipfs/bafybeidayyli6bpewkhgtwqpgubmo77kmgjn4r5zq2i7usoyadcmvynhhq",
		ExternalURL:     "https://gno.land/r/stefann/fomo3d:round/" + roundStr, // TODO: Add this render in main realm that shows details of specific round
		Attributes:      []grc721.Trait{},
		BackgroundColor: "2D2D2D", // Dark theme background
	}

	if err := fomo3dNFT.Mint(winner, tokenID); err != nil {
		return err
	}

	fomo3dNFT.SetTokenMetadata(tokenID, metadata)

	return nil
}
