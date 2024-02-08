package std

// faux impl to injected native funcs, for test sake
func AssertOriginCall() {
}

func IsOriginCall() (isOrigin bool) {
	return false
}

func Hash(bz []byte) (hash [20]byte) {
	return
}

func CurrentRealmPath() string {
	return ""
}

func GetChainID() string {
	return ""
}

func GetHeight() int64 {
	return -1
}

func GetOrigSend() Coins {
	return Coins{}
}

func CurrentRealm() Realm {
	return Realm{
		addr:    Address(""),
		pkgPath: "",
	}
}

func PrevRealm() Realm {
	return Realm{
		addr:    Address(""),
		pkgPath: "",
	}
}

func GetOrigCaller() Address {
	return Address("")
}

func GetOrigPkgAddr() Address {
	return Address("")
}

func GetCallerAt(n int) Address {
	return Address("")
}

func GetBanker(bankerType BankerType) Banker {
	return nil
}

func EncodeBech32(prefix string, bytes [20]byte) (addr Address) {
	return ""
}

func DecodeBech32(addr Address) (prefix string, bytes [20]byte, ok bool) {
	return "", [20]byte{}, false
}

func DerivePkgAddr(pkgPath string) (addr Address) {
	return Address("")
}

func TestSetOrigCaller(addr Address)           {} // injected
func TestSetOrigPkgAddr(addr Address)          {} // injected
func TestSetOrigSend(sent, spent Coins)        {} // injected
func TestIssueCoins(addr Address, coins Coins) {} // injected
