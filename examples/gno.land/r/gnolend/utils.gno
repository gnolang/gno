package gnolend

import (
	"gno.land/p/demo/grc/grc20"

	"gno.land/r/demo/grc20reg"
	"gno.land/r/gnoswap/v1/pool"
)

// GetTokenTeller returns a GRC20 token teller for the given token path
// Panics if the token is not registered
func GetTokenCallerTeller(path string) grc20.Teller {
	tokenGetter := grc20reg.MustGet(path) // panics if token not registered
	token := tokenGetter()
	return token.CallerTeller()
}

// GetToken returns a GRC20 token instance
// Panics if the token is not registered
func GetToken(path string) *grc20.Token {
	tokenGetter := grc20reg.MustGet(path)
	return tokenGetter()
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
