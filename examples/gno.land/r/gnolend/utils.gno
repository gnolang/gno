package gnolend

import (
	"std"
	
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/grc20reg"
	"gno.land/r/gnoswap/v1/pool"
)

// GetTokenTeller returns a GRC20 token teller for the given token path
// Panics if the token is not registered
func GetTokenCallerTeller(path string) grc20.Teller {
	token := grc20reg.MustGet(path) // panics if token not registered
	return token.CallerTeller()
}

// TransferFromSafe performs a safe token transfer with balance verification (similar to gnoswap)
func TransferFromSafe(cur realm, tokenPath string, from, to std.Address, amount int64) error {
	// Get token and check balances before transfer
	token := GetToken(tokenPath)
	beforeFromBalance := token.BalanceOf(from)
	beforeToBalance := token.BalanceOf(to)
	
	// Check if sender has sufficient balance
	if beforeFromBalance < amount {
		return ufmt.Errorf("insufficient balance: %d < %d", beforeFromBalance, amount)
	}
	
	// Check if allowance is sufficient
	allowance := token.Allowance(from, to)
	if allowance < amount {
		return ufmt.Errorf("insufficient allowance: %d < %d (from %s to %s)", allowance, amount, from.String(), to.String())
	}
	
	// Perform the transfer using teller
	teller := GetTokenCallerTeller(tokenPath)
	err := teller.TransferFrom(from, to, amount)
	if err != nil {
		return ufmt.Errorf("transfer failed: %v", err)
	}
	
	// Verify balances after transfer
	afterFromBalance := token.BalanceOf(from)
	afterToBalance := token.BalanceOf(to)
	
	expectedFromBalance := beforeFromBalance - amount
	expectedToBalance := beforeToBalance + amount
	
	if afterFromBalance != expectedFromBalance {
		return ufmt.Errorf("from balance verification failed: expected %d, got %d", expectedFromBalance, afterFromBalance)
	}
	
	if afterToBalance != expectedToBalance {
		return ufmt.Errorf("to balance verification failed: expected %d, got %d", expectedToBalance, afterToBalance)
	}
	
	return nil
}

// TransferSafe performs a safe token transfer from contract to user with balance verification
func TransferSafe(cur realm, tokenPath string, to std.Address, amount int64) error {
	// Get token and check balances before transfer
	token := GetToken(tokenPath)
	contractAddr := std.CurrentRealm().Address()
	beforeContractBalance := token.BalanceOf(contractAddr)
	beforeToBalance := token.BalanceOf(to)
	
	// Check if contract has sufficient balance
	if beforeContractBalance < amount {
		return ufmt.Errorf("insufficient contract balance: %d < %d", beforeContractBalance, amount)
	}
	
	// Perform the transfer using teller
	teller := GetTokenCallerTeller(tokenPath)
	err := teller.Transfer(to, amount)
	if err != nil {
		return ufmt.Errorf("transfer failed: %v", err)
	}
	
	// Verify balances after transfer
	afterContractBalance := token.BalanceOf(contractAddr)
	afterToBalance := token.BalanceOf(to)
	
	expectedContractBalance := beforeContractBalance - amount
	expectedToBalance := beforeToBalance + amount
	
	if afterContractBalance != expectedContractBalance {
		return ufmt.Errorf("contract balance verification failed: expected %d, got %d", expectedContractBalance, afterContractBalance)
	}
	
	if afterToBalance != expectedToBalance {
		return ufmt.Errorf("to balance verification failed: expected %d, got %d", expectedToBalance, afterToBalance)
	}
	
	return nil
}

// GetToken returns a GRC20 token instance
// Panics if the token is not registered
func GetToken(path string) *grc20.Token {
	return grc20reg.MustGet(path)
}

// Helper function to check if two tokens exist together in a Gnoswap pool
func areTokensPairedInGnoswap(token0, token1 string) bool {
	poolPaths := pool.PoolGetPoolList()
	for _, poolPath := range poolPaths {
		poolToken0 := pool.PoolGetToken0Path(poolPath)
		poolToken1 := pool.PoolGetToken1Path(poolPath)

		// Check if both tokens are in this pool (in either order)
		if (poolToken0 == token0 && poolToken1 == token1) ||
			(poolToken0 == token1 && poolToken1 == token0) {
			return true
		}
	}
	return false
}

// Min returns the minimum of two uint256 numbers
func Min(a, b *u256.Uint) *u256.Uint {
	if a.Lt(b) {
		return a
	}
	return b
}
