package tokenhub

import (
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
)

var (
	// Initialize basic NFTs using the public constructor
	nftA = grc721.NewBasicNFT("Collection A", "NFTA")
	nftB = grc721.NewBasicNFT("Collection B", "NFTB")
	nftC = grc721.NewBasicNFT("Collection C", "NFTC")
)

func init() {

	// Test address to receive NFTs
	testAddr := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	// Mint test NFTs
	MintTestNFTs(testAddr)

	// Register all NFT collections using the Getter method
	RegisterNFT(func() grc721.IGRC721 { return nftA })
	RegisterNFT(func() grc721.IGRC721 { return nftB })
	RegisterNFT(func() grc721.IGRC721 { return nftC })
}

func MintTestNFTs(to std.Address) error {
	if !to.IsValid() {
		return grc721.ErrInvalidAddress
	}

	// Mint NFTs for each collection
	for i := 1; i <= 2; i++ {
		// Create TokenID for each NFT
		tokenIDA := grc721.TokenID(ufmt.Sprintf("NFTA_%d", i))
		if err := nftA.Mint(to, tokenIDA); err != nil {
			return err
		}

		tokenIDB := grc721.TokenID(ufmt.Sprintf("NFTB_%d", i))
		if err := nftB.Mint(to, tokenIDB); err != nil {
			return err
		}

		tokenIDC := grc721.TokenID(ufmt.Sprintf("NFTC_%d", i))
		if err := nftC.Mint(to, tokenIDC); err != nil {
			return err
		}
	}

	return nil
}
