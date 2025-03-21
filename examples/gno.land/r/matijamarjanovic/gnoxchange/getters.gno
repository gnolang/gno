package gnoxchange

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/r/matijamarjanovic/tokenhub"
)

// --------------------------------------------- Ticket Getters ---------------------------------------------

// GetTicket allows a user to get a ticket by its ID. All tickets are named swap-<number>.
func GetTicket(ticketID string) (*Ticket, error) {
	ticketInterface, exists := ticketRegistry.allTickets.Get(ticketID)
	if !exists {
		return nil, errors.New("ticket not found")
	}

	return ticketInterface.(*Ticket), nil
}

// GetAllTickets returns a paginated list of all tickets.
func GetAllTickets(path string) *pager.Page {
	pager := pager.NewPager(ticketRegistry.allTickets, 1000000, true)
	return pager.MustGetPageByPath(path)
}

// GetOpenTickets returns a paginated list of open, non-expired tickets.
func GetOpenTickets(path string) *pager.Page {
	pager := pager.NewPager(ticketRegistry.openTickets, 1000000, true)
	return pager.MustGetPageByPath(path)
}

func GetAllNFTTickets(path string) *pager.Page {
	nftTickets := avl.NewTree()

	ticketRegistry.allTickets.Iterate("", "", func(key string, value interface{}) bool {
		ticket := value.(*Ticket)
		if ticket.AssetIn.Type == AssetTypeNFT {
			nftTickets.Set(key, ticket)
		}
		return false
	})

	pager := pager.NewPager(nftTickets, 5, true)
	return pager.MustGetPageByPath(path)
}

func GetOpenNFTTickets(path string) *pager.Page {
	pager := pager.NewPager(ticketRegistry.openNFTTickets, 1000000, true)
	return pager.MustGetPageByPath(path)
}

func GetAllOpenNonNFTTickets(path string) *pager.Page {
	nonNFTTickets := avl.NewTree()

	ticketRegistry.allTickets.Iterate("", "", func(key string, value interface{}) bool {
		ticket := value.(*Ticket)
		if ticket.Status == "open" &&
			!time.Now().After(ticket.ExpiresAt) &&
			ticket.AssetIn.Type != AssetTypeNFT &&
			ticket.AssetOut.Type != AssetTypeNFT {
			nonNFTTickets.Set(key, ticket)
		}
		return false
	})

	pager := pager.NewPager(nonNFTTickets, 1000000, true)
	return pager.MustGetPageByPath(path)
}

// GetTicketsPageInfoString returns a string representation of ticket information for the given page
// e.g. ticketsStr, err := GetTicketsPageInfoString("?page=1&size=10")
func GetTicketsPageInfoString(path string) (string, error) {
	pager := pager.NewPager(ticketRegistry.allTickets, 1000000, true)
	page := pager.MustGetPageByPath(path)
	var result string

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		assetInInfo := getDetailedAssetInfo(ticket.AssetIn)
		assetOutInfo := getDetailedAssetInfo(ticket.AssetOut)

		info := ufmt.Sprintf("%s>Creator:%s,AssetIn:{%s},AssetOut:{%s},AmountIn:%d,MinAmountOut:%d,CreatedAt:%s,ExpiresAt:%s,Status:%s;",
			ticket.ID,
			ticket.Creator,
			assetInInfo,
			assetOutInfo,
			ticket.AmountIn,
			ticket.MinAmountOut,
			ticket.CreatedAt.String(),
			ticket.ExpiresAt.String(),
			ticket.Status,
		)
		result += info
	}

	return result, nil
}

// GetOpenTicketsPageInfoString returns a string representation of open ticket information for the given page
// e.g. ticketsStr, err := GetOpenTicketsPageInfoString("?page=1&size=10")
func GetOpenTicketsPageInfoString(path string) (string, error) {
	pager := pager.NewPager(ticketRegistry.openTickets, 1000000, true)
	page := pager.MustGetPageByPath(path)
	var result string

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		assetInInfo := getDetailedAssetInfo(ticket.AssetIn)
		assetOutInfo := getDetailedAssetInfo(ticket.AssetOut)

		info := ufmt.Sprintf("%s>Creator:%s,AssetIn:{%s},AssetOut:{%s},AmountIn:%d,MinAmountOut:%d,CreatedAt:%s,ExpiresAt:%s,Status:%s;",
			ticket.ID,
			ticket.Creator,
			assetInInfo,
			assetOutInfo,
			ticket.AmountIn,
			ticket.MinAmountOut,
			ticket.CreatedAt.String(),
			ticket.ExpiresAt.String(),
			ticket.Status,
		)
		result += info
	}

	return result, nil
}

