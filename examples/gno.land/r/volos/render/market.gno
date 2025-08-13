package render

import (
	"std"
	"net/url"
	"strings"
	"strconv"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avl"
	"gno.land/r/demo/grc20reg"
	"gno.land/r/sys/users"
	volos "gno.land/r/volos/core"
)


func renderMarketPage(path string) string {
	marketId := strings.TrimPrefix(path, "?market=")
	marketId = strings.SplitN(marketId, "&", 2)[0]
	if marketId == "" {
		return md.Paragraph("No market ID provided.")
	}
	
	marketId, _ = url.QueryUnescape(marketId)

	market, params := volos.GetMarket(marketId)
	
	loanPath := params.GetLoanToken()
	collateralPath := params.GetCollateralToken()
	loanToken := grc20reg.Get(loanPath)
	collateralToken := grc20reg.Get(collateralPath)
	
	loanSymbol := loanPath
	collateralSymbol := collateralPath
	if loanToken != nil {
		loanSymbol = loanToken.GetSymbol()
	}
	if collateralToken != nil {
		collateralSymbol = collateralToken.GetSymbol()
	}

	out := md.H1("🏦 " + loanSymbol + "/" + collateralSymbol + " Market")
	out += md.Paragraph("**Market ID:** " + md.InlineCode(marketId))
	
	out += md.H2("📊 Market Overview")
	overviewTable := mdtable.Table{
		Headers: []string{"Parameter", "Value"},
	}
	overviewTable.Append([]string{"Pool Path", md.InlineCode(params.PoolPath)})
	overviewTable.Append([]string{"Loan Token", loanSymbol + " (" + md.InlineCode(loanPath) + ")"})
	overviewTable.Append([]string{"Collateral Token", collateralSymbol + " (" + md.InlineCode(collateralPath) + ")"})
	overviewTable.Append([]string{"Interest Rate Model", md.InlineCode(params.IRM)})
	overviewTable.Append([]string{"Liquidation LTV", params.LLTV.ToString() + "%"})
	overviewTable.Append([]string{"Market Fee", market.Fee.ToString()})
	out += overviewTable.String()
	
	out += md.H2("💰 Market Statistics & Rates")
	statsTable := mdtable.Table{
		Headers: []string{"Metric", "Value", "Rate Type", "APR"},
	}
	
	borrowAPR := volos.GetBorrowAPR(marketId)
	supplyAPR := volos.GetSupplyAPR(marketId)
	
	statsTable.Append([]string{
		"Total Supply Assets", formatMoney(parseFloat(volos.GetMarketTotalSupplyAssets(marketId))),
		"Borrow Rate", formatRate(borrowAPR) + "%",
	})
	statsTable.Append([]string{
		"Total Borrow Assets", formatMoney(parseFloat(volos.GetMarketTotalBorrowAssets(marketId))),
		"Supply Rate", formatRate(supplyAPR) + "%",
	})
	statsTable.Append([]string{
		"Utilization Rate", calculateUtilization(marketId) + "%",
		"Last Update", formatTimestamp(market.LastUpdate),
	})
	out += statsTable.String()
	
	out += md.H2("💱 Current Price")
	price := volos.GetMarketPrice(marketId)
	out += md.Paragraph("**Oracle Price:** " + price + " " + loanSymbol + " per " + collateralSymbol)
	out += md.Blockquote("Price sourced from Gnoswap pool: " + params.PoolPath)
	
	out += md.H2("👥 Top Positions")
	positions := volos.GetPositionList(marketId)
	if len(positions) == 0 {
		out += md.Paragraph("🦌 No positions in this market yet.")
	} else {
		positionsTree := avl.NewTree()
		for _, userAddr := range positions {
			positionsTree.Set(userAddr, userAddr)
		}
		
		positionPager := pager.NewPager(rotree.Wrap(positionsTree, nil), 5, false)
		positionPage := positionPager.MustGetPageByPath(path)
		
		positionsTable := mdtable.Table{
			Headers: []string{"User", "Supplied", "Borrowed", "Health Factor"},
		}
		
		for _, item := range positionPage.Items {
			userAddr := item.Key
			pos := volos.GetPosition(marketId, userAddr)
			if pos.SupplyShares.IsZero() && pos.BorrowShares.IsZero() && pos.Collateral.IsZero() {
				continue
			}
			
			usAddr := std.Address(userAddr)
			userData := users.ResolveAddress(usAddr)
			displayName := userAddr
			if userData != nil && userData.Name() != "" {
				displayName = userData.Name()
			}
			
			userLink := md.Link(displayName, "?user="+userAddr)
			supplied := pos.SupplyShares.ToString()
			borrowed := pos.BorrowShares.ToString()
			healthFactor := volos.GetHealthFactor(marketId, userAddr)
			
			hfFloat, _ := strconv.ParseFloat(healthFactor, 64)
			if hfFloat < 1.0 {
				healthFactor = "⚠️ " + healthFactor
			}
			
			positionsTable.Append([]string{userLink, supplied, borrowed, healthFactor})
		}
		
		if len(positionsTable.Rows) > 0 {
			out += positionsTable.String()
			out += md.HorizontalRule()
			picker := positionPage.Picker(path)
			if picker != "" {
				out += md.Paragraph(picker)
			}
		} else {
			out += md.Paragraph("🦌 No active positions in this market.")
		}
	}
	
	return out
}

