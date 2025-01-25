package std

func AssertOriginCall()           // injected
func IsOriginCall() bool          // injected
func TestSkipHeights(count int64) // injected

// TestSetRealm sets the realm for the current frame.
// After calling TestSetRealm, calling CurrentRealm() in the test function will yield the value of
// rlm, while if a realm function is called, using PrevRealm() will yield rlm.
func TestSetRealm(rlm Realm) {
	testSetRealm(string(rlm.addr), rlm.pkgPath)
}

func TestIssueCoins(addr Address, coins Coins) {
	denom, amt := coins.expandNative()
	testIssueCoins(string(addr), denom, amt)
}

// GetCallerAt calls callerAt, which we overwrite
func callerAt(n int) string

// native bindings
func testSetRealm(addr, pkgPath string)
func testIssueCoins(addr string, denom []string, amt []int64)
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool
