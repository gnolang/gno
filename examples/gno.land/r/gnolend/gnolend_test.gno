package gnolend

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/p/gnoswap/uint256"
)

// Reset state before each test
func setupTest(t *testing.T) {
	markets = avl.NewTree()
	marketParams = avl.NewTree()
	positions = avl.NewTree()
	enabledIRMs = avl.NewTree()
	enabledLLTVs = avl.NewTree()
}

func TestGetGnolendAddress(t *testing.T) {
	// Get the package address
	pkgPath := "gno.land/r/gnolend"
	addr := std.DerivePkgAddr(pkgPath)

	// Print the address for use in tests
	println("Gnolend package address:", addr.String())
}

func TestEnableIRM(t *testing.T) {
	setupTest(t)

	// Set up admin caller
	admin := testutils.TestAddress("admin")
	testing.SetOriginCaller(admin)

	// Test enabling IRM
	EnableIRM("r/gnolend/irm/linear")

	// Verify IRM is enabled
	enabled, _ := enabledIRMs.Get("r/gnolend/irm/linear")
	urequire.Equal(t, true, enabled.(bool))
}

func TestEnableLLTV(t *testing.T) {
	setupTest(t)

	// Set up admin caller
	admin := testutils.TestAddress("admin")
	testing.SetOriginCaller(admin)

	// Test enabling LLTV
	lltv := uint256.NewUint(75) // 75%
	EnableLLTV(lltv)

	// Verify LLTV is enabled
	enabled, _ := enabledLLTVs.Get(lltv.String())
	urequire.Equal(t, true, enabled.(bool))
}

func TestGetIRM(t *testing.T) {
	setupTest(t)

	// Register test IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000))) // 10% APR

	// Test getting registered IRM
	irmInstance := GetIRM("r/gnolend/irm/linear")
	urequire.NotNil(t, irmInstance)

	// Test getting non-existent IRM
	irmInstance = GetIRM("r/gnolend/irm/nonexistent")
	urequire.Nil(t, irmInstance)
}

func TestGetMarket(t *testing.T) {
	setupTest(t)

	// Create test tokens
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"

	// Register and enable IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Test getting existing market
	market, params := GetMarket(loanToken.String() + "/" + collateralToken.String())
	urequire.NotNil(t, market)
	urequire.NotNil(t, params)
	urequire.Equal(t, loanToken.String(), params.LoanToken)
	urequire.Equal(t, collateralToken.String(), params.CollateralToken)

	// Test getting non-existent market
	market, params = GetMarket("nonexistent")
	urequire.Nil(t, market)
	urequire.Nil(t, params)
}

func TestGetPosition(t *testing.T) {
	setupTest(t)

	// Create test tokens and addresses
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")

	// Register and enable IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Test getting non-existent position
	position := GetPosition(loanToken.String()+"/"+collateralToken.String(), user.String())
	urequire.Nil(t, position)

	// TODO: Add test for getting existing position once deposit/borrow functions are implemented
}

func TestCreateMarket(t *testing.T) {
	setupTest(t)

	// Create test tokens
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"

	// Register test IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000))) // 10% APR

	// Enable IRM
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75) // 75%
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Verify market was created
	market, params := GetMarket(loanToken.String() + "/" + collateralToken.String())

	// Check market parameters
	urequire.Equal(t, loanToken.String(), params.LoanToken)
	urequire.Equal(t, collateralToken.String(), params.CollateralToken)
	urequire.Equal(t, oracle, params.Oracle)
	urequire.Equal(t, "r/gnolend/irm/linear", params.IRM)

	expectedLLTV := uint256.NewUint(75)
	expectedLLTVWad := uint256.WMulDown(expectedLLTV, uint256.WAD)
	urequire.Equal(t, expectedLLTVWad.String(), params.LLTV.String())

	// Check initial market state
	urequire.Equal(t, "0", market.TotalSupplyAssets.String())
	urequire.Equal(t, "0", market.TotalSupplyShares.String())
	urequire.Equal(t, "0", market.TotalBorrowAssets.String())
	urequire.Equal(t, "0", market.TotalBorrowShares.String())
	urequire.Equal(t, "0", market.Fee.String())
}

