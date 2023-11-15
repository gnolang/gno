func mintNNFT(owner std.Address, n uint64) {
	count := my.TokenCount()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		mynonfungibletoken.Mint(owner, tid)
	}
}

// Getters

func BalanceOf(user users.AddressOrName) uint64 {
	balance, err := mynonfungibletoken.BalanceOf(user.Resolve())
	if err != nil {
		panic(err)
	}

	return balance
}

func OwnerOf(tid grc721.TokenID) std.Address {
	owner, err := mynonfungibletoken.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	return owner
}

func IsApprovedForAll(owner, user users.AddressOrName) bool {
	return mynonfungibletoken.IsApprovedForAll(owner.Resolve(), user.Resolve())
}

func GetApproved(tid grc721.TokenID) std.Address {
	addr, err := mynonfungibletoken.GetApproved(tid)
	if err != nil {
		panic(err)
	}

	return addr
}

// Setters

func Approve(user users.AddressOrName, tid grc721.TokenID) {
	err := mynonfungibletoken.Approve(user.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(user users.AddressOrName, approved bool) {
	err := mynonfungibletoken.SetApprovalForAll(user.Resolve(), approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to users.AddressOrName, tid grc721.TokenID) {
	err := mynonfungibletoken.TransferFrom(from.Resolve(), to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to users.AddressOrName, tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mynonfungibletoken.Mint(to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func Burn(tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mynonfungibletoken.Burn(tid)
	if err != nil {
		panic(err)
	}
}

// Render

func Render(path string) string {
	switch {
	case path == "":
		return mynonfungibletoken.RenderHome()
	default:
		return "404\n"
	}
}

// Util

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
