package std

func AssertOriginCall()  // injected
func IsOriginCall() bool // injected

func TestIssueCoins(addr Address, coins Coins) {
	denom, amt := coins.expandNative()
	testIssueCoins(string(addr), denom, amt)
}

// GetCallerAt calls callerAt, which we overwrite
func callerAt(n int) string

// native bindings
func testIssueCoins(addr string, denom []string, amt []int64)
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool
