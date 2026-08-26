// Realm foo721 is the reference GRC721 consumer.
package foo721

import (
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/ufmt/v0"
)

const admin address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh" // govdao t1 multisig

var (
	token      *grc721.Token
	ledger     *grc721.PrivateLedger
	callerTelr grc721.Teller
)

func init(cur realm) {
	token, ledger = grc721.NewToken("Foo NFT", "FNFT", 0, cur)
	callerTelr = token.CallerTeller()

	mintN(admin, 10)
}

func mintN(owner address, n int64) {
	count := token.TotalSupply()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		if err := ledger.Mint(owner, tid); err != nil {
			panic(err)
		}
	}
}

// Getters

func BalanceOf(user address) int64 {
	balance, err := token.BalanceOf(user)
	if err != nil {
		panic(err)
	}
	return balance
}

func OwnerOf(tid grc721.TokenID) address {
	owner, err := token.OwnerOf(tid)
	if err != nil {
		panic(err)
	}
	return owner
}

func IsApprovedForAll(owner, user address) bool { return token.IsApprovedForAll(owner, user) }

func GetApproved(tid grc721.TokenID) address {
	addr, err := token.GetApproved(tid)
	if err != nil {
		panic(err)
	}
	return addr
}

// Setters

// The Approve / SetApprovalForAll / TransferFrom wrappers are the
// owning realm's trusted boundary: they hand `cur` to the caller-scoped
// Teller, which derives `caller` from `cur.Previous().Address()`
// (unforgeable per-frame identity) and applies it to the unexported
// `ledger`. The Teller rejects an rlm that is not the live crossing
// frame, so a caller cannot act as an account it does not control.
// See `p/demo/tokens/grc721/tellers.gno` for the Teller doctrine.
func Approve(cur realm, to address, tid grc721.TokenID) {
	if err := callerTelr.Approve(0, cur, to, tid); err != nil {
		panic(err)
	}
}

func SetApprovalForAll(cur realm, operator address, approved bool) {
	if err := callerTelr.SetApprovalForAll(0, cur, operator, approved); err != nil {
		panic(err)
	}
}

func TransferFrom(cur realm, from, to address, tid grc721.TokenID) {
	if err := callerTelr.TransferFrom(0, cur, from, to, tid); err != nil {
		panic(err)
	}
}

// Admin

func Mint(cur realm, to address, tid grc721.TokenID) {
	assertIsAdmin(cur.Previous().Address())
	if err := ledger.Mint(to, tid); err != nil {
		panic(err)
	}
}

func Burn(cur realm, tid grc721.TokenID) {
	assertIsAdmin(cur.Previous().Address())
	if err := ledger.Burn(tid); err != nil {
		panic(err)
	}
}

// Render

func Render(path string) string {
	switch path {
	case "":
		return token.RenderHome()
	default:
		return "404\n"
	}
}

// Util

func assertIsAdmin(address_XXX address) {
	if address_XXX != admin {
		panic("restricted access")
	}
}
