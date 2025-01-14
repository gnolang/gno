package tokenhub

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestTokenRegistration(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))
	std.TestSetOrigPkgAddr(std.DerivePkgAddr("gno.land/r/matijamarjanovic/home"))

	token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(token.Getter(), "test_token")

	getter := GetGRC20("gno.land/r/matijamarjanovic/tokenhub.test_token")
	urequire.True(t, getter != nil, "Should retrieve registered token")

	retrievedToken := getter()
	uassert.Equal(t, "Test Token", retrievedToken.GetName(), "Token name should match")
	uassert.Equal(t, "TEST", retrievedToken.GetSymbol(), "Token symbol should match")
}

func TestNFTRegistration(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	err := RegisterNFT(nft.Getter(), "test_nft")
	urequire.NoError(t, err, "Should register NFT without error")

	getter := GetGRC721("gno.land/r/matijamarjanovic/home.test_nft")
	urequire.True(t, getter != nil, "Should retrieve registered NFT")

	retrievedNFT := getter()
	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Test NFT", metadata.Name(), "NFT name should match")
	uassert.Equal(t, "TNFT", metadata.Symbol(), "NFT symbol should match")

	err = RegisterNFT(nft.Getter(), "test_nft")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestGRC1155Registration(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterGRC1155Token(mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	info, exists := GetGRC1155("gno.land/r/matijamarjanovic/home.1")
	uassert.True(t, exists, "Should find registered multi-token")
	urequire.True(t, info.Collection != nil, "Should retrieve multi-token collection")
	uassert.Equal(t, "1", info.TokenID, "Token ID should match")

	err = RegisterGRC1155Token(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestBalanceRetrieval(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	addr := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	token, ledger := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(token.Getter(), "test_tokenn")
	ledger.Mint(addr, 1000)

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	RegisterNFT(nft.Getter(), "test_nftt")
	nft.Mint(addr, grc721.TokenID("str"))

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterGRC1155Token(mt.Getter(), "11")
	mt.SafeMint(addr, grc1155.TokenID("11"), 5)

	balances := GetBalancesAll(string(addr))
	uassert.True(t, strings.Contains(balances, "Token:gno.land/r/matijamarjanovic/tokenhub.test_tokenn:1000"), "Should show correct GRC20 balance")
	uassert.True(t, strings.Contains(balances, "NFT:gno.land/r/matijamarjanovic/home.test_nftt:1"), "Should show correct NFT balance")
	uassert.True(t, strings.Contains(balances, "MultiToken:gno.land/r/matijamarjanovic/home.11:5"), "Should show multi-token balance")

	nonZeroBalances := GetBalances(string(addr))
	uassert.True(t, strings.Contains(nonZeroBalances, "Token:gno.land/r/matijamarjanovic/tokenhub.test_tokenn:1000"), "Should show non-zero GRC20 balance")
}

func TestErrorCases(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	err := RegisterNFT(nft.Getter(), "test_nft")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	err = RegisterNFT(nft.Getter(), "test_nft")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err = RegisterGRC1155Token(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	err = RegisterGRC1155Token(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	uassert.PanicsWithMessage(t, "invalid address or username: invalid", func() {
		GetBalances("invalid")
	})
}

func TestTokenListingFunctions(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	grc20Token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(grc20Token.Getter(), "listing_token")

	nftToken := grc721.NewBasicNFT("Test NFT", "TNFT")
	RegisterNFT(nftToken.Getter(), "listing_nft")

	multiToken := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterGRC1155Token(multiToken.Getter(), "listing_mt")

	nftList := GetAllNFTs()
	uassert.True(t, strings.Contains(nftList, "NFT:gno.land/r/matijamarjanovic/home.listing_nft"), 
		"GetAllNFTs should list registered NFT")

	grc20List := GetAllGRC20Tokens()
	uassert.True(t, strings.Contains(grc20List, "Token:gno.land/r/matijamarjanovic/tokenhub.listing_token"), 
		"GetAllGRC20Tokens should list registered token")

	grc1155List := GetAllMultiTokens()
	uassert.True(t, strings.Contains(grc1155List, "MultiToken:gno.land/r/matijamarjanovic/home.listing_mt"), 
		"GetAllMultiTokens should list registered multi-token")

	completeList := GetAllTokens()
	uassert.True(t, strings.Contains(completeList, "NFT:gno.land/r/matijamarjanovic/home.listing_nft"), 
		"GetAllTokens should list NFTs")
	uassert.True(t, strings.Contains(completeList, "Token:gno.land/r/matijamarjanovic/tokenhub.listing_token"), 
		"GetAllTokens should list GRC20 tokens")
	uassert.True(t, strings.Contains(completeList, "MultiToken:gno.land/r/matijamarjanovic/home.listing_mt"), 
		"GetAllTokens should list multi-tokens")
}
