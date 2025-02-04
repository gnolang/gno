package poap

import (
	"strconv"
	"time"
	"std"
	"errors"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
)

type POAP struct {
	ID          uint64    // Unique identifier
	Name        string    // Event name
	Description string    // Event description
	Date        time.Time // Event date
	MaxMints    int       // Maximum number of POAPs
	MintCount   int       // Current mint count
}

type token struct {
	TokenCounter int
	Tokens       *avl.Tree 
}

type NFToken struct {
	Owner    std.Address
	Approved std.Address
	TokenID  grc721.TokenID
	Data     string
}

func (grc *token) Mint(to std.Address, tokenId grc721.TokenID) {
	grc.Tokens.Set(string(tokenId), &NFToken{
		Owner:   to,
		TokenID: tokenId,
	})
}

func (grc *token) OwnerOf(tokenId grc721.TokenID) (std.Address, error) {
	token, exists := grc.Tokens.Get(string(tokenId))
	if !exists {
		return "", errors.New("token not found")
	}
	nft := token.(*NFToken)
	return nft.Owner, nil
}

func (grc *token) BalanceOf(owner std.Address) int64 {
	count := int64(0)
	grc.Tokens.Iterate("", "", func(key string, value interface{}) bool {
		nft := value.(*NFToken)
		if nft.Owner == owner {
			count++
		}
		return false
	})
	return count
}

func (grc *token) GetApproved(tokenId grc721.TokenID) (std.Address, error) {
	token, exists := grc.Tokens.Get(string(tokenId))
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
		Tokens: avl.NewTree(),
	}
}

func CreatePOAP(name, description string, date time.Time, maxMints int) uint64 {
	if maxMints <= 0 {
		panic("Maximum mints must be greater than 0")
	}

	poapCount++
	poap := POAP{
		ID:          poapCount,
		Name:        name,
		Description: description,
		Date:        date,
		MaxMints:    maxMints,
		MintCount:   0,
	}

	poaps.Set(strconv.FormatUint(poapCount, 10), &poap)
	return poapCount
}

// retrieve POAP by ID
func GetPOAP(id uint64) (POAP, bool) {
	value, exists := poaps.Get(strconv.FormatUint(id, 10))
	if !exists {
		return POAP{}, false
	}
	return *(value.(*POAP)), true
}

// MintPOAP mints a POAP token for the caller
func MintPOAP(poapId uint64) {
	poap, exists := GetPOAP(poapId)
	if !exists {
		panic("POAP does not exist")
	}
	
	if poap.MintCount >= poap.MaxMints {
		panic("POAP minting limit reached")
	}
	
	caller := std.PrevRealm().Addr()
	tokenId := grc721.TokenID(strconv.FormatUint(poapId, 10) + "-" + strconv.Itoa(poap.MintCount+1))
	
	// Mint NFT
	nftTokens.Mint(caller, tokenId)
	
	// Update mint count
	poap.MintCount++
	poaps.Set(strconv.FormatUint(poapId, 10), &poap)
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

// GetPOAPsByOwner returns all POAPs owned by an address
func GetPOAPsByOwner(owner std.Address) []POAP {
	poaps := make([]POAP, 0)
	nftTokens.Tokens.Iterate("", "", func(key string, value interface{}) bool {
		nft := value.(*NFToken)
		if nft.Owner == owner {
			poapIdStr := strings.Split(string(nft.TokenID), "-")[0]
			poapId, _ := strconv.ParseUint(poapIdStr, 10, 64)
			if poap, exists := GetPOAP(poapId); exists {
				poaps = append(poaps, poap)
			}
		}
		return false
	})
	return poaps
}

// ListPOAPs returns all POAPs in the system
func ListPOAPs() []POAP {
	poapList := make([]POAP, 0)
	poaps.Iterate("", "", func(key string, value interface{}) bool {
		poap := value.(*POAP)
		poapList = append(poapList, *poap)
		return false
	})
	return poapList
}

// GetPOAPTokens returns all NFT tokens for a specific POAP
func GetPOAPTokens(poapId uint64) []NFToken {
	tokens := make([]NFToken, 0)
	nftTokens.Tokens.Iterate("", "", func(key string, value interface{}) bool {
		nft := value.(*NFToken)
		poapIdStr := strings.Split(string(nft.TokenID), "-")[0]
		tokenPoapId, _ := strconv.ParseUint(poapIdStr, 10, 64)
		if tokenPoapId == poapId {
			tokens = append(tokens, *nft)
		}
		return false
	})
	return tokens
}
