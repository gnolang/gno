package std

func AssertOriginCall()           // injected
func IsOriginCall() bool          // injected
func TestCurrentRealm() string    // injected
func TestSkipHeights(count int64) // injected
func ClearStoreCache()            // injected

func TestSetOrigCaller(addr Address)  { testSetOrigCaller(string(addr)) }
func TestSetOrigPkgAddr(addr Address) { testSetOrigPkgAddr(string(addr)) }
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
func testSetOrigSend(
	sentDenom []string, sentAmt []int64,
	spentDenom []string, spentAmt []int64)
func testIssueCoins(addr string, denom []string, amt []int64)