func TestCreateMarketErrors(t *testing.T) {
	setupTest(t)

	// Create test tokens
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"

	// Test zero address error
	urequire.PanicsWithMessage(t, ErrZeroAddress.Error(), func() {
		CreateMarket("", collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	})

	// Test same token error
	urequire.PanicsWithMessage(t, ErrSameToken.Error(), func() {
		CreateMarket(loanToken.String(), loanToken.String(), oracle, "r/gnolend/irm/linear", 75)
	})

	// Test market already created error
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	urequire.PanicsWithMessage(t, ErrMarketAlreadyCreated.Error(), func() {
		CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	})

	// Test IRM not found error
	urequire.PanicsWithMessage(t, "IRM not found: r/gnolend/irm/nonexistent", func() {
		CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/nonexistent", 75)
	})
}

func TestBorrow(t *testing.T) {
	setupTest(t)

	// Create test tokens and addresses
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")
	marketId := loanToken.String() + "/" + collateralToken.String()

	// Register and enable IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Set up user with collateral
	testing.SetOriginCaller(user)
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000000000}}) // 1000 GNOT
	testing.OriginSend(std.Coins{{"ugnot", 1000000000}}, nil)
	testing.IssueCoins(std.OriginPkgAddress(), std.Coins{{"ugnot", 1000000000}})

	// Supply collateral
	Supply(marketId, 1000000000, 0)

	// Test borrowing
	borrowAmount := uint256.NewUint(500000000) // 500 GNOT
	Borrow(marketId, borrowAmount)

	// Verify position after borrow
	position := GetPosition(marketId, user.String())
	urequire.Equal(t, borrowAmount.String(), position.BorrowAssets.String())
	urequire.Equal(t, borrowAmount.String(), position.BorrowShares.String())

	// Verify market state
	market, _ := GetMarket(marketId)
	urequire.Equal(t, borrowAmount.String(), market.TotalBorrowAssets.String())
	urequire.Equal(t, borrowAmount.String(), market.TotalBorrowShares.String())
}

func TestBorrowErrors(t *testing.T) {
	setupTest(t)

	// Create test tokens and addresses
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")
	marketId := loanToken.String() + "/" + collateralToken.String()

	// Register and enable IRM
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Test borrowing from non-existent market
	urequire.PanicsWithMessage(t, ErrMarketNotCreated.Error(), func() {
		Borrow("nonexistent", uint256.NewUint(100000000))
	})

	// Test borrowing without collateral
	testing.SetOriginCaller(user)
	urequire.PanicsWithMessage(t, ErrInsufficientCollateral.Error(), func() {
		Borrow(marketId, uint256.NewUint(100000000))
	})

	// Test borrowing too much
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000000000}})
	testing.SetOriginSend(std.Coins{{"ugnot", 1000000000}}, nil)
	testing.IssueCoins(std.OriginPkgAddress(), std.Coins{{"ugnot", 1000000000}})
	Supply(marketId, 1000000000, 0)

	// Try to borrow more than collateral value allows
	urequire.PanicsWithMessage(t, ErrInsufficientCollateral.Error(), func() {
		Borrow(marketId, uint256.NewUint(2000000000))
	})
}

