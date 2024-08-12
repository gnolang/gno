package openr

import (
	"std"
	"errors"
	"gno.land/p/demo/avl"
	"gno.land/p/molaryy/openocean"
	"gno.land/p/demo/json"
)

// change to another address
const (
	openoceanAdr std.Address = "g125fhetg9kc3surkxh78avppegc3997l3jr4wdf"
)

var (
	collections *avl.Tree
	collectionCounter uint64
	nftID uint64
)

func init() {
	collections = avl.NewTree()
}

var count int

func AddCollection(
	name string,
	symbol string,
	description string,
	logo string,
	avaiableTokens uint64,
) {
	caller := std.GetOrigCaller()
	coinSent := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	realmAddr := std.CurrentRealm().Addr()

	if coinSent.AmountOf("ugnot") < 5 {
		panic(errors.New("You don't have enough founds to create a collection, it costs 5 ugnots and you have " + string(coinSent.AmountOf("ugnot")) + " ugnots"))
	}

	clt, err := openocean.CreateCollection(
		collectionCounter,
		&nftID,
		name,
		symbol,
		caller,
		description,
		logo,
		avaiableTokens,
	)

	if clt == nil || err != nil {
		panic(err)
	}

	banker.SendCoins(realmAddr, openoceanAdr, std.Coins{coinSent[0]})
	id := clt.GetId()
    collections.Set(id, clt)
	collectionCounter++
}

