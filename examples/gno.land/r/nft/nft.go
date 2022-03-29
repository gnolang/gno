package nft

import (
	"std"
	"strconv"

	"gno.land/p/avl"
)

//----------------------------------------
// types

type TokenID string

type GRC721 interface {
	BalanceOf(owner std.Address) (count int64)
	OwnerOf(tid TokenID) std.Address
	SafeTransferFrom(from, to std.Address, tid TokenID)
	TransferFrom(from, to std.Address, tid TokenID)
	Approve(approved std.Address, tid TokenID)
	SetApprovalForAll(operator std.Address, approved bool)
	GetApproved(tid TokenID) std.Address
	IsApprovedForAll(owner, operator std.Address) bool
}

// TODO use
type TransferEvent struct {
	From    std.Address
	To      std.Address
	TokenID TokenID
}

// TODO use
type ApprovalEvent struct {
	Owner    std.Address
	Approved std.Address
	TokenID  TokenID
}

// TODO use
type ApprovalForAllEvent struct {
	Owner    std.Address
	Operator std.Address
	Approved bool
}

type NFToken struct {
	owner    std.Address
	approved std.Address
	tokenID  TokenID
	data     string
}

//----------------------------------------
// impl

type grc721 struct {
	tokenCounter int
	tokens       *avl.Tree // TokenID -> *NFToken{}
	operators    *avl.Tree // owner std.Address -> operator std.Address
}

var gGRC721 = &grc721{}

func GetGRC721() *grc721 { return gGRC721 }

func (grc *grc721) nextTokenID() TokenID {
	grc.tokenCounter++
	s := strconv.Itoa(grc.tokenCounter)
	return TokenID(s)
}

func (grc *grc721) getToken(tid TokenID) (*NFToken, bool) {
	_, token, ok := grc.tokens.Get(tid)
	if !ok {
		return nil, false
	}
	return token.(*NFToken), true
}

func (grc *grc721) Mint(to std.Address, data string) TokenID {
	tid := grc.nextTokenID()
	newTokens, _ := grc.tokens.Set(tid, &NFToken{
		owner:   to,
		tokenID: tid,
		data:    data,
	})
	grc.tokens = newTokens
	return tid
}

func (grc *grc721) BalanceOf(owner std.Address) (count int64) {
	panic("not yet implemented")
}

func (grc *grc721) OwnerOf(tid TokenID) std.Address {
	token, ok := grc.getToken(tid)
	if !ok {
		panic("token does not exist")
	}
	return token.owner
}

// XXX not fully implemented yet.
func (grc *grc721) SafeTransferFrom(from, to std.Address, tid TokenID) {
	grc.TransferFrom(from, to, tid)
	// When transfer is complete, this function checks if `_to` is a smart
	// contract (code size > 0). If so, it calls `onERC721Received` on
	// `_to` and throws if the return value is not
	// `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`.
	// XXX ensure "to" is a realm with onERC721Received() signature.
}

func (grc *grc721) TransferFrom(from, to std.Address, tid TokenID) {
	caller := std.GetCallerAt(2)
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	// Throws unless `msg.sender` is the current owner, an authorized
	// operator, or the approved address for this NFT.
	if caller != token.owner && caller != token.approved {
		_, operator, ok := grc.operators.Get(token.owner)
		if !ok || caller != operator.(std.Address) {
			panic("unauthorized")
		}
	}
	// Throws if `_from` is not the current owner.
	if from != token.owner {
		panic("from is not the current owner")
	}
	// Throws if `_to` is the zero address.
	if to == "" {
		panic("to cannot be empty")
	}
	// Good.
	token.owner = to
}

func (grc *grc721) Approve(approved std.Address, tid TokenID) {
	caller := std.GetCallerAt(2)
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	// Throws unless `msg.sender` is the current owner,
	// or an authorized operator.
	if caller != token.owner {
		_, operator, ok := grc.operators.Get(token.owner)
		if !ok || caller != operator.(std.Address) {
			panic("unauthorized")
		}
	}
	// Good.
	token.approved = approved
}

// XXX make it work for set of operators.
func (grc *grc721) SetApprovalForAll(operator std.Address, approved bool) {
	caller := std.GetCallerAt(2)
	newOperators, _ := grc.operators.Set(caller, operator)
	grc.operators = newOperators
}

func (grc *grc721) GetApproved(tid TokenID) std.Address {
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	return token.approved
}

// XXX make it work for set of operators
func (grc *grc721) IsApprovedForAll(owner, operator std.Address) bool {
	_, operator2, ok := grc.operators.Get(owner)
	if !ok {
		return false
	}
	return operator == operator2.(std.Address)
}
