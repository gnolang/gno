package gnoxchange

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/r/matijamarjanovic/tokenhub"
)

func TestCreatePool(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/pools"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	_, err := CreatePool("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b", 1000, 1000)
	uassert.Error(t, err, "Should fail with identical tokens")

	tellerA := tokenA.CallerTeller()
	tellerB := tokenB.CallerTeller()

	err = tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 10000)
	urequire.NoError(t, err, "Should approve token A")
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 10000)
	urequire.NoError(t, err, "Should approve token B")

	poolKey, err := CreatePool(
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a",
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b",
		1000,
		1000,
	)
	urequire.NoError(t, err, "Should create pool successfully")
	uassert.NotEmpty(t, poolKey, "Pool key should not be empty")

	tokenAName, tokenBName, reserveA, reserveB, err := GetPoolInfo(poolKey)
	urequire.NoError(t, err, "Should get pool info")
	uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a", tokenAName)
	uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b", tokenBName)
	uassert.Equal(t, uint64(1000), reserveA)
	uassert.Equal(t, uint64(1000), reserveB)
}

func TestAddLiquidity(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/pools"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a2")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b2")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	tellerA := tokenA.CallerTeller()
	tellerB := tokenB.CallerTeller()

	err := tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 10000)
	urequire.NoError(t, err, "Should approve token A")
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 10000)
	urequire.NoError(t, err, "Should approve token B")

	poolKey, err := CreatePool("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a2", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b2", 1000, 1000)
	urequire.NoError(t, err, "Should create pool")

	err = tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 500)
	urequire.NoError(t, err)
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	err = AddLiquidity(poolKey, 500, 1000)
	uassert.Error(t, err, "Should fail with incorrect ratio")

	err = tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 500)
	urequire.NoError(t, err)
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 500)
	urequire.NoError(t, err)

	err = AddLiquidity(poolKey, 500, 500)
	urequire.NoError(t, err, "Should add liquidity successfully")

	_, _, reserveA, reserveB, err := GetPoolInfo(poolKey)
	urequire.NoError(t, err)
	uassert.Equal(t, uint64(1500), reserveA)
	uassert.Equal(t, uint64(1500), reserveB)
}

func TestWithdrawLiquidity(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/pools"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a3")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b3")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	tellerA := tokenA.CallerTeller()
	tellerB := tokenB.CallerTeller()

	err := tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	poolKey, err := CreatePool("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a3", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b3", 1000, 1000)
	urequire.NoError(t, err, "Should create pool")

	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	urequire.True(t, exists, "Pool should exist")
	pool := poolInterface.(*Pool)

	initialLPBalance := pool.GetLPBalance(std.CurrentRealm().Address())
	uassert.True(t, initialLPBalance > 0, "Should have LP tokens")

	_, _, err = WithdrawLiquidity(poolKey, initialLPBalance+1)
	uassert.Error(t, err, "Should fail withdrawing more than available")

	amountA, amountB, err := WithdrawLiquidity(poolKey, initialLPBalance/2)
	urequire.NoError(t, err, "Should withdraw successfully")
	uassert.Equal(t, uint64(500), amountA)
	uassert.Equal(t, uint64(500), amountB)

	newLPBalance := pool.GetLPBalance(std.CurrentRealm().Address())
	uassert.Equal(t, initialLPBalance/2, newLPBalance)
}

