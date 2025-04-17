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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(token.Getter(), "test_token")

	retrievedToken := GetToken("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.test_token")
	urequire.True(t, retrievedToken != nil, "Should retrieve registered token")

	uassert.Equal(t, "Test Token", retrievedToken.GetName(), "Token name should match")
	uassert.Equal(t, "TEST", retrievedToken.GetSymbol(), "Token symbol should match")
}

func TestNFTRegistration(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	nft.Mint(std.CurrentRealm().Address(), grc721.TokenID("1"))
	err := RegisterNFT(nft.Getter(), "test_nft", "1")
	urequire.NoError(t, err, "Should register NFT without error")

	retrievedNFT := GetNFT("gno.land/r/matijamarjanovic/home.test_nft.1")
	urequire.True(t, retrievedNFT != nil, "Should retrieve registered NFT")

	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Test NFT", metadata.Name(), "NFT name should match")
	uassert.Equal(t, "TNFT", metadata.Symbol(), "NFT symbol should match")
}

func TestGRC1155Registration(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	multiToken := GetMultiToken("gno.land/r/matijamarjanovic/home.1")
	urequire.True(t, multiToken != nil, "Should retrieve multi-token")
	_, ok := multiToken.(grc1155.IGRC1155)
	urequire.True(t, ok, "Retrieved multi-token should implement IGRC1155")

	err = RegisterMultiToken(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestBalanceRetrieval(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	token, ledger := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(token.Getter(), "test_tokenn")
	ledger.Mint(std.CurrentRealm().Address(), 1000)

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterMultiToken(mt.Getter(), "11")
	mt.SafeMint(std.CurrentRealm().Address(), grc1155.TokenID("11"), 5)

	balances := GetUserTokenBalances(std.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "Token:gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.test_tokenn:1000"), "Should show correct GRC20 balance")

	balances = GetUserNFTBalances(std.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "NFT:gno.land/r/matijamarjanovic/home.test_nft.1"), "Should show correct NFT balance") //already minted in test register

	balances = GetUserMultiTokenBalances(std.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "MultiToken:gno.land/r/matijamarjanovic/home.11:5"), "Should show multi-token balance")

	nonZeroBalances := GetUserTokenBalancesNonZero(std.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(nonZeroBalances, "Token:gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.test_tokenn:1000"), "Should show non-zero GRC20 balance")
}

func TestErrorCases(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	err := RegisterNFT(nft.Getter(), "test_nft", "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	err = RegisterNFT(nft.Getter(), "test_nft", "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err = RegisterMultiToken(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	err = RegisterMultiToken(mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestTokenListingFunctions(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	grc20Token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(grc20Token.Getter(), "listing_token")

	nftToken := grc721.NewBasicNFT("Listing NFT", "LNFT")
	nftToken.Mint(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), grc721.TokenID("1"))
	RegisterNFT(nftToken.Getter(), "listing_nft", "1")

	multiToken := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterMultiToken(multiToken.Getter(), "listing_mt")

	nftList := GetAllNFTs()
	uassert.True(t, strings.Contains(nftList, "NFT:gno.land/r/matijamarjanovic/home.listing_nft.1"),
		"GetAllNFTs should list registered NFT")

	grc20List := GetAllTokens()
	uassert.True(t, strings.Contains(grc20List, "Token:gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.listing_token"),
		"GetAllGRC20Tokens should list registered token")

	grc1155List := GetAllMultiTokens()
	uassert.True(t, strings.Contains(grc1155List, "MultiToken:gno.land/r/matijamarjanovic/home.listing_mt"),
		"GetAllMultiTokens should list registered multi-token")

	completeList := GetAllRegistered()
	uassert.True(t, strings.Contains(completeList, "NFT:gno.land/r/matijamarjanovic/home.listing_nft.1"),
		"GetAllTokens should list NFTs")
	uassert.True(t, strings.Contains(completeList, "Token:gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.listing_token"),
		"GetAllTokens should list GRC20 tokens")
	uassert.True(t, strings.Contains(completeList, "MultiToken:gno.land/r/matijamarjanovic/home.listing_mt"),
		"GetAllTokens should list multi-tokens")
}

func TestMustGetFunctions(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	token, _ := grc20.NewToken("Must Token", "MUST", 6)
	RegisterToken(token.Getter(), "must_token")

	retrievedToken := MustGetToken("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/home.must_token")
	uassert.Equal(t, "Must Token", retrievedToken.GetName(), "Token name should match")

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "MustGetToken should panic for non-existent token")
		uassert.True(t, strings.Contains(r.(string), "unknown token"), "Panic message should mention unknown token")
	}()
	MustGetToken("non_existent_token")

	nft := grc721.NewBasicNFT("Must NFT", "MNFT")
	nft.Mint(std.CurrentRealm().Address(), grc721.TokenID("1"))
	RegisterNFT(nft.Getter(), "must_nft", "1")

	retrievedNFT := MustGetNFT("gno.land/r/matijamarjanovic/home.must_nft.1")
	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Must NFT", metadata.Name(), "NFT name should match")

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "MustGetNFT should panic for non-existent NFT")
		uassert.True(t, strings.Contains(r.(string), "unknown NFT"), "Panic message should mention unknown NFT")
	}()
	MustGetNFT("non_existent_nft")

	mt := grc1155.NewBasicGRC1155Token("must-uri")
	RegisterMultiToken(mt.Getter(), "must_mt")

	retrievedMT := MustGetMultiToken("gno.land/r/matijamarjanovic/home.must_mt")
	_, ok = retrievedMT.(grc1155.IGRC1155)
	urequire.True(t, ok, "Retrieved multi-token should implement IGRC1155")

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "MustGetMultiToken should panic for non-existent multi-token")
		uassert.True(t, strings.Contains(r.(string), "unknown multi-token"), "Panic message should mention unknown multi-token")
	}()
	MustGetMultiToken("non_existent_mt")
}

func TestGetAddressForUsername(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/matijamarjanovic/home"))

	validAddr := "g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"
	addr := getAddressForUsername(validAddr)
	uassert.Equal(t, validAddr, addr.String(), "Should return same address for valid address input")

	invalidInput := "invalid_input"
	addr = getAddressForUsername(invalidInput)
	uassert.Equal(t, "", addr.String(), "Should return empty address for invalid input")
}
