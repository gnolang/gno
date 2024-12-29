package openocean

import (
	"std"
	"errors"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721"
)

type Collection struct {
	id string
	name string
	symbol string
	addrOwner std.Address
	description string
	logo string // ipfs path
	avaiableTokens uint64
	nftsBought uint64
	nfts *avl.Tree
	stars *avl.Tree
}

var (
	invalidAdrErr = errors.New("Invalid address")
	idNotFoundErr = errors.New("Couldn't find the id that you are searching for")
)

func (c *Collection) BalanceOf(owner std.Address) (uint64, bool) {
	var blc uint64 = 0

	c.nfts.Iterate("", "", func (id string, value interface{}) bool {
		nft := value.(*NFTtoken721)
		if nft.GetOwner() == owner {
			blc += 1
		}
		return false
	})
	if blc > 0 {
		return blc, true
	}
	return blc, false
}

func (c *Collection) OwnerOf(tokenId string) (std.Address, bool) {
	i, found := c.nfts.Get(tokenId)

	if found == false {
		panic(idNotFoundErr)
	}

	nft := i.(*NFTtoken721)
	return nft.GetOwner(), found
}

func (c *Collection) SafeTransferFrom(
	from std.Address,
	to std.Address,
	tokenId string,
)  {
	if to.IsValid() == false {
		panic(invalidAdrErr)
	}

	i, ok := c.nfts.Get(tokenId)

	if ok == false {
		panic(idNotFoundErr)
	}

	nft := i.(*NFTtoken721)
	nft.SetOwner(to)
}

func (c *Collection) SafeTransferFromWithData(
	from std.Address,
	to std.Address,
	tokenId string,
	data string,
) {
	if to.IsValid() == false {
		panic(invalidAdrErr)
	}

	i, ok := c.nfts.Get(tokenId)

	if ok == false {
		panic(idNotFoundErr)
	}

	nft := i.(*NFTtoken721)
	nft.SetOwner(to)
}

func (c *Collection) TransferFrom(
	from std.Address,
	to std.Address,
	tokenId string,
) {
	i, ok := c.nfts.Get(tokenId)

	if ok == false {
		panic(idNotFoundErr)
	}

	nft := i.(*NFTtoken721)
	nft.SetOwner(to)
}

func (c *Collection) Approve(to std.Address, tokenId string) error {
	caller := std.GetOrigCaller()

	if to.IsValid() == false {
		panic(invalidAdrErr)
	}

	nft, found := c.GetNFTByID(tokenId)
	
	if found == false {
		panic(idNotFoundErr)
	}
	return nft.Approve(caller, to)
}

func (c *Collection) GetApproved(tokenId string) (std.Address, error) {
	nft, found := c.GetNFTByID(tokenId)
	
	if found == false {
		panic(idNotFoundErr)
	}
	return nft.GetApproved()
}

func (c *Collection) SetApprovalForAll(tokenId string, operator std.Address, _approved std.Address) error {
	caller := std.GetOrigCaller()

	if _approved.IsValid() == false || operator.IsValid() == false {
		panic(invalidAdrErr)
	}

	nft, found := c.GetNFTByID(tokenId)
	
	if found == false {
		panic(idNotFoundErr)
	}
	return nft.SetApprovalForAll(caller, operator, _approved)
}

func (c *Collection) IsApprovedForAll(tokenId string, owner std.Address, operator std.Address) bool {
	if owner.IsValid() == false || operator.IsValid() == false {
		panic(invalidAdrErr)
	}

	nft, found := c.GetNFTByID(tokenId)
	
	if found == false {
		panic(idNotFoundErr)
	}
	return nft.IsApprovedForAll(owner, operator)
}

func (c *Collection) GetId() string {
	return c.id
}

func (c *Collection) GetName() string {
	return c.name
}

func (c *Collection) GetSymbol() string {
	return c.symbol
}

func (c *Collection) GetOwner() std.Address {
	return c.addrOwner
}

func (c *Collection) IncreateNFTSBought() {
	c.nftsBought += 1
}

func (c *Collection) GetNFTSBought() uint64 {
	return c.nftsBought
}

func (c *Collection) GetDescription() string {
	return c.description
}

func (c *Collection) GetLogo() string {
	return c.logo
}

func (c *Collection) GetNFTS() *avl.Tree {
	return c.nfts
}

func (c *Collection) GetAvaiableTokens() uint64 {
	return c.avaiableTokens
}

func (c *Collection) GetStars() *avl.Tree {
	return c.stars
}

func (c *Collection) getAvailableNFT() (*NFTtoken721, bool) {
	var nft *NFTtoken721
	var found bool = false

	c.nfts.Iterate("", "", func(id string, n interface{}) bool {
		nft = n.(*NFTtoken721)

		if nft.IsMinted() == false {
			found = true
			return true
		}
		return false
	})
	return nft, found
}

func (c *Collection) Mint(
	name string,
	cid string,
	description string,
	owner std.Address,
	price uint64,
) bool {
	if c.avaiableTokens <= 0 {
		panic(errors.New("Collection hasn`t any available tokens"))
	}
	c.avaiableTokens -= 1
	nft, found := c.getAvailableNFT()

	if found == false {
		panic(errors.New("There are no more available nfts"))
	}
	nft.Mint(name, cid, description, owner, price)
	return true
}

func (c *Collection) GetNFTByID(nftID string) (*NFTtoken721, bool)  {
	nft, found :=  c.nfts.Get(nftID)

	return nft.(*NFTtoken721), found
}

func GetCollectionByID(clts *avl.Tree, idToFind string) (*Collection, bool) {
	value, exists := clts.Get(idToFind)

	if exists == false {
		panic(idNotFoundErr)
	}
	return value.(*Collection), true
}

func (c *Collection) IsStarred() bool {
	addr := std.GetOrigCaller()
	_, exists := c.stars.Get(addr.String())
	return exists
}

func (c *Collection) Star() bool {
	addr := std.GetOrigCaller()
	if c.IsStarred() == true {
		return false
	}
	c.stars.Set(addr.String(), true)
	return true;
}

func (c *Collection) Unstar() bool {
	addr := std.GetOrigCaller()
	if c.IsStarred() == false {
		return false
	}
	c.stars.Remove(addr.String())
	return true
}

func CreateCollection(
	collectionId uint64,
	nftID *uint64,
	name string,
	symbol string,
	addrOwner std.Address,
	description string,
	logo string,
	avaiableTokens uint64,
) (*Collection, error) {
	var nfts *avl.Tree = avl.NewTree()
	var stars *avl.Tree = avl.NewTree()

	if addrOwner.IsValid() == false {
		return nil, errors.New("Invalid address")
	}
	for i := uint64(0); i < avaiableTokens; i += 1 {
		idStr := ufmt.Sprintf("%d", *nftID)
		tkn, err := InitToken(addrOwner, grc721.TokenID(idStr))
		if err != nil {
			panic(err)
		}
		nfts.Set(string(tkn.GetId()), tkn)
		*nftID += 1
	}
	idClt := ufmt.Sprintf("%d", collectionId)
	clt := &Collection{
		id: idClt,
		name: name,
		symbol: symbol,
		addrOwner: addrOwner,
		description: description,
		logo: logo,
		avaiableTokens: avaiableTokens,
		nftsBought: 0,
		nfts: nfts,
		stars: stars,
	}
	clt.Star()
	return clt, nil
}
