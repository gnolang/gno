// TotalSupply returns the total supply of mytoken
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

// Decimals returns the number of decimals of mytoken
func Decimals() uint {
	return mytoken.GetDecimals()
}

// BalanceOf returns the balance mytoken for `account`
func BalanceOf(account std.Address) uint64 {
	balance, err := mytoken.BalanceOf(account)
	if err != nil {
		panic(err)
	}

	return balance
}

// Allowance returns the allowance of spender on owner's balance
func Allowance(owner, spender std.Address) uint64 {
	allowance, err := mytoken.Allowance(owner, spender)
	if err != nil {
		panic(err)
	}

	return allowance
}

// Transfer transfers amount from caller to recipient
func Transfer(recipient std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Transfer(caller, recipient, amount); err != nil {
		panic(err)
	}
}

// Approve approves amount of caller's tokens to be spent by spender
func Approve(spender std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Approve(caller, spender, amount); err != nil {
		panic(err)
	}
}

// TransferFrom transfers `amount` of tokens from `from` to `to`
func TransferFrom(from, to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()

	if amount <= 0 {
		panic("transfer amount must be greater than zero")
	}

	if err := mytoken.TransferFrom(caller, from, to, amount); err != nil {
		panic(err)
	}
}

// Mint mints amount of tokens to address. Callable only by admin of token
func Mint(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("mint amount must be greater than zero")
	}

	if err := mytoken.Mint(address, amount); err != nil {
		panic(err)
	}
}

// Burn burns amount of tokens from address. Callable only by admin of token
func Burn(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("burn amount must be greater than zero")
	}

	if err := mytoken.Burn(address, amount); err != nil {
		panic(err)
	}
}

// assertIsAdmin asserts the address is the admin of token
func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}

// Render renders the state of the realm
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		// Default GRC20 render
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		// Render balance of specific address
		owner := std.Address(parts[1])
		balance, _ := mytoken.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
