package gnoxchange

import (
	"math"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/matijamarjanovic/tokenhub"
)

var statusEmojis = map[string]string{
	"open":      "🟢 Open",
	"fulfilled": "✅ Fulfilled",
	"expired":   "⏰ Expired",
	"cancelled": "❌ Cancelled",
}

func Render(path string) string {
	basePath := path
	if idx := strings.Index(path, "?"); idx != -1 {
		basePath = path[:idx]
	}

	switch basePath {
	case "ticketing":
		return renderTickets(path)
	case "ticketHistory":
		return renderTicketHistory(path)
	case "nftmarket":
		return renderNFTMarket(path)
	case "pools":
		return renderPools(path)
	default:
		return renderHome()
	}
}

func renderHome() string {
	var str string
	str += md.H1("GnoXchange")
	str += md.Paragraph("Welcome to GnoXchange - Decentralized token exchange system on Gno.land")

	str += md.H2("Navigation")
	links := []string{
		"[Liquidity Pools](/r/matijamarjanovic/gnoxchange:pools) - View and interact with liquidity pools", // TODO: change to gno.land
		"[Ticket System](/r/matijamarjanovic/gnoxchange:ticketing) - Create and manage swap tickets",       // TODO: change to gno.land
		"[NFT Market](/r/matijamarjanovic/gnoxchange:nftmarket) - Create and manage NFT tickets",           // TODO: change to gno.land
	}
	str += md.BulletList(links)

	str += md.H2("Quick Stats")
	stats := []string{
		ufmt.Sprintf("Total Pools: %d", GetAllPoolNamesCount()),
		ufmt.Sprintf("Open Tickets: %d", GetOpenTicketsCount()),
		ufmt.Sprintf("Total NFTs for sale: %d", GetOpenNFTTicketsCount()),
	}
	str += md.BulletList(stats)

	str += md.Paragraph("")
	str += md.Paragraph("[Ticket History](/r/matijamarjanovic/gnoxchange:ticketHistory) - View all tickets (including expired, fulfilled, and cancelled).")

	str += md.HorizontalRule()
	str += md.Paragraph("Start by exploring the available pools or creating a new swap ticket.")

	return str
}

func renderTickets(path string) string {
	var str string
	str += md.H2("Ticket System")
	str += md.Paragraph("View all open tickets for coin and token swaps.")

	page := GetAllOpenNonNFTTickets(path)
	if len(page.Items) == 0 {
		str += md.Blockquote("No open tickets available.")
		return str
	}

	str += md.H2("Open Tickets")

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		swapType := ""
		var amountInStr, amountOutStr string

		if ticket.AssetIn.Type == AssetTypeCoin {
			swapType = ticket.AssetIn.Denom + " → " + ticket.AssetOut.Token
			amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/1000000) + " " + ticket.AssetIn.Denom

			tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
			decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))
			amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()

		} else if ticket.AssetOut.Type == AssetTypeCoin {
			tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
			swapType = tokenIn.GetSymbol() + " → " + ticket.AssetOut.Denom

			decimalAdjustIn := math.Pow(10, float64(tokenIn.GetDecimals()))
			amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/decimalAdjustIn) + " " + tokenIn.GetSymbol()
			amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/1000000) + " " + ticket.AssetOut.Denom

		} else {
			tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
			tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
			swapType = tokenIn.GetSymbol() + " → " + tokenOut.GetSymbol()

			decimalAdjustIn := math.Pow(10, float64(tokenIn.GetDecimals()))
			decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))

			amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/decimalAdjustIn) + " " + tokenIn.GetSymbol()
			amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()
		}

		str += md.H3("#" + ticket.ID + " (" + swapType + ")")

		details := []string{
			"Publisher Offers: " + amountInStr,
			"Publisher Wants (min): " + amountOutStr,
			"Created At: " + ticket.CreatedAt.Format(time.RFC3339),
			"Expires In: " + formatDuration(ticket.ExpiresAt.Sub(time.Now())),
			"Status: " + statusEmojis[ticket.Status],
		}

		str += md.BulletList(details)
		str += md.HorizontalRule()
	}

	if page.TotalPages > 1 {
		str += "\n\nPages: " + page.Picker()
	}

	return str
}

