package grc20

import "github.com/gnolang/gno/gnovm/stdlibs/std"

type IGRC20 interface {

	// Name returns the name of the token - e.g. "MyToken"
	Name() (name string)

	// Symbol returns the symbol of the token, for example, “WUGNOT”
	Symbol() (symbol string)

	// Decimals returns the number of decimals the token uses - e.g. 8, means to divide the token amount by 100000000 to get its user representation
	Decimals() (decimals uint8)

	// TotalSupply returns the total token supply
	TotalSupply() (totalSupply uint64)

	// BalanceOf returns the account balance of another account with address _owner
	BalanceOf(owner std.Address) (balance uint64)

	// Transfer transfers `value` amount of tokens to address `to`, and MUST fire the Transfer event
	// The function SHOULD throw if the message caller’s account balance does not have enough tokens to spend
	Transfer(to std.Address, value uint64) (success bool)

	// TransferFrom transfers `value` amount of tokens from address `from` to address `to`, and MUST fire the Transfer event
	// Note: Transfers of 0 values MUST be treated as normal transfers and fire the Transfer event.
	TransferFrom(from, to std.Address, value uint64) (success bool)

	// Approve allows `spender` to withdraw from your account multiple times, up to the `value` amount
	// If this function is called again it overwrites the current allowance with value
	Approve(spender std.Address, value uint64) (success bool)

	Allowance(owner, spender std.Address) (remaining uint64)

	// Events
	// Transfer(from, to std.Address, value uint64) - MUST trigger when tokens are transferred, including zero value transfers
	// Approval(spender std.Address, value uint64) - MUST trigger on any successful call to Approve
}
