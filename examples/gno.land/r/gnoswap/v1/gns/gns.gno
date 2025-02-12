package gns

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/grc20reg"

	"gno.land/p/gnoswap/consts"
)

const (
	MAXIMUM_SUPPLY      = uint64(1_000_000_000_000_000)
	INITIAL_MINT_AMOUNT = uint64(100_000_000_000_000)
	MAX_EMISSION_AMOUNT = uint64(900_000_000_000_000) // MAXIMUM_SUPPLY - INITIAL_MINT_AMOUNT
)

var (
	owner         *ownable.Ownable
	token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	UserTeller    grc20.Teller

	leftEmissionAmount   uint64 // amount of GNS can be minted for emission
	mintedEmissionAmount uint64 // amount of GNS that has been minted for emission
	lastMintedHeight     int64  // last block height that gns was minted for emission

	burnAmount uint64 // amount of GNS that has been burned
)

func init() {
	owner = ownable.NewWithAddress(consts.ADMIN)
	token, privateLedger = grc20.NewToken("Gnoswap", "GNS", 6)
	UserTeller = token.CallerTeller()

	privateLedger.Mint(owner.Owner(), INITIAL_MINT_AMOUNT)
	getter := func() *grc20.Token { return token }
	grc20reg.Register(getter, "")

	// Initial amount set to 900_000_000_000_000 (MAXIMUM_SUPPLY - INITIAL_MINT_AMOUNT).
	// leftEmissionAmount will decrease as tokens are minted.
	setLeftEmissionAmount(MAX_EMISSION_AMOUNT)
	setMintedEmissionAmount(uint64(0))
	setLastMintedHeight(std.GetHeight())
	burnAmount = uint64(0)
}

func TotalSupply() uint64 {
	return token.TotalSupply()
}

func GetName() string {
	return token.GetName()
}

func GetSymbol() string {
	return token.GetSymbol()
}

func GetDecimals() uint {
	return token.GetDecimals()
}

func KnownAccounts() int {
	return token.KnownAccounts()
}

func BalanceOf(owner std.Address) uint64 {
	return token.BalanceOf(owner)
}

func Allowance(owner, spender std.Address) uint64 {
	return token.Allowance(owner, spender)
}

func MintGns(address std.Address) uint64 {
	lastGNSMintedHeight := GetLastMintedHeight()
	currentHeight := std.GetHeight()

	// skip minting process if following conditions are met
	// - if gns for current block is already minted
	// - if last minted height is same or later than emission end height
	if lastGNSMintedHeight == currentHeight || lastGNSMintedHeight >= GetEndHeight() {
		return 0
	}

	assertShouldNotBeHalted()
	assertCallerIsEmission()

	// calculate gns amount to mint
	amountToMint := calculateAmountToMint(lastGNSMintedHeight+1, currentHeight)

	// update
	setLastMintedHeight(currentHeight)
	setMintedEmissionAmount(GetMintedEmissionAmount() + amountToMint)
	setLeftEmissionAmount(GetLeftEmissionAmount() - amountToMint)

	// mint calculated amount to address
	err := privateLedger.Mint(address, amountToMint)
	if err != nil {
		panic(err.Error())
	}

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"MintGNS",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"mintedBlockHeight", formatInt(currentHeight),
		"mintedGNSAmount", formatUint(amountToMint),
		"accumMintedGNSAmount", formatUint(GetMintedEmissionAmount()),
		"accumLeftMintGNSAmount", formatUint(GetLeftEmissionAmount()),
	)

	return amountToMint
}

func Burn(from std.Address, amount uint64) {
	owner.AssertCallerIsOwner()
	checkErr(privateLedger.Burn(from, amount))

	burnAmount += amount

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Burn",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"burnedBlockHeight", formatInt(std.GetHeight()),
		"burnFrom", from.String(),
		"burnedGNSAmount", formatUint(amount),
		"accumBurnedGNSAmount", formatUint(GetBurnAmount()),
	)
}

func Transfer(to std.Address, amount uint64) {
	checkErr(UserTeller.Transfer(to, amount))
}

func Approve(spender std.Address, amount uint64) {
	checkErr(UserTeller.Approve(spender, amount))
}