func renderNFTMarket(path string) string {
	var str string
	str += md.H2("NFT Market")
	str += md.Paragraph("View all NFT tickets for sale.")

	page := GetOpenNFTTickets(path)
	if len(page.Items) == 0 {
		str += md.Blockquote("No NFTs available for sale.")
		return str
	}

	str += md.H2("NFT Tickets")

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		nft := tokenhub.GetNFT(ticket.AssetIn.NFTPath)
		if nft == nil {
			continue
		}

		parts := strings.Split(ticket.AssetIn.NFTPath, ".")
		collection := parts[len(parts)-2]
		tokenID := parts[len(parts)-1]

		var priceStr string
		if ticket.AssetOut.Type == AssetTypeCoin {
			priceStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/1000000) + " " + ticket.AssetOut.Denom
		} else {
			tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
			decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))
			priceStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()
		}

		str += md.H3("#" + ticket.ID + " (" + collection + " #" + tokenID + ")")

		details := []string{
			"Collection: " + collection,
			"Token ID: " + tokenID,
			"Price: " + priceStr,
			"Created At: " + ticket.CreatedAt.Format(time.RFC3339),
			"Expires In: " + formatDuration(ticket.ExpiresAt.Sub(time.Now())),
			"Status: " + statusEmojis[ticket.Status],
		}

		str += md.BulletList(details)
		str += md.HorizontalRule()
	}

	if page.TotalPages > 1 {
		str += "\n\nPages: " + page.Picker()
	}

	return str
}

func renderPools(path string) string {
	var str string
	str += md.H2("Available Pools")
	str += md.Paragraph("View and interact with liquidity pools.")

	page := GetPoolsPage(path)
	if len(page.Items) == 0 {
		str += md.Blockquote("No pools available.")
		return str
	}

	for _, item := range page.Items {
		poolKey := item.Key
		tokenA, tokenB, reserveA, reserveB, err := GetPoolInfo(poolKey)
		if err != nil {
			continue
		}

		tokenAInstance := tokenhub.GetToken(tokenA)
		tokenBInstance := tokenhub.GetToken(tokenB)

		decimalAdjustA := math.Pow(10, float64(tokenAInstance.GetDecimals()))
		decimalAdjustB := math.Pow(10, float64(tokenBInstance.GetDecimals()))

		tokenAName := tokenAInstance.GetName()
		tokenBName := tokenBInstance.GetName()
		tokenASymbol := tokenAInstance.GetSymbol()
		tokenBSymbol := tokenBInstance.GetSymbol()

		str += md.H3(tokenAName + " (" + tokenASymbol + ") <-> " + tokenBName + " (" + tokenBSymbol + ")")
		details := []string{
			"First token: " + tokenASymbol + " - **" + ufmt.Sprintf("%f", float64(reserveA)/decimalAdjustA) + "**" +
				" (1 " + tokenASymbol + " ≈ " + ufmt.Sprintf("%f", float64(reserveB)/float64(reserveA)) + " " + tokenBSymbol + ")",
			"Second token: " + tokenBSymbol + " - **" + ufmt.Sprintf("%f", float64(reserveB)/decimalAdjustB) + "**" +
				" (1 " + tokenBSymbol + " ≈ " + ufmt.Sprintf("%f", float64(reserveA)/float64(reserveB)) + " " + tokenASymbol + ")",
		}
		str += md.BulletList(details)
		str += md.HorizontalRule()
	}

	if page.TotalPages > 1 {
		str += "\n\nPages: " + page.Picker()
	}

	return str
}

func renderTicketHistory(path string) string {
	var str string
	str += md.H2("Ticket History")
	str += md.Paragraph("View all tickets (including expired, fulfilled, and cancelled).")

	page := GetAllTickets(path)
	if len(page.Items) == 0 {
		str += md.Blockquote("No tickets available.")
		return str
	}

	str += md.H2("All Tickets")

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		if ticket.AssetIn.Type == AssetTypeNFT {
			nft := tokenhub.GetNFT(ticket.AssetIn.NFTPath)
			if nft == nil {
				continue
			}

			parts := strings.Split(ticket.AssetIn.NFTPath, ".")
			collection := parts[len(parts)-2]
			tokenID := parts[len(parts)-1]

			var priceStr string
			if ticket.AssetOut.Type == AssetTypeCoin {
				priceStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/1000000) + " " + ticket.AssetOut.Denom
			} else {
				tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
				decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))
				priceStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()
			}

			str += md.H3("#" + ticket.ID + " (NFT: " + collection + " #" + tokenID + ")")

			details := []string{
				"Type: NFT Sale",
				"Collection: " + collection,
				"Token ID: " + tokenID,
				"Price: " + priceStr,
				"Created At: " + ticket.CreatedAt.Format(time.RFC3339),
				"Expires In: " + formatDuration(ticket.ExpiresAt.Sub(time.Now())),
				"Status: " + statusEmojis[ticket.Status],
			}

			str += md.BulletList(details)

		} else {
			swapType := ""
			var amountInStr, amountOutStr string

			if ticket.AssetIn.Type == AssetTypeCoin {
				swapType = ticket.AssetIn.Denom + " → " + ticket.AssetOut.Token
				amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/1000000) + " " + ticket.AssetIn.Denom

				tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
				decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))
				amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()

			} else if ticket.AssetOut.Type == AssetTypeCoin {
				tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
				swapType = tokenIn.GetSymbol() + " → " + ticket.AssetOut.Denom

				decimalAdjustIn := math.Pow(10, float64(tokenIn.GetDecimals()))
				amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/decimalAdjustIn) + " " + tokenIn.GetSymbol()
				amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/1000000) + " " + ticket.AssetOut.Denom

			} else {
				tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
				tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
				swapType = tokenIn.GetSymbol() + " → " + tokenOut.GetSymbol()

				decimalAdjustIn := math.Pow(10, float64(tokenIn.GetDecimals()))
				decimalAdjustOut := math.Pow(10, float64(tokenOut.GetDecimals()))

				amountInStr = ufmt.Sprintf("%f", float64(ticket.AmountIn)/decimalAdjustIn) + " " + tokenIn.GetSymbol()
				amountOutStr = ufmt.Sprintf("%f", float64(ticket.MinAmountOut)/decimalAdjustOut) + " " + tokenOut.GetSymbol()
			}

			str += md.H3("#" + ticket.ID + " (" + swapType + ")")

			details := []string{
				"Type: Token Swap",
				"Publisher Offers: " + amountInStr,
				"Publisher Wants (min): " + amountOutStr,
				"Created At: " + ticket.CreatedAt.Format(time.RFC3339),
				"Expires In: " + formatDuration(ticket.ExpiresAt.Sub(time.Now())),
				"Status: " + statusEmojis[ticket.Status],
			}

			str += md.BulletList(details)
		}

		str += md.HorizontalRule()
	}

	if page.TotalPages > 1 {
		str += "\n\nPages: " + page.Picker()
	}

	return str
}

func formatAsset(asset Asset, amount uint64) string {
	if asset.Type == AssetTypeCoin {
		return ufmt.Sprintf("**%d** %s", amount, asset.Denom)
	}
	return ufmt.Sprintf("**%d** %s", amount, asset.Token)
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "Expired"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return ufmt.Sprintf("%dd %dh", days, hours)
	}

	if hours > 0 {
		return ufmt.Sprintf("%dh %dm", hours, minutes)
	}

	return ufmt.Sprintf("%dm", minutes)
}

func getTokenName(tokenPath string) string {
	parts := strings.Split(tokenPath, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return tokenPath
}
