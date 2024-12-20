package std

// AssertOriginCall panics if [IsOriginCall] returns false.
func AssertOriginCall() // injected

// IsOriginCall returns true only if the calling method is invoked via a direct
// MsgCall. It returns false for all other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also returns false every time when the transaction is broadcasted via
// MsgRun.
func IsOriginCall() bool // injected

func GetChainID() string     // injected
func GetChainDomain() string // injected
func GetHeight() int64       // injected

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

func GetOriginPkgAddr() Address {
	return Address(originPkgAddr())
}

func CallerAt(n int) Address {
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

// Variations which don't use named types.
func originSend() (denoms []string, amounts []int64)
func originCaller() string
func originPkgAddr() string
func callerAt(n int) string
func getRealm(height int) (address string, pkgPath string)
func derivePkgAddr(pkgPath string) string
func encodeBech32(prefix string, bz [20]byte) string
func decodeBech32(addr string) (prefix string, bz [20]byte, ok bool)
func assertCallerIsRealm()
