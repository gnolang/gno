package tokenhub

import (
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/demo/fqname"
	"gno.land/r/demo/grc20reg"
)

const (
	token = "token" // grc20
	nft   = "nft"   // grc721
	mt    = "mt"    // grc1155
)

func Render(path string) string {
	var out string

	switch {
	case path == "":
		out = renderHome()

	case strings.HasPrefix(path, token):
		out = renderToken(path)

	case strings.HasPrefix(path, nft):
		out = renderNFT(path)

	case strings.HasPrefix(path, mt):
		out = renderMT(path)
	}

	return out
}

func renderHome() string {
	out := md.H1("Token Hub")
	out += md.Paragraph("Token Hub provides listings of all existing token types on Gno.land - GRC20 tokens, GRC721 NFTs, and GRC1155 multi-tokens. You can browse these listings to find available tokens, check balances, and access token metadata. If you're developing wallets or interfaces, you can query this registry to display token information to your users.")

	links := []string{
		"[GRC20 Tokens](/r/matijamarjanovic/tokenhub:tokens)",
		"[GRC721 NFTs](/r/matijamarjanovic/tokenhub:nfts)",
		"[GRC1155 Multi-Tokens](/r/matijamarjanovic/tokenhub:mts)",
	}
	out += md.BulletList(links)

	out += md.H2("How to Register Your Tokens")
	out += md.Paragraph("You can register your tokens with the following import and function calls:")

	registerCode := `// Import the tokenhub package
import (
	"gno.land/r/matijamarjanovic/tokenhub"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc1155"
)

// Create and register a GRC20 token
// Create a new token with name, symbol, and decimal precision
myToken, myLedger := grc20.NewToken("My Token", "MTK", 6)
// Register the token with TokenHub
myTokenPath := tokenhub.RegisterToken(myToken.Getter(), "my_token")

// Create and register a GRC721 NFT
// Create a new NFT collection with name and symbol
myNFT := grc721.NewBasicNFT("My NFT Collection", "MNFT")
// Mint an NFT (TokenID = "1") to your address
myNFT.Mint("g1your_address_here", "1")
// Register the NFT collection with TokenHub
err := tokenhub.RegisterNFT(myNFT.Getter(), "my_collection", "1")

// Create and register a GRC1155 multi-token
// Create a new multi-token with base URI
myMultiToken := grc1155.NewBasicGRC1155Token("https://metadata.example.com/")
// Mint a token (TokenID = "123") to your address with quantity 10
myMultiToken.SafeMint("g1your_address_here", "123", 10)
// Register the multi-token with TokenHub
err := tokenhub.RegisterMultiToken(myMultiToken.Getter(), "123")`

	out += md.LanguageCodeBlock("go", registerCode)

	out += "\n"
	out += md.H2("Querying Token Information")
	out += md.Paragraph("You can query token information and balances using functions like:")

	queryCode := `// Get all registered tokens
allTokens := tokenhub.GetAllTokens()

// Get token balances for a user
balances := tokenhub.GetUserTokenBalances("g1...")

// Get non-zero token balances
nonZeroBalances := tokenhub.GetUserTokenBalancesNonZero("g1...")`

	out += md.LanguageCodeBlock("go", queryCode)

	return out
}

func renderToken(path string) string {
	out := md.H1("GRC20 Tokens")
	var tokenItems []string

	tokenPager := pager.NewPager(grc20reg.GetRegistry(), pageSize, false)
	page := tokenPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		tokenGetter := item.Value.(grc20.TokenGetter)
		token := tokenGetter()
		pkgPath, _ := fqname.Parse(item.Key)
		linkURL := strings.ReplaceAll(pkgPath, "gno.land/r/matijamarjanovic/tokenhub.", "")
		linkURL = strings.ReplaceAll(linkURL, "gno.land", "")
		tokenItems = append(tokenItems, ufmt.Sprintf("%s (%s), registered at key: %s",
			md.Link(token.GetName(), linkURL),
			md.InlineCode(token.GetSymbol()),
			md.InlineCode(item.Key)))
	}

	out += renderItemsList(tokenItems, page, "No tokens registered yet")
	return out
}

func renderNFT(path string) string {
	out := md.H1("GRC721 NFTs")

	var nftItems []string
	nftPager := pager.NewPager(registeredNFTs, pageSize, false)
	page := nftPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		nftGetter := item.Value.(grc721.NFTGetter)
		nft := nftGetter()
		metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
		if !ok {
			continue
		}

		linkURL := strings.ReplaceAll(item.Key, "gno.land", "")
		parts := strings.Split(linkURL, ".")
		if len(parts) > 2 {
			linkURL = strings.Join(parts[:len(parts)-2], ".")
		}
		
		nftItems = append(nftItems, ufmt.Sprintf("%s (%s), registered at [%s]",
			md.Link(metadata.Name(), linkURL),
			md.InlineCode(metadata.Symbol()),
			md.InlineCode(item.Key)))
	}

	out += renderItemsList(nftItems, page, "No NFTs registered yet")
	return out
}

func renderMT(path string) string {
	out := md.H1("GRC1155 Multi-Tokens")
	var mtItems []string

	mtPager := pager.NewPager(registeredMTs, pageSize, false)
	page := mtPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		info := item.Value.(GRC1155TokenInfo)
		linkURL := strings.ReplaceAll(item.Key, "gno.land", "")
		parts := strings.Split(linkURL, ".")
		if len(parts) > 1 {
			linkURL = strings.Join(parts[:len(parts)-1], ".")
		}

		mtItems = append(mtItems, ufmt.Sprintf("%s: %s, registered at [%s]",
			md.Bold("TokenID"),
			md.Link(md.InlineCode(info.TokenID), linkURL),
			md.InlineCode(item.Key)))
	}

	out += renderItemsList(mtItems, page, "No multi-tokens registered yet")
	return out
}

func renderItemsList(items []string, page *pager.Page, emptyMessage string) string {
	var out string
	if len(items) == 0 {
		out += md.Italic(emptyMessage)
		out += "\n"
		return out
	}

	out += md.BulletList(items)
	out += "\n"
	out += md.HorizontalRule()

	picker := page.Picker(page.Pager.PageQueryParam)
	if picker != "" {
		out += md.Paragraph(picker)
	}

	return out
}
