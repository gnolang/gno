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
	std.TestSetOriginCaller(admin)

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
	std.TestSetOriginCaller(admin)

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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Set up user with collateral
	std.TestSetOriginCaller(user)
	std.TestIssueCoins(user, std.Coins{{"ugnot", 1000000000}}) // 1000 GNOT
	std.TestSetOriginSend(std.Coins{{"ugnot", 1000000000}}, nil)
	std.TestIssueCoins(std.OriginPkgAddress(), std.Coins{{"ugnot", 1000000000}})

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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
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
	std.TestSetOriginCaller(user)
	urequire.PanicsWithMessage(t, ErrInsufficientCollateral.Error(), func() {
		Borrow(marketId, uint256.NewUint(100000000))
	})

	// Test borrowing too much
	std.TestIssueCoins(user, std.Coins{{"ugnot", 1000000000}})
	std.TestSetOriginSend(std.Coins{{"ugnot", 1000000000}}, nil)
	std.TestIssueCoins(std.OriginPkgAddress(), std.Coins{{"ugnot", 1000000000}})
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
	std.TestSetOriginCaller(testutils.TestAddress("admin"))
	EnableIRM("r/gnolend/irm/linear")

	// Enable LLTV
	lltv := uint256.NewUint(75)
	EnableLLTV(lltv)

	// Create market
	CreateMarket(loanToken.String(), collateralToken.String(), oracle, "r/gnolend/irm/linear", 75)

	// Set up lender
	std.TestSetOriginCaller(lender)
	Supply(loanToken.String()+"/"+collateralToken.String(), 1000000, 0) // Supply 1M tokens

	// Set up borrower
	std.TestSetOriginCaller(borrower)
	SupplyCollateral(loanToken.String()+"/"+collateralToken.String(), 1000000) // Supply 1M collateral
	Borrow(loanToken.String()+"/"+collateralToken.String(), 500000, 0) // Borrow 500K tokens

	// Get initial state
	market, _ := GetMarket(loanToken.String() + "/" + collateralToken.String())
	initialBorrowed := market.TotalBorrowAssets.String()
	initialShares := market.TotalBorrowShares.String()

	// Simulate time passing (1 year = 31536000 seconds)
	std.TestSetTime(std.TestGetTime() + 31536000)

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
