package render

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/matijamarjanovic/charts"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
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

	markets := volos.GetUserMarkets(userAddr)
	loans := volos.GetUserLoans(userAddr)

	if len(markets) == 0 && len(loans) == 0 {
		out += md.Paragraph("🦌 This user has no positions or loans in any market.")
		return out
	}

	totalSupplied := 0.0
	totalBorrowed := 0.0

	table := mdtable.Table{
		Headers: []string{"Market", "Supplied", "Borrowed", "Collateral", "Health Factor"},
	}

	anyRisk := false

	for _, marketId := range markets {
		pos := volos.GetPosition(marketId, userAddr)
		_, params := volos.GetMarket(marketId)
		loanToken := grc20reg.Get(params.GetLoanToken())
		collateralToken := grc20reg.Get(params.GetCollateralToken())
		marketLabel := marketId
		if loanToken != nil && collateralToken != nil {
			marketLabel = loanToken.GetSymbol() + "/" + collateralToken.GetSymbol()
		}
		marketLink := md.Link(marketLabel, "?market="+marketId)
		supplied := pos.SupplyShares.ToString()
		borrowed := pos.BorrowShares.ToString()
		collateral := pos.Collateral.ToString()
		hf := volos.GetHealthFactor(marketId, userAddr)

		suppliedF, _ := strconv.ParseFloat(supplied, 64)
		borrowedF, _ := strconv.ParseFloat(borrowed, 64)
		hfF, _ := strconv.ParseFloat(hf, 64)
		totalSupplied += suppliedF
		totalBorrowed += borrowedF
		if hfF < 1.0 {
			anyRisk = true
		}
		table.Append([]string{marketLink, supplied, borrowed, collateral, hf})
	}

	if len(markets) > 0 {
		out += md.H2("📈 Positions")
		out += table.String()
		out += "\n"
		out += md.Blockquote("These fiat values are mocked: 1 token denom = $1.")
		out += "\n"
		out += md.Paragraph("💰 **Total Supplied:** $" + formatMoney(totalSupplied))
		out += md.Paragraph("💰 **Total Borrowed:** $" + formatMoney(totalBorrowed))
		if anyRisk {
			out += md.Paragraph("⚠️ " + md.Bold("Warning: One or more positions are at risk of liquidation!"))
		}
	}

	if len(loans) > 0 {
		var (
			labels []string
			values []float64
		)
		for i := 0; i < len(loans); i += 2 {
			tokenPath := loans[i]
			amountStr := loans[i+1]
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				continue
			}
			token := grc20reg.Get(tokenPath)
			label := tokenPath
			if token != nil {
				label = token.GetSymbol()
			}
			labels = append(labels, label)
			values = append(values, amount)
		}
		chart := charts.GenerateBarChart(values, labels, 30, "Loans for "+displayName)
		out += md.H2("📊 Loans by Token")
		out += chart
	}

	return out
}
