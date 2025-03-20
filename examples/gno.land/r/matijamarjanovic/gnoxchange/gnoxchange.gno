// GnoXchange is a simple decentralized exchange for Gno.land.
// It allows swapping tokens in 2 ways while keeping the spirit of decentralization.
//
//  1. AMM - allows users to create pools with any 2 tokens.
//     - When pools are created, creator gets minted LP tokens.
//     - Anyone can add liquidity to any existing pool but the liquidity has to be in the same ratio as the pool reserves.
//     When adding liquidity, the amount of LP tokens received depends on the ratio of the pool reserves.
//     - LP holders are able to withdraw their liquidity and receive their share of the pool's tokens.
//     Share is in the ratio of the pool's reserves.
//     - Swapping tokens adjusts their price in the pool automatically.
//     - Small fee of 0.3% is taken when tokens are swapped, the fee remains in the pool making LP tokens more valuable.
//
//  2. Peer to peer - allows users to create tickets for a fixed amount of tokens for a certain token
//     - Anyone can create a ticket for a fixed amount of tokens for a certain token.
//     - On ticket creation, the funds that creator wishes to swap are locked in the realm,
//     this ensures that the swaps are atomic and provides insurance for the fulfiller.
//     - There are 3 types of tickets: Coin (ugnot) to Token, Token to Coin, Token to Token.
//     - When creating tickets that aim to swap Coin to Token, the amount of Coin to be swapped is sent in the tx.
//     - All tickets are set to have an expiration time set on creation (in hours); all tickets are public
//     - If user wants to fulfill a ticket, he triggers FulfilTicket() with amount he wishes to offer
//     (has to be bigger than requested miniumum)
//     - If fulfilling a ticket requires coins to be sent, user has to match the amount of coins sent
//     in the transaction and the amount argument (in denominations)
//     - Tickets can be cancelled by the creator before they are fulfilled or expired which refunds the funds
//     to the creator.
//
// GnoXchange offers one more functionality - NFT market.
// - Anyone can create a ticket for a fixed amount of tokens for a certain NFT.
// - When creating a ticket for an NFT, the NFT has to be approved first by the creator.
// - On ticket creation NFT is transferred from the creator to the realm.
// - When fulfilling a ticket for an NFT, the NFT is transferred from the realm to the fulfiller (buyer).
//
// Notes:
// All actions with tokens require user's allowance approval of the GnoXchange realm for all tokens to be used:
// <your-token>.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), <amount>)
// (in case of NFTs, the approval is done on the NFT contract)
// <your-nft>.SetApprovalForAll(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), true) (for all NFTs in the collection), or:
// <your-nft>.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), <token-id>) (for specific NFT)
//
// Both token and nft instances can be accessed through the tokenhub realm (GetToken() and GetNFT() functions).
//
// Since tokens can have different decimals, 1 unit of a token with less decimals is the minimum amount of both tokens.
// Meaning if user wants to p2p swap/create a pool with tokenA (4 decimalns) and tokenB (six decimals)
// he would have to use at least 1 tokenA (to be more precise 0.0001 becasue of the decimals) and 100 tokenB (0.000100).
//
// All actions are done optimistically - for instance, if user wants to add liquidiy of tokenA and tokenB,
// after passing all the checks 2 tokens are to be added to the pool. Since this cannot happen at once,
// after adding the first token, if adding of the second token fails, the first token is refunded.
// This is applied to all operations in the system.
package gnoxchange

import (
	"gno.land/p/demo/avl"
	"gno.land/r/leon/hof"
)

var (
	poolRegistry   *PoolRegistry
	ticketRegistry *TicketRegistry
)

func init() {
	initHelper()
}

func initHelper() {
	poolRegistry = &PoolRegistry{
		pools: avl.NewTree(),
	}

	ticketRegistry = &TicketRegistry{
		allTickets:     avl.NewTree(),
		openTickets:    avl.NewTree(),
		openNFTTickets: avl.NewTree(),
	}

	hof.Register("gnoXchange", "Decentralized exchange for gno.land for p2p, amm and nft-to-token swaps")
}
