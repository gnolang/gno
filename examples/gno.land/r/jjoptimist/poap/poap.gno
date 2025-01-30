package poap

import (
	"strconv"
	"time"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
)

type POAP struct {
	id          uint64    // Unique identifier
	name        string    // Event name
	description string    // Event description
	date        time.Time // Event date
	maxMints    int       // Maximum number of POAPs
	mintCount   int       // Current mint count
}

var (
	poaps     *avl.Tree                            // Store POAPs
	poapCount uint64                               // Counter for POAP IDs
	nftTokens = grc721.NewBasicNFT("POAP", "POAP") // NFT implementation
)

func init() {
	poaps = avl.NewTree()
}

func CreatePOAP(name, description string, date time.Time, maxMints int) uint64 {
	if maxMints <= 0 {
		panic("Maximum mints must be greater than 0")
	}

	poapCount++
	poap := POAP{
		id:          poapCount,
		name:        name,
		description: description,
		date:        date,
		maxMints:    maxMints,
		mintCount:   0,
	}

	poaps.Set(strconv.FormatUint(poapCount, 10), poap)
	return poapCount
}

// retrieve POAP by ID
func GetPOAP(id uint64) (POAP, bool) {
	value, exists := poaps.Get(strconv.FormatUint(id, 10))
	if !exists {
		return POAP{}, false
	}
	return value.(POAP), true
}

func MintPOAP(poapId uint64) {
	poap, exists := GetPOAP(poapId)
	if !exists {
		panic("POAP does not exist")
	}
	
	if poap.mintCount >= poap.maxMints {
		panic("POAP minting limit reached")
	}
	
	caller := std.PrevRealm().Addr()
	tokenId := grc721.TokenID(strconv.FormatUint(poapId, 10) + "-" + strconv.Itoa(poap.mintCount+1))
	nftTokens.Mint(caller, tokenId)
	
	poap.mintCount++
	poaps.Set(strconv.FormatUint(poapId, 10), poap)
}
