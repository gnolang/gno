package minidex

/*
var testRun bool

func TestPlaceOrder(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestPlaceOrder because another test has already run.")
	}
	testRun = true

	// Setup
	trader := std.Address("trader")
	tokenFrom := &grc20.Bank{
		Name:     "TokenA",
		Symbol:   "TA",
		Decimals: 18,
	}
	tokenTo := &grc20.Bank{
		Name:     "TokenB",
		Symbol:   "TB",
		Decimals: 18,
	}
	amount := uint64(100)
	price := uint64(2)

	// Mock std functions
	std.GetOrigCaller = func() std.Address { return trader }
	std.CurrentRealm = func() std.Realm { return std.Realm{Addr: std.Address("dex")} }

	// Create a new DEX
	dex := NewDEX()

	// Place an order
	dex.PlaceOrder(tokenFrom, tokenTo, amount, price)

	// Test initial state
	ordersTree, ok := dex.orders.Get(tokenFrom.GetName() + "/" + tokenTo.GetName())
	uassert.True(t, ok, "expected orders tree to exist")
	uassert.Equal(t, 1, ordersTree.(*avl.Tree).Size(), "expected one order")
}

func TestMatchOrders(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestMatchOrders because another test has already run.")
	}
	testRun = true

	// Setup
	trader1 := std.Address("trader1")
	trader2 := std.Address("trader2")
	tokenA := &grc20.Bank{
		Name:     "TokenA",
		Symbol:   "TA",
		Decimals: 18,
	}
	tokenB := &grc20.Bank{
		Name:     "TokenB",
		Symbol:   "TB",
		Decimals: 18,
	}
	amount := uint64(100)
	price := uint64(2)

	// Mock std functions
	std.GetOrigCaller = func() std.Address { return trader1 }
	std.CurrentRealm = func() std.Realm { return std.Realm{Addr: std.Address("dex")} }

	// Create a new DEX
	dex := NewDEX()

	// Place orders
	dex.PlaceOrder(tokenA, tokenB, amount, price)
	std.GetOrigCaller = func() std.Address { return trader2 }
	dex.PlaceOrder(tokenB, tokenA, amount, price)

	// Test state after matching
	ordersTree, ok := dex.orders.Get(tokenA.GetName() + "/" + tokenB.GetName())
	uassert.True(t, ok, "expected orders tree to exist")
	uassert.Equal(t, 0, ordersTree.(*avl.Tree).Size(), "expected no orders after matching")
}

func TestCancelOrders(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestCancelOrders because another test has already run.")
	}
	testRun = true

	// Setup
	trader := std.Address("trader")
	tokenFrom := &grc20.Bank{
		Name:     "TokenA",
		Symbol:   "TA",
		Decimals: 18,
	}
	tokenTo := &grc20.Bank{
		Name:     "TokenB",
		Symbol:   "TB",
		Decimals: 18,
	}
	amount := uint64(100)
	price := uint64(2)

	// Mock std functions
	std.GetOrigCaller = func() std.Address { return trader }
	std.CurrentRealm = func() std.Realm { return std.Realm{Addr: std.Address("dex")} }

	// Create a new DEX
	dex := NewDEX()

	// Place an order
	dex.PlaceOrder(tokenFrom, tokenTo, amount, price)

	// Cancel orders
	dex.CancelOrders()

	// Test state after cancelling
	ordersTree, ok := dex.orders.Get(tokenFrom.GetName() + "/" + tokenTo.GetName())
	uassert.True(t, ok, "expected orders tree to exist")
	uassert.Equal(t, 0, ordersTree.(*avl.Tree).Size(), "expected no orders after cancelling")
}

func TestAdminClearOrders(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestAdminClearOrders because another test has already run.")
	}
	testRun = true

	// Setup
	trader := std.Address("trader")
	tokenFrom := &grc20.Bank{
		Name:     "TokenA",
		Symbol:   "TA",
		Decimals: 18,
	}
	tokenTo := &grc20.Bank{
		Name:     "TokenB",
		Symbol:   "TB",
		Decimals: 18,
	}
	amount := uint64(100)
	price := uint64(2)

	// Mock std functions
	std.GetOrigCaller = func() std.Address { return trader }
	std.CurrentRealm = func() std.Realm { return std.Realm{Addr: std.Address("dex")} }

	// Create a new DEX
	dex := NewDEX()

	// Place an order
	dex.PlaceOrder(tokenFrom, tokenTo, amount, price)

	// Clear orders
	dex.AdminClearOrders()

	// Test state after clearing
	uassert.Equal(t, 0, dex.orders.Size(), "expected no orders after clearing")
}
*/
