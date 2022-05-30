package gnousd

func UpdateCoinPrice(price uint64) {
	// FIXME: check origin call
	// TODO: msg.sender == oracle
	controller.UpdateCoinPrice(price)
}

func UpdateSharePrice(price uint64) {
	// FIXME: check origin call
	// TODO: msg.sender == oracle
	controller.UpdateSharePrice(price)
}

// func: UPDATE ORACLE

/*
import "std"

// method proxies as public functions.
//

// getters.

func TotalSupply() uint64                         { return coins.TotalSupply() }
func BalanceOf(address std.Address) uint64        { return coins.BalanceOf(address) }
func Allowance(owner, spender std.Address) uint64 { return coins.Allowance(owner, spender) }

// setters.

func Transfer(to std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	coins.Transfer(to, amount)
}

func Approve(spender std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	coins.Approve(spender, amount)
}

func TransferFrom(from, to std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	coins.TransferFrom(from, to, amount)
}

// administration.

func Mint(address std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	coins.Mint(address, amount)
}

func Burn(address std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	coins.Burn(address, amount)
}
*/
