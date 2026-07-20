package tba

import (
	"strconv"
	"strings"

	"chain"
	"chain/banker"
	"chain/runtime/unsafe"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

const selfPath = "gno.land/r/zeycan1/tba"

type tokenMeta struct {
	name  string
	image string
}

var (
	nft       *grc721.BasicNFT
	admin     address
	nextID    int64 = 1
	allTokens []grc721.TokenID
	metaByID  avl.Tree // tokenID string -> *tokenMeta
)

func init() {
	admin = address(unsafe.PreviousRealm().Address())
}

func Setup(cur realm) {
	if nft != nil {
		panic("already initialized")
	}
	caller := cur.Previous().Address()
	if caller != admin {
		panic("only admin can initialize")
	}
	nft = grc721.NewBasicNFT(0, cur, "Token Bound Account", "TBA")
}

func Mint(cur realm, to address, name string, image string) grc721.TokenID {
	caller := cur.Previous().Address()
	if caller != admin {
		panic("only admin can mint")
	}
	tid := grc721.TokenID(strconv.FormatInt(nextID, 10))
	nextID++
	if err := nft.Mint(to, tid); err != nil {
		panic(err)
	}
	allTokens = append(allTokens, tid)
	metaByID.Set(string(tid), &tokenMeta{name: name, image: image})
	return tid
}

func vaultSubpath(tid grc721.TokenID) string {
	return "vault/" + string(tid)
}

func VaultAddress(tid grc721.TokenID) address {
	return chain.DerivePkgSubAddr(selfPath, vaultSubpath(tid))
}

func Balance(tid grc721.TokenID) int64 {
	coins := banker.NewReadonlyBanker().GetCoins(VaultAddress(tid))
	return coins.AmountOf("ugnot")
}

func Withdraw(cur realm, tid grc721.TokenID, to address, amount int64) {
	caller := cur.Previous().Address()
	owner, err := nft.OwnerOf(tid)
	if err != nil {
		panic(err)
	}
	if caller != owner {
		panic("caller is not the current owner of this token")
	}

	sub := cur.Sub(vaultSubpath(tid))
	b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
	b.SendCoins(sub.Address(), to, chain.NewCoins(chain.NewCoin("ugnot", amount)))
}

func OwnerOf(tid grc721.TokenID) (address, error) {
	return nft.OwnerOf(tid)
}

func BalanceOfOwner(owner address) (int64, error) {
	return nft.BalanceOf(owner)
}

func TransferFrom(cur realm, from, to address, tid grc721.TokenID) error {
	caller := cur.Previous().Address()
	return nft.TransferFrom(caller, from, to, tid)
}

func AllTokens() []grc721.TokenID {
	return allTokens
}

func TokensOf(owner address) []grc721.TokenID {
	var out []grc721.TokenID
	for _, tid := range allTokens {
		o, err := nft.OwnerOf(tid)
		if err == nil && o == owner {
			out = append(out, tid)
		}
	}
	return out
}

func TokenInfo(tid grc721.TokenID) (name string, image string, owner address, vault address, balance int64, err error) {
	owner, err = nft.OwnerOf(tid)
	if err != nil {
		return "", "", "", "", 0, err
	}
	vault = VaultAddress(tid)
	balance = Balance(tid)
	if v := metaByID.Get(string(tid)); v != nil {
		m := v.(*tokenMeta)
		name = m.name
		image = m.image
	}
	return name, image, owner, vault, balance, nil
}

func Render(path string) string {
	if nft == nil {
		return "# Token Bound Accounts\n\nNot initialized yet. Call Setup() first.\n"
	}
	return nft.RenderHome()
}

// --- Frontend-friendly CSV variants (avoid slice-ref parsing issues) ---

func AllTokensCSV() string {
	ids := make([]string, len(allTokens))
	for i, t := range allTokens {
		ids[i] = string(t)
	}
	return strings.Join(ids, ",")
}

func TokensOfCSV(owner address) string {
	toks := TokensOf(owner)
	ids := make([]string, len(toks))
	for i, t := range toks {
		ids[i] = string(t)
	}
	return strings.Join(ids, ",")
}
