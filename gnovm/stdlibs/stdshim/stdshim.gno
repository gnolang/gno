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

func ChainID() string {
	panic(shimWarn)
	return ""
}

func Height() int64 {
	panic(shimWarn)
	return -1
}

func OrigSend() Coins {
	panic(shimWarn)
	return Coins{}
}

func OrigCaller() Address {
	panic(shimWarn)
	return Address("")
}

func OrigPkgAddr() Address {
	panic(shimWarn)
	return Address("")
}

func CallerAt(n int) Address {
	panic(shimWarn)
	return Address("")
}

func Banker(bankerType BankerType) Banker {
	panic(shimWarn)
	return nil
}

func Timestamp() Time {
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
