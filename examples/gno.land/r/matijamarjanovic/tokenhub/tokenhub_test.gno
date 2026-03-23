package tokenhub

import (
	"chain/runtime"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc1155"
	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const testRealmPkgPath = "gno.land/r/matijamarjanovic/testrealm"

func TestTokenRegistration(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(cross, token, "test_token")

	retrievedToken := GetToken("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/matijamarjanovic/testrealm.test_token")

	urequire.True(t, retrievedToken != nil, "Should retrieve registered token")
	uassert.Equal(t, "Test Token", retrievedToken.GetName(), "Token name should match")
	uassert.Equal(t, "TEST", retrievedToken.GetSymbol(), "Token symbol should match")
}

func TestNFTRegistration(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	nft.Mint(runtime.CurrentRealm().Address(), grc721.TokenID("1"))
	err := RegisterNFT(cross, nft.Getter(), "test_nft", "1")

	urequire.NoError(t, err, "Should register NFT without error")

	retrievedNFT := GetNFT(testRealmPkgPath + ".test_nft.1")

	urequire.True(t, retrievedNFT != nil, "Should retrieve registered NFT")

	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Test NFT", metadata.Name(), "NFT name should match")
	uassert.Equal(t, "TNFT", metadata.Symbol(), "NFT symbol should match")
}

func TestGRC1155Registration(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(cross, mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	multiToken := GetMultiToken(testRealmPkgPath + ".1")
	urequire.True(t, multiToken != nil, "Should retrieve multi-token")
	_, ok := multiToken.(grc1155.IGRC1155)
	urequire.True(t, ok, "Retrieved multi-token should implement IGRC1155")

	err = RegisterMultiToken(cross, mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestBalanceRetrieval(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	token, ledger := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(cross, token, "test_token")
	ledger.Mint(runtime.CurrentRealm().Address(), 1000)

	balances := GetUserTokenBalances(runtime.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances,
		"Token:gno.land/r/matijamarjanovic/tokenhub."+testRealmPkgPath+".test_token:1000"), "Should show correct GRC20 balance")

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	nft.Mint(runtime.CurrentRealm().Address(), grc721.TokenID("1"))
	uassert.NoError(t, RegisterNFT(cross, nft.Getter(), "test_nft", "1"))

	balances = GetUserNFTBalances(runtime.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "NFT:"+testRealmPkgPath+".test_nft.1"), "Should show correct NFT balance")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(cross, mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	balances = GetUserMultiTokenBalances(runtime.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "MultiToken:"+testRealmPkgPath+".1:0"), "Should show multi-token balance")

	nonZeroBalances := GetUserTokenBalancesNonZero(runtime.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(nonZeroBalances,
		"Token:gno.land/r/matijamarjanovic/tokenhub."+testRealmPkgPath+".test_token:1000"), "Should show non-zero GRC20 balance")
}

func TestErrorCases(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	err := RegisterNFT(cross, nft.Getter(), "test_nft", "1")

	err = RegisterNFT(cross, nft.Getter(), "test_nft", "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err = RegisterMultiToken(cross, mt.Getter(), "1")

	err = RegisterMultiToken(cross, mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registrasdasdation")
}

func TestTokenListingFunctions(t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	grc20Token, _ := grc20.NewToken("Test Token", "TEST", 6)
	RegisterToken(cross, grc20Token, "listing_token")

	grc20List := GetAllTokens()
	uassert.True(t, strings.Contains(grc20List, "Token:gno.land/r/matijamarjanovic/tokenhub."+testRealmPkgPath+".listing_token"),
		"GetAllGRC20Tokens should list registered token")

	nftToken := grc721.NewBasicNFT("Listing NFT", "LNFT")
	nftToken.Mint(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), grc721.TokenID("1"))
	RegisterNFT(cross, nftToken.Getter(), "listing_nft", "1")

	nftList := GetAllNFTs()
	uassert.True(t, strings.Contains(nftList, "NFT:"+testRealmPkgPath+".listing_nft.1"),
		"GetAllNFTs should list registered NFT")

	multiToken := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterMultiToken(cross, multiToken.Getter(), "listing_mt")

	grc1155List := GetAllMultiTokens()
	uassert.True(t, strings.Contains(grc1155List, "MultiToken:"+testRealmPkgPath+".listing_mt"),
		"GetAllMultiTokens should list registered multi-token")

	completeList := GetAllRegistered()
	uassert.True(t, strings.Contains(completeList, "NFT:"+testRealmPkgPath+".listing_nft.1"),
		"GetAllTokens should list NFTs")
	uassert.True(t, strings.Contains(completeList, "Token:gno.land/r/matijamarjanovic/tokenhub."+testRealmPkgPath+".listing_token"),
		"GetAllTokens should list GRC20 tokens")
	uassert.True(t, strings.Contains(completeList, "MultiToken:"+testRealmPkgPath+".listing_mt"),
		"GetAllTokens should list multi-tokens")
}

func TestMustGetFunctions(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	token, _ := grc20.NewToken("Must Token", "MUST", 6)
	RegisterToken(cross, token, "must_token")

	retrievedToken := MustGetToken("gno.land/r/matijamarjanovic/tokenhub." + testRealmPkgPath + ".must_token")
	uassert.Equal(t, "Must Token", retrievedToken.GetName(), "Token name should match")

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "MustGetToken should panic for non-existent token")
		uassert.True(t, strings.Contains(r.(string), "unknown token"), "Panic message should mention unknown token")
	}()
	MustGetToken("non_existent_token")

	nft := grc721.NewBasicNFT("Must NFT", "MNFT")
	nft.Mint(runtime.CurrentRealm().Address(), grc721.TokenID("1"))
	RegisterNFT(cross, nft.Getter(), "must_nft", "1")

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
	RegisterMultiToken(cross, mt.Getter(), "must_mt")

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

func resetState(t *testing.T) {
	t.Helper()

	registeredTokens = avl.NewTree()
	registeredNFTs = avl.NewTree()
	registeredMTs = avl.NewTree()
}
