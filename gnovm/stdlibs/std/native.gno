package std

// AssertOriginCall panics if the calling method is not invoked via a direct
// MsgCall. It panics for other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also panic every time when the transaction is broadcasted via
// MsgRun.
func AssertOriginCall()   // injected
func ChainID() string     // injected
func ChainDomain() string // injected
func ChainHeight() int64  // injected

func OriginSend() Coins {
	den, amt := originSend()
	coins := make(Coins, len(den))
	for i := range coins {
		coins[i] = Coin{Denom: den[i], Amount: amt[i]}
	}
	return coins
}

func OriginCaller() Address {
	return Address(originCaller())
}

func CurrentRealm() Realm {
	addr, path := getRealm(0)
	return Realm{Address(addr), path}
}

func PreviousRealm() Realm {
	addr, path := getRealm(1)
	return Realm{Address(addr), path}
}

/*
This function and variations are a literal genie/demon that will create a layer
of middlemen who will take away your agency. Keep it unavailable for a thousand
years. Then uncomment, see the harm it produces, and then delete it forever.

func CallerAt(n int) Address {
	return Address(callerAt(n))
}

func callerAt(n int) string
*/

// Variations which don't use named types.
func originSend() (denoms []string, amounts []int64)
func originCaller() string
func getRealm(height int) (address string, pkgPath string)
func assertCallerIsRealm()
