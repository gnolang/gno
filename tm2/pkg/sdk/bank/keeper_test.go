package bank

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	addr1 = crypto.AddressFromPreimage([]byte("addr1"))
	addr2 = crypto.AddressFromPreimage([]byte("addr2"))
	addr3 = crypto.AddressFromPreimage([]byte("addr3"))
)

func TestKeeper(t *testing.T) {
	t.Parallel()

	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc3 := env.acck.NewAccountWithAddress(ctx, addr3)

	// Initialize accounts with initial balances
	env.acck.SetAccount(ctx, acc1)
	env.acck.SetAccount(ctx, acc2)
	env.acck.SetAccount(ctx, acc3)

	// Test GetCoins and AddCoins for multiple accounts
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins()), "Initial balance for addr1 should be zero")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins()), "Initial balance for addr2 should be zero")
	require.True(t, env.bank.GetCoins(ctx, addr3).IsEqual(std.NewCoins()), "Initial balance for addr3 should be zero")

	// Initial TotalCoin check (should be zero for all denominations)
	require.Equal(t, int64(0), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin for 'foocoin' should be zero")
	require.Equal(t, int64(0), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin for 'barcoin' should be zero")

	// Add coins for account 1
	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.NoError(t, err, "Error should be nil when adding coins to addr1")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))),
		"Balance for addr1 should be updated to 10 foocoins")
	require.Equal(t, int64(10), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin for 'foocoin' should be 10")

	// Add coins for account 2
	_, err = env.bank.AddCoins(ctx, addr2, std.NewCoins(std.NewCoin("barcoin", 20)))
	require.NoError(t, err, "Error should be nil when adding coins to addr2")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20))),
		"Balance for addr2 should be updated to 20 barcoins")
	require.Equal(t, int64(20), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin for 'barcoin' should be 20")

	// Add coins for account 3
	_, err = env.bank.AddCoins(ctx, addr3, std.NewCoins(std.NewCoin("foocoin", 15), std.NewCoin("barcoin", 5)))
	require.NoError(t, err, "Error should be nil when adding coins to addr3")
	require.True(t, env.bank.GetCoins(ctx, addr3).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15), std.NewCoin("barcoin", 5))),
		"Balance for addr3 should be updated to 15 foocoins and 5 barcoins")
	require.Equal(t, int64(25), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin for 'foocoin' should be 25")
	require.Equal(t, int64(25), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin for 'barcoin' should be 25")

	// Test HasCoins for account 1
	require.True(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10))), "addr1 should have exactly 10 foocoins")
	require.False(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 15))), "addr1 should not have 15 foocoins")
	require.False(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("barcoin", 1))), "addr1 should not have any barcoins")

	// Add more coins to account 1
	_, err = env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.NoError(t, err, "Error should be nil when adding more coins to addr1")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 20))),
		"Balance for addr1 should be updated to 20 foocoins")
	require.Equal(t, int64(35), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin for 'foocoin' should be 35")

	// Subtract coins from account 2
	_, err = env.bank.SubtractCoins(ctx, addr2, std.NewCoins(std.NewCoin("barcoin", 10)))
	require.NoError(t, err, "Error should be nil when subtracting coins from addr2")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10))),
		"Balance for addr2 should be updated to 10 barcoins")
	require.Equal(t, int64(15), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin for 'barcoin' should be 15")

	// Test SendCoins between accounts
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.NoError(t, err, "Error should be nil when sending coins from addr1 to addr2")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15))),
		"Balance for addr1 should be updated to 15 foocoins")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5))),
		"Balance for addr2 should include 5 foocoins")
	require.Equal(t, int64(35), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin for 'foocoin' should be 35")
	require.Equal(t, int64(15), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin for 'barcoin' should be 15")

	// Test InputOutputCoins involving all accounts
	input1 := NewInput(addr1, std.NewCoins(std.NewCoin("foocoin", 5)))
	input2 := NewInput(addr2, std.NewCoins(std.NewCoin("barcoin", 5)))

	output1 := NewOutput(addr3, std.NewCoins(std.NewCoin("foocoin", 5)))
	output2 := NewOutput(addr1, std.NewCoins(std.NewCoin("barcoin", 5)))

	err = env.bank.InputOutputCoins(ctx, []Input{input1, input2}, []Output{output1, output2})
	require.NoError(t, err, "Error should be nil when processing input/output coins")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10), std.NewCoin("barcoin", 5))),
		"Balance for addr1 should be updated correctly after InputOutputCoins")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5), std.NewCoin("barcoin", 5))),
		"Balance for addr2 should be updated correctly after InputOutputCoins")
	require.True(t, env.bank.GetCoins(ctx, addr3).IsEqual(std.NewCoins(std.NewCoin("foocoin", 20), std.NewCoin("barcoin", 5))),
		"Balance for addr3 should be updated correctly after InputOutputCoins")

	// Final TotalCoin check
	require.Equal(t, int64(35), env.bank.TotalCoin(ctx, "foocoin"), "Final TotalCoin for 'foocoin' should be 35")
	require.Equal(t, int64(15), env.bank.TotalCoin(ctx, "barcoin"), "Final TotalCoin for 'barcoin' should be 15")
}