// GetAllNFTTicketsPageInfoString returns a string representation of all NFT ticket information for the given page
// e.g. ticketsStr, err := GetAllNFTTicketsPageInfoString("?page=1&size=10")
func GetAllNFTTicketsPageInfoString(path string) (string, error) {
	nftTickets := avl.NewTree()

	ticketRegistry.allTickets.Iterate("", "", func(key string, value interface{}) bool {
		ticket := value.(*Ticket)
		if ticket.AssetIn.Type == AssetTypeNFT {
			nftTickets.Set(key, ticket)
		}
		return false
	})

	pager := pager.NewPager(nftTickets, 1000000, true)
	page := pager.MustGetPageByPath(path)
	var result string

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		assetInInfo := getDetailedAssetInfo(ticket.AssetIn)
		assetOutInfo := getDetailedAssetInfo(ticket.AssetOut)

		info := ufmt.Sprintf("%s>Creator:%s,AssetIn:{%s},AssetOut:{%s},AmountIn:%d,MinAmountOut:%d,CreatedAt:%s,ExpiresAt:%s,Status:%s;",
			ticket.ID,
			ticket.Creator,
			assetInInfo,
			assetOutInfo,
			ticket.AmountIn,
			ticket.MinAmountOut,
			ticket.CreatedAt.String(),
			ticket.ExpiresAt.String(),
			ticket.Status,
		)
		result += info
	}

	return result, nil
}

// GetOpenNFTTicketsPageInfoString returns a string representation of open NFT ticket information for the given page
// e.g. ticketsStr, err := GetOpenNFTTicketsPageInfoString("?page=1&size=10")
func GetOpenNFTTicketsPageInfoString(path string) (string, error) {
	pager := pager.NewPager(ticketRegistry.openNFTTickets, 1000000, true)
	page := pager.MustGetPageByPath(path)
	var result string

	for _, item := range page.Items {
		ticket := item.Value.(*Ticket)

		assetInInfo := getDetailedAssetInfo(ticket.AssetIn)
		assetOutInfo := getDetailedAssetInfo(ticket.AssetOut)

		info := ufmt.Sprintf("%s>Creator:%s,AssetIn:{%s},AssetOut:{%s},AmountIn:%d,MinAmountOut:%d,CreatedAt:%s,ExpiresAt:%s,Status:%s;",
			ticket.ID,
			ticket.Creator,
			assetInInfo,
			assetOutInfo,
			ticket.AmountIn,
			ticket.MinAmountOut,
			ticket.CreatedAt.String(),
			ticket.ExpiresAt.String(),
			ticket.Status,
		)
		result += info
	}

	return result, nil
}

func GetOpenTicketsCount() int {
	return ticketRegistry.openTickets.Size()
}

func GetOpenNFTTicketsCount() int {
	return ticketRegistry.openNFTTickets.Size()
}

func GetAllTicketsCount() int {
	return ticketRegistry.allTickets.Size()
}

// --------------------------------------------- Pool Getters ---------------------------------------------

func (p *Pool) GetLPBalance(address std.Address) uint64 {
	teller := p.lpToken.CallerTeller()
	return teller.BalanceOf(address)
}

func (p *Pool) GetSharePercentage(address std.Address) float64 {
	if p.totalSupplyLp == 0 {
		return 0
	}
	lpBalance := p.GetLPBalance(address)
	return float64(lpBalance) / float64(p.totalSupplyLp) * 100
}

func GetPoolsPage(path string) *pager.Page {
	pager := pager.NewPager(poolRegistry.pools, 5, false)
	return pager.MustGetPageByPath(path)
}

func GetAllPoolNames() []string {
	var poolNames []string

	poolRegistry.pools.Iterate("", "", func(key string, value interface{}) bool {
		poolNames = append(poolNames, key)
		return false
	})

	return poolNames
}

