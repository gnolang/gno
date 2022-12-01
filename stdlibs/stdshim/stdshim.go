//go:build gno
// +build gno

package std

const shimWarn = "stdshim cannot be used to run code"

func AssertOriginCall() {
	panic(shimWarn)
}

func IsOriginCall() (isOrigin bool) {
	panic(shimWarn)
	return false
}

func Hash(bz []byte) (hash [20]byte) {
	panic(shimWarn)
	return
}

func CurrentRealmPath() string {
	panic(shimWarn)
	return ""
}

func GetChainID() string {
	panic(shimWarn)
	return ""
}

func GetHeight() int64 {
	panic(shimWarn)
	return -1
}

func GetOrigSend() Coins {
	panic(shimWarn)
	return Coins{}
}

func GetOrigCaller() Address {
	panic(shimWarn)
	return Address("")
}

func GetOrigPkgAddr() Address {
	panic(shimWarn)
	return Address("")
}

func GetCallerAt(n int) Address {
	panic(shimWarn)
	return Address("")
}

func GetBanker(bankerType BankerType) Banker {
	panic(shimWarn)
	return nil
}

func GetTimestamp() Time {
	panic(shimWarn)
	return 0
}

func FormatTimestamp(timestamp Time, format string) string {
	panic(shimWarn)
	return ""
}

func EncodeBech32(prefix string, bytes [20]byte) (addr Address) {
	panic(shimWarn)
	return ""
}

func DecodeBech32(addr Address) (prefix string, bytes [20]byte, ok bool) {
	panic(shimWarn)
}

func DerivePkgAddr(pkgPath string) (addr Address) {
	panic(shimWarn)
}
