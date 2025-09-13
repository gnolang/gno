package render

import (
	"net/url"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/avl/rotree"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/r/demo/grc20reg"
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

	out := md.H1("🏦 " + loanSymbol + "/" + collateralSymbol)

	out += md.H2("📊 Market Overview")
	overviewTable := mdtable.Table{
		Headers: []string{"Parameter", "Value"},
	}
	overviewTable.Append([]string{"Pool Path", md.InlineCode(params.PoolPath)})
	overviewTable.Append([]string{"Loan Token", md.Link(loanSymbol, strings.ReplaceAll(loanPath, "gno.land/", ""))})
	overviewTable.Append([]string{"Collateral Token", md.Link(collateralSymbol, strings.ReplaceAll(collateralPath, "gno.land/", ""))})
	overviewTable.Append([]string{"Interest Rate Model", md.InlineCode(params.IRM)})
	overviewTable.Append([]string{"Liquidation LTV", formatPercentage(params.LLTV) + "%"})
	overviewTable.Append([]string{"Market Fee", market.Fee.ToString()})
	out += overviewTable.String()

	coreRealm := txlink.Realm("gno.land/r/volos/core")
	loanRealm := txlink.Realm(loanPath)
	collateralRealm := txlink.Realm(collateralPath)

	out += md.H2("💰 Market Statistics & Actions")
	statsTable := mdtable.Table{
		Headers: []string{"Metric", "Value", "Rate Type", "APR", "Action", "Action"},
	}

	borrowAPR := u256.MustFromDecimal(volos.GetBorrowAPR(marketId))
	supplyAPR := u256.MustFromDecimal(volos.GetSupplyAPR(marketId))
	totalSupplyAssets := market.TotalSupplyAssets
	totalBorrowAssets := market.TotalBorrowAssets
	utilization := volos.CalculateUtilization(marketId)

	statsTable.Append([]string{
		"Total Supply Assets",
		formatTokenAmount(totalSupplyAssets, loanToken.GetDecimals()),
		"Borrow Rate",
		formatPercentage(borrowAPR) + "%",
		md.Link("Borrow", coreRealm.NewLink("Borrow").AddArgs("marketId", marketId, "assets", "", "shares", "").URL()),
		md.Link("Repay", coreRealm.NewLink("Repay").AddArgs("marketId", marketId, "assets", "", "shares", "").URL()),
	})

	statsTable.Append([]string{
		"Supply Rate",
		formatPercentage(supplyAPR) + "%",
		"Total Borrow Assets",
		formatTokenAmount(totalBorrowAssets, loanToken.GetDecimals()),
		md.Link("Supply", coreRealm.NewLink("Supply").AddArgs("marketId", marketId, "assets", "", "shares", "").URL()),
		md.Link("Withdraw", coreRealm.NewLink("Withdraw").AddArgs("marketId", marketId, "assets", "", "shares", "").URL()),
	})

	statsTable.Append([]string{
		"Utilization Rate",
		formatPercentage(utilization) + "%",
		"Last Update: " + formatTimestamp(market.LastUpdate),
		"",
		md.Link("Approve", loanRealm.NewLink("Approve").AddArgs("spender", std.DerivePkgAddr("gno.land/r/volos/core").String(), "amount", "").URL()),
		"",
		"",
	})
	out += statsTable.String()

	out += md.H2("🛠 Collateral Operations")
	colTable := mdtable.Table{Headers: []string{"Operation", "Action", "Approve"}}
	colTable.Append([]string{
		"Supply Collateral (" + collateralSymbol + ")",
		md.Link("Supply Collateral", coreRealm.NewLink("SupplyCollateral").AddArgs("marketId", marketId, "amount", "").URL()),
		md.Link("Approve Collateral", collateralRealm.NewLink("Approve").AddArgs("spender", std.DerivePkgAddr("gno.land/r/volos/core").String(), "amount", "").URL()),
	})
	colTable.Append([]string{
		"Withdraw Collateral (" + collateralSymbol + ")",
		md.Link("Withdraw Collateral", coreRealm.NewLink("WithdrawCollateral").AddArgs("marketId", marketId, "amount", "").URL()),
		md.Link("Approve Collateral", collateralRealm.NewLink("Approve").AddArgs("spender", std.DerivePkgAddr("gno.land/r/volos/core").String(), "amount", "").URL()),
	})
	out += colTable.String()

	out += md.H2("💱 Current Price")
	price := volos.GetMarketPrice(marketId)
	out += md.Paragraph("**Oracle Price:** " + formatPrice(u256.MustFromDecimal(price), loanToken.GetDecimals(), collateralToken.GetDecimals()) + " " + collateralSymbol + " per " + loanSymbol)
	out += md.Blockquote("Price sourced from Gnoswap pool: " + params.PoolPath)
	
	return out
}
