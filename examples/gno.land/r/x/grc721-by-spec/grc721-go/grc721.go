package grc721

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"strconv"
)

type Token struct {
	name              string
	symbol            string
	owners            *avl.Tree // tokenID > std.Address
	balances          *avl.Tree // std.Address > # of owned tokens
	tokenApprovals    *avl.Tree // tokenID > std.Address
	operatorApprovals *avl.Tree // "OwnerAddress:OperatorAddress" -> bool
	tokenURIs         *avl.Tree // tokenID > URI
}

var _ IGRC721 = (*Token)(nil)

func NewGRC721Token(name, symbol string) *Token {
	return &Token{
		name:     name,
		symbol:   symbol,
		owners:   avl.NewTree(),
		balances: avl.NewTree(),
		// give an address permission to a specific tokenID
		tokenApprovals: avl.NewTree(),
		// give any addresses permissions for all owners' assets
		operatorApprovals: avl.NewTree(),
		tokenURIs:         avl.NewTree(),
	}
}

func (nft Token) Name() string   { return nft.name }
func (nft Token) Symbol() string { return nft.symbol }

func (nft Token) BalanceOf(owner std.Address) uint64 {
	mustBeValid(owner)

	balance, found := nft.balances.Get(owner.String())
	if !found {
		return 0
	}

	return balance.(uint64)
}

func (nft Token) OwnerOf(tokenId string) std.Address {
	return nft.mustBeOwned(tokenId)
}

func (nft Token) TransferFrom(from, to std.Address, tokenId string) {
	caller := std.PrevRealm().Addr()
	mustBeValid(to)

	prevOwner := nft.update(to, tokenId, caller)
	if prevOwner != from {
		panic("GRC721: incorrect owner")
	}

	std.Emit("Transfer", "from", from.String(), "operator", to.String(), "tokenID", tokenId)
}

func (nft Token) Approve(to std.Address, tokenId string) {
	caller := std.PrevRealm().Addr()

	if caller == to {
		panic("GRC721: cannot approve yourself")
	}

	mustBeValid(to)
	nft.requireOwner(caller, tokenId)

	nft.tokenApprovals.Set(tokenId, to)
	std.Emit("Approval", "owner", caller.String(), "approved", to.String(), "tokenID", tokenId)
}

func (nft Token) SetApprovalForAll(operator std.Address, approved bool) {
	caller := std.PrevRealm().Addr()
	mustBeValid(operator)

	if caller == operator {
		panic("GRC721: cannot set operator to yourself")
	}

	nft.operatorApprovals.Set(operatorKey(caller, operator), approved)
	std.Emit("ApprovalForAll", "owner", caller.String(), "operator", operator.String(), "approved", strconv.FormatBool(approved))
}

func (nft Token) GetApproved(tokenId string) std.Address {
	_ = nft.mustBeOwned(tokenId)
	return nft.getApproved(tokenId)
}

func (nft Token) IsApprovedForAll(owner, operator std.Address) bool {
	approved, exists := nft.operatorApprovals.Get(operatorKey(owner, operator))
	if !exists || approved == false {
		return false
	}

	return true
}

func (nft Token) TokenURI(tokenId string) string {
	nft.mustBeOwned(tokenId)
	uri, exists := nft.tokenURIs.Get(tokenId)
	if !exists {
		return ""
	}

	return uri.(string)
}

func (nft Token) SetTokenURI(tokenId string, tokenURI string) string {
	nft.requireOwner(std.PrevRealm().Addr(), tokenId)
	nft.tokenURIs.Set(tokenId, tokenURI)
	return tokenURI
}

// Helpers
func (nft Token) requireOwner(caller std.Address, tokenId string) {
	if caller != nft.mustBeOwned(tokenId) {
		panic("GRC721: not owner")
	}
}

func (nft Token) getApproved(tokenId string) std.Address {
	approved, exists := nft.tokenApprovals.Get(tokenId)
	if !exists {
		return "" // panic instead?
	}

	return approved.(std.Address)
}

// mustBeValid panics if the given address is not valid
func mustBeValid(address std.Address) {
	if !address.IsValid() {
		err := ufmt.Sprintf("GRC721: invalid address %s", address)
		panic(err)
	}
}

// mustBeOwned panics if token is not owned by an address (does not exist)
// If the token is owned, mustBeOwned returns the owner of the token
func (nft Token) mustBeOwned(tokenId string) std.Address {
	owner, exists := nft.balances.Get(tokenId)
	if !exists {
		err := ufmt.Sprintf("GRC721: token with id %s does not exist", tokenId)
		panic(err)
	}

	return owner.(std.Address)
}

// checkAuthorized checks if spender is authorized to spend specified token on behalf of owner
// Panics if token doesn't exist, or if spender is not authorized in any way
func (nft Token) checkAuthorized(owner, spender std.Address, tokenId string) {
	_ = nft.mustBeOwned(tokenId)

	if !nft.isAuthorized(owner, spender, tokenId) {
		str := ufmt.Sprintf("GRC721: %s is not authorized for %s", spender, tokenId)
		panic(str)
	}
}

// isAuthorized returns if the spender is authorized to transfer the specified token
// Assumes addresses are valid and the token exists
func (nft Token) isAuthorized(owner, spender std.Address, tokenId string) bool {
	return owner == spender ||
		nft.IsApprovedForAll(owner, spender) ||
		nft.getApproved(tokenId) == owner
}

func (nft Token) update(to std.Address, tokenId string, auth std.Address) std.Address {
	owner := nft.mustBeOwned(tokenId)

	if auth != "" {
		nft.checkAuthorized(owner, auth, tokenId)
	}

	// Clear approval for this token
	nft.tokenApprovals.Set(tokenId, "")

	// Set new balances
	newOwnerBalance, _ := nft.balances.Get(owner.String())
	nft.balances.Set(owner.String(), newOwnerBalance.(int64)-1)

	toBalance, _ := nft.balances.Get(to.String())
	nft.balances.Set(to.String(), toBalance.(int64)+1)

	// Set new ownership
	nft.owners.Set(tokenId, to.String())

	return owner
}

// operatorKey is a helper to create the key for the operatorApproval tree
func operatorKey(owner, operator std.Address) string {
	return owner.String() + ":" + operator.String()
}
