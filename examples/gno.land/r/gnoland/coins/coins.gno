package coins

import (
	"std"
	"strings"

	"gno.land/p/moul/md"
)

func Render(path string) string {
	banker := std.NewBanker(std.BankerTypeReadonly)

	parts := strings.Split(path, "/")

	switch {
	case path == "":
		return renderHomepage()
	case len(parts) == 1:
		return renderCoinInfo(banker, parts[0])
	case len(parts) == 2:
		return renderAddressBalance(banker, parts[0], parts[1])
	default:
		return md.Paragraph("Invalid path. Usage:") +
			md.BulletList([]string{
				"/r/gnoland/coins - Show usage instructions",
				"/r/gnoland/coins:<denom> - Show total supply for DENOM",
				"/r/gnoland/coins:<denom>/<address> - Show balance for ADDRESS",
			})
	}
}

func renderHomepage() string {
	return md.H1("gno.land coin information") +
		md.Paragraph("This realm provides information about coins on the Gno.land blockchain.") +
		md.H2("Usage") +
		md.Paragraph("View coin supply: `/r/gnoland/coins:<denom>`\n\n"+
			"Check address balance: `/r/gnoland/coins:<denom>/<address>`\n\n"+
			"[Check total ugnot supply](/r/gnoland/coins:ugnot)")
}

func renderCoinInfo(banker std.Banker, denom string) string {
	//TODO : uncomment after banker.TotalCoin is implemented implementation
	//total := banker.TotalCoin(denom)

	return md.H1("Coin Information: "+denom) +
		//"Total Supply: " + std.NewCoin(denom, total).String() + "\n\n" +
		md.Paragraph("[← Back to homepage](/r/gnoland/coins)")
}

func renderAddressBalance(banker std.Banker, denom string, addrStr string) string {
	addr := std.Address(addrStr)
	coins := banker.GetCoins(addr)
	balance := coins.AmountOf(denom)

	return md.H1("Address Balance") +
		md.Paragraph("Address: **"+addrStr+"**") +
		md.Paragraph("Balance: **"+std.NewCoin(denom, balance).String()+"**") +
		md.Paragraph("[← Back to coin info](/r/gnoland/coins:"+denom+")") +
		md.Paragraph("[← Back to homepage](/r/gnoland/coins)")
}