func TestBankKeeper(t *testing.T) {
	t.Parallel()

	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	acc := env.acck.NewAccountWithAddress(ctx, addr1)

	// Set the account and verify initial balance
	env.acck.SetAccount(ctx, acc)
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins()), "Initial balance should be zero")

	// Test AddCoins
	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.NoError(t, err, "Error should be nil when adding coins")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))), "Balance should reflect added coins")

	// Test HasCoins
	require.True(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10))), "Account should have exactly 10 foocoins")
	require.True(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 5))), "Account should have at least 5 foocoins")
	require.False(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 15))), "Account should not have 15 foocoins")
	require.False(t, env.bank.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("barcoin", 5))), "Account should not have barcoin")

	// Add more coins and test SendCoins
	_, err = env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 15)))
	require.NoError(t, err, "Error should be nil when adding more coins")
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.NoError(t, err, "Error should be nil when sending coins")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 20))),
		"Balance should be updated correctly after sending coins")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))),
		"Recipient's balance should reflect the sent amount")

	// Attempt sending more coins than available
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.Error(t, err, "Should return an error when trying to send more coins than available")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 20))),
		"Balance should remain unchanged after failed send")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))),
		"Recipient's balance should remain unchanged after failed send")

	// Test sending multiple denominations
	_, err = env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("barcoin", 30)))
	require.NoError(t, err, "Error should be nil when adding multiple denominations")
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.NoError(t, err, "Error should be nil when sending multiple denominations")
	require.True(t, env.bank.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 15))),
		"Balance should be updated correctly after sending multiple denominations")
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))),
		"Recipient's balance should reflect the sent amounts")

	// Validate coins with invalid denominations or negative values cannot be sent
	err = env.bank.SendCoins(ctx, addr1, addr2, std.Coins{sdk.Coin{"foocoin", -5}})
	require.Error(t, err, "Should return an error when trying to send negative coin values")
}

func TestViewKeeper(t *testing.T) {
	t.Parallel()

	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx
	view := NewViewKeeper(env.acck, env.tck)

	acc := env.acck.NewAccountWithAddress(ctx, addr1)

	// Set the account and verify initial balance
	env.acck.SetAccount(ctx, acc)
	require.True(t, view.GetCoins(ctx, addr1).IsEqual(std.NewCoins()), "Initial balance should be zero")

	// Test GetCoins and SetCoins
	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.NoError(t, err, "Error should be nil when adding coins")
	require.True(t, view.GetCoins(ctx, addr1).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))), "ViewKeeper should reflect added coins")

	// Test HasCoins
	require.True(t, view.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 10))), "Account should have exactly 10 foocoins")
	require.True(t, view.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 5))), "Account should have at least 5 foocoins")
	require.False(t, view.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 15))), "Account should not have 15 foocoins")
	require.False(t, view.HasCoins(ctx, addr1, std.NewCoins(std.NewCoin("barcoin", 5))), "Account should not have barcoin")

	// Test TotalCoin
	require.Equal(t, int64(10), env.bank.TotalCoin(ctx, "foocoin"), "TotalCoin should return the correct amount of foocoin")
	require.Equal(t, int64(0), env.bank.TotalCoin(ctx, "barcoin"), "TotalCoin should return 0 for coins not present")
}

func TestInputOutputCoins(t *testing.T) {
	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	// Initialize account balances
	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 200)))
	require.NoError(t, err)

	// Test valid input and output
	inputs := []Input{
		{Address: addr1, Coins: std.NewCoins(std.NewCoin("foocoin", 100))},
	}
	outputs := []Output{
		{Address: addr2, Coins: std.NewCoins(std.NewCoin("foocoin", 100))},
	}

	err = env.bank.InputOutputCoins(ctx, inputs, outputs)
	require.NoError(t, err)

	// Verify balances after transaction
	remainingCoins := env.bank.GetCoins(ctx, addr1)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 100)), remainingCoins)

	receivedCoins := env.bank.GetCoins(ctx, addr2)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 100)), receivedCoins)

	// Test invalid input (e.g., insufficient funds)
	inputs = []Input{
		{Address: addr1, Coins: std.NewCoins(std.NewCoin("foocoin", 200))}, // more than available
	}
	outputs = []Output{
		{Address: addr2, Coins: std.NewCoins(std.NewCoin("foocoin", 200))},
	}

	err = env.bank.InputOutputCoins(ctx, inputs, outputs)
	require.Error(t, err)
}

func TestSendCoins(t *testing.T) {
	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 100)))
	require.NoError(t, err)
	_, err = env.bank.AddCoins(ctx, addr2, std.NewCoins(std.NewCoin("foocoin", 0)))
	require.NoError(t, err)

	// Test valid send
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.NoError(t, err)

	// Verify balances after send
	balance1 := env.bank.GetCoins(ctx, addr1)
	balance2 := env.bank.GetCoins(ctx, addr2)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 50)), balance1)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 50)), balance2)

	// Test send with insufficient funds
	err = env.bank.SendCoins(ctx, addr1, addr2, std.NewCoins(std.NewCoin("foocoin", 100))) // more than available
	require.Error(t, err)
}

func TestSubtractCoins(t *testing.T) {
	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	// Initialize account with coins
	_, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 100)))
	require.NoError(t, err)

	// Test valid subtraction
	newCoins, err := env.bank.SubtractCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.NoError(t, err)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 50)), newCoins)

	// Test subtraction with insufficient funds
	_, err = env.bank.SubtractCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 100))) // more than available
	require.Error(t, err)
}

func TestAddCoins(t *testing.T) {
	// Setup test environment
	env := setupTestEnv()
	ctx := env.ctx

	// Test valid addition
	newCoins, err := env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.NoError(t, err)
	require.Equal(t, std.NewCoins(std.NewCoin("foocoin", 50)), newCoins)

	// Test negative coins should cause a panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("AddCoins did not panic on negative coin value")
		}
	}()
	env.bank.AddCoins(ctx, addr1, std.NewCoins(std.NewCoin("foocoin", -50)))
}
