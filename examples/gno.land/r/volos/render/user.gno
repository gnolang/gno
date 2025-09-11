package render

import (
	"std"
	"strconv"
	"strings"

	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/matijamarjanovic/charts"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/volos/consts"
	"gno.land/p/volos/math"
	"gno.land/r/demo/grc20reg"
	"gno.land/r/sys/users"
	volos "gno.land/r/volos/core"
)

func renderUserPage(path string) string {
	userAddr := strings.TrimPrefix(path, "?user=")
	userAddr = strings.SplitN(userAddr, "&", 2)[0]
	if userAddr == "" {
		return md.Paragraph("No user address provided.")
	}

	var displayName string
	usAddr := std.Address(userAddr)
	userData := users.ResolveAddress(usAddr)
	if userData != nil && userData.Name() != "" {
		displayName = userData.Name()
	} else {
		displayName = userAddr
	}

	out := md.H1("👤 User Page")
	out += md.H4(displayName)

	markets := volos.GetUserMarkets(userAddr)
	loans := volos.GetUserLoans(userAddr)

	if len(markets) == 0 && len(loans) == 0 {
		out += md.Paragraph("🦌 This user has no positions or loans in any market.")
		return out
	}

	table := mdtable.Table{
		Headers: []string{"Market", "Supplied", "Borrowed", "Collateral", "Health Factor"},
	}

	anyRisk := false

	var marketLabels []string
	var chartSupplied []float64
	var chartBorrowed []float64

	for _, marketId := range markets {
		pos := volos.GetPosition(marketId, userAddr)
		market, params := volos.GetMarket(marketId)
		loanToken := grc20reg.Get(params.GetLoanToken())
		collateralToken := grc20reg.Get(params.GetCollateralToken())
		marketLabel := marketId
		if loanToken != nil && collateralToken != nil {
			marketLabel = loanToken.GetSymbol() + "/" + collateralToken.GetSymbol()
		}
		marketLink := md.Link(marketLabel, "?market="+marketId)

		suppliedAssets := math.ToAssetsUp(pos.SupplyShares, market.TotalSupplyAssets, market.TotalSupplyShares)
		borrowedAssets := math.ToAssetsUp(pos.BorrowShares, market.TotalBorrowAssets, market.TotalBorrowShares)
		collateralAssets := pos.Collateral

		supplied := formatTokenAmount(suppliedAssets, loanToken.GetDecimals())
		borrowed := formatTokenAmount(borrowedAssets, loanToken.GetDecimals())
		collateral := formatTokenAmount(collateralAssets, collateralToken.GetDecimals())

		hfRaw := u256.MustFromDecimal(volos.GetHealthFactor(marketId, userAddr))
		hf := formatPercentage(hfRaw)
		if hfRaw.Lt(consts.WAD) { // < 1.00
			anyRisk = true
		}
		hfString := hf

		threeWad := new(u256.Uint).Mul(consts.WAD, u256.NewUint(3))
		if hfRaw.Gt(threeWad) {
			hfString = "🔥 3.00+"
		}

		oneFiveWad := new(u256.Uint).Mul(consts.WAD, u256.NewUint(15))
		oneFiveWad = new(u256.Uint).Div(oneFiveWad, u256.NewUint(10))
		if hfRaw.Lt(oneFiveWad) {
			hfString = "⚠️ " + hf
		}

		table.Append([]string{marketLink, supplied, borrowed, collateral, hfString})

		marketLabels = append(marketLabels, marketLabel)
		if v, err := strconv.ParseFloat(supplied, 64); err == nil {
			chartSupplied = append(chartSupplied, v)
		} else {
			chartSupplied = append(chartSupplied, 0)
		}
		if v, err := strconv.ParseFloat(borrowed, 64); err == nil {
			chartBorrowed = append(chartBorrowed, v)
		} else {
			chartBorrowed = append(chartBorrowed, 0)
		}
	}

	if len(loans) > 0 {
		out += md.H2("📊 Loans by Token")
		loansTable := mdtable.Table{Headers: []string{"Token", "Amount"}}
		for i := 0; i < len(loans); i += 2 {
			tokenPath := loans[i]
			amountStr := loans[i+1]
			token := grc20reg.Get(tokenPath)
			label := tokenPath
			dec := 0
			if token != nil {
				label = token.GetSymbol()
				dec = token.GetDecimals()
			}
			amt := formatTokenAmount(u256.MustFromDecimal(amountStr), dec)
			loansTable.Append([]string{label, amt})
		}
		out += loansTable.String()
	}

	if len(markets) > 0 {
		out += md.H2("📈 Positions")
		out += table.String()
		out += "\n"
		if anyRisk {
			out += md.Paragraph("⚠️ " + md.Bold("Warning: One or more positions are at risk of liquidation!"))
		}

		if len(chartSupplied) > 0 {
			out += md.H3("Supplied per Market")
			out += charts.GenerateBarChart(chartSupplied, marketLabels, 30, "")
		}
		if len(chartBorrowed) > 0 {
			out += md.H3("Borrowed per Market")
			out += charts.GenerateBarChart(chartBorrowed, marketLabels, 30, "")
		}
	}

	return out
}
