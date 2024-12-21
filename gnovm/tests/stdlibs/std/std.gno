package std

func AssertOriginCall()           // injected
func IsOriginCall() bool          // injected
func TestSkipHeights(count int64) // injected

func TestSetOriginCaller(addr Address)  { testSetOriginCaller(string(addr)) }
func TestSetOriginPkgAddress(addr Address) { testSetOriginPkgAddr(string(addr)) }

// TestSetRealm sets the realm for the current frame.
// After calling TestSetRealm, calling CurrentRealm() in the test function will yield the value of
// rlm, while if a realm function is called, using PreviousRealm() will yield rlm.
func TestSetRealm(rlm Realm) {
	testSetRealm(string(rlm.addr), rlm.pkgPath)
}

func TestSetOriginSend(sent, spent Coins) {
	sentDenom, sentAmt := sent.expandNative()
	spentDenom, spentAmt := spent.expandNative()
	testSetOriginSend(sentDenom, sentAmt, spentDenom, spentAmt)
}

func TestIssueCoins(addr Address, coins Coins) {
	denom, amt := coins.expandNative()
	testIssueCoins(string(addr), denom, amt)
}

// CallerAt calls callerAt, which we overwrite
func callerAt(n int) string

// native bindings
func testSetOriginCaller(s string)
func testSetOriginPkgAddr(s string)
func testSetRealm(addr, pkgPath string)
func testSetOriginSend(
	sentDenom []string, sentAmt []int64,
	spentDenom []string, spentAmt []int64)
func testIssueCoins(addr string, denom []string, amt []int64)
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool
