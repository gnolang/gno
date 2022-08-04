package nft

import (
	"std"
	"strconv"

	"gno.land/p/avl"
	"gno.land/p/grc/grc721"
)

type grc721 struct {
	grc721.IGRC721 // implements the GRC721 interface

	tokenCounter int
	tokens       *avl.Tree // igrc721.TokenID -> *NFToken{}
	operators    *avl.Tree // owner std.Address -> operator std.Address
}

type NFToken struct {
	owner    std.Address
	approved std.Address
	tokenID  igrc721.TokenID
	data     string
}

var gGRC721 = &grc721{}

func GetGRC721() *grc721 { return gGRC721 }

func (grc *grc721) nextTokenID() igrc721.TokenID {
	grc.tokenCounter++
	s := strconv.Itoa(grc.tokenCounter)
	return igrc721.TokenID(s)
}

func (grc *grc721) getToken(tid igrc721.TokenID) (*NFToken, bool) {
	_, token, ok := grc.tokens.Get(string(tid))
	if !ok {
		return nil, false
	}
	return token.(*NFToken), true
}

func (grc *grc721) Mint(to std.Address, data string) igrc721.TokenID {
	tid := grc.nextTokenID()
	newTokens, _ := grc.tokens.Set(string(tid), &NFToken{
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

func (grc *grc721) OwnerOf(tid igrc721.TokenID) std.Address {
	token, ok := grc.getToken(tid)
	if !ok {
		panic("token does not exist")
	}
	return token.owner
}

// XXX not fully implemented yet.
func (grc *grc721) SafeTransferFrom(from, to std.Address, tid igrc721.TokenID) {
	grc.TransferFrom(from, to, tid)
	// When transfer is complete, this function checks if `_to` is a smart
	// contract (code size > 0). If so, it calls `onERC721Received` on
	// `_to` and throws if the return value is not
	// `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`.
	// XXX ensure "to" is a realm with onERC721Received() signature.
}

func (grc *grc721) TransferFrom(from, to std.Address, tid igrc721.TokenID) {
	caller := std.GetCallerAt(2)
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	// Throws unless `msg.sender` is the current owner, an authorized
	// operator, or the approved address for this NFT.
	if caller != token.owner && caller != token.approved {
		_, operator, ok := grc.operators.Get(token.owner.String())
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

func (grc *grc721) Approve(approved std.Address, tid igrc721.TokenID) {
	caller := std.GetCallerAt(2)
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	// Throws unless `msg.sender` is the current owner,
	// or an authorized operator.
	if caller != token.owner {
		_, operator, ok := grc.operators.Get(token.owner.String())
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
	newOperators, _ := grc.operators.Set(caller.String(), operator)
	grc.operators = newOperators
}

func (grc *grc721) GetApproved(tid igrc721.TokenID) std.Address {
	token, ok := grc.getToken(tid)
	// Throws if `_tokenId` is not a valid NFT.
	if !ok {
		panic("token does not exist")
	}
	return token.approved
}

// XXX make it work for set of operators
func (grc *grc721) IsApprovedForAll(owner, operator std.Address) bool {
	_, operator2, ok := grc.operators.Get(owner.String())
	if !ok {
		return false
	}
	return operator == operator2.(std.Address)
}