func TransferFrom(from, to std.Address, amount uint64) {
	checkErr(UserTeller.TransferFrom(from, to, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance := token.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// calculateAmountToMint calculates the amount of gns to mint
// It calculates the amount of gns to mint for each halving year for block range.
// It also handles the left emission amount if the current block range includes halving year end block.
func calculateAmountToMint(fromHeight, toHeight int64) uint64 {
	// if toHeight is greater than emission end height, set toHeight to emission end height
	endH := GetEndHeight()
	if toHeight > endH {
		toHeight = endH
	}

	if fromHeight > toHeight {
		return 0
	}

	fromYear := GetHalvingYearByHeight(fromHeight)
	toYear := GetHalvingYearByHeight(toHeight)

	totalAmountToMint := uint64(0)

	curFrom := fromHeight

	for year := fromYear; year <= toYear; year++ {
		yearEndHeight := GetHalvingYearEndBlock(year)
		mintUntilHeight := i64Min(yearEndHeight, toHeight)

		// how many blocks to calculate
		blocks := uint64(mintUntilHeight - curFrom + 1)
		if blocks <= 0 {
			break
		}

		// amount of gns to mint for each block for current year
		singleBlockAmount := GetAmountPerBlockPerHalvingYear(year)

		// amount of gns to mint for current year
		yearAmountToMint := singleBlockAmount * blocks

		// if last block of halving year, handle left emission amount
		if mintUntilHeight >= yearEndHeight {
			leftover := handleLeftEmissionAmount(year, yearAmountToMint)
			yearAmountToMint += leftover
		}

		totalAmountToMint += yearAmountToMint

		setHalvingYearMintAmount(year, GetHalvingYearMintAmount(year)+yearAmountToMint)
		setHalvingYearLeftAmount(year, GetHalvingYearLeftAmount(year)-yearAmountToMint)

		std.Emit(
			"CalculateAmountToMint",
			"fromHeight", formatInt(curFrom),
			"toHeight", formatInt(mintUntilHeight),
			"year", formatInt(year),
			"singleBlockAmount", formatUint(singleBlockAmount),
		)

		// update fromHeight for next year (if necessary)
		curFrom = mintUntilHeight + 1
		if curFrom > toHeight {
			break
		}
	}

	assertTooManyEmission(totalAmountToMint)
	return totalAmountToMint
}

// isLastBlockOfHalvingYear returns true if the current block is the last block of a halving year.
func isLastBlockOfHalvingYear(height int64) bool {
	year := GetHalvingYearByHeight(height)
	lastBlock := GetHalvingYearEndBlock(year)

	return height == lastBlock
}

// handleLeftEmissionAmount handles the left emission amount for a halving year.
// It calculates the left emission amount by subtracting the halving year mint amount from the halving year amount.
func handleLeftEmissionAmount(year int64, amount uint64) uint64 {
	return GetHalvingYearLeftAmount(year) - amount
}

// skipIfSameHeight returns true if the current block height is the same as the last minted height.
// This prevents multiple gns minting inside the same block.
func skipIfSameHeight(lastMintedHeight, currentHeight int64) bool {
	return lastMintedHeight == currentHeight
}

// skipIfEmissionEnded returns true if the emission has ended.
func skipIfEmissionEnded(height int64) bool {
	if isEmissionEnded(height) {
		return true
	}

	return false
}

// GetLastMintedHeight returns the last block height that gns was minted.
func GetLastMintedHeight() int64 {
	return lastMintedHeight
}

// GetLeftEmissionAmount returns the amount of GNS can be minted.
func GetLeftEmissionAmount() uint64 {
	return leftEmissionAmount
}

// GetBurnAmount returns the amount of GNS that has been burned.
func GetBurnAmount() uint64 {
	return burnAmount
}

// GetMintedEmissionAmount returns the amount of GNS that has been minted by the emission contract.
// It does not include initial minted amount.
func GetMintedEmissionAmount() uint64 {
	return mintedEmissionAmount
}

func setLastMintedHeight(height int64) {
	lastMintedHeight = height
}

func setLeftEmissionAmount(amount uint64) {
	leftEmissionAmount = amount
}

func setMintedEmissionAmount(amount uint64) {
	mintedEmissionAmount = amount
}

// assertTooManyEmission asserts if the amount of gns to mint is too many.
// It checks if the amount of gns to mint is greater than the left emission amount or the total emission amount.
func assertTooManyEmission(amount uint64) {
	if (amount + GetMintedEmissionAmount()) > MAX_EMISSION_AMOUNT {
		panic(addDetailToError(
			errTooManyEmission,
			ufmt.Sprintf("amount: %d", amount),
		))
	}
}