func GetAllPoolNamesCount() int {
	return poolRegistry.pools.Size()
}

// GetPoolInfo returns basic information about a specific pool
func GetPoolInfo(poolKey string) (string, string, uint64, uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return "", "", 0, 0, errors.New("pool not found")
	}

	pool := poolInterface.(*Pool)
	return pool.tokenA, pool.tokenB, pool.reserveA, pool.reserveB, nil
}

// GetPoolsPageInfo returns detailed information about pools for the given page
func GetPoolsPageInfo(path string) ([]PoolInfo, string, error) {
	pager := pager.NewPager(poolRegistry.pools, 1000000, false)
	page := pager.MustGetPageByPath(path)
	pools := make([]PoolInfo, 0, len(page.Items))

	for _, item := range page.Items {
		pool := item.Value.(*Pool)

		tokenA := tokenhub.MustGetToken(pool.tokenA)
		tokenB := tokenhub.MustGetToken(pool.tokenB)

		pools = append(pools, PoolInfo{
			PoolKey: item.Key,
			TokenAInfo: TokenInfo{
				Path:     pool.tokenA,
				Name:     tokenA.GetName(),
				Symbol:   tokenA.GetSymbol(),
				Decimals: tokenA.GetDecimals(),
			},
			TokenBInfo: TokenInfo{
				Path:     pool.tokenB,
				Name:     tokenB.GetName(),
				Symbol:   tokenB.GetSymbol(),
				Decimals: tokenB.GetDecimals(),
			},
			ReserveA:      pool.reserveA,
			ReserveB:      pool.reserveB,
			TotalSupplyLP: pool.totalSupplyLp,
		})
	}

	return pools, page.Picker(page.Pager.PageQueryParam), nil
}

// GetPoolsPageInfoString returns a string representation of pool information for the given page
// e.g. poolsStr, err := GetPoolsPageInfoString("?page=1&size=10")
func GetPoolsPageInfoString(path string) (string, error) {
	pager := pager.NewPager(poolRegistry.pools, 1000000, false)
	page := pager.MustGetPageByPath(path)
	var result string

	for _, item := range page.Items {
		pool := item.Value.(*Pool)

		tokenA := tokenhub.MustGetToken(pool.tokenA)
		tokenB := tokenhub.MustGetToken(pool.tokenB)

		info := ufmt.Sprintf("%s>TokenA:{Path:%s,Name:%s,Symbol:%s,Decimals:%d},TokenB:{Path:%s,Name:%s,Symbol:%s,Decimals:%d},ReserveA:%d,ReserveB:%d,TotalSupplyLP:%d;",
			item.Key,
			pool.tokenA, tokenA.GetName(), tokenA.GetSymbol(), tokenA.GetDecimals(),
			pool.tokenB, tokenB.GetName(), tokenB.GetSymbol(), tokenB.GetDecimals(),
			pool.reserveA,
			pool.reserveB,
			pool.totalSupplyLp,
		)
		result += info
	}

	return result, nil
}

// --------------------------------------------- Helper Functions ---------------------------------------------

// Helper function to get detailed asset information including token metadata when applicable
func getDetailedAssetInfo(asset Asset) string {
	switch asset.Type {
	case AssetTypeCoin:
		return ufmt.Sprintf("Type:coin,Denom:%s", asset.Denom)
	case AssetTypeToken:
		token := tokenhub.GetToken(asset.Token)
		if token == nil {
			return ufmt.Sprintf("Type:token,Path:%s,Error:TokenNotFound", asset.Token)
		}
		return ufmt.Sprintf("Type:token,Path:%s,Name:%s,Symbol:%s,Decimals:%d",
			asset.Token,
			token.GetName(),
			token.GetSymbol(),
			token.GetDecimals(),
		)
	case AssetTypeNFT:
		nft := tokenhub.GetNFT(asset.NFTPath)
		if nft == nil {
			return ufmt.Sprintf("Type:nft,Path:%s,Error:NFTNotFound", asset.NFTPath)
		}
		return ufmt.Sprintf("Type:nft,TokenhubPath:%s", asset.NFTPath)
	default:
		return "Type:unknown"
	}
}
