package tokenhub

import (
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/demo/grc20reg"
	"gno.land/r/demo/users"
)

type GRC1155TokenInfo struct {
	Collection grc1155.MultiTokenGetter
	TokenID    string
}

var (
	registeredNFTs = avl.NewTree() // rlmPath[.slug] -> NFTGetter
	registeredMTs  = avl.NewTree() // rlmPath[.slug] -> GRC1155TokenInfo

	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound          = errors.New("NFT not found")
	ErrMTAlreadyRegistered  = errors.New("Multi-token already registered")
	ErrMTNotFound           = errors.New("Multi-token not found")
	ErrMTInfoNotFound       = errors.New("Multi-token info not found")

	pageSize = 2 // tokens per page
	test     = ""
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

// for grc1155 is a little bit different, since the one collection can have multiple tokens, this is why don't use slugs but tokenIDs,
// and the collection is saved as a part of the :value in the node
func RegisterGRC1155Token(mtGetter grc1155.MultiTokenGetter, tokenID string) error {
	rlmPath := std.PrevRealm().PkgPath()

	key := fqname.Construct(rlmPath, tokenID)

	if registeredMTs.Has(key) {
		return ErrMTAlreadyRegistered
	}

	registeredMTs.Set(key, GRC1155TokenInfo{
		Collection: mtGetter,
		TokenID:    tokenID,
	})
	return nil
}

func GetMyBalancesAll() string {
	return getBalances(std.GetOrigCaller())
}

func GetMyBalances() string {
	return getNonZeroBalances(std.GetOrigCaller())
}

func GetGRC20(key string) grc20.TokenGetter {
	return grc20reg.Get(key)
}

func MustGetGRC20(key string) grc20.TokenGetter {
	return grc20reg.MustGet(key)
}

func GetGRC721(key string) grc721.NFTGetter {
	nftGetter, ok := registeredNFTs.Get(key)
	if !ok {
		return nil
	}
	return nftGetter.(grc721.NFTGetter)
}

func MustGetGRC721(key string) grc721.NFTGetter {
	nftGetter := GetGRC721(key)
	if nftGetter == nil {
		panic("unknown NFT: " + key)
	}
	return nftGetter
}

func GetGRC1155(key string) (GRC1155TokenInfo, bool) {
	info, ok := registeredMTs.Get(key)
	if !ok {
		return GRC1155TokenInfo{}, false
	}
	return info.(GRC1155TokenInfo), true
}

func MustGetGRC1155(key string) GRC1155TokenInfo {
	info, ok := GetGRC1155(key)
	if !ok {
		panic("unknown multi-token: " + key)
	}
	return info
}

func GetBalancesAll(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}
	return getBalances(addr)
}

func GetBalances(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}
	return getNonZeroBalances(addr)
}

