package test

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/ufmt"
	"gno.land/r/matijamarjanovic/tokenhub"
)

var (
	nftA = grc721.NewBasicNFT("Collection A", "NFTA")
	nftB = grc721.NewBasicNFT("Collection B", "NFTB")
	nftC = grc721.NewBasicNFT("Collection C", "NFTC")

	tokenA, tokenAAdmin = grc20.NewToken("Token A", "TOKA", 6)
	tokenB, tokenBAdmin = grc20.NewToken("Token B", "TOKB", 6)
	tokenC, tokenCAdmin = grc20.NewToken("Token C", "TOKC", 6)

	mtGameItems   = grc1155.NewBasicGRC1155Token("https://game.example.com/items/{id}.json")
	mtArtworks    = grc1155.NewBasicGRC1155Token("https://art.example.com/pieces/{id}.json")
	mtCollectibles = grc1155.NewBasicGRC1155Token("https://collect.example.com/cards/{id}.json")
)

func init() {
	testAddr := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	tokenhub.RegisterNFT(func() grc721.IGRC721 { return nftA }, "nft-a")
	tokenhub.RegisterNFT(func() grc721.IGRC721 { return nftB }, "nft-b")
	tokenhub.RegisterNFT(func() grc721.IGRC721 { return nftC }, "nft-c")

	tokenhub.RegisterToken(tokenA.Getter(), "token-a")
	tokenhub.RegisterToken(tokenB.Getter(), "token-b")
	tokenhub.RegisterToken(tokenC.Getter(), "token-c")

	tokenhub.RegisterGRC1155Token(mtGameItems.Getter(), "sword")
	tokenhub.RegisterGRC1155Token(mtGameItems.Getter(), "potion")
	tokenhub.RegisterGRC1155Token(mtArtworks.Getter(), "artwork_1")
	tokenhub.RegisterGRC1155Token(mtCollectibles.Getter(), "rare_card")

	MintTestNFTs(testAddr)
	MintTestTokens(testAddr)
	MintTestMultiTokens(testAddr)

}

func MintTestNFTs(to std.Address) error {
	if !to.IsValid() {
		return grc721.ErrInvalidAddress
	}

	for i := 1; i <= 2; i++ {
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

	if err := tokenAAdmin.Mint(to, 1000_000000); err != nil {
		return err
	}
	if err := tokenBAdmin.Mint(to, 10000_000000); err != nil {
		return err
	}
	if err := tokenCAdmin.Mint(to, 100_000000); err != nil {
		return err
	}

	return nil
}

func MintTestMultiTokens(to std.Address) error {
	if !to.IsValid() {
		return grc1155.ErrInvalidAddress
	}

	if err := mtGameItems.SafeMint(to, "sword", 10); err != nil {
		return err
	}
	if err := mtGameItems.SafeMint(to, "potion", 50); err != nil {
		return err
	}
	if err := mtArtworks.SafeMint(to, "artwork_1", 1); err != nil {
		return err
	}
	if err := mtCollectibles.SafeMint(to, "rare_card", 1); err != nil {
		return err
	}

	return nil
}
