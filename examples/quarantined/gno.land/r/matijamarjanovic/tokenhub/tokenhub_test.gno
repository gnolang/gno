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

const (
	tokenhubPkgPath  = "gno.land/r/matijamarjanovic/tokenhub"
	testRealmPkgPath = "gno.land/r/matijamarjanovic/testrealm"
)

// newTestNFT constructs an NFT through a same-realm cross so NewBasicNFT's
// IsCurrent + non-empty PkgPath checks pass.
func newTestNFT(_ int, rlm realm, name, symbol string) *grc721.BasicNFT {
	var nft *grc721.BasicNFT
	func(cur realm) {
		nft = grc721.NewBasicNFT(0, cur, name, symbol)
	}(cross(rlm))
	return nft
}

func TestTokenRegistration(cur realm, t *testing.T) {
	resetState(t)

	token, _ := grc20.NewToken(0, cur, "Test Token", "TEST", 6)
	key := RegisterToken(cross(cur), token, "test_token")

	uassert.Equal(t, tokenhubPkgPath+".TEST", key)

	retrievedToken := GetToken(tokenhubPkgPath + ".TEST")

	urequire.True(t, retrievedToken != nil, "Should retrieve registered token")
	uassert.Equal(t, "Test Token", retrievedToken.GetName(), "Token name should match")
	uassert.Equal(t, "TEST", retrievedToken.GetSymbol(), "Token symbol should match")
}

func TestNFTRegistration(cur realm, t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := newTestNFT(0, cur, "Test NFT", "TNFT")
	nft.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	err := RegisterNFT(cross(cur), nft.Getter(), "test_nft", "1")

	urequire.NoError(t, err, "Should register NFT without error")

	retrievedNFT := GetNFT(testRealmPkgPath + ".test_nft.1")

	urequire.True(t, retrievedNFT != nil, "Should retrieve registered NFT")

	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Test NFT", metadata.Name(), "NFT name should match")
	uassert.Equal(t, "TNFT", metadata.Symbol(), "NFT symbol should match")
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

	token, ledger := grc20.NewToken(0, cur, "Balance Token", "BAL", 6)
	RegisterToken(cross(cur), token, "balance_token")
	holder := unsafe.CurrentRealm().Address()
	ledger.Mint(holder, 1000)

	balances := GetUserTokenBalances(holder.String())
	uassert.True(t, strings.Contains(balances,
		"Token:"+tokenhubPkgPath+".BAL:1000"), "Should show correct GRC20 balance")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := newTestNFT(0, cur, "Test NFT", "TNFT")
	nft.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	uassert.NoError(t, RegisterNFT(cross(cur), nft.Getter(), "test_nft", "1"))

	balances = GetUserNFTBalances(unsafe.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "NFT:"+testRealmPkgPath+".test_nft.1"), "Should show correct NFT balance")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err := RegisterMultiToken(cross(cur), mt.Getter(), "1")
	urequire.NoError(t, err, "Should register multi-token without error")

	balances = GetUserMultiTokenBalances(unsafe.CurrentRealm().Address().String())
	uassert.True(t, strings.Contains(balances, "MultiToken:"+testRealmPkgPath+".1:0"), "Should show multi-token balance")

	nonZeroBalances := GetUserTokenBalancesNonZero(holder.String())
	uassert.True(t, strings.Contains(nonZeroBalances,
		"Token:"+tokenhubPkgPath+".BAL:1000"), "Should show non-zero GRC20 balance")
}

func TestErrorCases(cur realm, t *testing.T) {
	resetState(t)
	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := newTestNFT(0, cur, "Test NFT", "TNFT")
	err := RegisterNFT(cross(cur), nft.Getter(), "test_nft", "1")

	err = RegisterNFT(cross(cur), nft.Getter(), "test_nft", "1")
	uassert.True(t, err != nil, "Should not allow duplicate registration")

	mt := grc1155.NewBasicGRC1155Token("test-uri")
	err = RegisterMultiToken(cross(cur), mt.Getter(), "1")

	err = RegisterMultiToken(cross(cur), mt.Getter(), "1")
	uassert.True(t, err != nil, "Should not allow duplicate registrasdasdation")
}

func TestTokenListingFunctions(cur realm, t *testing.T) {
	resetState(t)

	grc20Token, _ := grc20.NewToken(0, cur, "Listing Token", "LST", 6)
	RegisterToken(cross(cur), grc20Token, "listing_token")

	grc20List := GetAllTokens()
	uassert.True(t, strings.Contains(grc20List, "Token:"+tokenhubPkgPath+".LST"),
		"GetAllGRC20Tokens should list registered token")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nftToken := newTestNFT(0, cur, "Listing NFT", "LNFT")
	nftToken.Mint(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), grc721.TokenID("1"))
	RegisterNFT(cross(cur), nftToken.Getter(), "listing_nft", "1")

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
	uassert.True(t, strings.Contains(completeList, "Token:"+tokenhubPkgPath+".LST"),
		"GetAllTokens should list GRC20 tokens")
	uassert.True(t, strings.Contains(completeList, "MultiToken:"+testRealmPkgPath+".listing_mt"),
		"GetAllTokens should list multi-tokens")
}

func TestMustGetFunctions(cur realm, t *testing.T) {
	resetState(t)

	token, _ := grc20.NewToken(0, cur, "Must Token", "MUST", 6)
	RegisterToken(cross(cur), token, "must_token")

	retrievedToken := MustGetToken(tokenhubPkgPath + ".MUST")
	uassert.Equal(t, "Must Token", retrievedToken.GetName(), "Token name should match")

	// Under the unified declaring-realm borrow, calling MustGetToken
	// (which dispatches to grc20reg.MustGet in /r/demo/defi/grc20reg)
	// crosses a soft realm boundary. Panics across borrows are aborts
	// not catchable by plain defer/recover — use revive() to capture.
	r := revive(func() { MustGetToken("non_existent_token") })
	uassert.True(t, r != nil, "MustGetToken should panic for non-existent token")
	uassert.True(t, strings.Contains(r.(string), "unknown token"), "Panic message should mention unknown token")

	testing.SetRealm(testing.NewCodeRealm(testRealmPkgPath))

	nft := newTestNFT(0, cur, "Must NFT", "MNFT")
	nft.Mint(unsafe.CurrentRealm().Address(), grc721.TokenID("1"))
	RegisterNFT(cross(cur), nft.Getter(), "must_nft", "1")

	// NFT registration key is cur.Previous().PkgPath() + "." + collection + "." + tokenId
	// where cur.Previous() at the cross-call site is the realm set via SetRealm above (testRealmPkgPath).
	retrievedNFT := MustGetNFT(testRealmPkgPath + ".must_nft.1")
	metadata, ok := retrievedNFT.(grc721.IGRC721CollectionMetadata)
	urequire.True(t, ok, "NFT should implement IGRC721CollectionMetadata")
	uassert.Equal(t, "Must NFT", metadata.Name(), "NFT name should match")

	// MustGetNFT's panic is local (panic site is in this package, no foreign
	// realm call), so defer/recover catches it without crossing any borrow.
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

	// MultiToken registration key is cur.Previous().PkgPath() + "." + tokenID.
	retrievedMT := MustGetMultiToken(testRealmPkgPath + ".must_mt")
	_, ok = retrievedMT.(grc1155.IGRC1155)
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
