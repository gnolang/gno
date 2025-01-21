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
	var sb strings.Builder

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