func TestInterestAccrual(t *testing.T) {
	setupTest(t)

	// Create test tokens and addresses
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	lender := testutils.TestAddress("lender")
	borrower := testutils.TestAddress("borrower")

	// Register and enable IRM with 10% APR
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Set up lender
	testing.SetOriginCaller(lender)
	Supply(loanToken.String()+"/"+collateralToken.String(), 1000000, 0) // Supply 1M tokens

	// Set up borrower
	testing.SetOriginCaller(borrower)
	SupplyCollateral(loanToken.String()+"/"+collateralToken.String(), 1000000) // Supply 1M collateral
	Borrow(loanToken.String()+"/"+collateralToken.String(), 500000, 0) // Borrow 500K tokens

	// Get initial state
	market, _ := GetMarket(loanToken.String() + "/" + collateralToken.String())
	initialBorrowed := market.TotalBorrowAssets.String()
	initialShares := market.TotalBorrowShares.String()

	// Simulate time passing (1 year = 31536000 seconds)
	testing.SetTime(testing.GetTime() + 31536000)

	// Accrue interest
	accrueInterest(loanToken.String() + "/" + collateralToken.String())

	// Get final state
	market, _ = GetMarket(loanToken.String() + "/" + collateralToken.String())
	finalBorrowed := market.TotalBorrowAssets.String()
	finalShares := market.TotalBorrowShares.String()

	// Verify interest has accrued (should be ~10% more)
	// Note: Due to rounding and the way interest is calculated, it might not be exactly 10%
	urequire.Greater(t, finalBorrowed, initialBorrowed)
	urequire.Greater(t, finalShares, initialShares)

	// Print values for debugging
	println("Initial borrowed:", initialBorrowed)
	println("Final borrowed:", finalBorrowed)
	println("Initial shares:", initialShares)
	println("Final shares:", finalShares)
}

func TestMarketAPI(t *testing.T) {
	setupTest(t)
	
	// Create test market
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")
	
	// Setup admin
	admin := testutils.TestAddress("admin")
	testing.SetOriginCaller(admin)
	
	// Register and enable IRM/LLTV
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	EnableIRM("r/gnolend/irm/linear")
	EnableLLTV(uint256.NewUint(75))
	
	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	
	// Fund user and supply assets
	testing.SetOriginCaller(user)
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000000000}})
	Supply(loanToken.String()+"/"+collateralToken.String(), 500000000, 0) // Supply 500 GNOT
	
	// Test GetMarketJSON
	t.Run("GetMarketJSON", func(t *testing.T) {
		marketID := loanToken.String() + "/" + collateralToken.String()
		jsonResponse := GetMarketJSON(marketID)
		
		// Parse JSON
		var result map[string]interface{}
		err := json.Unmarshal([]byte(jsonResponse), &result)
		urequire.NoError(t, err)
		
		// Verify market data
		market := result["market"].(map[string]interface{})
		urequire.Equal(t, "500000000", market["totalSupplyAssets"])
		urequire.Equal(t, "500000000", market["totalSupplyShares"]) // 1:1 initial ratio
		urequire.Equal(t, "0", market["totalBorrowAssets"])
		urequire.Equal(t, "0", market["totalBorrowShares"])
		
		// Verify params
		params := result["params"].(map[string]interface{})
		urequire.Equal(t, oracle, params["poolPath"])
		urequire.Equal(t, "r/gnolend/irm/linear", params["irm"])
		urequire.Equal(t, "750000000000000000000000", params["lltv"]) // 75% in WAD
	})
}

func TestPositionAPI(t *testing.T) {
	setupTest(t)
	
	// Create test environment
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")
	
	// Setup market
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	EnableIRM("r/gnolend/irm/linear")
	EnableLLTV(uint256.NewUint(75))
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	
	// Fund user and create position
	testing.SetOriginCaller(user)
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000000000}})
	Supply(loanToken.String()+"/"+collateralToken.String(), 500000000, 0)
	SupplyCollateral(loanToken.String()+"/"+collateralToken.String(), 300000000)
	Borrow(loanToken.String()+"/"+collateralToken.String(), 200000000, 0)
	
	t.Run("GetPositionJSON", func(t *testing.T) {
		positionJSON := GetPositionJSON(
			loanToken.String()+"/"+collateralToken.String(),
			user.String(),
		)
		
		var position map[string]interface{}
		err := json.Unmarshal([]byte(positionJSON), &position)
		urequire.NoError(t, err)
		
		urequire.Equal(t, "500000000", position["supplyShares"])
		urequire.Equal(t, "200000000", position["borrowShares"])
		urequire.Equal(t, "300000000", position["collateral"])
	})
}

