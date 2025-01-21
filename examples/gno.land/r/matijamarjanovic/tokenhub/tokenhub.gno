package tokenhub

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/r/demo/grc20reg"
)

type GRC1155TokenInfo struct {
	Collection grc1155.MultiTokenGetter
	TokenID    string
}

var (
	registeredTokens = avl.NewTree() // rlmPath[.slug] -> grc20.TokenGetter
	registeredNFTs = avl.NewTree() // rlmPath[.slug] -> grc721.NFTGetter
	registeredMTs  = avl.NewTree() // rlmPath[.slug] -> GRC1155TokenInfo

	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound          = errors.New("NFT not found")
	ErrMTAlreadyRegistered  = errors.New("multi-token already registered")
	ErrMTNotFound           = errors.New("multi-token not found")
	ErrMTInfoNotFound       = errors.New("multi-token info not found")
	ErrNFTtokIDNotExists    = errors.New("NFT token ID does not exists")
	ErrNFTNotMetadata       = errors.New("NFT must implement IGRC721CollectionMetadata")
)

const pageSize = 1	
const token = "token" // grc20
const nft = "nft" // grc721
const mt = "mt" // grc1155

func RegisterToken(tokenGetter grc20.TokenGetter, slug string) {
	grc20reg.Register(tokenGetter, slug)
}

// registering a collection of NFTs
func RegisterNFT(nftGetter grc721.NFTGetter, collection string, tokenId string) error {
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

	rlmPath := std.PrevRealm().PkgPath()
	key := rlmPath + "." + collection + "." + tokenId

	if registeredNFTs.Has(key) {
		return ErrNFTAlreadyRegistered
	}

	registeredNFTs.Set(key, nftGetter)
	return nil
}

// for grc1155 is a little bit different, since the one collection can have multiple tokens, this is why don't use slugs but tokenIDs,
// and the collection is saved as a part of the :value in the node
func RegisterMultiToken(mtGetter grc1155.MultiTokenGetter, tokenID string) error {
	rlmPath := std.PrevRealm().PkgPath()

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