func GetCollections() string {
	cltsJson := json.ArrayNode("", []*json.Node{})

	collections.Iterate("", "", func (id string, value interface{}) bool {
		clt := value.(*openocean.Collection)

		if clt == nil {
			return true
		}

		nfts := clt.GetNFTS()
		nftsJson := json.ArrayNode("", []*json.Node{})
		cltField := json.ObjectNode("", map[string]*json.Node{
			"id": json.StringNode("id", clt.GetId()),
			"name": json.StringNode("name", clt.GetName()),
			"logo": json.StringNode("logo", clt.GetLogo()),
			"owner": json.StringNode("owner", clt.GetOwner().String()),
			"description": json.StringNode("description", clt.GetDescription()),
			"volume": json.NumberNode("volume", float64(clt.GetAvaiableTokens())),
			"sales": json.NumberNode("sales", float64(clt.GetNFTSBought())),
		})
		nfts.Iterate("", "", func (id string, valueNft interface{}) bool {
			nft := valueNft.(* openocean.NFTtoken721)
			nftField := json.ObjectNode("", map[string]*json.Node{
				"id": json.StringNode("id",string(nft.GetId())),
				"owner": json.StringNode("owner", nft.GetOwner().String()),
				"isMinted": json.BoolNode("isMinted", nft.IsMinted()),
				"price": json.NumberNode("price", float64(nft.GetPrice())),
				"isForSale": json.BoolNode("isForSale", nft.IsForSale()),
				"collectionId": json.StringNode("collectionId", clt.GetId()),
			})
			cltField.AppendObject(string(nft.GetId()), nftField)
			metadata := nft.GetMetadata()

			if metadata != "" {
				metadataJson, err := json.Unmarshal([]byte(metadata))
				if err != nil {
					panic(err)
				}
				nftField.AppendObject("metadata", metadataJson)
			}

			nftsJson.AppendArray(nftField)
			return false
		})
		cltField.AppendObject("nfts", nftsJson)
		starsJson := json.ArrayNode("", []*json.Node{})
		stars := clt.GetStars()

		stars.Iterate("", "", func (id string, value interface{}) bool {
			starsJson.AppendArray(json.StringNode("", id))
			return false
		})

		cltField.AppendObject("stars", starsJson)
		cltsJson.AppendArray(cltField)
		return false
	})
	encoded, err := json.Marshal(cltsJson)

	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func MintInCollectionById(
	collectionID string,
	nftName string,
	cid string,
	description string,
	price uint64,
) bool {
	clt, ok := openocean.GetCollectionByID(collections, collectionID)
	caller := std.GetOrigCaller()

	if ok == false {
		panic(errors.New("Couldn't find collection: " + collectionID))
	}

	if caller != clt.GetOwner() {
		panic(errors.New("You are not the owner of the collection"))
	}

	return clt.Mint(nftName, cid, description, caller, price)
}

func BuyNFT(collectionID string, nftID string) {
	caller := std.GetOrigCaller()
	clt, ok := openocean.GetCollectionByID(collections, collectionID)
	coinSent := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	realmAddr := std.CurrentRealm().Addr()

	if ok == false {
		panic(errors.New("Couldn't find collection: " + collectionID))
	}
	
	nft, ok := clt.GetNFTByID(nftID)

	if ok == false {
		panic(errors.New("Couldn't find an nft with this iD"))
	}

	if nft.GetOwner() == caller {
		panic(errors.New("You already have this nft: " + nftID))
	}

	if nft.IsForSale() == false {
		panic(errors.New("The nft: " + nftID + " is not for sale"))
	}


	if nft.GetPrice() > uint64(coinSent.AmountOf("ugnot")) {
		panic(errors.New("You don't have enough ugnots to buy the nft: " + nftID))
	}

	clt.IncreateNFTSBought()
	nft.SetOwner(caller)
	nft.SetForSale(false)
	banker.SendCoins(realmAddr, clt.GetOwner(), std.Coins{coinSent[0]})
}

func SetNFTForSale(collectionID string, nftID string, sellingState bool) {
	caller := std.GetOrigCaller()
	clt, ok := openocean.GetCollectionByID(collections, collectionID)

	if ok == false {
		panic(errors.New("Couldn't find collection: " + collectionID))
	}

	nft, ok := clt.GetNFTByID(nftID)

	if ok == false {
		panic(errors.New("Couldn't find nft:" + nftID + " in collection: " + collectionID))
	}

	if nft.GetOwner() != caller {
		panic(errors.New("You can't set the sell state of an NFT if you are not the owner!"))
	}

	if nft.IsForSale() == sellingState {
		panic(errors.New("This NFT is already has the selling state that you are trying to set"))
	}

	nft.SetForSale(sellingState)
}

func GetCollectionStats(collectionID string) string {
	clt, ok := openocean.GetCollectionByID(collections, collectionID)

	if ok == false {
		panic(errors.New("Couldn't find collection: " + collectionID))
	}

	collectionJSON := json.ObjectNode("", map[string]*json.Node{
		"id": json.StringNode("id", clt.GetId()),
		"name": json.StringNode("name", clt.GetName()),
		"logo": json.StringNode("logo", clt.GetLogo()),
		"volume": json.NumberNode("volume", float64(clt.GetAvaiableTokens())),
		"sales": json.NumberNode("sales", float64(clt.GetNFTSBought())),
	})

	nftsJson := json.ArrayNode("", []*json.Node{})
	nfts := clt.GetNFTS()
	nfts.Iterate("", "", func (id string, value interface{}) bool {
		nft := value.(* openocean.NFTtoken721)

		if nft.IsMinted() == false {
			return false
		}

		nftJSON := json.ObjectNode("", map[string]*json.Node{
			"id": json.StringNode("id", string(nft.GetId())),
			"name": json.StringNode("name", nft.GetName()),
			"price": json.NumberNode("price", float64(nft.GetPrice())),
		})
		nftsJson.AppendArray(nftJSON)
		return false
	})
	collectionJSON.AppendObject("nfts", nftsJson)

	encoded, err := json.Marshal(collectionJSON)

	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func StarCollection(collectionID string, starred string) {
	clt, ok := openocean.GetCollectionByID(collections, collectionID)

	if ok == false {
		panic(errors.New("Couldn't find collection: " + collectionID))
	}

	if starred == "true" {
		clt.Star()
	} else if starred == "false" {
		clt.Unstar()
	}
}

func Render(path string) string {
	return GetCollections()
}