func TestPoolRegistry(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/pools"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a4")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b4")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	tellerA := tokenA.CallerTeller()
	tellerB := tokenB.CallerTeller()

	err := tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 2000)
	urequire.NoError(t, err)
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 2000)
	urequire.NoError(t, err)

	poolKey1, err := CreatePool("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_a4", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/pools.test_token_b4", 1000, 1000)
	urequire.NoError(t, err)

	poolNames := GetAllPoolNames()
	uassert.True(t, len(poolNames) > 0, "Should have at least one pool")
	uassert.True(t, contains(poolNames, poolKey1), "Should contain created pool")

	page := GetPoolsPage("")
	uassert.True(t, len(page.Items) > 0, "Page should contain entries")
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func TestCreateCoinToTokenTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets"))
	token, ledger := grc20.NewToken("Test Token", "TT", 6)
	tokenhub.RegisterToken(token.Getter(), "test_token")

	{
		_, err := CreateCoinToTokenTicket("ugnot", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", 1000, 24)
		uassert.Error(t, err, "Should fail without sending coins")
	}

	{
		std.TestSetOriginSend(std.Coins{{"wrongcoin", 1000}}, nil)
		_, err := CreateCoinToTokenTicket("ugnot", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", 1000, 24)
		uassert.Error(t, err, "Should fail with wrong coin denomination")
	}

	{
		std.TestSetOriginSend(std.Coins{{"ugnot", 1000}, {"other", 500}}, nil)
		_, err := CreateCoinToTokenTicket("ugnot", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", 1000, 24)
		uassert.Error(t, err, "Should fail with multiple coins")
	}

	{
		std.TestSetOriginSend(std.Coins{{"ugnot", 1000}}, nil)
		ticketID, err := CreateCoinToTokenTicket("ugnot", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", 1000, 24)
		urequire.NoError(t, err, "Should create ticket successfully")
		uassert.NotEmpty(t, ticketID, "Ticket ID should not be empty")

		ticket, err := GetTicket(ticketID)
		urequire.NoError(t, err)
		uassert.Equal(t, "ugnot", ticket.AssetIn.Denom)
		uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", ticket.AssetOut.Token)
		uassert.Equal(t, uint64(1000), ticket.MinAmountOut)
		uassert.Equal(t, "open", ticket.Status)
	}
}

func TestCreateTokenToCoinTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets"))
	token, ledger := grc20.NewToken("Test Token", "TT", 6)
	tokenhub.RegisterToken(token.Getter(), "test_token")

	ledger.Mint(std.CurrentRealm().Address(), 10000)

	_, err := CreateTokenToCoinTicket("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", "ugnot", 1000, 1000, 24)
	uassert.Error(t, err, "Should fail without approval")

	teller := token.CallerTeller()
	err = teller.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	ticketID, err := CreateTokenToCoinTicket("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", "ugnot", 1000, 1000, 24)
	urequire.NoError(t, err, "Should create ticket successfully")

	ticket, err := GetTicket(ticketID)
	urequire.NoError(t, err)
	uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", ticket.AssetIn.Token)
	uassert.Equal(t, "ugnot", ticket.AssetOut.Denom)
	uassert.Equal(t, uint64(1000), ticket.MinAmountOut)
}

func TestCreateTokenToTokenTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)

	tellerA := tokenA.CallerTeller()
	err := tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	ticketID, err := CreateTokenToTokenTicket(
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_a",
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_b",
		1000,
		1000,
		24,
	)
	urequire.NoError(t, err, "Should create ticket successfully")

	ticket, err := GetTicket(ticketID)
	urequire.NoError(t, err)
	uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_a", ticket.AssetIn.Token)
	uassert.Equal(t, "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_b", ticket.AssetOut.Token)
}

func TestFulfillTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets"))
	tokenA, ledgerA := grc20.NewToken("Test Token A", "TTA", 6)
	tokenB, ledgerB := grc20.NewToken("Test Token B", "TTB", 6)

	tokenhub.RegisterToken(tokenA.Getter(), "test_token_a")
	tokenhub.RegisterToken(tokenB.Getter(), "test_token_b")

	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	tellerA := tokenA.CallerTeller()
	err := tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	ticketID, err := CreateTokenToTokenTicket(
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_a",
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token_b",
		1000,
		1000,
		24,
	)
	urequire.NoError(t, err)

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets2"))
	ledgerA.Mint(std.CurrentRealm().Address(), 10000)
	ledgerB.Mint(std.CurrentRealm().Address(), 10000)

	err = tellerA.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	err = FulfillTicket(ticketID, 500)
	uassert.Error(t, err, "Should fail with insufficient amount")

	tellerB := tokenB.CallerTeller()
	err = tellerB.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	err = FulfillTicket(ticketID, 1000)
	urequire.NoError(t, err, "Should fulfill ticket successfully")

	ticket, err := GetTicket(ticketID)
	urequire.NoError(t, err)
	uassert.Equal(t, "fulfilled", ticket.Status)
}

func TestCancelTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/tickets"))
	token, ledger := grc20.NewToken("Test Token", "TT", 6)
	tokenhub.RegisterToken(token.Getter(), "test_token")

	ledger.Mint(std.CurrentRealm().Address(), 10000)

	teller := token.CallerTeller()
	err := teller.Approve(std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange"), 1000)
	urequire.NoError(t, err)

	ticketID, err := CreateTokenToCoinTicket("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/tickets.test_token", "ugnot", 1000, 1000, 24)
	urequire.NoError(t, err)

	err = CancelTicket(ticketID)
	urequire.NoError(t, err, "Should cancel ticket successfully")

	ticket, err := GetTicket(ticketID)
	urequire.NoError(t, err)
	uassert.Equal(t, "cancelled", ticket.Status)
}

func TestCreateNFTToTokenTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/nfts"))

	token, ledger := grc20.NewToken("Test Token", "TT", 6)
	tokenhub.RegisterToken(token.Getter(), "test_token")
	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	tokenhub.RegisterNFT(nft.Getter(), "test_nft", "1")

	{
		_, err := CreateNFTToTokenTicket("gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_nft.nonexistent.1", "gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_token", 1000, 24)
		uassert.Error(t, err, "Should fail with non-existent NFT")
	}

	{
		otherAddr := std.DerivePkgAddr("some.other.address")
		err := nft.Mint(otherAddr, "1")
		urequire.NoError(t, err)

		_, err = CreateNFTToTokenTicket(
			"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_nft.1",
			"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_token",
			1000,
			24,
		)
		uassert.Error(t, err, "Should fail when caller is not owner")
	}

	{
		err := nft.Mint(std.CurrentRealm().Address(), "2")
		urequire.NoError(t, err)

		_, err = CreateNFTToTokenTicket(
			"gno.land/r/test/nfts.test_nft.2",
			"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_token",
			1000,
			24,
		)
		uassert.Error(t, err, "Should fail without approval")
	}
}

func TestCreateNFTToCoinTicket(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/nftsss"))
	nft := grc721.NewBasicNFT("Test NFT", "TNFT")

	tokenhub.RegisterNFT(nft.Getter(), "test_nft", "1")

	{
		_, err := CreateNFTToCoinTicket("gno.land/r/test/nftsss.test_nft.nonexistent.1", "ugnot", 1000, 24)
		uassert.Error(t, err, "Should fail with non-existent NFT")
	}

	{
		otherAddr := std.DerivePkgAddr("some.other.address")
		err := nft.Mint(otherAddr, "1")
		urequire.NoError(t, err)

		_, err = CreateNFTToCoinTicket(
			"gno.land/r/test/nftsss.test_nft.1",
			"ugnot",
			1000,
			24,
		)
		uassert.Error(t, err, "Should fail when caller is not owner")
	}

	{
		err := nft.Mint(std.CurrentRealm().Address(), "2")
		tokenhub.RegisterNFT(nft.Getter(), "test_nft", "2")
		urequire.NoError(t, err)

		contractAddr := std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange")
		err = nft.Approve(contractAddr, "2")
		urequire.NoError(t, err)

		ticketID, err := CreateNFTToCoinTicket(
			"gno.land/r/test/nftsss.test_nft.2",
			"ugnot",
			1000,
			24,
		)
		urequire.NoError(t, err, "Should create ticket successfully")

		ticket, err := GetTicket(ticketID)
		urequire.NoError(t, err)
		uassert.Equal(t, "gno.land/r/test/nftsss.test_nft.2", ticket.AssetIn.NFTPath)
		uassert.Equal(t, "ugnot", ticket.AssetOut.Denom)
		uassert.Equal(t, uint64(1000), ticket.MinAmountOut)

		owner, err := nft.OwnerOf("2")
		urequire.NoError(t, err)
		uassert.Equal(t, contractAddr, owner, "NFT should be transferred to contract")
	}
}

func TestBuyNFT(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/nfts"))

	token, ledger := grc20.NewToken("Test Token", "TT", 6)
	tokenhub.RegisterToken(token.Getter(), "test_token")
	nft := grc721.NewBasicNFT("Test NFT", "TNFT")
	tokenhub.RegisterNFT(nft.Getter(), "test_nft", "1")

	contractAddr := std.DerivePkgAddr("gno.land/r/matijamarjanovic/gnoxchange")

	err := nft.Mint(std.CurrentRealm().Address(), "1")
	tokenhub.RegisterNFT(nft.Getter(), "test_nft", "1")
	urequire.NoError(t, err)

	err = nft.Approve(contractAddr, "1")
	urequire.NoError(t, err)

	ticketID, err := CreateNFTToTokenTicket(
		"gno.land/r/test/nfts.test_nft.1",
		"gno.land/r/matijamarjanovic/tokenhub.gno.land/r/test/nfts.test_token",
		1000,
		24,
	)
	urequire.NoError(t, err)

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/test/nftss"))

	{
		err = BuyNFT(ticketID, 500)
		uassert.Error(t, err, "Should fail with insufficient payment")
	}

	{
		ledger.Mint(std.CurrentRealm().Address(), 2000)
		teller := token.CallerTeller()
		err = teller.Approve(contractAddr, 1000)
		urequire.NoError(t, err)

		err = BuyNFT(ticketID, 1000)
		urequire.NoError(t, err, "Should buy NFT successfully")

		owner, err := nft.OwnerOf("1")
		urequire.NoError(t, err)
		uassert.Equal(t, std.CurrentRealm().Address(), owner, "NFT should be transferred to buyer")

		ticket, err := GetTicket(ticketID)
		urequire.NoError(t, err)
		uassert.Equal(t, "fulfilled", ticket.Status)
	}
}
