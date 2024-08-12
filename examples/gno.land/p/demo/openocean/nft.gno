package openocean

import (
	"std"
	"errors"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/json"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/avl"
)

type NFTtoken721 struct {
	id 			   grc721.TokenID
	name 		   string
	owner 		   std.Address
	metadata 	   string
	isMinted 	   bool
	isForSale 	   bool
	price		   uint64
	tokenApprovals *avl.Tree
}

func (n *NFTtoken721) GetId() grc721.TokenID {
	return n.id
}

func (n *NFTtoken721) GetName() string {
	return n.name
}

func (n *NFTtoken721) GetOwner() std.Address {
	return n.owner
}

func (n *NFTtoken721) GetMetadata() string { 
	return n.metadata
}

func (n *NFTtoken721) GetPrice() uint64 {
	return n.price
}

func (n *NFTtoken721) SetForSale(state bool) {
	n.isForSale = state
}

func (n *NFTtoken721) IsForSale() bool {
	return n.isForSale
}

func (n *NFTtoken721) SetOwner(owner std.Address) bool {
	if owner.IsValid() == false {
		return false
	}
	n.owner = owner
	return true
}

func (n *NFTtoken721) IsMinted() bool {
	return n.isMinted
}

func (n *NFTtoken721) Mint(
	name string,
	cid string,
	description string,
	owner std.Address,
	price uint64,
) bool {
	if n.isMinted == true {
		panic(errors.New(ufmt.Sprintf("NFT: %d has already been minted!", n.id)))
	}
	nftJson := json.ObjectNode("", map[string]*json.Node{
		"owner": json.StringNode("owner", owner.String()),
		"name": json.StringNode("name", name),
		"cid": json.StringNode("cid", cid),
		"description": json.StringNode("description", description),
	})

	encoded, err := json.Marshal(nftJson)

	if err != nil {
		panic(err)
	}
	n.metadata = string(encoded)
	n.isMinted = true
	n.owner = owner
	n.name = name
	n.price = price
	n.isForSale = true
	return true
}

func (n *NFTtoken721) Approve(caller std.Address, to std.Address) error  {
	if !n.isMinted {
		return errors.New("token does not exist")
	}

	if caller != n.owner && !n.IsApprovedForAll(n.owner, caller) {
		return errors.New("caller is not owner nor approved for all")
	}
	n.tokenApprovals.Set(string(n.GetId()), to)
	return nil
}

func (n *NFTtoken721) GetApproved() (std.Address, error) {
	var adr std.Address

	if !n.isMinted {
		return adr, errors.New("token does not exist")
	}

	approvedAddr, exists := n.tokenApprovals.Get(string(n.GetId()))
	if !exists {
		return adr, errors.New("no approved address for this token")
	}

	return approvedAddr.(std.Address), nil
}

func (n *NFTtoken721) SetApprovalForAll(caller std.Address, operator std.Address, _approved std.Address) error {
	if caller != n.owner {
		return errors.New("caller is not the owner")
	}

	n.tokenApprovals.Set(operator.String(), _approved)
	return nil
}

func (n *NFTtoken721) IsApprovedForAll(owner std.Address, operator std.Address) bool {
	approved, exists := n.tokenApprovals.Get(operator.String())

	if !exists {
		return false
	}
	return approved.(bool)
}

func InitToken(
	addrOwner std.Address, 
	nftId grc721.TokenID,
) (*NFTtoken721, error) {
	if addrOwner.IsValid() == false {
		return nil, errors.New("Invalid owner address")
	}
	return &NFTtoken721 {
		id: nftId,
		name: "",
		owner: addrOwner,
		metadata: "",
		isMinted: false,
		isForSale: true,
		price: 0,
		tokenApprovals: nil,
	}, nil
}
