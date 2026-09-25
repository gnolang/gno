package tba

import (
	"strconv"
	"strings"

	"chain"
	"chain/banker"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

// admin is fixed to the deployer's address. Unlike deriving it from
// unsafe.PreviousRealm() at package init, this stays correct regardless of
// which key signs the addpkg transaction (e.g. examples/ deployments signed
// by a genesis key).
const admin = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

type tokenMeta struct {
	name  string
	image string
}

var (
	nft *grc721.BasicNFT

	// realmPath is captured from cur.PkgPath() the first time Setup runs,
	// i.e. it reflects wherever this code is actually deployed. VaultAddress
	// and Balance derive from this instead of a hardcoded constant, so a
	// copy of this file deployed at another path cannot advertise a vault
	// address it does not control.
	realmPath string

	nextID    int64 = 1
	allTokens []grc721.TokenID
	metaByID  avl.Tree // tokenID string -> *tokenMeta
)

// Setup initializes the realm. Must be called once, by admin, before Mint,
// Withdraw, or TransferFrom can be used.
func Setup(cur realm) {
	if nft != nil {
		panic("already initialized")
	}
	caller := cur.Previous().Address()
	if caller != admin {
		panic("only admin can initialize")
	}
	realmPath = cur.PkgPath()
	nft = grc721.NewBasicNFT(0, cur, "Token Bound Account", "TBA")
}

// Mint creates a new token-bound-account NFT and assigns it to `to`.
// Only admin can mint. Must be called after Setup.
func Mint(cur realm, to address, name string, image string) grc721.TokenID {
	if nft == nil {
		panic("not initialized: call Setup first")
	}
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

// VaultAddress returns the on-chain vault address for a given token.
func VaultAddress(tid grc721.TokenID) address {
	return chain.DerivePkgSubAddr(realmPath, vaultSubpath(tid))
}

// Balance returns the ugnot balance held in a token's vault.
func Balance(tid grc721.TokenID) int64 {
	coins := banker.NewReadonlyBanker().GetCoins(VaultAddress(tid))
	return coins.AmountOf("ugnot")
}

// Withdraw sends `amount` ugnot from a token's vault to `to`. Only the
// current owner of the token may withdraw. Note: the vault is drainable by
// the current owner up to the moment a transfer lands, since Withdraw and
// TransferFrom are separate transactions (the same property ERC-6551 has).
func Withdraw(cur realm, tid grc721.TokenID, to address, amount int64) {
	if amount <= 0 {
		panic("amount must be positive")
	}
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

// OwnerOf returns the current owner of a token.
func OwnerOf(tid grc721.TokenID) (address, error) {
	return nft.OwnerOf(tid)
}

// BalanceOfOwner returns how many tokens an address owns.
func BalanceOfOwner(owner address) (int64, error) {
	return nft.BalanceOf(owner)
}

// TransferFrom transfers a token (and control of its vault) from one owner
// to another. Panics if the caller is not the owner or an approved operator,
// so an unauthorized attempt aborts the transaction instead of committing a
// silent no-op.
func TransferFrom(cur realm, from, to address, tid grc721.TokenID) {
	caller := cur.Previous().Address()
	if err := nft.TransferFrom(caller, from, to, tid); err != nil {
		panic(err)
	}
}

// AllTokens returns every minted token ID.
func AllTokens() []grc721.TokenID {
	return allTokens
}

// TokensOf returns every token ID owned by `owner`.
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

// TokenInfo returns the metadata, owner, vault address, and balance for a
// token in one call.
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

// Render implements the realm's default render function. When path names a
// specific token id, it shows that token's vault details instead of just
// the collection header.
func Render(path string) string {
	if nft == nil {
		return "# Token Bound Accounts\n\nNot initialized yet. Call Setup() first.\n"
	}
	if path == "" {
		return nft.RenderHome()
	}

	tid := grc721.TokenID(strings.TrimPrefix(path, "token/"))
	name, image, owner, vault, balance, err := TokenInfo(tid)
	if err != nil {
		return "# Token Bound Accounts\n\nToken not found: " + string(tid) + "\n"
	}

	var b strings.Builder
	b.WriteString("# " + name + "\n\n")
	if image != "" {
		b.WriteString("![" + name + "](" + image + ")\n\n")
	}
	b.WriteString("- Owner: " + owner.String() + "\n")
	b.WriteString("- Vault: " + vault.String() + "\n")
	b.WriteString("- Balance: " + strconv.FormatInt(balance, 10) + " ugnot\n")
	return b.String()
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
