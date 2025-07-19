package volos

import (
	"std"
	
	"gno.land/p/demo/grc/grc20"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/grc20reg"
	"gno.land/r/gnoswap/v1/pool"
)

// GetToken returns a GRC20 token instance
// Panics if the token is not registered
func GetToken(path string) *grc20.Token {
	return grc20reg.MustGet(path)
}

// GetTokenTeller returns a GRC20 token Caller teller for the given token path
// Panics if the token is not registered
func GetTokenCallerTeller(path string) grc20.Teller {
	token := GetToken(path) // panics if token not registered
	return token.CallerTeller()
}

// GetTokenTeller returns a GRC20 token Realm teller for the given token path
// Panics if the token is not registered
func GetTokenRealmTeller(path string) grc20.Teller {
	token := GetToken(path) // panics if token not registered
	return token.RealmTeller()
}

// safeTransferFrom performs a safe token transfer with balance verification
func safeTransferFrom(tokenPath string, from std.Address, amount int64) {
	// Get token and check balances before transfer
	token := GetToken(tokenPath)
	contractAddress := std.CurrentRealm().Address()
	beforeFromBalance := token.BalanceOf(from)
	beforeToBalance := token.BalanceOf(contractAddress)
	
	// Check if sender has sufficient balance
	if beforeFromBalance < amount {
		panic(ErrInsufficientBalance)
	}
	
	// Check if allowance is sufficient
	allowance := token.Allowance(from, contractAddress)
	if allowance < amount {
		panic(ErrInsufficientAllowance)
	}
	
	// Perform the transfer using teller
	teller := GetTokenRealmTeller(tokenPath)
	err := teller.TransferFrom(from, contractAddress, amount)
	if err != nil {
		panic(ErrTransferFailed)
	}
	
	// Verify balances after transfer
	afterFromBalance := token.BalanceOf(from)
	afterToBalance := token.BalanceOf(contractAddress)
	
	expectedFromBalance := beforeFromBalance - amount
	expectedToBalance := beforeToBalance + amount
	
	if afterFromBalance != expectedFromBalance {
		panic(ErrBalanceVerificationFailed)
	}
	
	if afterToBalance != expectedToBalance {
		panic(ErrBalanceVerificationFailed)
	}
}

// safeTransferTo performs a safe token transfer from contract to user with balance verification
func safeTransferTo(tokenPath string, to std.Address, amount int64) {
	// Get token and check balances before transfer
	token := GetToken(tokenPath)
	contractAddr := std.CurrentRealm().Address()
	beforeContractBalance := token.BalanceOf(contractAddr)
	beforeToBalance := token.BalanceOf(to)
	
	// Check if contract has sufficient balance
	if beforeContractBalance < amount {
		panic(ErrInsufficientBalance)
	}
	
	// Perform the transfer using teller
	teller := GetTokenRealmTeller(tokenPath)
	err := teller.Transfer(to, amount)
	if err != nil {
		panic(ErrTransferFailed)
	}
	
	// Verify balances after transfer
	afterContractBalance := token.BalanceOf(contractAddr)
	afterToBalance := token.BalanceOf(to)
	
	expectedContractBalance := beforeContractBalance - amount
	expectedToBalance := beforeToBalance + amount
	
	if afterContractBalance != expectedContractBalance {
		panic(ErrBalanceVerificationFailed)
	}
	
	if afterToBalance != expectedToBalance {
		panic(ErrBalanceVerificationFailed)
	}
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
