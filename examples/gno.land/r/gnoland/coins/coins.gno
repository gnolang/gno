// Package coins provides simple helpers to retrieve information about coins
// on the Gno.land blockchain.
//
// The primary goal of this realm is to allow users to check their token balances without
// relying on external tools or services. This is particularly valuable for new networks
// that aren't yet widely supported by public explorers or wallets. By using this realm,
// users can always access their balance information directly through the gnodev.
//
// While currently focused on basic balance checking functionality, this realm could
// potentially be extended to support other banker-related workflows in the future.
// However, we aim to keep it minimal and focused on its core purpose.
//
// This is a "Render-only realm" - it exposes only a Render function as its public
// interface and doesn't maintain any state of its own. This pattern allows for
// simple, stateless information retrieval directly through the blockchain's
// rendering capabilities.
//
// Example usage:
//
//	/r/gnoland/coins:ugnot - shows the total supply of ugnot
//	/r/gnoland/coins:ugnot/g1... - shows the ugnot balance of a specific address
package coins

import (
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/moul/md"
)

var (
	errEmptyDenom     = errors.New("empty denom")
	errInvalidAddress = errors.New("invalid address")
	errInvalidPath    = errors.New("invalid path")
)

// TotalSupply returns the total supply of the specified denomination
func TotalSupply(denom string) int64 {
	if denom == "" {
		panic(errEmptyDenom)
	}
	// make the coin's denom to follow the correct format
	qualifiedDenom := std.CurrentRealm().CoinDenom(denom)
	return std.NewBanker(std.BankerTypeReadonly).TotalCoin(qualifiedDenom)
}

// IsEligible checks if an address is eligible for the airdrop
func IsEligible(denom string, addr std.Address) bool {
	return AddressBalance(denom, addr) > 0
}

// AddressBalance returns the balance of the specified token for the given address
func AddressBalance(denom string, address std.Address) int64 {
	if denom == "" {
		panic(errEmptyDenom)
	}
	if !address.IsValid() {
		panic(errInvalidAddress)
	}
	return std.NewBanker(std.BankerTypeReadonly).
		GetCoins(address).
		AmountOf(denom)
}

// Path parsing logic:
// - Empty path: Show homepage with general info and usage instructions
// - "<denom>": Show total supply and info about the specified coin
// - "<denom>/<address>": Show specific address balance for the given coin
func Render(path string) string {
	if path == "" {
		return renderHomepage()
	}

	if !strings.Contains(path, "/") {
		denom := path
		totalSupply := TotalSupply(denom)
		return renderCoinInfo(denom, totalSupply)
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		panic(errInvalidPath)
	}

	denom, addr := parts[0], parts[1]
	return renderAddressBalance(denom, std.Address(addr))

	// TODO: implement an explanatory homepage that shows usage instructions
	// TODO: implement a coin info page showing total supply, and other metrics if we have them
	// TODO: implement an address page showing balance for specified coin
	// TODO: create navigation links between pages to enable intuitive exploration without typing URLs
}

func renderHomepage() string {
	var sb strings.Builder

	sb.WriteString(md.H1("Gno.land Coins"))
	sb.WriteString(md.Paragraph("This realm provides an interface to query coin information on the Gno.land blockchain."))

	sb.WriteString(md.H2("Usage"))
	sb.WriteString(md.Paragraph("You can query information using the following URL patterns:"))

	urlPatterns := []string{
		md.Link("/r/gnoland/coins:ugnot", "/r/gnoland/coins:ugnot") + " - Check total supply of ugnot",
		"/r/gnoland/coins:ugnot/{address} - Check ugnot balance of a specific address",
	}
	sb.WriteString(md.BulletList(urlPatterns))

	sb.WriteString(md.H2("Supported Coins"))
	supportedCoins := []string{
		md.Link("ugnot", "/r/gnoland/coins:ugnot") + " - The native token of Gno.land",
	}
	sb.WriteString(md.BulletList(supportedCoins))

	return sb.String()
}

func renderCoinInfo(denom string, totalSupply int64) string {
	var sb strings.Builder

	sb.WriteString(md.H1("Coin Information"))

	sb.WriteString(md.H2("Denomination"))
	sb.WriteString(md.Paragraph(denom))

	sb.WriteString(md.H2("Total Supply"))
	sb.WriteString(md.Paragraph(std.NewCoin(denom, totalSupply).String()))

	sb.WriteString(md.H2("Statistics"))
	stats := []string{
		"Total Supply: " + std.NewCoin(denom, totalSupply).String(),
		"Chain ID: " + std.ChainID(),
		"Current Block Height: " + strconv.Itoa(int(std.ChainHeight())),
	}
	sb.WriteString(md.BulletList(stats))

	sb.WriteString(md.Paragraph(md.Link("Back to Home", "/r/gnoland/coins")))

	return sb.String()
}

func renderAddressBalance(denom string, address std.Address) string {
	var sb strings.Builder
	balance := AddressBalance(denom, address)
	isEligible := IsEligible(denom, address)

	sb.WriteString(md.H1("Address Balance Information"))

	sb.WriteString(md.H2("Address"))
	sb.WriteString(md.Paragraph(address.String()))

	sb.WriteString(md.H2("Balance"))
	sb.WriteString(md.Paragraph(std.NewCoin(denom, balance).String()))

	sb.WriteString(md.H2("Additional Information"))
	info := []string{
		"Airdrop Eligibility: " + (map[bool]string{true: "Eligible", false: "Not Eligible"})[isEligible],
	}
	sb.WriteString(md.BulletList(info))

	sb.WriteString(md.H2("Links"))
	links := []string{
		md.Link("Back to Coin Information", "/r/gnoland/coins:"+denom),
		md.Link("Back to Home", "/r/gnoland/coins"),
	}
	sb.WriteString(md.BulletList(links))

	return sb.String()
}
