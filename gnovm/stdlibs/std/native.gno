package std

// AssertOriginCall panics if the calling method is not invoked via a direct
// MsgCall. It doesn't panic for any other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also panic every time when the transaction is broadcasted via
// MsgRun.
func AssertOriginCall() {
	if !PrevRealm().IsUser() || isMsgRun() {
		panic("invalid non-origin call")
	}
}
func GetChainID() string     // injected
func GetChainDomain() string // injected
func GetHeight() int64       // injected

func GetOrigSend() Coins {
	den, amt := origSend()
	coins := make(Coins, len(den))
	for i := range coins {
		coins[i] = Coin{Denom: den[i], Amount: amt[i]}
	}
	return coins
}

func GetOrigCaller() Address {
	return Address(origCaller())
}

func CurrentRealm() Realm {
	addr, path := getRealm(0)
	return Realm{Address(addr), path}
}

func PrevRealm() Realm {
	addr, path := getRealm(1)
	return Realm{Address(addr), path}
}

func GetOrigPkgAddr() Address {
	return Address(origPkgAddr())
}

func GetCallerAt(n int) Address {
	return Address(callerAt(n))
}

// Variations which don't use named types.
func origSend() (denoms []string, amounts []int64)
func origCaller() string
func origPkgAddr() string
func callerAt(n int) string
func getRealm(height int) (address string, pkgPath string)
func assertCallerIsRealm()
func isMsgRun() bool
