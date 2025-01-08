package fomo3d

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"
)

// Fomo3DNFT manages the NFTs for Fomo3D winners
type Fomo3DNFT struct {
	nft          *grc721.basicNFT
	roundCounter int
}

// Global instance
var gNFT = &Fomo3DNFT{
	nft:          grc721.NewBasicNFT("Fomo3D Winner", "FOMO"),
	roundCounter: 0,
}

// GetNFT returns the global NFT instance
func GetNFT() *Fomo3DNFT {
	return gNFT
}

// MintRoundWinnerNFT mints a new NFT for the round winner
func (f *Fomo3DNFT) MintRoundWinnerNFT(winner std.Address) error {
	f.roundCounter++
	tokenID := grc721.TokenID(strconv.Itoa(f.roundCounter))

	// Mint NFT to winner
	err := f.nft.Mint(winner, tokenID)
	if err != nil {
		return err
	}

	// Set token URI with round information
	uri := "Fomo3D Round " + strconv.Itoa(f.roundCounter) + " Winner"
	_, err = f.nft.SetTokenURI(tokenID, grc721.TokenURI(uri))
	if err != nil {
		return err
	}

	return nil
}

// GetWinnerNFTs returns all NFTs owned by an address
func (f *Fomo3DNFT) GetWinnerNFTs(owner std.Address) (uint64, error) {
	return f.nft.BalanceOf(owner)
}

// GetTokenURI returns the URI for a specific token
func (f *Fomo3DNFT) GetTokenURI(tokenID grc721.TokenID) (string, error) {
	return f.nft.TokenURI(tokenID)
}