func Render(path string) string {
	var sb strings.Builder
	sb.WriteString(test)

	switch {
	case path == "":
		// home page
		sb.WriteString(md.H1("Token Hub"))
		sb.WriteString(md.Paragraph("A central registry for GRC721 NFTs, GRC20 tokens, and GRC1155 multi-tokens on gno.land"))

		links := []string{
			md.Link("GRC20 Tokens", "http://localhost:8888/r/matijamarjanovic/tokenhub:token_1"),
			md.Link("GRC721 NFTs", "http://localhost:8888/r/matijamarjanovic/tokenhub:nft_1"),
			md.Link("GRC1155 Multi-Tokens", "http://localhost:8888/r/matijamarjanovic/tokenhub:mt_1"),
		}
		sb.WriteString(md.BulletList(links))

	case strings.HasPrefix(path, "token"): // grc20
		sb.WriteString(md.H1("GRC20 Tokens"))
		var tokenItems []string

		page := 1
		if strings.HasPrefix(path, "token_") { // get page number from path
			pageStr := strings.TrimPrefix(path, "token_")
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		startKey := ""
		if page > 1 {
			lastKey := grc20reg.IterateN("", (page-1)*pageSize, func(key string, tokenGetter grc20.TokenGetter) bool { // start position
				return false
			})
			startKey = lastKey
		}

		lastKey := grc20reg.IterateN(startKey, pageSize, func(key string, tokenGetter grc20.TokenGetter) bool { // get current page tokens
			token := tokenGetter()
			tokenItems = append(tokenItems, ufmt.Sprintf("%s (%s) [%s]",
				md.Bold(token.GetName()),
				md.InlineCode(token.GetSymbol()),
				md.InlineCode(key)))
			return false
		})

		if len(tokenItems) > 0 {
			sb.WriteString(md.BulletList(tokenItems))

			sb.WriteString("\n")
			sb.WriteString(md.HorizontalRule())

			if page > 1 {
				prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:token_%d", page-1)
				sb.WriteString(md.Link("Previous", prevPage))
				sb.WriteString(" | ")
			}

			if len(tokenItems) == pageSize && lastKey != "" {
				nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:token_%d", page+1)
				sb.WriteString(md.Link("Next", nextPage))
			}
		} else {
			sb.WriteString(md.Italic("No tokens registered yet"))
			sb.WriteString("\n")
		}

	case strings.HasPrefix(path, "nft"): // grc721
		sb.WriteString(md.H1("GRC721 NFTs"))
		var nftItems []string

		page := 1
		if strings.HasPrefix(path, "nft_") {
			pageStr := strings.TrimPrefix(path, "nft_")
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		startIdx := (page - 1) * pageSize
		count := 0
		currentIdx := 0

		registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
			if currentIdx < startIdx {
				currentIdx++
				return false
			}
			if count >= pageSize {
				return true
			}

			nftGetter := value.(grc721.NFTGetter)
			nft := nftGetter()
			metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
			if !ok {
				return false
			}

			nftItems = append(nftItems, ufmt.Sprintf("%s (%s) [%s]",
				md.Bold(metadata.Name()),
				md.InlineCode(metadata.Symbol()),
				md.InlineCode(key)))

			count++
			currentIdx++
			return false
		})

		if len(nftItems) > 0 {
			sb.WriteString(md.BulletList(nftItems))
			sb.WriteString("\n")
			sb.WriteString(md.HorizontalRule())

			if page > 1 {
				prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:nft_%d", page-1)
				sb.WriteString(md.Link("Previous", prevPage))
				sb.WriteString(" | ")
			}

			if len(nftItems) == pageSize {
				nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:nft_%d", page+1)
				sb.WriteString(md.Link("Next", nextPage))
			}
		} else {
			sb.WriteString(md.Italic("No NFTs registered yet"))
			sb.WriteString("\n")
		}

	case strings.HasPrefix(path, "mt"): // grc1155
		sb.WriteString(md.H1("GRC1155 Multi-Tokens"))
		var mtItems []string

		page := 1
		if strings.HasPrefix(path, "mt_") {
			pageStr := strings.TrimPrefix(path, "mt_")
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		startIdx := (page - 1) * pageSize
		count := 0
		currentIdx := 0

		registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
			if currentIdx < startIdx {
				currentIdx++
				return false
			}
			if count >= pageSize {
				return true
			}

			info := value.(GRC1155TokenInfo)
			mtItems = append(mtItems, ufmt.Sprintf("%s: %s [%s]",
				md.Bold("TokenID"),
				md.InlineCode(info.TokenID),
				md.InlineCode(key)))

			count++
			currentIdx++
			return false
		})

		if len(mtItems) > 0 {
			sb.WriteString(md.BulletList(mtItems))
			sb.WriteString("\n")
			sb.WriteString(md.HorizontalRule())

			if page > 1 {
				prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:mt_%d", page-1)
				sb.WriteString(md.Link("Previous", prevPage))
				sb.WriteString(" | ")
			}

			if len(mtItems) == pageSize {
				nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:mt_%d", page+1)
				sb.WriteString(md.Link("Next", nextPage))
			}
		} else {
			sb.WriteString(md.Italic("No multi-tokens registered yet"))
			sb.WriteString("\n")
		}

	default:
		sb.WriteString(md.H1("404 Not Found"))
		sb.WriteString(md.Paragraph("The requested page does not exist."))
		sb.WriteString(md.Link("Back to home", "http://localhost:8888/r/matijamarjanovic/tokenhub"))
	}

	// navigation footer if not on home page
	if path != "" {
		sb.WriteString("\n")
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.Link("Back to home", "http://localhost:8888/r/matijamarjanovic/tokenhub"))
	}

	return sb.String()
}

func getBalances(addr std.Address) string {
	if addr == "" {
		panic("invalid address")
	}
	balances := avl.NewTree()
	var sb strings.Builder

	// GRC721
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		balance, _ := nft.BalanceOf(addr)
		balances.Set("nft:"+key, balance)
		return false
	})

	// GRC20
	grc20reg.Iterate(func(key string, tokenGetter grc20.TokenGetter) bool {
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		balances.Set("token:"+key, balance)
		return false
	})

	// GRC1155
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
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s:%d,", symbol, balance))
			sb.WriteString(ufmt.Sprintf("Token:%s:%d,", symbol, balance))
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

func getNonZeroBalances(addr std.Address) string {
	if addr == "" {
		panic("invalid address")
	}
	balances := avl.NewTree()
	var sb strings.Builder

	// GRC721
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		balance, _ := nft.BalanceOf(addr)
		if balance > 0 {
			balances.Set("nft:"+key, balance)
		}
		return false
	})

	// GRC20
	grc20reg.Iterate(func(key string, tokenGetter grc20.TokenGetter) bool {
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		if balance > 0 {
			balances.Set("token:"+key, balance)
		}
		return false
	})

	// GRC1155
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
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s:%d,", symbol, balance))
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
