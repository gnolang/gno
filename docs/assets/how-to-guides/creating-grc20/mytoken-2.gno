// Name returns the name of the token
func Name() string {
	return mytoken.Name()
}

// Symbol returns the name of the token
func Symbol() string {
	return mytoken.Symbol()
}

// TotalSupply returns the total supply of the token
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

// Decimals returns the decimals of the token
func Decimals() uint8 {
	return mytoken.Decimals()
}

// BalanceOf returns the balance of a specific username or address
func BalanceOf(owner std.Address) uint64 {
	return mytoken.BalanceOf(owner)
}

// Allowance returns the allowance of a spender for the owner's tokens
func Allowance(owner, spender std.Address) uint64 {
	return mytoken.Allowance(owner, spender)
}

// Setters

// Transfer transfers `value` amount of tokens to address `to`, and MUST fire the Transfer event
// The function SHOULD throw if the message caller’s account balance does not have enough tokens to spend
func Transfer(to std.Address, amount uint64) {
	mytoken.Transfer(to, amount)
}

// Approve allows `spender` to withdraw from your account multiple times, up to the `value` amount
// If this function is called again it overwrites the current allowance with value
func Approve(spender std.Address, amount uint64) {
	mytoken.Approve(spender, amount)
}

// TransferFrom transfers `value` amount of tokens from address `from` to address `to`, and MUST fire the Transfer event
func TransferFrom(from, to std.Address, amount uint64) {
	mytoken.TransferFrom(from, to, amount)
}

// Administration

// Mint mints `amount` of tokens to `address`, only callable by owner
func Mint(address std.Address, amount uint64) {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint tokens")
	}

	mytoken.Mint(address, amount)
}

// Burn burns `amount` of tokens from `address`, only callable by owner
func Burn(address std.Address, amount uint64) {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can burn tokens")
	}

	mytoken.Burn(address, amount)
}

// Rendering

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance": // pkgpath:balance/address
		owner := std.Address(parts[1])
		return ufmt.Sprintf("%d\n", mytoken.BalanceOf(owner))
	default:
		return "404\n"
	}
}
