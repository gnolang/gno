package std

func AssertOriginCall()                        // injected
func IsOriginCall() bool                       // injected
func TestCurrentRealm() string                 // injected
func TestSkipHeights(count int64)              // injected
func ClearStoreCache()                         // injected
func GetCallerAt(n int) Address                // injected
func TestSetOrigCaller(addr Address)           // injected
func TestSetOrigPkgAddr(addr Address)          // injected
func TestSetOrigSend(sent, spent Coins)        // injected
func TestIssueCoins(addr Address, coins Coins) // injected
