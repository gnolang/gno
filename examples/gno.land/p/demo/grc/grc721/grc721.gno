package grc721

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Collection struct {
	name              string
	symbol            string
	balances          *avl.Tree // std.Address > # of owned tokens
	operatorApprovals *avl.Tree // "OwnerAddress:OperatorAddress" -> bool
	tokenData         *avl.Tree // tokenID > *tokenData
}

type tokenData struct {
	owner    std.Address
	approved std.Address
	uri      string
}

var _ IGRC721 = (*Collection)(nil)

const emptyAddress = std.Address("")

func NewCollection(name, symbol string) *Collection {
	return &Collection{
		name:     name,
		symbol:   symbol,
		balances: avl.NewTree(),
		// give any addresses permissions for all owners' assets
		operatorApprovals: avl.NewTree(),
		tokenData:         avl.NewTree(),
	}
}

func (c Collection) Name() string   { return c.name }
func (c Collection) Symbol() string { return c.symbol }

func (c Collection) BalanceOf(owner std.Address) uint64 {
	mustBeValid(owner)

	balance, found := c.balances.Get(owner.String())
	if !found {
		return 0
	}

	return balance.(uint64)
}

func (c Collection) OwnerOf(tokenId string) std.Address {
	return c.mustBeOwned(tokenId)
}

func (c *Collection) TransferFrom(from, to std.Address, tokenId string) {
	caller := std.PrevRealm().Addr()
	mustBeValid(to)

	prevOwner := c.update(to, tokenId, caller)
	if prevOwner != from {
		panic("GRC721: incorrect owner")
	}
}

func (c *Collection) Approve(to std.Address, tokenId string) {
	caller := std.PrevRealm().Addr()

	if caller == to {
		panic("GRC721: cannot approve yourself")
	}

	mustBeValid(to)
	c.approve(to, tokenId, caller, true)
}

func (c *Collection) SetApprovalForAll(operator std.Address, approved bool) {
	caller := std.PrevRealm().Addr()
	mustBeValid(operator)

	if caller == operator {
		panic("GRC721: cannot set operator to yourself")
	}

	c.operatorApprovals.Set(operatorKey(caller, operator), approved)

	if approved {
		std.Emit("ApprovalForAll",
			"owner", caller.String(),
			"operator", operator.String(),
			"approved", "true",
		)
	} else {
		std.Emit("ApprovalForAll",
			"owner", caller.String(),
			"operator", operator.String(),
			"approved", "false", // We do not support strconv.FormatBool yet
		)
	}
}

func (c Collection) GetApproved(tokenId string) std.Address {
	_ = c.mustBeOwned(tokenId)
	return c.getApproved(tokenId)
}

func (c *Collection) IsApprovedForAll(owner, operator std.Address) bool {
	approved, exists := c.operatorApprovals.Get(operatorKey(owner, operator))
	if !exists || approved == false {
		return false
	}

	return true
}

func (c Collection) TokenURI(tokenId string) string {
	c.mustBeOwned(tokenId)
	rawData, _ := c.tokenData.Get(tokenId)
	return rawData.(*tokenData).uri
}

func (c *Collection) SetTokenURI(tokenId string, tokenURI string) string {
	rawData, exists := c.tokenData.Get(tokenId)
	if !exists {
		// Instantiate token data if it does not exist
		c.tokenData.Set(tokenId, &tokenData{
			uri: tokenURI,
		})
	} else {
		// Update URI
		data := rawData.(*tokenData)
		data.uri = tokenURI
		c.tokenData.Set(tokenId, data)
	}

	return tokenURI
}

func (c *Collection) Mint(to std.Address, tokenId string) {
	mustBeValid(to)
	prevOwner := c.update(to, tokenId, emptyAddress)
	if prevOwner != emptyAddress {
		str := ufmt.Sprintf("GRC721: token with id %s has already been minted", tokenId)
		panic(str)
	}
}

func (c *Collection) Burn(tokenId string) {
	prevOwner := c.update(emptyAddress, tokenId, emptyAddress)

	if prevOwner == emptyAddress {
		str := ufmt.Sprintf("GRC721: Token with ID %s does not exist", tokenId)
		panic(str)
	}
}

// Helpers
func (c Collection) requireOwner(caller std.Address, tokenId string) {
	if caller != c.mustBeOwned(tokenId) {
		panic("GRC721: not owner")
	}
}

func (c Collection) getApproved(tokenId string) std.Address {
	rawData, exists := c.tokenData.Get(tokenId)
	if !exists {
		err := ufmt.Sprintf("GRC721: token with ID %s does not exist", tokenId)
		panic(err)
	}

	return rawData.(*tokenData).approved
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
func (c Collection) mustBeOwned(tokenId string) std.Address {
	rawData, exists := c.tokenData.Get(tokenId)

	if !exists {
		err := ufmt.Sprintf("GRC721: token with ID %s does not exist", tokenId)
		panic(err)
	}

	return rawData.(*tokenData).owner
}

// checkAuthorized checks if spender is authorized to spend specified token on behalf of owner
// Panics if token doesn't exist, or if spender is not authorized in any way
func (c Collection) checkAuthorized(owner, spender std.Address, tokenId string) {
	_ = c.mustBeOwned(tokenId)

	if !c.isAuthorized(owner, spender, tokenId) {
		str := ufmt.Sprintf("GRC721: %s is not authorized for %s", spender, tokenId)
		panic(str)
	}
}

// isAuthorized returns if the spender is authorized to transfer the specified token
// Assumes addresses are valid and the token exists
func (c Collection) isAuthorized(owner, spender std.Address, tokenId string) bool {
	return owner == spender ||
		c.IsApprovedForAll(owner, spender) ||
		c.getApproved(tokenId) == owner
}

func (c *Collection) update(to std.Address, tokenId string, auth std.Address) std.Address {
	from := c.ownerOf(tokenId)

	if auth != emptyAddress {
		c.checkAuthorized(from, auth, tokenId)
	}

	// If token exists
	if from != emptyAddress {
		// Clear approval for this token
		c.approve(emptyAddress, tokenId, emptyAddress, false)

		// Set new balances
		ownerNewBalance, _ := c.balances.Get(from.String())
		c.balances.Set(from.String(), ownerNewBalance.(uint64)-1)
	}

	if to != emptyAddress {
		toBalance, initialized := c.balances.Get(to.String())
		if !initialized {
			c.balances.Set(to.String(), uint64(1))
		} else {
			c.balances.Set(to.String(), toBalance.(uint64)+1)
		}

		rawData, initialized := c.tokenData.Get(tokenId)
		if !initialized {
			// Initialize new tokenData
			c.tokenData.Set(tokenId, &tokenData{
				owner: to,
			})
		} else {
			// If tokenData exists, change only what's needed
			data := rawData.(*tokenData)
			data.owner = to
			c.tokenData.Set(tokenId, data)
		}
	} else {
		// Burn case
		// Remove token data completely
		c.tokenData.Remove(tokenId)
	}

	std.Emit("Transfer",
		"from", from.String(),
		"to", to.String(),
		"tokenID", tokenId,
	)

	return from
}

func (c *Collection) approve(to std.Address, tokenId string, auth std.Address, emitEvent bool) {
	if emitEvent || auth != emptyAddress {
		owner := c.mustBeOwned(tokenId)

		if auth != emptyAddress && owner != auth && !c.IsApprovedForAll(owner, auth) {
			panic("GRC721: invalid approver")
		}

		if emitEvent {
			std.Emit("Approval",
				"owner", owner.String(),
				"approved", to.String(),
				"tokenID", tokenId,
			)
		}
	}

	// No need for exist check, done by mustBeOwned already
	rawData, _ := c.tokenData.Get(tokenId)
	data := rawData.(*tokenData)
	data.approved = to

	c.tokenData.Set(tokenId, data)
}

func (c Collection) ownerOf(tokenId string) std.Address {
	data, exists := c.tokenData.Get(tokenId)
	if !exists {
		return emptyAddress
	}

	return data.(*tokenData).owner
}

// operatorKey is a helper to create the key for the operatorApproval tree
func operatorKey(owner, operator std.Address) string {
	return owner.String() + ":" + operator.String()
}
