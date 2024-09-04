package std

// AssertOriginCall panics if [IsOriginCall] returns false.
func AssertOriginCall() // injected

// IsOriginCall returns true only if the calling method is invoked via a direct
// MsgCall. It returns false for all other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also returns false every time when the transaction is broadcasted via
// MsgRun.
func IsOriginCall() bool // injected

func GetChainID() string // injected
func GetHeight() int64   // injected

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

func DerivePkgAddr(pkgPath string) Address {
	return Address(derivePkgAddr(pkgPath))
}

func EncodeBech32(prefix string, bz [20]byte) Address {
	return Address(encodeBech32(prefix, bz))
}

func DecodeBech32(addr Address) (prefix string, bz [20]byte, ok bool) {
	return decodeBech32(string(addr))
}

func VerifySignature(pubKeySigner string,msg string, signature string) bool {
	return verifySignature(pubKeySigner, msg, signature)
}

// Variations which don't use named types.
func origSend() (denoms []string, amounts []int64)
func origCaller() string
func origPkgAddr() string
func callerAt(n int) string
func getRealm(height int) (address string, pkgPath string)
func derivePkgAddr(pkgPath string) string
func encodeBech32(prefix string, bz [20]byte) string
func decodeBech32(addr string) (prefix string, bz [20]byte, ok bool)
func verifySignature(pubKeySigner string,msg string, signature string) bool
