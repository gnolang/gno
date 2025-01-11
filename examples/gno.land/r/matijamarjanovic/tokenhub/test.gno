package tokenhub

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
)

var (
	// Initialize basic NFTs using the public constructor
	nftA = grc721.NewBasicNFT("Collection A", "NFTA")
	nftB = grc721.NewBasicNFT("Collection B", "NFTB")
	nftC = grc721.NewBasicNFT("Collection C", "NFTC")

	// Initialize GRC20 tokens
	tokenA, tokenAAdmin = grc20.NewToken("Token A", "TOKA", 6)
	tokenB, tokenBAdmin = grc20.NewToken("Token B", "TOKB", 6)
	tokenC, tokenCAdmin = grc20.NewToken("Token C", "TOKC", 6)
)

func init() {
	// Test address to receive tokens
	testAddr := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	// Mint test NFTs
	MintTestNFTs(testAddr)
	// Mint test tokens
	MintTestTokens(testAddr)

	// Register all NFT collections using the Getter method
	RegisterNFT(func() grc721.IGRC721 { return nftA }, "nft-a")
	RegisterNFT(func() grc721.IGRC721 { return nftB }, "nft-b")
	RegisterNFT(func() grc721.IGRC721 { return nftC }, "nft-c")

	// Register all tokens using their Getter methods
	RegisterToken(tokenA.Getter(), "token-a")
	RegisterToken(tokenB.Getter(), "token-b")
	RegisterToken(tokenC.Getter(), "token-c")
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

func MintTestTokens(to std.Address) error {
	if !to.IsValid() {
		return grc20.ErrInvalidAddress
	}

	// Mint some test tokens to the address
	if err := tokenAAdmin.Mint(to, 1000_000000); err != nil { // 1000 TOKA
		return err
	}
	if err := tokenBAdmin.Mint(to, 10000_000000); err != nil { // 10000 TOKB
		return err
	}
	if err := tokenCAdmin.Mint(to, 100_000000); err != nil { // 100 TOKC
		return err
	}

	return nil
}
