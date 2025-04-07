package gnoxchange

import (
	"std"
	"time"

	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
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

var (
	ErrIdenticalTokens              = errors.New("identical tokens")
	ErrPoolAlreadyExists            = errors.New("pool already exists")
	ErrSentCoinsMismatch            = errors.New("sent coins don't match ticket parameters")
	ErrTicketNotFound               = errors.New("ticket not found")
	ErrTicketNotOpen                = errors.New("ticket is not open")
	ErrTicketExpired                = errors.New("ticket has expired")
	ErrOnlyCreatorCanCancel         = errors.New("only ticket creator can cancel")
	ErrPoolNotFound                 = errors.New("pool not found")
	ErrInvalidPoolType              = errors.New("invalid pool type in registry")
	ErrCannotWithdrawZero           = errors.New("cannot withdraw zero liquidity")
	ErrInsufficientLPBalance        = errors.New("insufficient LP token balance")
	ErrInsufficientLiquidity        = errors.New("insufficient liquidity provided")
	ErrInsufficientInitialLiquidity = errors.New("insufficient initial liquidity")
	ErrInvalidNFTPathFormat         = errors.New("invalid NFT path format: expected path.collection.tokenID")
	ErrInvalidTokenID               = errors.New("invalid token ID")
	ErrCallerNotOwner               = errors.New("caller is not the owner of the NFT")

	ErrNFTNotFound = func(nftPath string) error {
		return errors.New("NFT not found: " + nftPath)
	}

	ErrFailedRefundTokens = func(err error) error {
		return errors.New(ufmt.Sprintf("failed to refund tokens: %v", err))
	}

	ErrFailedNFTTransfer = func(err error) error {
		return errors.New("failed to transfer NFT: " + err.Error())
	}

	ErrTokenNotFound = func(token string) error {
		return errors.New(ufmt.Sprintf("token %s not found", token))
	}

	ErrInsufficientAllowance = func(token string, have, need uint64) error {
		return errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", token, have, need))
	}

	ErrTokenDecimalsExceedRange = func(token string, decimals uint) error {
		return errors.New(ufmt.Sprintf("token decimals exceed safe range: %s has %d decimals, maximum allowed is 6", token, decimals))
	}

	ErrFailedTokenTransfer = func(token string, err error) error {
		return errors.New(ufmt.Sprintf("failed to transfer token %s: %v", token, err))
	}

	ErrIncorrectTokenRatio = func(reserveA, reserveB uint64) error {
		return errors.New(ufmt.Sprintf("incorrect token ratio, should be %d:%d", reserveA, reserveB))
	}

	ErrFailedMintLPTokens = func(err error) error {
		return errors.New(ufmt.Sprintf("failed to mint LP tokens: %v", err))
	}

	ErrInvalidInputToken        = errors.New("invalid input token")
	ErrInsufficientOutputAmount = func(minAmountOut, amountOut uint64) error {
		return errors.New(ufmt.Sprintf("insufficient output amount: expected minimum %d, got %d", minAmountOut, amountOut))
	}

	ErrInvalidInputAmount = func(token string, expectedScale uint64, decimalsIn, minDecimals uint) error {
		return errors.New(ufmt.Sprintf(
			"invalid input amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
			token, expectedScale, decimalsIn, minDecimals,
		))
	}
)
