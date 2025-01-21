package tokenhub

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
	"gno.land/r/demo/users"
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

func getBalances(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}

	balances := avl.NewTree()
	var sb strings.Builder

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner { // show only the nfts owner owns
			balances.Set("nft:"+key, 1)
		}
		return false
	})

	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		tokenGetter := value.(grc20.TokenGetter)
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		balances.Set("token:"+key, balance)
		return false
	})

	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil {
			balances.Set("mt:"+key, balance)
		}
		return false
	})

	sb.WriteString("Balances: ")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		switch {
		case strings.HasPrefix(key, "nft:"):
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s,", symbol))
		case strings.HasPrefix(key, "token:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token:%s:%d,", symbol, balance))
		case strings.HasPrefix(key, "mt:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "mt:")
			sb.WriteString(ufmt.Sprintf("MultiToken:%s:%d,",
				symbol,
				balance))
		}
		return false
	})

	return sb.String()
}

func getNonZeroBalances(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}

	balances := avl.NewTree()
	var sb strings.Builder

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner { // show only the nfts owner owns
			balances.Set("nft:"+key, 1)
		}
		return false
	})

	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		tokenGetter := value.(grc20.TokenGetter)
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		if balance > 0 {
			balances.Set("token:"+key, balance)
		}
		return false
	})

	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil && balance > 0 {
			balances.Set("mt:"+key, balance)
		}
		return false
	})

	sb.WriteString("Non-Zero Balances: ")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		switch {
		case strings.HasPrefix(key, "nft:"):
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s,", symbol))
		case strings.HasPrefix(key, "token:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token:%s:%d,", symbol, balance))
		case strings.HasPrefix(key, "mt:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "mt:")
			sb.WriteString(ufmt.Sprintf("MultiToken:%s:%d,",
				symbol,
				balance))
		}
		return false
	})

	return sb.String()
}

func getAddressForUsername(addrOrName string) std.Address {
	addr := std.Address(addrOrName)

	if addr.IsValid() {
		return addr
	}
	if user := users.GetUserByName(addrOrName); user != nil {
		return user.Address
	}

	return ""
}
