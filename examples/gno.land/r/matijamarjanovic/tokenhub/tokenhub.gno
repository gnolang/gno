package tokenhub

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/demo/grc20reg"
	"gno.land/p/demo/fqname"
	"gno.land/p/moul/md"
)

// Add this struct to store GRC1155 token info
type GRC1155TokenInfo struct {
	Collection grc1155.MultiTokenGetter
	TokenID    string
}

var (
	registeredNFTs = avl.NewTree() // rlmPath[.slug] -> NFTGetter
	registeredMTs = avl.NewTree()  // rlmPath[.slug] -> GRC1155TokenInfo

	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound         = errors.New("NFT not found")
	ErrMTAlreadyRegistered = errors.New("Multi-token already registered")
	ErrMTNotFound         = errors.New("Multi-token not found")
	ErrMTInfoNotFound     = errors.New("Multi-token info not found")
)

func RegisterToken(tokenGetter grc20.TokenGetter, slugs ...string) {
	slug := ""
	if len(slugs) > 0 {
			slug = slugs[0]
	}
	grc20reg.Register(tokenGetter, slug)
}

func RegisterNFT(nftGetter grc721.NFTGetter, slugs ...string) error {
	nft := nftGetter()
	_, ok := nft.(grc721.IGRC721CollectionMetadata)
	if !ok {
		return errors.New("NFT must implement IGRC721CollectionMetadata")
	}

	rlmPath := std.PrevRealm().PkgPath()
	slug := ""
	if len(slugs) > 0 {
		slug = slugs[0]
	}
	key := fqname.Construct(rlmPath, slug)

	if registeredNFTs.Has(key) {
		return ErrNFTAlreadyRegistered
	}

	registeredNFTs.Set(key, nftGetter)
	return nil
}

func RegisterGRC1155Token(mtGetter grc1155.MultiTokenGetter, tokenID string, slugs ...string) error {
	rlmPath := std.PrevRealm().PkgPath()
	slug := ""
	if len(slugs) > 0 {
		slug = slugs[0]
	}
	key := fqname.Construct(rlmPath, slug)

	if registeredMTs.Has(key) {
		return ErrMTAlreadyRegistered
	}

	registeredMTs.Set(key, GRC1155TokenInfo{
		Collection: mtGetter,
		TokenID:    tokenID,
	})
	return nil
}

func getBalances(addr std.Address) string {
	balances := avl.NewTree()
	var sb strings.Builder

	// NFT balances
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		balance, _ := nft.BalanceOf(addr)
		balances.Set("nft:"+key, balance)
		return false
	})

	// GRC20 balances
	grc20reg.Iterate(func(key string, tokenGetter grc20.TokenGetter) bool {
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		balances.Set("token:"+key, balance)
		return false
	})

	// GRC1155 token balances
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil {
			balances.Set("mt:"+key, balance)
		}
		return false
	})

	sb.WriteString("Balances:")
	balances.Iterate("", "", func(key string, value interface{}) bool { //should we display all even if the balance is 0?
		switch {
		case strings.HasPrefix(key, "nft:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT %s: %d ", symbol, balance))
		case strings.HasPrefix(key, "token:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token %s: %d ", symbol, balance))
		case strings.HasPrefix(key, "mt:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "mt:")
			info, exists := registeredMTs.Get(strings.TrimPrefix(key, "mt:"))
			if !exists {
				return errors.New(ErrMTInfoNotFound)
			}
			collectionPath := strings.TrimPrefix(symbol, std.PrevRealm().PkgPath()+".")
			sb.WriteString(ufmt.Sprintf("MultiToken %s.%s: %d ", 
				collectionPath, 
				info.(GRC1155TokenInfo).TokenID, 
				balance))
		}
		return false
	})

	return sb.String()
}

func GetMyBalances() string {
	return getBalances(std.GetOrigCaller())
}

func Render(path string) string {
	var sb strings.Builder

	sb.WriteString(md.H1("Token Hub"))
	sb.WriteString(md.Paragraph("A central registry for GRC721 NFTs, GRC20 tokens, and GRC1155 multi-tokens on gno.land"))

	// NFT section
	sb.WriteString(md.H2("Registered NFTs"))
	var nftItems []string
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
		if !ok {
			return false
		}
		nftItems = append(nftItems, ufmt.Sprintf("%s (%s)",
			metadata.Name(),
			metadata.Symbol()))
		return false
	})

	if len(nftItems) > 0 {
		sb.WriteString(md.BulletList(nftItems))
	} else {
		sb.WriteString(md.Italic("No NFTs registered yet"))
		sb.WriteString("\n")
	}

	// GRC20 token section
	sb.WriteString(md.H2("Registered Tokens"))
	var tokenItems []string
	grc20reg.Iterate(func(key string, tokenGetter grc20.TokenGetter) bool {
		token := tokenGetter()
		tokenItems = append(tokenItems, ufmt.Sprintf("%s (%s)",
			token.GetName(),
			token.GetSymbol()))
		return false
	})

	if len(tokenItems) > 0 {
		sb.WriteString(md.BulletList(tokenItems))
	} else {
		sb.WriteString(md.Italic("No tokens registered yet"))
		sb.WriteString("\n")
	}

	// GRC1155 section
	sb.WriteString(md.H2("Registered Multi-Tokens"))
	var mtItems []string
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mtItems = append(mtItems, ufmt.Sprintf("TokenID: %s (Collection: %s)", 
			info.TokenID, 
			strings.TrimPrefix(key, std.PrevRealm().PkgPath()+"."),
		))
		return false
	})

	if len(mtItems) > 0 {
		sb.WriteString(md.BulletList(mtItems))
	} else {
		sb.WriteString(md.Italic("No multi-tokens registered yet"))
		sb.WriteString("\n")
	}

	return sb.String()
}