func TestListMarketsAPI(t *testing.T) {
	setupTest(t)
	
	// Create 3 test markets
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	EnableIRM("r/gnolend/irm/linear")
	EnableLLTV(uint256.NewUint(75))
	
	tokens := []struct {
		loan       string
		collateral string
	}{
		{"token0", "token1"},
		{"tokenA", "tokenB"},
		{"tokenX", "tokenY"},
	}
	
	for _, pair := range tokens {
		oracle := "r/gnoswap/v1/pool/" + pair.loan + "/" + pair.collateral + "/0"
		CreateMarket(pair.loan, pair.collateral, oracle, "r/gnolend/irm/linear", 75)
	}
	
	t.Run("ListMarketsJSON", func(t *testing.T) {
		marketsJSON := ListMarketsJSON()
		
		var markets []interface{}
		err := json.Unmarshal([]byte(marketsJSON), &markets)
		urequire.NoError(t, err)
		
		urequire.Equal(t, 3, len(markets))
		
		// Verify first market structure
		firstMarket := markets[0].(map[string]interface{})
		urequire.Contains(t, firstMarket, "totalSupplyAssets")
		urequire.Contains(t, firstMarket, "totalBorrowAssets")
		urequire.Contains(t, firstMarket, "fee")
	})
}

func TestAPIErrorHandling(t *testing.T) {
	setupTest(t)
	
	t.Run("NonExistentMarket", func(t *testing.T) {
		urequire.Panics(t, func() {
			GetMarketJSON("nonexistent-market")
		})
	})
	
	t.Run("InvalidPositionRequest", func(t *testing.T) {
		// Should return empty position instead of panicking
		positionJSON := GetPositionJSON("invalid-market", "invalid-user")
		var position map[string]interface{}
		err := json.Unmarshal([]byte(positionJSON), &position)
		urequire.NoError(t, err)
		
		urequire.Equal(t, "0", position["supplyShares"])
		urequire.Equal(t, "0", position["borrowShares"])
		urequire.Equal(t, "0", position["collateral"])
	})
}

func TestHTTPHandlers(t *testing.T) {
	setupTest(t)
	
	// Create test market
	loanToken := testutils.TestAddress("loanToken")
	collateralToken := testutils.TestAddress("collateralToken")
	oracle := "r/gnoswap/v1/pool/" + loanToken.String() + "/" + collateralToken.String() + "/0"
	user := testutils.TestAddress("user")
	
	testing.SetOriginCaller(testutils.TestAddress("admin"))
	irm.RegisterIRM(irm.NewLinearIRM("TestIRM", uint256.NewUint(100000000000000000)))
	EnableIRM("r/gnolend/irm/linear")
	EnableLLTV(uint256.NewUint(75))
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)
	
	t.Run("MarketEndpoint", func(t *testing.T) {
		testing.SetRequest(&std.Request{
			Query: "name=" + loanToken.String() + "/" + collateralToken.String(),
		})
		response := Render("/gnolend/market")
		urequire.Contains(t, response, `"totalSupplyAssets"`)
	})
	
	t.Run("PositionEndpoint", func(t *testing.T) {
		testing.SetRequest(&std.Request{
			Query: "market=" + loanToken.String() + "/" + collateralToken.String() + "&user=" + user.String(),
		})
		response := Render("/gnolend/position")
		urequire.Contains(t, response, `"supplyShares":"0"`)
	})
	
	t.Run("ListMarketsEndpoint", func(t *testing.T) {
		response := Render("/gnolend/markets")
		urequire.Contains(t, response, `"totalSupplyAssets"`)
	})
	
	t.Run("DefaultEndpoint", func(t *testing.T) {
		response := Render("/unknown")
		urequire.Equal(t, "Gnolend API - available endpoints: /market, /position, /markets", response)
	})
}
