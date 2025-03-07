package gnoxchange

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
)

// PoolRegistry is a registry for swap pools. It is a singleton that is initialized in the init function.
// It is private to only this realm and cannot and should not be accessed in other realms.
type PoolRegistry struct {
	pools *avl.Tree // key is "tokenA:tokenB", value is *Pool
}

// Pool is a struct that represents a swap pool. It contains the two tokens in the pool,
// the reserves of each token, the liquidity provider token, and the total supply of the liquidity provider token.
// None of the pool's fields are to be accessed externally, only through the PoolRegistry in the gnoXchange realm.
type Pool struct {
	tokenA        string // tokenhub tokenKey
	tokenB        string // tokenhub tokenKey
	reserveA      uint64
	reserveB      uint64
	lpToken       *grc20.Token
	lpLedger      *grc20.PrivateLedger
	totalSupplyLp uint64
}

// TokenInfo represents the public information about a token
type TokenInfo struct {
	Path     string
	Name     string
	Symbol   string
	Decimals uint
}

// PoolInfo represents the public information about a pool
type PoolInfo struct {
	PoolKey       string
	TokenAInfo    TokenInfo
	TokenBInfo    TokenInfo
	ReserveA      uint64
	ReserveB      uint64
	TotalSupplyLP uint64
}

// Ticket is a struct that represents a swap ticket.
// It holds information about the available swaps
type Ticket struct {
	ID           string
	Creator      std.Address
	AssetIn      Asset
	AssetOut     Asset
	AmountIn     uint64
	MinAmountOut uint64
	CreatedAt    time.Time
	ExpiresAt    time.Time
	Status       string
}

// TicketRegisty is consisted of 3 trees only for better gas efficiency.
// Most of the time, user would only want to see open tickets or nfts currently for sale.
// There is no need to iterate over all tickets everry time.
type TicketRegistry struct {
	allTickets     *avl.Tree // all tickets (including NFTs)
	openTickets    *avl.Tree // open non-NFT tickets
	openNFTTickets *avl.Tree // open NFT tickets
}

// Types of assets
type AssetType uint8

const (
	AssetTypeCoin  AssetType = iota // Native coins (like ugnot)
	AssetTypeToken                  // GRC20 tokens
	AssetTypeNFT                    // GRC721 NFTs
)

// Asset is a struct shared between coins, tokens and NFTs
type Asset struct {
	Type    AssetType
	Denom   string // for coins (e.g. "ugnot")
	Token   string // for GRC20 tokens (e.g. "gno.land/r/test.testtokena")
	NFTPath string // for NFTs (e.g. "gno.land/r/test.mycollection.1")
}

// Helper functions for creating assets
func NewCoinAsset(denom string) Asset {
	return Asset{
		Type:  AssetTypeCoin,
		Denom: denom,
	}
}

func NewTokenAsset(token string) Asset {
	return Asset{
		Type:  AssetTypeToken,
		Token: token,
	}
}

func NewNFTAsset(nftFullPath string) Asset {
	return Asset{
		Type:    AssetTypeNFT,
		NFTPath: nftFullPath,
	}
}
