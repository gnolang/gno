package tokenhub

import (
	"strings"
	"strconv"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/demo/grc20reg"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
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
		out += md.Link("Back to home", "http://localhost:8888/r/matijamarjanovic/tokenhub")
	}
	

	// navigation footer if not on home page
	if path != "" {
		out += "\n"
		out += md.HorizontalRule()
		out += md.Link("Back to home", "http://localhost:8888/r/matijamarjanovic/tokenhub")
	}

	return out
}

func RenderHome() string {
	out := md.H1("Token Hub")
	out += md.Paragraph("A central registry for GRC721 NFTs, GRC20 tokens, and GRC1155 multi-tokens on gno.land")

	links := []string{
		md.Link("GRC20 Tokens", "http://localhost:8888/r/matijamarjanovic/tokenhub:token_1"),
		md.Link("GRC721 NFTs", "http://localhost:8888/r/matijamarjanovic/tokenhub:nft_1"),
		md.Link("GRC1155 Multi-Tokens", "http://localhost:8888/r/matijamarjanovic/tokenhub:mt_1"),
	}
	out +=md.BulletList(links)

	return out
}

func RenderToken(path string) string {
	out := md.H1("GRC20 Tokens")
	var tokenItems []string

	page := 1
	if strings.HasPrefix(path, token + "_") { // get page number from path
		pageStr := strings.TrimPrefix(path, token + "_")
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
		out += md.BulletList(tokenItems)

		out += "\n"

		if page > 1 {
			out += md.HorizontalRule()
			prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:token_%d", page-1)
			out += md.Link("Previous", prevPage)
			out += " | "
		}

		if len(tokenItems) == pageSize && lastKey != "" {
			nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:token_%d", page+1)
			out += md.Link("Next", nextPage)
		}
	} else {
		out += md.Italic("No tokens registered yet")
		out += "\n"
	}

	return out
}

func RenderNFT(path string) string {
	var out string
	out = md.H1("GRC721 NFTs")
	var nftItems []string

	page := 1
	if strings.HasPrefix(path, nft + "_") {
		pageStr := strings.TrimPrefix(path, nft + "_")
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
			out += md.BulletList(nftItems)
			out += "\n"
			out += md.HorizontalRule()

			if page > 1 {
				prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:nft_%d", page-1)
				out += md.Link("Previous", prevPage)
				out += " | "
			}

			if len(nftItems) == pageSize {
				nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:nft_%d", page+1)
				out += md.Link("Next", nextPage)
			}
		} else {
			out += md.Italic("No NFTs registered yet")
			out += "\n"
		}

		return out
}

func RenderMT(path string) string {
	var out string
	out = md.H1("GRC1155 Multi-Tokens")
	var mtItems []string

		page := 1
		if strings.HasPrefix(path, mt + "_") {
			pageStr := strings.TrimPrefix(path, mt + "_")
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
			out += md.BulletList(mtItems)
			out += "\n"
			out += md.HorizontalRule()

			if page > 1 {
				prevPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:mt_%d", page-1)
				out += md.Link("Previous", prevPage)
				out += " | "
			}

			if len(mtItems) == pageSize {
				nextPage := ufmt.Sprintf("http://localhost:8888/r/matijamarjanovic/tokenhub:mt_%d", page+1)
				out += md.Link("Next", nextPage)
			}
		} else {
		out += md.Italic("No multi-tokens registered yet")
		out += "\n"
	}

	return out
}
