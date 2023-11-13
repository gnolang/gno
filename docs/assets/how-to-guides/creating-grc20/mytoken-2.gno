func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

func BalanceOf(owner users.AddressOrName) uint64 {
	balance, err := mytoken.BalanceOf(owner.Resolve())
	if err != nil {
		panic(err)
	}
	return balance
}

func Allowance(owner, spender users.AddressOrName) uint64 {
	allowance, err := mytoken.Allowance(owner.Resolve(), spender.Resolve())
	if err != nil {
		panic(err)
	}
	return allowance
}

func Transfer(to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.Transfer(caller, to.Resolve(), amount)
	if err != nil {
		panic(err)
	}
}

func Approve(spender users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.Approve(caller, spender.Resolve(), amount)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.TransferFrom(caller, from.Resolve(), to.Resolve(), amount)
	if err != nil {
		panic(err)
	}
}

func Mint(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mytoken.Mint(address.Resolve(), amount)
	if err != nil {
		panic(err)
	}
}

func Burn(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mytoken.Burn(address.Resolve(), amount)
	if err != nil {
		panic(err)
	}
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := users.AddressOrName(parts[1])
		balance, _ := mytoken.BalanceOf(owner.Resolve())
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
