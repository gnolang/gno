package std

func AssertOriginCall()           // injected
func IsOriginCall() bool          // injected
func TestSkipHeights(count int64) // injected
func ClearStoreCache()            // injected

func TestSetOrigCaller(addr Address)  { testSetOrigCaller(string(addr)) }
func TestSetOrigPkgAddr(addr Address) { testSetOrigPkgAddr(string(addr)) }

// TestSetRealm sets the realm for the current frame.
// After calling TestSetRealm, calling CurrentRealm() will yield the value of
// rlm, while if a realm function is called, using PrevRealm() will yield rlm.
func TestSetRealm(rlm Realm) {
	testSetRealm(string(rlm.addr), rlm.pkgPath)
}

func TestSetOrigSend(sent, spent Coins) {
	sentDenom, sentAmt := sent.expandNative()
	spentDenom, spentAmt := spent.expandNative()
	testSetOrigSend(sentDenom, sentAmt, spentDenom, spentAmt)
}

func TestIssueCoins(addr Address, coins Coins) {
	denom, amt := coins.expandNative()
	testIssueCoins(string(addr), denom, amt)
}

// GetCallerAt calls callerAt, which we overwrite
func callerAt(n int) string

// native bindings
func testSetOrigCaller(s string)
func testSetOrigPkgAddr(s string)
func testSetRealm(addr, pkgPath string)
func testSetOrigSend(
	sentDenom []string, sentAmt []int64,
	spentDenom []string, spentAmt []int64)
func testIssueCoins(addr string, denom []string, amt []int64)
func getRealm(height int) (address string, pkgPath string)
