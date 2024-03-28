package mynft

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/grc/grc721"
	"std"
	//"gno.land/p/demo/grc/grc721"
)

var (
	mynft *grc721.IGRC721
	admin std.Address
)

// init is called once at time of deployment
func init() {
	// Set deployer of Realm to admin
	admin = std.PrevRealm().Addr()

	// Set token name, symbol and number of decimals
	mynft = grc721.NewBasicNFT("My NFT", "MNFT")
}

func BalanceOf(owner std.Address) uint64 {
	balance, err := mynft.BalanceOf(owner)
	if err != nil {
		panic(err)
	}

	return balance
}

func OwnerOf(tid grc721.TokenID) std.Address {
	owner, err := mynft.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	return owner
}

func IsApprovedForAll(owner, operator std.Address) bool {
	return mynft.IsApprovedForAll(owner, operator)
}

func GetApproved(tid grc721.TokenID) std.Address {
	addr, err := mynft.GetApproved(tid)
	if err != nil {
		panic(err)
	}

	return addr
}

func Approve(to std.Address, tid grc721.TokenID) {
	err := mynft.Approve(to, tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(operator std.Address, approved bool) {
	err := mynft.SetApprovalForAll(operator, approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to std.Address, tid grc721.TokenID) {
	err := mynft.TransferFrom(from, to, tid)
	if err != nil {
		panic(err)
	}
}

func Mint(to std.Address, tid grc721.TokenID) {
	assertIsAdmin(std.PrevRealm().Addr())
	err := mynft.Mint(to, tid)
	if err != nil {
		panic(err)
	}
}

func Burn(tid grc721.TokenID) {
	assertIsAdmin(std.PrevRealm().Addr())
	err := mynft.Burn(tid)
	if err != nil {
		panic(err)
	}
}

func Render(path string) string {
	switch {
	case path == "":
		return mynft.RenderHome()
	default:
		return "404\n"
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
