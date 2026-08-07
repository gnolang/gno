package tokenhub

import (
	"gno.land/p/demo/tokens/grc20"
	"gno.land/r/demo/defi/grc20reg"

	"gno.land/p/demo/tokens/grc1155"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"

	"gno.land/r/leon/hor"
)

type GRC1155TokenInfo struct {
	Collection grc1155.MultiTokenGetter
	TokenID    string
}

var (
	registeredTokens = avl.NewTree() // rlmPath[.slug] -> *grc20.Token
	registeredNFTs   = avl.NewTree() // rlmPath[.slug] -> *grc721.Token
	registeredMTs    = avl.NewTree() // rlmPath[.slug] -> GRC1155TokenInfo
)

const pageSize = 10

func init(cur realm) {
	hor.Register(cross(cur), "Token Hub", "Registry for tokens and NFTs on gno.land")
}

// RegisterToken returns grc20reg's canonical key (rlmPath.symbol), not the
// token's full ID (which carries a trailing sequence id).
func RegisterToken(cur realm, token *grc20.Token, slug string) string {
	return grc20reg.Register(cross(cur), token, slug)
}

// RegisterNFT stores the concrete *grc721.Token (not an interface getter), so
// an entry cannot be a lying reader implementation.
func RegisterNFT(cur realm, token *grc721.Token, collection string, tokenId string) error {
	if token == nil {
		return ErrNFTNotFound
	}

	nftOwner, err := token.OwnerOf(grc721.TokenID(tokenId))
	if err != nil {
		return err
	}
	if !nftOwner.IsValid() {
		return ErrNFTtokIDNotExists
	}

	rlmPath := cur.Previous().PkgPath()
	key := rlmPath + "." + collection + "." + tokenId

	if registeredNFTs.Has(key) {
		return ErrNFTAlreadyRegistered
	}

	registeredNFTs.Set(key, token)
	return nil
}

func RegisterMultiToken(cur realm, mtGetter grc1155.MultiTokenGetter, tokenID string) error {
	rlmPath := cur.Previous().PkgPath()
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
