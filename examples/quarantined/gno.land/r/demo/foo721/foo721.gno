// Realm foo721 is the reference GRC721 consumer.
package foo721

import (
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/demo/tokens/grc721/collection"
	"gno.land/p/demo/tokens/grc721/metadata"
	"gno.land/p/demo/tokens/grc721/royalty"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/demo/grc721reg"
)

const admin address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh" // govdao t1 multisig

// Trust boundary: nft and the write-ledgers stay private; public writes go
// through the caller-resolved Teller, so no caller address can be spoofed.
var (
	nft        *collection.NFT
	meta       *metadata.Metadata
	metaLedger *metadata.Ledger
	roy        *royalty.Royalty
	royLedger  *royalty.Ledger
	callerTelr grc721.Teller
)

func init(cur realm) {
	nft = collection.New("Foo NFT", "FNFT", 0, cur)
	meta, metaLedger = metadata.NewMetadata(nft.Ledger())
	roy, royLedger = royalty.NewRoyalty(nft.Ledger(), 1000) // cap royalties at 10%
	nft.Attach(meta).Attach(roy)
	callerTelr = nft.CallerTeller()

	royLedger.SetDefaultRoyalty(admin, 500)

	mintN(admin, 10)

	grc721reg.Register(cross(cur), nft.Token(), "", meta, roy)
}

func mintN(owner address, n int64) {
	count := nft.Token().TotalSupply()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		if err := nft.Mint(owner, tid); err != nil {
			panic(err)
		}
	}
}

// -- reads ----------------------------------------------------------------

func BalanceOf(user address) int64 { return nft.Token().BalanceOf(user) }

func OwnerOf(tid grc721.TokenID) address {
	owner, err := nft.Token().OwnerOf(tid)
	if err != nil {
		panic(err)
	}
	return owner
}

func IsApprovedForAll(owner, user address) bool { return nft.Token().IsApprovedForAll(owner, user) }

func GetApproved(tid grc721.TokenID) address {
	addr, err := nft.Token().GetApproved(tid)
	if err != nil {
		panic(err)
	}
	return addr
}

func TokenURI(tid grc721.TokenID) string {
	uri, err := meta.TokenURI(tid)
	if err != nil {
		panic(err)
	}
	return uri
}

func RoyaltyInfo(tid grc721.TokenID, salePrice int64) (address, int64) {
	recv, amount, err := roy.RoyaltyInfo(tid, salePrice)
	if err != nil {
		panic(err)
	}
	return recv, amount
}

// -- caller-resolved writes -----------------------------------------------

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

// -- admin (issuer-only) --------------------------------------------------

func Mint(cur realm, to address, tid grc721.TokenID) {
	assertIsAdmin(cur.Previous().Address())
	if err := nft.Mint(to, tid); err != nil {
		panic(err)
	}
}

func Burn(cur realm, tid grc721.TokenID) {
	assertIsAdmin(cur.Previous().Address())
	if err := nft.Burn(tid); err != nil {
		panic(err)
	}
}

func SetTokenURI(cur realm, tid grc721.TokenID, uri string) {
	assertIsAdmin(cur.Previous().Address())
	if err := metaLedger.SetTokenURI(tid, uri); err != nil {
		panic(err)
	}
}

func SetTokenRoyalty(cur realm, tid grc721.TokenID, receiver address, bps int64) {
	assertIsAdmin(cur.Previous().Address())
	if err := royLedger.SetTokenRoyalty(tid, receiver, bps); err != nil {
		panic(err)
	}
}

// -- render ---------------------------------------------------------------

func Render(path string) string {
	switch path {
	case "":
		return nft.Token().RenderHome()
	default:
		return "404\n"
	}
}

func assertIsAdmin(caller address) {
	if caller != admin {
		panic("restricted access")
	}
}
