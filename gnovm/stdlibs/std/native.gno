package std

func AssertOriginCall()        // injected
func IsOriginCall() bool       // injected
func CurrentRealmPath() string // injected
func GetChainID() string       // injected
func GetHeight() int64         // injected

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

// Variations which don't use named types.
func origSend() (denoms []string, amounts []int64)
func origCaller() string
func origPkgAddr() string
func callerAt(n int) string
func getRealm(height int) (address string, pkgPath string)
func derivePkgAddr(pkgPath string) string
func encodeBech32(prefix string, bz [20]byte) string
func decodeBech32(addr string) (prefix string, bz [20]byte, ok bool)
