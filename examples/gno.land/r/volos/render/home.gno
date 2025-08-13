package render

import (
	volos "gno.land/r/volos/core"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avl"
	"gno.land/r/demo/grc20reg"
)

func renderHome(path string) string {
	out := md.H1("Volos🦌")
	out += md.Paragraph("Volos is a decentralized lending protocol built on gno.land. It enables users to lend and borrow assets through configurable markets, with real-time risk management and on-chain price oracles powered by Gnoswap pools. The protocol uses a shares-based accounting system, supports variable interest rates, and ensures solvency through continuous health monitoring and liquidation mechanisms—all within the gno.land ecosystem.")
	out += "\n"
	out += md.Paragraph("This is a minimalistic gnoweb app. To access all Volos features, please use the " + md.Link("Official Volos Website", "https://officialvolos.com") + ".")
	out += md.H2("👤 Check Your Position")
	out += md.Blockquote("To view your lending and borrowing positions on Volos, navigate to `?user=g1xxx` e.g. `?user=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`")

	marketList := volos.GetMarketList()
	marketsTree := avl.NewTree()
	for _, marketId := range marketList {
		marketsTree.Set(marketId, marketId)
	}
	pager := pager.NewPager(rotree.Wrap(marketsTree, nil), pageSize, false)
	page := pager.MustGetPageByPath(path)

	out += md.H2("📊 Markets")
	out += md.Paragraph("Browse all lending markets on Volos. Each market is defined by a Gnoswap pool, interest rate model, and collateralization parameters.")

	table := mdtable.Table{
		Headers: []string{"Loan Token", "Collateral Token", "IRM", "Total Borrow Assets", "Total Supply Assets", "LLTV", ""},
	}
	for _, item := range page.Items {
		marketId := item.Key
		_, params := volos.GetMarket(marketId)
		loanPath := params.GetLoanToken()
		collateralPath := params.GetCollateralToken()
		irm := params.IRM
		lltv := params.LLTV.ToString()

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

		totalBorrowAssets := volos.GetMarketTotalBorrowAssets(marketId)
		totalSupplyAssets := volos.GetMarketTotalSupplyAssets(marketId)

		viewLink := md.Link("View", "?market="+marketId)
		table.Append([]string{loanSymbol, collateralSymbol, irm, totalBorrowAssets, totalSupplyAssets, lltv, viewLink})
	}

	if len(table.Rows) == 0 {
		out += md.Italic("No markets have been created yet. 🦌")
	} else {
		out += table.String()
		out += md.HorizontalRule()
		picker := page.Picker(path)
		if picker != "" {
			out += md.Paragraph(picker)
		}
	}

	return out
}
