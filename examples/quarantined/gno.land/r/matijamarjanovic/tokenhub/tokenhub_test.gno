package tokenhub

import (
	"chain/runtime/unsafe"
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

// Same-realm cross so NewToken's IsCurrent + non-empty PkgPath checks pass.
func newTestNFT(name, symbol string, rlm realm) (tok *grc721.Token, led *grc721.PrivateLedger) {
	func(cur realm) {
		tok, led = grc721.NewToken(name, symbol, 0, cur)
	}(cross(rlm))
	return
}

func TestTokenRegistration(cur realm, t *testing.T) {
	resetState(t)

	token, _ := grc20.NewToken("Test Token", "TEST", 6, cur)
	key := RegisterToken(cross(cur), token, "test_token")

	uassert.Equal(t, token.GetOriginRealm(), "gno.land/r/matijamarjanovic/tokenhub")

	retrievedToken := GetToken(key)

	urequire.True(t, retrievedToken != nil, "Should retrieve registered token")
	uassert.Equal(t, "Test Token", retrievedToken.GetName(), "Token name should match")
	uassert.Equal(t, "TEST", retrievedToken.GetSymbol(), "Token symbol should match")
}

func TestNFTRegistration(cur realm, t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft, nftLedger := newTestNFT("Test NFT", "TNFT", cur)
	nftLedger.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	err := RegisterNFT(cross(cur), nft, "test_nft", "1")

	urequire.NoError(t, err, "Should register NFT without error")

	retrievedNFT := GetNFT(testRealmPkgPath + ".test_nft.1")

	urequire.True(t, retrievedNFT != nil, "Should retrieve registered NFT")

	uassert.Equal(t, "Test NFT", retrievedNFT.GetName(), "NFT name should match")
	uassert.Equal(t, "TNFT", retrievedNFT.GetSymbol(), "NFT symbol should match")
}

func TestGRC1155Registration(cur realm, t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(cross(cur), mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	multiToken := GetMultiToken(testRealmPkgPath + ".1")
	urequire.True(t, multiToken != nil, "Should retrieve multi-token")
	_, ok := multiToken.(grc1155.IGRC1155)
	urequire.True(t, ok, "Retrieved multi-token should implement IGRC1155")

	err = RegisterMultiToken(cross(cur), mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestBalanceRetrieval(cur realm, t *testing.T) {
	resetState(t)

	token, ledger := grc20.NewToken("Balance Token", "BAL", 6, cur)
	key := RegisterToken(cross(cur), token, "balance_token")
	holder := unsafe.CurrentRealm().Address()
	ledger.Mint(holder, 1000)

	balances := GetUserTokenBalances(holder.String())
	uassert.True(t, strings.Contains(balances,
		"Token:"+key+":1000"), "Should show correct GRC20 balance")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft, nftLedger := newTestNFT("Test NFT", "TNFT", cur)
	nftLedger.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	uassert.NoError(t, RegisterNFT(cross(cur), nft, "test_nft", "1"))

	balances = GetUserNFTBalances(unsafe.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "NFT:"+testRealmPkgPath+".test_nft.1"), "Should show correct NFT balance")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(cross(cur), mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	balances = GetUserMultiTokenBalances(unsafe.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "MultiToken:"+testRealmPkgPath+".1:0"), "Should show multi-token balance")

	nonZeroBalances := GetUserTokenBalancesNonZero(holder.String())
	uassert.True(t, strings.Contains(nonZeroBalances,
		"Token:"+key+":1000"), "Should show non-zero GRC20 balance")
}

func TestErrorCases(cur realm, t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft, nftLedger := newTestNFT("Test NFT", "TNFT", cur)
	nftLedger.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	err := RegisterNFT(cross(cur), nft, "test_nft", "1")
	uassert.NoError(t, err, "First registration should succeed")

	err = RegisterNFT(cross(cur), nft, "test_nft", "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err = RegisterMultiToken(cross(cur), mt.Getter(), "1")
	uassert.NoError(t, err, "First multi-token registration should succeed")

	err = RegisterMultiToken(cross(cur), mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")
}

func TestTokenListingFunctions(cur realm, t *testing.T) {
	resetState(t)

	grc20Token, _ := grc20.NewToken("Listing Token", "LST", 6, cur)
	grc20Key := RegisterToken(cross(cur), grc20Token, "listing_token")

	grc20List := GetAllTokens()
	uassert.True(t, strings.Contains(grc20List, "Token:"+grc20Key),
		"GetAllGRC20Tokens should list registered token")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nftToken, nftLedger := newTestNFT("Listing NFT", "LNFT", cur)
	nftLedger.Mint(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), grc721.TokenID("1"))
	RegisterNFT(cross(cur), nftToken, "listing_nft", "1")

	nftList := GetAllNFTs()
	uassert.True(t, strings.Contains(nftList, "NFT:"+testRealmPkgPath+".listing_nft.1"),
		"GetAllNFTs should list registered NFT")

	multiToken := grc1155.NewBasicGRC1155Token("test-uri")
	RegisterMultiToken(cross(cur), multiToken.Getter(), "listing_mt")

	grc1155List := GetAllMultiTokens()
	uassert.True(t, strings.Contains(grc1155List, "MultiToken:"+testRealmPkgPath+".listing_mt"),
		"GetAllMultiTokens should list registered multi-token")

	completeList := GetAllRegistered()
	uassert.True(t, strings.Contains(completeList, "NFT:"+testRealmPkgPath+".listing_nft.1"),
		"GetAllTokens should list NFTs")
	uassert.True(t, strings.Contains(completeList, "Token:"+grc20Key),
		"GetAllTokens should list GRC20 tokens")
	uassert.True(t, strings.Contains(completeList, "MultiToken:"+testRealmPkgPath+".listing_mt"),
		"GetAllTokens should list multi-tokens")
}

func TestMustGetFunctions(cur realm, t *testing.T) {
	resetState(t)

	token, _ := grc20.NewToken("Must Token", "MUST", 6, cur)
	key := RegisterToken(cross(cur), token, "must_token")

	retrievedToken := MustGetToken(key)
	uassert.Equal(t, "Must Token", retrievedToken.GetName(), "Token name should match")

	// MustGetToken dispatches into grc20reg across a borrow; that abort is not
	// catchable by defer/recover, so capture it with revive().
	r := revive(func() { MustGetToken("non_existent_token") })
	uassert.True(t, r != nil, "MustGetToken should panic for non-existent token")
	uassert.True(t, strings.Contains(r.(string), "unknown token"), "Panic message should mention unknown token")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft, nftLedger := newTestNFT("Must NFT", "MNFT", cur)
	nftLedger.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	RegisterNFT(cross(cur), nft, "must_nft", "1")

	retrievedNFT := MustGetNFT(testRealmPkgPath + ".must_nft.1")
	uassert.Equal(t, "Must NFT", retrievedNFT.GetName(), "NFT name should match")

	// MustGetNFT's panic is local (no borrow crossed), so defer/recover catches it.
	func() {
		defer func() {
			r := recover()
			uassert.True(t, r != nil, "MustGetNFT should panic for non-existent NFT")
			uassert.True(t, strings.Contains(r.(string), "unknown NFT"), "Panic message should mention unknown NFT")
		}()
		MustGetNFT("non_existent_nft")
	}()

	mt := grc1155.NewBasicGRC1155Token("must-uri")
	RegisterMultiToken(cross(cur), mt.Getter(), "must_mt")

	retrievedMT := MustGetMultiToken(testRealmPkgPath + ".must_mt")
	_, ok := retrievedMT.(grc1155.IGRC1155)
	urequire.True(t, ok, "Retrieved multi-token should implement IGRC1155")

	// Local panic, same as MustGetNFT above.
	func() {
		defer func() {
			r := recover()
			uassert.True(t, r != nil, "MustGetMultiToken should panic for non-existent multi-token")
			uassert.True(t, strings.Contains(r.(string), "unknown multi-token"), "Panic message should mention unknown multi-token")
		}()
		MustGetMultiToken("non_existent_mt")
	}()
}

func resetState(t *testing.T) {
	t.Helper()

	registeredTokens = avl.NewTree()
	registeredNFTs = avl.NewTree()
	registeredMTs = avl.NewTree()
}
