package poap

import (
	"strconv"
	"time"
	"std"
	"errors"

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

type token struct {
	tokenCounter int
	tokens       avl.Tree // grc721.TokenID -> *NFToken{}
}

type NFToken struct {
	Owner    std.Address
	Approved std.Address
	TokenID  grc721.TokenID
	Data     string
}

func (grc *token) Mint(to std.Address, tokenId grc721.TokenID) {
	grc.tokens.Set(string(tokenId), &NFToken{
		Owner:   to,
		TokenID: tokenId,
	})
}

func (grc *token) OwnerOf(tokenId grc721.TokenID) (std.Address, error) {
	token, exists := grc.tokens.Get(string(tokenId))
	if !exists {
		return "", errors.New("token not found")
	}
	nft := token.(*NFToken)
	return nft.Owner, nil
}

func (grc *token) BalanceOf(owner std.Address) int64 {
	count := int64(0)
	grc.tokens.Iterate("", "", func(key string, value interface{}) bool {
		nft := value.(*NFToken)
		if nft.Owner == owner {
			count++
		}
		return false
	})
	return count
}

func (grc *token) GetApproved(tokenId grc721.TokenID) (std.Address, error) {
	token, exists := grc.tokens.Get(string(tokenId))
	if !exists {
		return "", errors.New("token not found")
	}
	nft := token.(*NFToken)
	return nft.Approved, nil
}

func (grc *token) IsApprovedForAll(owner, operator std.Address) bool {
	return false // For now, we don't support operator approvals
}

var (
	poaps     *avl.Tree                            // Store POAPs
	poapCount uint64                               // Counter for POAP IDs
	nftTokens *token                              // NFT implementation
)

func init() {
	poaps = avl.NewTree()
	nftTokens = &token{
	tokens: *avl.NewTree(),
	}
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

// MintPOAP mints a POAP token for the caller
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
	
	// Mint NFT
	nftTokens.Mint(caller, tokenId)
	
	// Update mint count
	poap.mintCount++
	poaps.Set(strconv.FormatUint(poapId, 10), poap)
}

func (grc *token) Name() string {
	return "POAP"
}

func (grc *token) Symbol() string {
	return "POAP"
}

func (grc *token) TokenURI(id grc721.TokenID) string {
	return string(id)
}
