package tokenhub

import (
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/demo/grc20reg"
)

func Render(path string) string {
	var out string

	switch {
	case path == "":
		out = RenderHome()

	case strings.HasPrefix(path, token):
		out = RenderToken(path)

	case strings.HasPrefix(path, nft):
		out = RenderNFT(path)

	case strings.HasPrefix(path, mt):
		out = RenderMT(path)

	default:
		out = md.H1("404 Not Found")
		out += md.Paragraph("The requested page does not exist.")
		out += md.Link("Back to home", "http://localhost:8888/r/demo/tokenhub")
	}

	// navigation footer if not on home page
	if path != "" {
		out += "\n"
		out += md.Link("Back to home", "http://localhost:8888/r/demo/tokenhub")
	}

	return out
}

func RenderHome() string {
	out := md.H1("Token Hub")
	out += md.Paragraph("A central registry for GRC721 NFTs, GRC20 tokens, and GRC1155 multi-tokens on gno.land")

	links := []string{
		md.Link("GRC20 Tokens", "http://localhost:8888/r/demo/tokenhub:tokens"),
		md.Link("GRC721 NFTs", "http://localhost:8888/r/demo/tokenhub:nfts"),
		md.Link("GRC1155 Multi-Tokens", "http://localhost:8888/r/demo/tokenhub:mts"),
	}
	out += md.BulletList(links)

	return out
}

func RenderToken(path string) string {
	out := md.H1("GRC20 Tokens")
	var tokenItems []string

	tokenPager := pager.NewPager(grc20reg.GetRegistry(), pageSize, false)

	page := tokenPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		tokenGetter := item.Value.(grc20.TokenGetter)
		token := tokenGetter()
		tokenItems = append(tokenItems, ufmt.Sprintf("%s (%s) [%s]",
			md.Bold(token.GetName()),
			md.InlineCode(token.GetSymbol()),
			md.InlineCode(item.Key)))
	}

	if len(tokenItems) > 0 {
		out += md.BulletList(tokenItems)
		out += "\n"

		if picker := page.Picker(); picker != "" {
			out += md.HorizontalRule()
			out += md.Paragraph(picker)
		} else {
			out += md.HorizontalRule()
		}
	} else {
		out += md.Italic("No tokens registered yet")
		out += "\n"
	}

	return out
}

func RenderNFT(path string) string {
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

		nftItems = append(nftItems, ufmt.Sprintf("%s (%s) [%s]",
			md.Bold(metadata.Name()),
			md.InlineCode(metadata.Symbol()),
			md.InlineCode(item.Key)))
	}

	if len(nftItems) > 0 {
		out += md.BulletList(nftItems)
		out += "\n"

		if picker := page.Picker(); picker != "" {
			out += md.HorizontalRule()
			out += md.Paragraph(picker)
		} else {
			out += md.HorizontalRule()
		}
	} else {
		out += md.Italic("No NFTs registered yet")
		out += "\n"
	}

	return out
}

func RenderMT(path string) string {
	out := md.H1("GRC1155 Multi-Tokens")
	var mtItems []string

	mtPager := pager.NewPager(registeredMTs, pageSize, false)

	page := mtPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		info := item.Value.(GRC1155TokenInfo)
		mtItems = append(mtItems, ufmt.Sprintf("%s: %s [%s]",
			md.Bold("TokenID"),
			md.InlineCode(info.TokenID),
			md.InlineCode(item.Key)))
	}

	if len(mtItems) > 0 {
		out += md.BulletList(mtItems)
		out += "\n"

		if picker := page.Picker(); picker != "" {
			out += md.HorizontalRule()
			out += md.Paragraph(picker)
		} else {
			out += md.HorizontalRule()
		}
	} else {
		out += md.Italic("No multi-tokens registered yet")
		out += "\n"
	}

	return out
}
