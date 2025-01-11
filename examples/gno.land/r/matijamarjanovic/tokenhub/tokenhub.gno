package tokenhub

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/demo/grc20reg"
	"gno.land/p/demo/fqname"
)

var (
	registeredNFTs = avl.NewTree() // rlmPath[.slug] -> NFTGetter

	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound          = errors.New("NFT not found")
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
	metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
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

	sb.WriteString("Balances:")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		if strings.HasPrefix(key, "nft:") {
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT %s: %d ", symbol, balance))
		} else {
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token %s: %d ", symbol, balance))
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
	sb.WriteString(md.Paragraph("A central registry for GRC721 NFTs and GRC20 tokens on gno.land"))

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

	return sb.String()
}

