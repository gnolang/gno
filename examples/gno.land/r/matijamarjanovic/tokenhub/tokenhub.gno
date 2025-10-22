package tokenhub

import (
	"chain/runtime"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/fqname"
	"gno.land/r/demo/defi/grc20reg"

	"gno.land/p/demo/tokens/grc1155"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl"

	"gno.land/r/leon/hor"
)

type GRC1155TokenInfo struct {
	Collection grc1155.MultiTokenGetter
	TokenID    string
}

var (
	registeredTokens = avl.NewTree() // rlmPath[.slug] -> *grc20.Token
	registeredNFTs   = avl.NewTree() // rlmPath[.slug] -> grc721.NFTGetter
	registeredMTs    = avl.NewTree() // rlmPath[.slug] -> GRC1155TokenInfo
)

const pageSize = 10

func init() {
	hor.Register(cross, "Token Hub", "Registry for tokens and NFTs on gno.land")
}

// RegisterToken is a function that uses gno.land/r/demo/defi/grc20reg to register a token
// It uses the slug to construct a key and then registers the token in the registry
// The logic is the same as in grc20reg, but it's done here so the key path is callers pkgpath and not of this realm
// After doing so, the token hub realm uses grc20reg's registry as a read-only avl.Tree
//
// Note: register token returns the key path that can be used to retrieve the token
func RegisterToken(cur realm, token *grc20.Token, slug string) string {
	rlmPath := runtime.PreviousRealm().PkgPath()
	key := fqname.Construct(rlmPath, slug)

	grc20reg.Register(cross, token, key)

	return fqname.Construct(runtime.CurrentRealm().PkgPath(), key)
}

// RegisterNFT is a function that registers an NFT in an avl.Tree
func RegisterNFT(cur realm, nftGetter grc721.NFTGetter, collection string, tokenId string) error {
	nft := nftGetter()
	_, ok := nft.(grc721.IGRC721CollectionMetadata)
	if !ok {
		return ErrNFTNotMetadata
	}

	nftOwner, err := nft.OwnerOf(grc721.TokenID(tokenId))

	if err != nil {
		return err
	}
	if !nftOwner.IsValid() {
		return ErrNFTtokIDNotExists
	}

	rlmPath := runtime.PreviousRealm().PkgPath()
	key := rlmPath + "." + collection + "." + tokenId

	if registeredNFTs.Has(key) {
		return ErrNFTAlreadyRegistered
	}

	registeredNFTs.Set(key, nftGetter)
	return nil
}

// RegisterMultiToken is a function that registers a multi-token in an avl.Tree
// The avl.Tree value is a struct defined in this realm. It contains not only the getter (like other token types) but also the tokenID
func RegisterMultiToken(cur realm, mtGetter grc1155.MultiTokenGetter, tokenID string) error {
	rlmPath := runtime.PreviousRealm().PkgPath()
	key := rlmPath + "." + tokenID

	if registeredMTs.Has(key) {
		return ErrMTAlreadyRegistered
	}

	registeredMTs.Set(key, GRC1155TokenInfo{
		Collection: mtGetter,
		TokenID:    tokenID,
	})
	return nil
}
